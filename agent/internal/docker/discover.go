package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Check performs Docker discovery and returns container status with security analysis.
func Check() (*proto.DockerStatus, error) {
	// First verify docker is available
	if err := exec.Command("docker", "info", "--format", "{{.ServerVersion}}").Run(); err != nil {
		return nil, nil // Docker not available, skip
	}

	containers, err := listContainers()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	// Check if Docker socket is exposed on the host
	socketExposed := checkSocketExposed(containers)

	return &proto.DockerStatus{
		RunningContainers: containers,
		TotalContainers:   len(containers),
		SocketExposed:     socketExposed,
	}, nil
}

type dockerInspect struct {
	Name       string `json:"Name"`
	State      struct {
		Status string `json:"Status"`
	} `json:"State"`
	HostConfig struct {
		Privileged bool   `json:"Privileged"`
		NetworkMode string `json:"NetworkMode"`
	} `json:"HostConfig"`
	Config struct {
		User string `json:"User"`
	} `json:"Config"`
	NetworkSettings struct {
		Ports map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
	Mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
}

func listContainers() ([]proto.DockerContainer, error) {
	// List running containers with basic info as JSON
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	ids := strings.Fields(string(out))
	containers := make([]proto.DockerContainer, 0, len(ids))

	for _, id := range ids {
		c, err := inspectContainer(id)
		if err != nil {
			continue
		}
		containers = append(containers, c)
	}

	return containers, nil
}

func inspectContainer(id string) (proto.DockerContainer, error) {
	cmd := exec.Command("docker", "inspect", id)
	out, err := cmd.Output()
	if err != nil {
		return proto.DockerContainer{}, err
	}

	var details []dockerInspect
	if err := json.Unmarshal(out, &details); err != nil || len(details) == 0 {
		return proto.DockerContainer{}, fmt.Errorf("inspect parse failed")
	}

	di := details[0]
	c := proto.DockerContainer{
		ID:          id[:12],
		Name:        strings.TrimPrefix(di.Name, "/"),
		Status:      di.State.Status,
		Privileged:  di.HostConfig.Privileged,
		NetworkMode: di.HostConfig.NetworkMode,
		User:        di.Config.User,
	}

	// Extract image name from Name
	if di.Name != "" {
		// docker inspect format: Name is "/name"
	}

	// Parse ports
	for containerPort, hostBindings := range di.NetworkSettings.Ports {
		for _, binding := range hostBindings {
			hostPort := 0
			fmt.Sscanf(binding.HostPort, "%d", &hostPort)
			internalPort := 0
			fmt.Sscanf(containerPort, "%d", &internalPort)

			exposure := "none"
			if binding.HostIP == "0.0.0.0" {
				exposure = "0.0.0.0"
			} else if binding.HostIP == "127.0.0.1" || binding.HostIP == "::1" {
				exposure = "127.0.0.1"
			} else if binding.HostIP != "" {
				exposure = binding.HostIP
			}

			c.Ports = append(c.Ports, proto.DockerPort{
				Internal: internalPort,
				External: hostPort,
				Exposure: exposure,
			})
		}
	}

	if c.Ports == nil {
		c.Ports = []proto.DockerPort{}
	}

	// Parse mounts for Docker socket
	for _, mount := range di.Mounts {
		if mount.Source == "/var/run/docker.sock" || mount.Destination == "/var/run/docker.sock" {
			c.SocketExposed = true
		}
	}

	if di.Config.User == "" {
		c.User = "root"
	}

	return c, nil
}

func checkSocketExposed(containers []proto.DockerContainer) bool {
	for _, c := range containers {
		if c.SocketExposed {
			return true
		}
	}
	return false
}
