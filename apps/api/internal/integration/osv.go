package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const osvBaseURL = "https://api.osv.dev/v1"

// OSVClient queries the OSV (osv.dev) vulnerability database. Free, open,
// no API key required.
type OSVClient struct {
	baseURL string
	http    *http.Client
}

func NewOSVClient() *OSVClient {
	return &OSVClient{
		baseURL: osvBaseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// VulnQuery matches packages by name + ecosystem and optionally version.
type VulnQuery struct {
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	Version string `json:"version,omitempty"`
}

type VulnSummary struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Severity []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
	Modified string `json:"modified"`
}

type VulnQueryResponse struct {
	Vulns []VulnSummary `json:"vulns"`
}

// QueryVulnerabilities returns CVEs affecting a package@version in the given
// ecosystem. An empty result means no known vulnerabilities.
func (c *OSVClient) QueryVulnerabilities(ctx context.Context, name, version, ecosystem string) ([]VulnSummary, error) {
	var q VulnQuery
	q.Package.Name = name
	q.Package.Ecosystem = ecosystem
	q.Version = version

	body, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("encode osv query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/query", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osv query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("osv query status %d", resp.StatusCode)
	}

	var out VulnQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode osv response: %w", err)
	}
	if out.Vulns == nil {
		out.Vulns = []VulnSummary{}
	}
	return out.Vulns, nil
}
