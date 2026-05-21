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

type Reporter struct {
	serverURL string
	apiKey    string
	hostname  string
	client    *http.Client
}

func New(serverURL, apiKey, hostname string) *Reporter {
	return &Reporter{
		serverURL: serverURL,
		apiKey:    apiKey,
		hostname:  hostname,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *Reporter) Report(eventType, payload string) error {
	body, _ := json.Marshal(proto.AgentReport{
		Type: eventType,
		Payload: json.RawMessage(payload),
	})

	req, err := http.NewRequest("POST", r.serverURL+"/api/agent/report", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("send report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("report returned %d", resp.StatusCode)
	}

	return nil
}

func (r *Reporter) Close() {
	r.client.CloseIdleConnections()
}
