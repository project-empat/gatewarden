package cloudflare

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Check returns Cloudflare Tunnel status by checking cloudflared.
func Check() (*proto.CloudflareStatus, error) {
	if err := exec.Command("cloudflared", "--version").Run(); err != nil {
		return nil, nil // cloudflared not available
	}

	status := &proto.CloudflareStatus{
		Installed: true,
	}

	if out, err := exec.Command("cloudflared", "--version").Output(); err == nil {
		fields := strings.Fields(string(out))
		if len(fields) >= 3 {
			status.Version = fields[2]
		}
	}

	if err := exec.Command("pgrep", "-x", "cloudflared").Run(); err == nil {
		status.Running = true
	}

	tunnels, err := listTunnels()
	if err == nil {
		status.Tunnels = tunnels
	} else {
		tunnels, err := parseConfigIngress()
		if err == nil && len(tunnels) > 0 {
			status.Tunnels = tunnels
		}
	}

	if status.Tunnels == nil {
		status.Tunnels = []proto.CloudflareTunnel{}
	}

	return status, nil
}

func listTunnels() ([]proto.CloudflareTunnel, error) {
	cmd := exec.Command("cloudflared", "tunnel", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseTunnelList(string(out))
}

func parseTunnelList(output string) ([]proto.CloudflareTunnel, error) {
	// Expected output format (YAML-like or table):
	// ID                                   Name         Status
	// a1b2c3d4-e5f6-...                   prod-tunnel  active
	// or YAML:
	// - ID: a1b2c3d4-e5f6-...
	//   Name: prod-tunnel
	//   Status: active
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("not enough output lines")
	}

	// Check if it's YAML format (starts with dash)
	var tunnels []proto.CloudflareTunnel

	if strings.HasPrefix(strings.TrimSpace(lines[0]), "-") {
		// YAML-like format
		var current *proto.CloudflareTunnel
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ID:") {
				if current != nil {
					tunnels = append(tunnels, *current)
				}
				current = &proto.CloudflareTunnel{}
				current.ID = strings.TrimSpace(strings.TrimPrefix(trimmed, "- ID:"))
			} else if current != nil && strings.HasPrefix(trimmed, "Name:") {
				current.Name = strings.TrimSpace(strings.TrimPrefix(trimmed, "Name:"))
			} else if current != nil && strings.HasPrefix(trimmed, "Status:") {
				current.Status = strings.TrimSpace(strings.TrimPrefix(trimmed, "Status:"))
			}
		}
		if current != nil {
			tunnels = append(tunnels, *current)
		}
	} else {
		// Table format, skip header
		for _, line := range lines[1:] {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) >= 3 {
				tunnels = append(tunnels, proto.CloudflareTunnel{
					ID:     fields[0],
					Name:   fields[1],
					Status: fields[2],
				})
			}
		}
	}

	// Try to get ingress for each tunnel
	for i := range tunnels {
		if ingress, err := getIngress(tunnels[i].Name); err == nil {
			tunnels[i].Ingress = ingress
		}
	}

	return tunnels, nil
}

func getIngress(tunnelName string) ([]proto.CloudflareIngress, error) {
	cmd := exec.Command("cloudflared", "tunnel", "info", tunnelName)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseIngressFromInfo(string(out))
}

func parseIngressFromInfo(info string) ([]proto.CloudflareIngress, error) {
	lines := strings.Split(info, "\n")
	var ingress []proto.CloudflareIngress

	inIngress := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Ingress:") || strings.Contains(trimmed, "ingress:") {
			inIngress = true
			continue
		}

		if !inIngress {
			continue
		}

		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "hostname:") || strings.HasPrefix(trimmed, "service:") {
			if strings.Contains(trimmed, "hostname:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					ingress = append(ingress, proto.CloudflareIngress{
						Hostname: strings.TrimSpace(parts[1]),
					})
				}
			}
			if strings.Contains(trimmed, "service:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					svc := strings.TrimSpace(parts[1])
					if len(ingress) > 0 && ingress[len(ingress)-1].Service == "" {
						ingress[len(ingress)-1].Service = svc
					} else {
						ingress = append(ingress, proto.CloudflareIngress{
							Service: svc,
						})
					}
				}
			}
		} else if trimmed == "" {
			break
		}
	}

	if len(ingress) == 0 {
		return nil, fmt.Errorf("no ingress found in tunnel info")
	}

	return ingress, nil
}

func parseConfigIngress() ([]proto.CloudflareTunnel, error) {
	usr, _ := user.Current()
	homeDir := ""
	if usr != nil {
		homeDir = usr.HomeDir
	}

	paths := []string{
		"/etc/cloudflared/config.yml",
		filepath.Join(homeDir, ".cloudflared", "config.yml"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Simple YAML-like key-value parsing for common cloudflared config
		lines := strings.Split(string(data), "\n")
		var tunnelName string
		var ingress []proto.CloudflareIngress
		inIngress := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "tunnel:") {
				tunnelName = strings.TrimSpace(strings.TrimPrefix(trimmed, "tunnel:"))
			}
			if strings.HasPrefix(trimmed, "ingress:") {
				inIngress = true
				continue
			}
			if inIngress {
				if strings.HasPrefix(trimmed, "- hostname:") {
					hostname := strings.TrimSpace(strings.TrimPrefix(trimmed, "- hostname:"))
					ingress = append(ingress, proto.CloudflareIngress{
						Hostname: hostname,
					})
				} else if strings.HasPrefix(trimmed, "hostname:") {
					hostname := strings.TrimSpace(strings.TrimPrefix(trimmed, "hostname:"))
					ingress = append(ingress, proto.CloudflareIngress{
						Hostname: hostname,
					})
				} else if strings.HasPrefix(trimmed, "service:") {
					service := strings.TrimSpace(strings.TrimPrefix(trimmed, "service:"))
					if len(ingress) > 0 && ingress[len(ingress)-1].Service == "" {
						ingress[len(ingress)-1].Service = service
					}
				}
			}
		}

		if tunnelName != "" || len(ingress) > 0 {
			tunnels := []proto.CloudflareTunnel{
				{
					Name:    tunnelName,
					Status:  "configured",
					Ingress: ingress,
				},
			}
			if tunnels[0].Ingress == nil {
				tunnels[0].Ingress = []proto.CloudflareIngress{}
			}
			return tunnels, nil
		}
	}

	return nil, fmt.Errorf("no config file found")
}
