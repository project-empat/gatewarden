package docker

import (
	"os/exec"
	"strings"
)

func Discover() (map[string]interface{}, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Image}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "docker",
			"status":  "unavailable",
			"message": "Docker not found or not running",
		}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := []map[string]string{}
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) == 3 {
			containers = append(containers, map[string]string{
				"id":     parts[0][:12],
				"image":  parts[1],
				"status": parts[2],
			})
		}
	}

	return map[string]interface{}{
		"name":       "docker",
		"status":     "ok",
		"containers": containers,
		"count":      len(containers),
	}, nil
}
