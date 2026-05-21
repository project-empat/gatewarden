package fail2ban

import (
	"os/exec"
	"strings"
)

func Check() (map[string]interface{}, error) {
	cmd := exec.Command("fail2ban-client", "status")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "fail2ban",
			"status":  "unavailable",
			"message": "fail2ban not installed or not running",
		}, nil
	}

	jailOutput := strings.TrimSpace(string(output))
	jails := strings.Split(jailOutput, "\n")

	return map[string]interface{}{
		"name":   "fail2ban",
		"status": "running",
		"jails":  jails,
	}, nil
}
