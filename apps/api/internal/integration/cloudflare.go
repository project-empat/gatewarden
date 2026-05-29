package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const cfBaseURL = "https://api.cloudflare.com/client/v4"

type CFClient struct {
	apiToken string
	http     *http.Client
}

func NewCFClient(apiToken string) *CFClient {
	return &CFClient{
		apiToken: apiToken,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

// cfResponse is a generic Cloudflare API response wrapper.
type cfResponse struct {
	Success  bool            `json:"success"`
	Errors   []cfError       `json:"errors"`
	Result   json.RawMessage `json:"result"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Account represents a Cloudflare account.
type CFAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Tunnel represents a Cloudflare Tunnel (Argo Tunnel).
type CFTunnel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Created string `json:"created_at"`
	// Connectors shows active connections
	Connectors []CFConnector `json:"connectors,omitempty"`
}

// CFConnector represents a tunnel connector/agent.
type CFConnector struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Arch      string `json:"arch"`
	State     string `json:"state"` // "healthy", "degraded", "down"
	Hostname  string `json:"hostname"`
}

// GetAccounts returns the list of Cloudflare accounts accessible with this token.
func (c *CFClient) GetAccounts() ([]CFAccount, error) {
	body, err := c.get("/accounts?per_page=50")
	if err != nil {
		return nil, err
	}
	var resp cfResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse accounts: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("cloudflare api error: %s", extractCFError(resp.Errors))
	}
	var accounts []CFAccount
	if err := json.Unmarshal(resp.Result, &accounts); err != nil {
		return nil, fmt.Errorf("parse account list: %w", err)
	}
	return accounts, nil
}

// ListTunnels returns all CF Tunnels for an account.
func (c *CFClient) ListTunnels(accountID string) ([]CFTunnel, error) {
	body, err := c.get(fmt.Sprintf("/accounts/%s/cfd_tunnel?is_deleted=false&per_page=50", accountID))
	if err != nil {
		return nil, err
	}
	var resp cfResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse tunnels: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("cloudflare api error: %s", extractCFError(resp.Errors))
	}
	var tunnels []CFTunnel
	if err := json.Unmarshal(resp.Result, &tunnels); err != nil {
		return nil, fmt.Errorf("parse tunnel list: %w", err)
	}
	return tunnels, nil
}

// GetTunnel returns a single tunnel with its connector status.
func (c *CFClient) GetTunnel(accountID, tunnelID string) (*CFTunnel, error) {
	body, err := c.get(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s", accountID, tunnelID))
	if err != nil {
		return nil, err
	}
	var resp cfResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse tunnel: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("cloudflare api error: %s", extractCFError(resp.Errors))
	}
	var tunnel CFTunnel
	if err := json.Unmarshal(resp.Result, &tunnel); err != nil {
		return nil, fmt.Errorf("parse tunnel detail: %w", err)
	}
	return &tunnel, nil
}

func (c *CFClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", cfBaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cloudflare api returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func extractCFError(errors []cfError) string {
	for _, e := range errors {
		return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
	}
	return "unknown error"
}
