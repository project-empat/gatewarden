package fail2ban

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Check returns Fail2Ban status by querying fail2ban-client.
func Check() (*proto.Fail2BanStatus, error) {
	// Check if fail2ban-client is available
	if err := exec.Command("fail2ban-client", "--version").Run(); err != nil {
		return nil, nil // Fail2Ban not available
	}

	status := &proto.Fail2BanStatus{
		Installed: true,
	}

	// Get version
	if out, err := exec.Command("fail2ban-client", "--version").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			status.Version = strings.Fields(lines[0])[0]
		}
	}

	// Check if service is running
	if err := exec.Command("fail2ban-client", "ping").Run(); err == nil {
		status.Running = true
	}

	// Get jail list
	jails, err := parseJails(status)
	if err == nil {
		status.Jails = jails
	}
	if status.Jails == nil {
		status.Jails = []proto.Fail2BanJail{}
	}

	return status, nil
}

func parseJails(status *proto.Fail2BanStatus) ([]proto.Fail2BanJail, error) {
	out, err := exec.Command("fail2ban-client", "status").Output()
	if err != nil {
		return nil, err
	}

	// Parse output to find jail list
	// Format: "|- Jail list:\tsshd, nginx-http-auth"
	lines := strings.Split(string(out), "\n")
	var jailNames []string

	for _, line := range lines {
		if strings.Contains(line, "Jail list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				names := strings.Split(parts[1], ",")
				for _, name := range names {
					name = strings.TrimSpace(name)
					if name != "" {
						jailNames = append(jailNames, name)
					}
				}
			}
			break
		}
	}

	jails := make([]proto.Fail2BanJail, 0, len(jailNames))
	for _, name := range jailNames {
		jail, err := parseJailStatus(name)
		if err == nil {
			jails = append(jails, jail)
		}
	}

	return jails, nil
}

func parseJailStatus(name string) (proto.Fail2BanJail, error) {
	out, err := exec.Command("fail2ban-client", "status", name).Output()
	if err != nil {
		return proto.Fail2BanJail{}, err
	}

	jail := proto.Fail2BanJail{
		Name:   name,
		Active: true,
	}

	// Parse jail status output
	// Status for the jail: sshd
	// |- Filter
	// |  |- Currently failed: 12
	// |  |- Total failed:    234
	// |- Actions
	//    |- Currently banned: 3
	//    |- Total banned:     47
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Currently banned:") {
			val := extractValue(trimmed)
			if v, err := strconv.Atoi(val); err == nil {
				jail.CurrentlyBanned = v
			}
		}

		if strings.Contains(trimmed, "Total banned:") {
			val := extractValue(trimmed)
			if v, err := strconv.Atoi(val); err == nil {
				jail.TotalBanned = v
			}
		}

		if strings.Contains(trimmed, "Currently failed:") {
			// Count not stored in struct yet
		}

		if strings.Contains(trimmed, "Total failed:") {
			val := extractValue(trimmed)
			if v, err := strconv.Atoi(val); err == nil {
				jail.FailedCount = v
			}
		}
	}

	return jail, nil
}

func extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
