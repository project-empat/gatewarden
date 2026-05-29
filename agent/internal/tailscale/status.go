package tailscale

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// tsStatus represents the JSON output of `tailscale status --json`.
type tsStatus struct {
	Version   string           `json:"Version"`
	Peer      map[string]tsPeer `json:"Peer"`
	Self      tsSelf           `json:"Self"`
	CurrentTailnet struct {
		Name       string `json:"Name"`
		MagicDNS   bool   `json:"MagicDNS"`
	} `json:"CurrentTailnet"`
}

type tsSelf struct {
	ID        string `json:"ID"`
	PublicKey string `json:"PublicKey"`
	HostName  string `json:"HostName"`
	DNSName   string `json:"DNSName"`
	TailAddr  string `json:"TailAddr"`
	Online    bool   `json:"Online"`
}

type tsPeer struct {
	ID        string `json:"ID"`
	HostName  string `json:"HostName"`
	DNSName   string `json:"DNSName"`
	TailAddr  string `json:"TailAddr"`
	Online    bool   `json:"Online"`
	OS        string `json:"OS"`
}

// Check returns Tailscale status by querying tailscale CLI.
func Check() (*proto.TailscaleStatus, error) {
	// Check if tailscale is installed
	if err := exec.Command("tailscale", "version").Run(); err != nil {
		return nil, nil // Tailscale not available
	}

	status := &proto.TailscaleStatus{
		Installed: true,
	}

	// Get version
	if out, err := exec.Command("tailscale", "version").Output(); err == nil {
		status.Version = strings.TrimSpace(string(out))
	}

	// Check ACL version via tailscale debug
	if out, err := exec.Command("tailscale", "debug", "acls").Output(); err == nil {
		status.ACLVersion = strings.TrimSpace(string(out))
	}
 
	// Try JSON status for maximum detail
	if out, err := exec.Command("tailscale", "status", "--json").Output(); err == nil {
		var ts tsStatus
		if json.Unmarshal(out, &ts) == nil {
			status.NodeName = ts.Self.HostName
			status.NodeIP = ts.Self.TailAddr
			status.Online = ts.Self.Online
			status.Running = true
			status.PeersCount = len(ts.Peer)
			return status, nil
		}
	}

	// Fallback to plain text status parsing
	cmd := exec.Command("tailscale", "status")
	out, err := cmd.Output()
	if err != nil {
		return &proto.TailscaleStatus{
			Installed: true,
			Running:   false,
		}, nil
	}

	text := strings.TrimSpace(string(out))
	lines := strings.Split(text, "\n")

	// When running, tailscale status prints the current node as the first line
	// Format: 100.x.x.x    hostname           username@hostname    linux   active; direct
	if len(lines) > 0 && !strings.Contains(text, "Tailscale is stopped") {
		status.Running = true
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			status.NodeIP = parts[0]
			status.NodeName = parts[1]
			status.Online = strings.Contains(text, "active")
		}
		status.PeersCount = len(lines) - 1
	}

	return status, nil
}
