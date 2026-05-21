package firewall

import (
	"os/exec"
	"strings"
)

func CheckNFTables() (map[string]interface{}, error) {
	cmd := exec.Command("nft", "list", "ruleset")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"name":    "nftables",
			"status":  "unavailable",
			"message": "nftables not found or not running",
		}, nil
	}

	ruleset := strings.TrimSpace(string(output))
	if ruleset == "" {
		return map[string]interface{}{
			"name":    "nftables",
			"status":  "inactive",
			"message": "No nftables rules defined",
		}, nil
	}

	return map[string]interface{}{
		"name":    "nftables",
		"status":  "active",
		"message": ruleset,
	}, nil
}
