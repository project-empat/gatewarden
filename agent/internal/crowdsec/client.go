package crowdsec

import (
	"os/exec"
	"strings"
)

func Check() (map[string]interface{}, error) {
	cmd := exec.Command("cscli", "alerts", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "crowdsec",
			"status":  "unavailable",
			"message": "Crowdsec not installed or cscli not found",
		}, nil
	}

	return map[string]interface{}{
		"name":   "crowdsec",
		"status": "running",
		"output": strings.TrimSpace(string(output)),
	}, nil
}
