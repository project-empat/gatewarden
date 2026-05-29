package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatewarden/api/internal/model"
)

type NodeService struct {
	db *pgxpool.Pool
}

func NewNodeService(db *pgxpool.Pool) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) List(ctx context.Context) ([]model.Node, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, hostname, ip, os, status, labels, last_seen, created_at FROM nodes ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Hostname, &n.IP, &n.OS, &n.Status, &n.Labels, &n.LastSeen, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (s *NodeService) Get(ctx context.Context, id string) (*model.Node, error) {
	var n model.Node
	err := s.db.QueryRow(ctx,
		`SELECT id, name, hostname, ip, os, status, labels, last_seen, created_at FROM nodes WHERE id = $1`,
		id,
	).Scan(&n.ID, &n.Name, &n.Hostname, &n.IP, &n.OS, &n.Status, &n.Labels, &n.LastSeen, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	return &n, nil
}

func (s *NodeService) DashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	var stats model.DashboardStats
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&stats.TotalNodes)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM nodes WHERE status = 'online'`).Scan(&stats.OnlineNodes)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents`).Scan(&stats.TotalIncidents)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents WHERE status = 'open'`).Scan(&stats.OpenIncidents)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetLatestReport retrieves the most recent agent report JSON for a node.
func (s *NodeService) GetLatestReport(ctx context.Context, nodeID string) (json.RawMessage, error) {
	var report json.RawMessage
	err := s.db.QueryRow(ctx,
		`SELECT report FROM agent_reports WHERE node_id = $1 ORDER BY received_at DESC LIMIT 1`,
		nodeID,
	).Scan(&report)
	if err != nil {
		return nil, fmt.Errorf("get latest report: %w", err)
	}
	return report, nil
}

// NodeSecuritySummary aggregates security indicators from all nodes.
type NodeSecuritySummary struct {
	ExposedSSH     int `json:"exposed_ssh"`
	DockerExposed  int `json:"docker_exposed"`
	PasswordAuthSSH int `json:"password_auth_ssh"`
	TotalIncidents  int `json:"total_incidents"`
	OpenIncidents   int `json:"open_incidents"`
	TotalNodes      int `json:"total_nodes"`
	OnlineNodes     int `json:"online_nodes"`
	HighSeverity    int `json:"high_severity"`
	// Per-node CrowdSec
	CrowdSecNodes int `json:"crowdsec_nodes"`
	TotalDecisions int `json:"total_decisions"`
	TotalAlerts    int `json:"total_alerts"`
	// Per-node Fail2Ban
	Fail2BanJailsTotal int `json:"fail2ban_jails_total"`
	Fail2BanBansTotal  int `json:"fail2ban_bans_total"`
}

// SecuritySummary collects security indicators across all nodes.
func (s *NodeService) SecuritySummary(ctx context.Context) (*NodeSecuritySummary, error) {
	summary := &NodeSecuritySummary{}

	// Count nodes
	rows, err := s.db.Query(ctx,
		`SELECT id, status FROM nodes ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query nodes: %w", err)
	}
	defer rows.Close()

	type nodeInfo struct {
		id     string
		status string
	}
	var nodes []nodeInfo
	for rows.Next() {
		var n nodeInfo
		if err := rows.Scan(&n.id, &n.status); err != nil {
			continue
		}
		nodes = append(nodes, n)
	}

	summary.TotalNodes = len(nodes)
	for _, n := range nodes {
		if n.status == "online" {
			summary.OnlineNodes++
		}
	}

	// Count incidents
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents`).Scan(&summary.TotalIncidents)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents WHERE status = 'open'`).Scan(&summary.OpenIncidents)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM incidents WHERE severity IN ('critical','high') AND status = 'open'`).Scan(&summary.HighSeverity)

	// Process latest reports for security indicators
	type reportRow struct {
		nodeID string
		report json.RawMessage
	}
	reportRows, err := s.db.Query(ctx,
		`SELECT DISTINCT ON (ar.node_id) ar.node_id, ar.report
		 FROM agent_reports ar
		 ORDER BY ar.node_id, ar.received_at DESC`,
	)
	if err != nil {
		return summary, nil // Non-fatal; return what we have
	}
	defer reportRows.Close()

	var reports []reportRow
	for reportRows.Next() {
		var r reportRow
		if err := reportRows.Scan(&r.nodeID, &r.report); err != nil {
			continue
		}
		reports = append(reports, r)
	}

	for _, r := range reports {
		var parsed map[string]interface{}
		if err := json.Unmarshal(r.report, &parsed); err != nil {
			continue
		}

		// SSH
		if ssh, ok := parsed["ssh"].(map[string]interface{}); ok {
			if exposed, _ := ssh["publicly_exposed"].(bool); exposed {
				summary.ExposedSSH++
			}
			if pw, _ := ssh["password_auth"].(bool); pw {
				summary.PasswordAuthSSH++
			}
		}

		// Docker
		if docker, ok := parsed["docker"].(map[string]interface{}); ok {
			if socket, _ := docker["socket_exposed"].(bool); socket {
				summary.DockerExposed++
			}
		}

		// CrowdSec
		if cs, ok := parsed["crowdsec"].(map[string]interface{}); ok {
			if installed, _ := cs["installed"].(bool); installed {
				summary.CrowdSecNodes++
				if decisions, _ := cs["active_decisions"].(float64); decisions > 0 {
					summary.TotalDecisions += int(decisions)
				}
				if alerts, _ := cs["alerts_last_hour"].(float64); alerts > 0 {
					summary.TotalAlerts += int(alerts)
				}
			}
		}

		// Fail2Ban
		if f2b, ok := parsed["fail2ban"].(map[string]interface{}); ok {
			if jails, ok := f2b["jails"].([]interface{}); ok {
				summary.Fail2BanJailsTotal += len(jails)
				for _, j := range jails {
					if jm, ok := j.(map[string]interface{}); ok {
						if bans, _ := jm["currently_banned"].(float64); bans > 0 {
							summary.Fail2BanBansTotal += int(bans)
						}
					}
				}
			}
		}
	}

	return summary, nil
}

func (s *NodeService) Upsert(ctx context.Context, n model.Node) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO nodes (id, name, hostname, ip, os, status, labels, last_seen)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (hostname) DO UPDATE SET
		   name = EXCLUDED.name,
		   ip = EXCLUDED.ip,
		   os = EXCLUDED.os,
		   status = EXCLUDED.status,
		   labels = EXCLUDED.labels,
		   last_seen = NOW()`,
		uuid.New().String(), n.Name, n.Hostname, n.IP, n.OS, n.Status, n.Labels,
	)
	return err
}
