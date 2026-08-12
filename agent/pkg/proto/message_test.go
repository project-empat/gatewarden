package proto

import (
	"encoding/json"
	"testing"
)

func TestAgentReportJSON(t *testing.T) {
	report := &AgentReport{
		Hostname:      "test-server",
		OS:            "Ubuntu 22.04",
		Kernel:        "5.15.0-generic",
		UptimeSeconds: 3600,
		SSH: &SSHStatus{
			Port:            22,
			PasswordAuth:    false,
			RootLogin:       "prohibit-password",
			PubkeyAuth:      true,
			PubliclyExposed: true,
			MaxAuthTries:    3,
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded AgentReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Hostname != "test-server" {
		t.Errorf("Hostname = %q, want \"test-server\"", decoded.Hostname)
	}
	if decoded.OS != "Ubuntu 22.04" {
		t.Errorf("OS = %q, want \"Ubuntu 22.04\"", decoded.OS)
	}
	if decoded.SSH == nil {
		t.Fatal("SSH should not be nil")
	}
	if decoded.SSH.Port != 22 {
		t.Errorf("SSH.Port = %d, want 22", decoded.SSH.Port)
	}
	if decoded.SSH.PasswordAuth {
		t.Error("SSH.PasswordAuth should be false")
	}
}

func TestAgentReportOmitEmpty(t *testing.T) {
	report := &AgentReport{
		Hostname: "minimal",
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// These should be omitted (json tags with omitempty)
	if _, ok := decoded["docker"]; ok {
		t.Error("docker should be omitted when nil")
	}
	if _, ok := decoded["crowdsec"]; ok {
		t.Error("crowdsec should be omitted when nil")
	}
	if _, ok := decoded["auth_log"]; ok {
		t.Error("auth_log should be omitted when nil")
	}

	// These should be present
	if decoded["hostname"] != "minimal" {
		t.Errorf("hostname = %v, want \"minimal\"", decoded["hostname"])
	}
}

func TestHeartbeatJSON(t *testing.T) {
	hb := Heartbeat{
		Hostname:  "node-1",
		Version:   "0.1.0",
		Uptime:    86400,
		Timestamp: "2026-06-29T12:00:00Z",
	}

	data, err := json.Marshal(hb)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded Heartbeat
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Hostname != "node-1" {
		t.Errorf("Hostname = %q, want \"node-1\"", decoded.Hostname)
	}
	if decoded.Version != "0.1.0" {
		t.Errorf("Version = %q, want \"0.1.0\"", decoded.Version)
	}
	if decoded.Uptime != 86400 {
		t.Errorf("Uptime = %d, want 86400", decoded.Uptime)
	}
}

func TestListeningPortJSON(t *testing.T) {
	lp := ListeningPort{
		Port:     80,
		Protocol: "tcp",
		Process:  "nginx",
		Exposed:  true,
	}

	data, err := json.Marshal(lp)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded ListeningPort
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Port != 80 {
		t.Errorf("Port = %d, want 80", decoded.Port)
	}
	if decoded.Process != "nginx" {
		t.Errorf("Process = %q, want \"nginx\"", decoded.Process)
	}
	if !decoded.Exposed {
		t.Error("Exposed should be true")
	}
}
