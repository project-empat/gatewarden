package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportService generates security posture reports, incident summaries,
// and node health overviews from database data.
type ReportService struct {
	db *pgxpool.Pool
}

func NewReportService(db *pgxpool.Pool) *ReportService {
	return &ReportService{db: db}
}

// ---------------------------------------------------------------------------
// Security Posture Report
// ---------------------------------------------------------------------------

// PostureReport is a comprehensive snapshot of infrastructure security.
type PostureReport struct {
	GeneratedAt       time.Time            `json:"generated_at"`
	NodeCount         int                  `json:"node_count"`
	OnlineCount       int                  `json:"online_count"`
	OfflineCount      int                  `json:"offline_count"`
	IncidentCount     int                  `json:"incident_count"`
	OpenIncidentCount int                  `json:"open_incident_count"`
	ResolvedCount     int                  `json:"resolved_count"`
	Exposures         ExposureSummary      `json:"exposures"`
	IntegrationCov    IntegrationCoverage  `json:"integration_coverage"`
	FirewallSummary   FirewallSummary      `json:"firewall_summary"`
}

type ExposureSummary struct {
	SSHPublic      int `json:"ssh_public"`
	SSHPasswordAuth int `json:"ssh_password_auth"`
	DockerSocket    int `json:"docker_socket"`
}

type IntegrationCoverage struct {
	TotalNodes       int `json:"total_nodes"`
	CrowdSecNodes    int `json:"crowdsec_nodes"`
	Fail2BanNodes    int `json:"fail2ban_nodes"`
	TailscaleNodes   int `json:"tailscale_nodes"`
	CloudflareNodes  int `json:"cloudflare_nodes"`
}

type FirewallSummary struct {
	TotalWithFirewall int `json:"total_with_firewall"`
	ActiveFirewalls   int `json:"active_firewalls"`
	InactiveFirewalls int `json:"inactive_firewalls"`
}

// SecurityPosture produces a full posture report from the database.
func (s *ReportService) SecurityPosture(ctx context.Context) (*PostureReport, error) {
	r := &PostureReport{GeneratedAt: time.Now().UTC()}

	// Node counts
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&r.NodeCount)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM nodes WHERE status = 'online'`).Scan(&r.OnlineCount)
	r.OfflineCount = r.NodeCount - r.OnlineCount

	// Incident counts
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents`).Scan(&r.IncidentCount)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents WHERE status = 'open'`).Scan(&r.OpenIncidentCount)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents WHERE status = 'resolved'`).Scan(&r.ResolvedCount)

	// Report-level parsing for exposures and integrations
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT ON (ar.node_id) ar.node_id, ar.report
		FROM agent_reports ar
		ORDER BY ar.node_id, ar.received_at DESC
	`)
	if err != nil {
		return r, nil
	}
	defer rows.Close()

	var reportCount int
	for rows.Next() {
		var raw json.RawMessage
		var nodeID string
		if err := rows.Scan(&nodeID, &raw); err != nil {
			continue
		}
		reportCount++

		var parsed map[string]interface{}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			continue
		}

		// SSH
		if ssh, ok := parsed["ssh"].(map[string]interface{}); ok {
			if exposed, _ := ssh["publicly_exposed"].(bool); exposed {
				r.Exposures.SSHPublic++
			}
			if pw, _ := ssh["password_auth"].(bool); pw {
				r.Exposures.SSHPasswordAuth++
			}
		}

		// Docker socket
		if docker, ok := parsed["docker"].(map[string]interface{}); ok {
			if exposed, _ := docker["socket_exposed"].(bool); exposed {
				r.Exposures.DockerSocket++
			}
		}

		// Integrations
		if cs, ok := parsed["crowdsec"].(map[string]interface{}); ok {
			if installed, _ := cs["installed"].(bool); installed {
				r.IntegrationCov.CrowdSecNodes++
			}
		}
		if f2b, ok := parsed["fail2ban"].(map[string]interface{}); ok {
			if installed, _ := f2b["installed"].(bool); installed {
				r.IntegrationCov.Fail2BanNodes++
			}
		}
		if ts, ok := parsed["tailscale"].(map[string]interface{}); ok {
			if installed, _ := ts["installed"].(bool); installed {
				r.IntegrationCov.TailscaleNodes++
			}
		}
		if cf, ok := parsed["cloudflare_tunnel"].(map[string]interface{}); ok {
			if installed, _ := cf["installed"].(bool); installed {
				r.IntegrationCov.CloudflareNodes++
			}
		}

		// Firewall
		if fw, ok := parsed["firewall"].(map[string]interface{}); ok {
			r.FirewallSummary.TotalWithFirewall++
			if active, _ := fw["active"].(bool); active {
				r.FirewallSummary.ActiveFirewalls++
			} else {
				r.FirewallSummary.InactiveFirewalls++
			}
		}
	}
	r.IntegrationCov.TotalNodes = reportCount

	return r, nil
}

// ---------------------------------------------------------------------------
// Incident Summary
// ---------------------------------------------------------------------------

// IncidentSummary aggregates incident data for the reporting view.
type IncidentSummary struct {
	BySeverity       map[string]int        `json:"by_severity"`
	ByStatus         map[string]int        `json:"by_status"`
	Total            int                   `json:"total"`
	OpenCount        int                   `json:"open_count"`
	ResolvedCount    int                   `json:"resolved_count"`
	TopAffectedNodes []NodeIncidentCount   `json:"top_affected_nodes"`
}

type NodeIncidentCount struct {
	NodeID   string `json:"node_id"`
	Hostname string `json:"hostname"`
	Count    int    `json:"count"`
}

// IncidentSummaryReport returns aggregated incident statistics.
func (s *ReportService) IncidentSummaryReport(ctx context.Context) (*IncidentSummary, error) {
	summary := &IncidentSummary{
		BySeverity: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	rows, err := s.db.Query(ctx, `
		SELECT severity, status, COUNT(*) as cnt
		FROM incidents
		GROUP BY severity, status
		ORDER BY severity, status
	`)
	if err != nil {
		return nil, fmt.Errorf("query incident summary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var severity, status string
		var count int
		if err := rows.Scan(&severity, &status, &count); err != nil {
			continue
		}
		summary.BySeverity[severity] += count
		summary.ByStatus[status] += count
		summary.Total += count
	}
	summary.OpenCount = summary.ByStatus["open"]
	summary.ResolvedCount = summary.ByStatus["resolved"]

	// Top affected nodes
	topRows, err := s.db.Query(ctx, `
		SELECT i.node_id, n.hostname, COUNT(*) as cnt
		FROM incidents i
		JOIN nodes n ON n.id = i.node_id
		GROUP BY i.node_id, n.hostname
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return summary, nil
	}
	defer topRows.Close()

	for topRows.Next() {
		var n NodeIncidentCount
		if err := topRows.Scan(&n.NodeID, &n.Hostname, &n.Count); err != nil {
			continue
		}
		summary.TopAffectedNodes = append(summary.TopAffectedNodes, n)
	}
	if summary.TopAffectedNodes == nil {
		summary.TopAffectedNodes = []NodeIncidentCount{}
	}

	return summary, nil
}

// ---------------------------------------------------------------------------
// Node Health Overview
// ---------------------------------------------------------------------------

// NodeHealthOverview provides per-node health metrics.
type NodeHealthOverview struct {
	Nodes []NodeHealth `json:"nodes"`
}

// NodeHealth contains the health snapshot for a single node.
type NodeHealth struct {
	ID            string  `json:"id"`
	Hostname      string  `json:"hostname"`
	IP            string  `json:"ip"`
	Status        string  `json:"status"`
	LastSeen      string  `json:"last_seen"`
	OS            string  `json:"os"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}

// NodeHealthReport returns health data for all nodes, pulled from the
// latest agent report of each.
func (s *ReportService) NodeHealthReport(ctx context.Context) (*NodeHealthOverview, error) {
	overview := &NodeHealthOverview{}

	rows, err := s.db.Query(ctx, `
		SELECT n.id, n.hostname, n.ip, n.status, n.last_seen, n.os,
			   ar.report
		FROM nodes n
		LEFT JOIN LATERAL (
			SELECT report FROM agent_reports
			WHERE node_id = n.id
			ORDER BY received_at DESC
			LIMIT 1
		) ar ON true
		WHERE n.status != 'deleted'
		ORDER BY n.hostname
	`)
	if err != nil {
		return nil, fmt.Errorf("query node health: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var h NodeHealth
		var rawReport []byte
		if err := rows.Scan(&h.ID, &h.Hostname, &h.IP, &h.Status, &h.LastSeen, &h.OS, &rawReport); err != nil {
			continue
		}

		if rawReport != nil {
			var parsed map[string]interface{}
			if err := json.Unmarshal(rawReport, &parsed); err == nil {
				if sys, ok := parsed["system"].(map[string]interface{}); ok {
					h.CPUPercent, _ = sys["cpu_percent"].(float64)
					h.MemoryPercent, _ = sys["memory_percent"].(float64)
					h.DiskPercent, _ = sys["disk_percent"].(float64)
				}
				if uptime, ok := parsed["uptime_seconds"].(float64); ok {
					h.UptimeSeconds = uptime
				}
			}
		}

		overview.Nodes = append(overview.Nodes, h)
	}
	if overview.Nodes == nil {
		overview.Nodes = []NodeHealth{}
	}

	return overview, nil
}
