package cloudflare

import (
	"os/exec"
	"strings"
)

func Check() (map[string]interface{}, error) {
	cmd := exec.Command("cloudflared", "tunnel", "info")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "cloudflare_tunnel",
			"status":  "unavailable",
			"message": "cloudflared not installed or no tunnel configured",
		}, nil
	}

	return map[string]interface{}{
		"name":    "cloudflare_tunnel",
		"status":  "connected",
		"message": strings.TrimSpace(string(output)),
	}, nil
}
