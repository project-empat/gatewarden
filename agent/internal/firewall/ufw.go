package firewall

import (
	"os/exec"
	"strings"
)

func CheckUFW() (map[string]interface{}, error) {
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "ufw",
			"status":  "inactive",
			"message": "UFW is not active or not installed",
		}, nil
	}

	status := strings.TrimSpace(string(output))
	state := "active"
	if !strings.Contains(status, "active") {
		state = "inactive"
	}

	return map[string]interface{}{
		"name":    "ufw",
		"status":  state,
		"message": status,
	}, nil
}
