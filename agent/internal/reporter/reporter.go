package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gatewarden/agent/pkg/proto"
)

// Reporter handles sending reports and heartbeats to the API server.
type Reporter struct {
	serverURL string
	apiKey    string
	hostname  string
	nodeID    string
	client    *http.Client
}

// New creates a new Reporter.
func New(serverURL, apiKey, hostname string) *Reporter {
	return &Reporter{
		serverURL: serverURL,
		apiKey:    apiKey,
		hostname:  hostname,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SetNodeID stores the node ID received from registration.
func (r *Reporter) SetNodeID(id string) {
	r.nodeID = id
}

// SendReport sends a full agent status report to the API.
func (r *Reporter) SendReport(report *proto.AgentReport) error {
	report.NodeID = r.nodeID

	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	return r.post("/api/agent/report", body)
}

// SendHeartbeat sends a lightweight heartbeat to the API.
func (r *Reporter) SendHeartbeat() error {
	hb := proto.Heartbeat{
		Hostname:  r.hostname,
		Version:   agentVersion,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(hb)
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}

	return r.post("/api/agent/heartbeat", body)
}

// Register sends an initial registration request and stores the returned node ID.
func (r *Reporter) Register() error {
	reg := struct {
		Hostname string `json:"hostname"`
		Version  string `json:"version"`
	}{
		Hostname: r.hostname,
		Version:  agentVersion,
	}

	body, err := json.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal registration: %w", err)
	}

	resp, err := r.request("POST", "/api/agent/register", body)
	if err != nil {
		return fmt.Errorf("register request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		NodeID string `json:"node_id"`
		APIKey string `json:"api_key,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode registration response: %w", err)
	}

	r.nodeID = result.NodeID
	if result.APIKey != "" {
		r.apiKey = result.APIKey
	}

	log.Printf("registered as node %s", r.nodeID)
	return nil
}

func (r *Reporter) post(path string, body []byte) error {
	resp, err := r.request("POST", path, body)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		log.Printf("auth rejected on %s, re-registering", path)
		if err := r.Register(); err != nil {
			return fmt.Errorf("re-registration failed: %w", err)
		}
		// Retry the original request with new credentials
		resp, err = r.request("POST", path, body)
		if err != nil {
			return fmt.Errorf("retry after re-registration failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("%s returned %d after re-registration", path, resp.StatusCode)
		}
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s returned %d", path, resp.StatusCode)
	}

	return nil
}

func (r *Reporter) request(method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, r.serverURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return resp, nil
}

// Close cleans up the HTTP client.
func (r *Reporter) Close() {
	r.client.CloseIdleConnections()
}

// agentVersion is set at build time.
var agentVersion = "0.1.0"
