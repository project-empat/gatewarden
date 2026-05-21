package ssh

import (
	"bufio"
	"os"
	"strings"
)

func CheckHardening() (map[string]interface{}, error) {
	f, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		return map[string]interface{}{
			"name":    "ssh_hardening",
			"status":  "unavailable",
			"message": "Cannot read sshd_config",
		}, nil
	}
	defer f.Close()

	var issues []string
	var findings []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.Contains(line, "PermitRootLogin yes") {
			issues = append(issues, "Root login is permitted")
		}
		if strings.Contains(line, "PasswordAuthentication yes") {
			issues = append(issues, "Password authentication is enabled")
		}
		if strings.Contains(line, "Port 22") {
			findings = append(findings, "Default SSH port 22")
		}
		if strings.Contains(line, "PubkeyAuthentication yes") {
			findings = append(findings, "Public key authentication enabled")
		}
	}

	if len(issues) > 0 {
		return map[string]interface{}{
			"name":    "ssh_hardening",
			"status":  "warning",
			"issues":  issues,
			"message": strings.Join(issues, "; "),
		}, nil
	}

	return map[string]interface{}{
		"name":     "ssh_hardening",
		"status":   "ok",
		"findings": findings,
		"message":  "SSH is properly secured",
	}, nil
}
