package firewall

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Check detects the active firewall backend (UFW or nftables) and returns status.
func Check() *proto.FirewallStatus {
	status := &proto.FirewallStatus{}

	// Try UFW first
	ufw, ok := checkUFW()
	if ok {
		status.ActiveBackend = "ufw"
		status.UFW = ufw
		return status
	}

	// Fallback to nftables
	nft, ok := checkNFTables()
	if ok {
		status.ActiveBackend = "nftables"
		status.NFTables = nft
		return status
	}

	status.ActiveBackend = "none"
	return status
}

func checkUFW() (*proto.UFWStatus, bool) {
	cmd := exec.Command("ufw", "status", "verbose")
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return nil, false
	}

	u := &proto.UFWStatus{}
	// First line: "Status: active" or "Status: inactive"
	if len(lines) > 0 && strings.Contains(lines[0], "active") {
		u.Active = true
	}
	if !u.Active {
		return u, true
	}

	// Parse logging level from second line: "Logging: on (low)"
	for _, line := range lines {
		if strings.HasPrefix(line, "Logging:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				u.Logging = strings.TrimSpace(parts[1])
			}
		}
	}

	// Parse rules — look for lines with numbered rules
	// Format: "1234  ALLOW IN  22/tcp  from 0.0.0.0/0"
	inRules := false
	rules := []proto.FirewallRule{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "----") {
			inRules = true
			continue
		}
		if !inRules {
			continue
		}
		if trimmed == "" {
			continue
		}

		// First word is the number
		fields := strings.Fields(trimmed)
		if len(fields) < 4 {
			continue
		}

		// Check if first field is a number (rule number)
		if _, err := strconv.Atoi(fields[0]); err != nil {
			continue
		}

		rule := proto.FirewallRule{
			Action:   strings.ToLower(fields[1]),
		}

		if len(fields) >= 4 {
			// port/proto or just proto
			portProto := fields[3]
			if strings.Contains(portProto, "/") {
				parts := strings.SplitN(portProto, "/", 2)
				if port, err := strconv.Atoi(parts[0]); err == nil {
					rule.Port = port
				}
				rule.Protocol = parts[1]
			} else if proto := portProto; proto == "tcp" || proto == "udp" {
				rule.Protocol = proto
			}

			// Source is typically at the end
			for _, f := range fields {
				if f == "from" {
					// next field is source
				}
			}
			// Find source
			for i, f := range fields {
				if f == "from" && i+1 < len(fields) {
					rule.Source = fields[i+1]
				}
			}
		}

		rules = append(rules, rule)
	}

	if rules != nil {
		u.Rules = rules
	}

	return u, true
}

func checkNFTables() (*proto.NFTStatus, bool) {
	cmd := exec.Command("nft", "list", "ruleset")
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	ruleset := strings.TrimSpace(string(out))
	if ruleset == "" {
		return &proto.NFTStatus{Active: false}, true
	}

	return &proto.NFTStatus{
		Active: true,
		Rules:  ruleset,
	}, true
}
