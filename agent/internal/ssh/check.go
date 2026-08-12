package ssh

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// CheckHardening parses /etc/ssh/sshd_config and checks for security issues.
func CheckHardening() (*proto.SSHStatus, error) {
	// First check if SSH is running
	if err := exec.Command("systemctl", "is-active", "ssh", "sshd").Run(); err != nil {
		// Try alternate service name
		if err := exec.Command("pgrep", "-x", "sshd").Run(); err != nil {
			return nil, nil // SSH not available
		}
	}

	s := &proto.SSHStatus{
		Port:         22,
		PubkeyAuth:   true,
		PasswordAuth: false,
		RootLogin:    "no",
		MaxAuthTries: 6,
	}

	// Read effective config using sshd -T
	cmd := exec.Command("sshd", "-T")
	out, err := cmd.Output()
	if err != nil {
		// Fallback: parse config file manually
		return parseConfigFile(s)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				s.Port = port
			}
		case "passwordauthentication":
			s.PasswordAuth = value == "yes"
		case "permitrootlogin":
			s.RootLogin = value
		case "pubkeyauthentication":
			s.PubkeyAuth = value == "yes"
		case "maxauthtries":
			if tries, err := strconv.Atoi(value); err == nil {
				s.MaxAuthTries = tries
			}
		case "listenaddress":
			s.ListenAddresses = append(s.ListenAddresses, value)
		}
	}

	if s.ListenAddresses == nil {
		// No explicit ListenAddress means it listens on all
		s.ListenAddresses = []string{"0.0.0.0"}
	}

	// Determine public exposure
	for _, addr := range s.ListenAddresses {
		if addr == "0.0.0.0" || addr == "::" {
			s.PubliclyExposed = true
			break
		}
	}

	return s, nil
}

func parseConfigFile(s *proto.SSHStatus) (*proto.SSHStatus, error) {
	f, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		return nil, fmt.Errorf("cannot read sshd_config: %w", err)
	}
	defer f.Close()

	return parseSSHDConfig(s, f)
}

// parseSSHDConfig parses an SSH config from the given reader.
// Exposed for testing.
func parseSSHDConfig(s *proto.SSHStatus, r io.Reader) (*proto.SSHStatus, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				s.Port = port
			}
		case "passwordauthentication":
			s.PasswordAuth = strings.EqualFold(value, "yes")
		case "permitrootlogin":
			s.RootLogin = value
		case "pubkeyauthentication":
			s.PubkeyAuth = strings.EqualFold(value, "yes")
		case "maxauthtries":
			if tries, err := strconv.Atoi(value); err == nil {
				s.MaxAuthTries = tries
			}
		}
	}

	s.ListenAddresses = []string{"0.0.0.0"}
	s.PubliclyExposed = true

	return s, nil
}
