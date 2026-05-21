package journald

import (
	"os/exec"
	"strings"
)

func Check() (map[string]interface{}, error) {
	cmd := exec.Command("journalctl", "-n", "20", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "journald",
			"status":  "unavailable",
			"message": "journalctl not available",
		}, nil
	}

	logs := strings.TrimSpace(string(output))
	lines := strings.Split(logs, "\n")

	return map[string]interface{}{
		"name":   "journald",
		"status": "ok",
		"logs":   lines,
	}, nil
}
