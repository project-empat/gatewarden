package tailscale

import (
	"os/exec"
	"strings"
)

func Check() (map[string]interface{}, error) {
	cmd := exec.Command("tailscale", "status")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "tailscale",
			"status":  "unavailable",
			"message": "Tailscale not installed or not running",
		}, nil
	}

	status := strings.TrimSpace(string(output))
	lines := strings.Split(status, "\n")

	return map[string]interface{}{
		"name":   "tailscale",
		"status": "connected",
		"peers":  lines,
	}, nil
}
