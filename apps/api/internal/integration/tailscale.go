package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const tsBaseURL = "https://api.tailscale.com/api/v2"

type TSClient struct {
	apiKey  string
	tailnet string
	http    *http.Client
}

func NewTSClient(apiKey, tailnet string) *TSClient {
	return &TSClient{
		apiKey:  apiKey,
		tailnet: tailnet,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// TSDevice represents a Tailscale device/node.
type TSDevice struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Hostname        string   `json:"hostname"`
	User            string   `json:"user"`
	IPs             []string `json:"addresses"`
	OS              string   `json:"os"`
	Online          bool     `json:"online"`
	LastSeen        string   `json:"lastSeen"`
	Tags            []string `json:"tags"`
	ClientVersion   string   `json:"clientVersion"`
	UpdateAvailable bool     `json:"updateAvailable"`
}

// TSACL is the Tailscale ACL/HuJSON structure (simplified).
type TSACL struct {
	ACLs    []TSACLEntry    `json:"acls"`
	Groups  json.RawMessage `json:"groups,omitempty"`
	TagOwners json.RawMessage `json:"tagOwners,omitempty"`
	// ETag is for version checking
	ETag string `json:"-"`
}

type TSACLEntry struct {
	Action      string   `json:"action"`
	Users       []string `json:"users"`
	Ports       []string `json:"ports"`
}

// ListDevices returns all devices in the tailnet.
func (c *TSClient) ListDevices() ([]TSDevice, error) {
	body, err := c.get(fmt.Sprintf("/tailnet/%s/devices", c.tailnet))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Devices []TSDevice `json:"devices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}
	return resp.Devices, nil
}

// GetACL returns the ACL for the tailnet with its ETag for version tracking.
func (c *TSClient) GetACL() (*TSACL, error) {
	req, err := http.NewRequest("GET", tsBaseURL+"/tailnet/"+c.tailnet+"/acl", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var acl TSACL
	if err := json.Unmarshal(body, &acl); err != nil {
		return nil, fmt.Errorf("parse acl: %w", err)
	}

	acl.ETag = resp.Header.Get("ETag")
	return &acl, nil
}

func (c *TSClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", tsBaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		return nil, fmt.Errorf("tailscale api returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
