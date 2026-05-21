package proto

import "encoding/json"

type CheckResult struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Hostname   string `json:"hostname,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
}

type Heartbeat struct {
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
	Uptime    int64  `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

type AgentReport struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
