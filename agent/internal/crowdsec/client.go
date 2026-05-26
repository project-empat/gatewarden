package crowdsec

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

type csAlert struct {
	ID     int    `json:"id"`
	Source struct {
		IP string `json:"ip"`
	} `json:"source"`
	Scenario   string `json:"scenario"`
	CreatedAt  string `json:"created_at"`
	Remediation bool   `json:"remediation"`
}

type csDecision struct {
	ID        int    `json:"id"`
	Origin    string `json:"origin"`
	Scenario  string `json:"scenario"`
	Value     string `json:"value"`
	Duration  string `json:"duration"`
	CreatedAt string `json:"created_at"`
}

type csBouncer struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IP       string `json:"ip"`
	LastPull string `json:"last_pull"`
	Status   string `json:"status"`
}

// Check returns CrowdSec status by querying cscli.
func Check() (*proto.CrowdSecStatus, error) {
	// Check if cscli is available
	if err := exec.Command("cscli", "version").Run(); err != nil {
		return nil, nil // CrowdSec not available
	}

	status := &proto.CrowdSecStatus{
		Installed: true,
	}

	// Check if LAPI is reachable
	if err := exec.Command("cscli", "lapi", "status").Run(); err == nil {
		status.Running = true
	}

	// Get decisions count
	if out, err := exec.Command("cscli", "decisions", "list", "-o", "json").Output(); err == nil {
		var decisions []csDecision
		if json.Unmarshal(out, &decisions) == nil {
			status.ActiveDecisions = len(decisions)
		}
	}

	// Get alerts count (last hour)
	if out, err := exec.Command("cscli", "alerts", "list", "-o", "json").Output(); err == nil {
		var alerts []csAlert
		if json.Unmarshal(out, &alerts) == nil {
			// Count alerts from last hour
			status.AlertsLastHour = len(alerts)
		}
	}

	// Get bouncers
	if out, err := exec.Command("cscli", "bouncers", "list", "-o", "json").Output(); err == nil {
		var bouncers []csBouncer
		if json.Unmarshal(out, &bouncers) == nil {
			for _, b := range bouncers {
				if b.Status == "ok" {
					status.Bouncers = append(status.Bouncers, b.Name)
				}
			}
		}
	}

	if status.Bouncers == nil {
		status.Bouncers = []string{}
	}

	return status, nil
}

// Helper for checking if command produced valid output
func hasOutput(name string, args ...string) bool {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
