package ports

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Scan returns listening ports by parsing `ss -tlnp`.
func Scan() ([]proto.ListeningPort, error) {
	cmd := exec.Command("ss", "-tlnp", "4")
	output, err := cmd.Output()
	if err != nil {
		// fallback: try without ipv4 filter
		cmd = exec.Command("ss", "-tlnp")
		output, err = cmd.Output()
		if err != nil {
			return nil, nil
		}
	}

	return parseSS(output), nil
}

func parseSS(data []byte) []proto.ListeningPort {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var ports []proto.ListeningPort

	for _, line := range lines {
		// Skip header
		if strings.HasPrefix(line, "State") || strings.HasPrefix(line, "Netid") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// fields: State Recv-Q Send-Q LocalAddress:Port PeerAddress:Port Process
		// Example: LISTEN 0 128 0.0.0.0:22 0.0.0.0:* users:(("sshd",pid=1234,fd=3))
		localAddr := fields[3]
		process := extractProcess(fields)

		addr, portStr, ok := strings.Cut(localAddr, ":")
		if !ok {
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		exposed := addr == "0.0.0.0" || addr == "::"

		ports = append(ports, proto.ListeningPort{
			Port:     port,
			Protocol: "tcp",
			Process:  process,
			Exposed:  exposed,
		})
	}

	return ports
}

func extractProcess(fields []string) string {
	// Process info is in the last field, format: users:(("sshd",pid=1234,fd=3))
	for _, f := range fields {
		if strings.HasPrefix(f, "users:") {
			// Extract process name between first set of quotes
			start := strings.Index(f, "(\"")
			if start < 0 {
				start = strings.Index(f, "\"")
				if start < 0 {
					continue
				}
				start++
			} else {
				start += 2
			}
			end := strings.Index(f[start:], "\"")
			if end < 0 {
				continue
			}
			return f[start : start+end]
		}
	}
	return ""
}
