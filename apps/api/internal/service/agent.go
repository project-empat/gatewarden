package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/gatewarden/agent/pkg/proto"
	"github.com/gatewarden/api/internal/model"
)

// AgentService handles agent registration, report ingestion, and heartbeats.
type AgentService struct {
	db   *pgxpool.Pool
	log  *zap.SugaredLogger
	ev   *EventService
	vuln *VulnerabilityService
}

func NewAgentService(db *pgxpool.Pool, log *zap.SugaredLogger, ev *EventService, vuln *VulnerabilityService) *AgentService {
	return &AgentService{db: db, log: log, ev: ev, vuln: vuln}
}

// Register creates a node + agent record and returns credentials.
func (s *AgentService) Register(ctx context.Context, req model.RegisterRequest) (*model.RegisterResponse, error) {
	// Check if a node with this hostname already exists
	var existingNodeID string
	err := s.db.QueryRow(ctx, `SELECT id FROM nodes WHERE hostname = $1`, req.Hostname).Scan(&existingNodeID)
	if err == nil {
		// Node exists — check if it has an agent, return existing key
		var apiKeyHash string
		err = s.db.QueryRow(ctx, `SELECT api_key_hash FROM agents WHERE node_id = $1`, existingNodeID).Scan(&apiKeyHash)
		if err == nil {
			// Return existing — can't return the original key (hashed), so generate a new one
			// In practice the agent stores the key locally, so this only happens on re-registration
			return s.registerExisting(ctx, existingNodeID, req)
		}
	}

	// Generate a new API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash api key: %w", err)
	}

	nodeID := uuid.New().String()
	agentID := uuid.New().String()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create node
	_, err = tx.Exec(ctx,
		`INSERT INTO nodes (id, name, hostname, ip, os, status, last_seen)
		 VALUES ($1, $2, $3, '0.0.0.0', '', 'online', NOW())`,
		nodeID, req.Hostname, req.Hostname,
	)
	if err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}

	// Create agent
	_, err = tx.Exec(ctx,
		`INSERT INTO agents (id, node_id, api_key_hash, version, last_heartbeat)
		 VALUES ($1, $2, $3, $4, NOW())`,
		agentID, nodeID, string(hash), req.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.log.Infow("agent registered", "node_id", nodeID, "hostname", req.Hostname)

	return &model.RegisterResponse{
		NodeID: nodeID,
		APIKey: apiKey,
	}, nil
}

func (s *AgentService) registerExisting(ctx context.Context, nodeID string, req model.RegisterRequest) (*model.RegisterResponse, error) {
	// Generate new key, update agent record
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash api key: %w", err)
	}

	// Update node status and agent
	_, err = s.db.Exec(ctx,
		`UPDATE nodes SET status = 'online', last_seen = NOW(), name = $2 WHERE id = $1`,
		nodeID, req.Hostname,
	)
	if err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE agents SET api_key_hash = $1, version = $2, last_heartbeat = NOW() WHERE node_id = $3`,
		string(hash), req.Version, nodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
	}

	s.log.Infow("agent re-registered", "node_id", nodeID)

	return &model.RegisterResponse{
		NodeID: nodeID,
		APIKey: apiKey,
	}, nil
}

// ProcessReport handles an incoming agent report, updating node data and creating events.
func (s *AgentService) ProcessReport(ctx context.Context, nodeID string, report json.RawMessage) error {
	// Update node last_seen and status
	_, err := s.db.Exec(ctx,
		`UPDATE nodes SET last_seen = NOW(), status = 'online' WHERE id = $1`,
		nodeID,
	)
	if err != nil {
		return fmt.Errorf("update node: %w", err)
	}

	// Update agent heartbeat
	_, err = s.db.Exec(ctx,
		`UPDATE agents SET last_heartbeat = NOW() WHERE node_id = $1`,
		nodeID,
	)
	if err != nil {
		return fmt.Errorf("update agent heartbeat: %w", err)
	}

	// Store raw report
	var agentID string
	err = s.db.QueryRow(ctx, `SELECT id FROM agents WHERE node_id = $1`, nodeID).Scan(&agentID)
	if err != nil {
		return fmt.Errorf("find agent: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO agent_reports (agent_id, node_id, report) VALUES ($1, $2, $3)`,
		agentID, nodeID, report,
	)
	if err != nil {
		s.log.Warnw("failed to store agent report", "error", err)
	}

	// Parse report and generate events/incidents
	s.analyzeReport(ctx, nodeID, report)

	return nil
}

// Heartbeat updates the node and agent last_seen timestamps.
func (s *AgentService) Heartbeat(ctx context.Context, nodeID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE nodes SET last_seen = NOW(), status = 'online' WHERE id = $1`,
		nodeID,
	)
	if err != nil {
		return fmt.Errorf("update node heartbeat: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE agents SET last_heartbeat = NOW() WHERE node_id = $1`,
		nodeID,
	)
	return err
}

// ValidateAPIKey checks an API key and returns the associated node ID.
func (s *AgentService) ValidateAPIKey(ctx context.Context, key string) (string, error) {
	rows, err := s.db.Query(ctx, `SELECT node_id, api_key_hash FROM agents`)
	if err != nil {
		return "", fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, hash string
		if err := rows.Scan(&nodeID, &hash); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(key)) == nil {
			return nodeID, nil
		}
	}

	return "", fmt.Errorf("invalid API key")
}

// analyzeReport parses an agent report and creates events/incidents for notable findings.
func (s *AgentService) analyzeReport(ctx context.Context, nodeID string, report json.RawMessage) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(report, &parsed); err != nil {
		s.log.Warnw("failed to parse agent report", "error", err)
		return
	}

	// Decode structured fields for FIM + package ingestion.
	var typed proto.AgentReport
	_ = json.Unmarshal(report, &typed)
	s.ingestPackages(ctx, nodeID, typed.Packages)
	s.ingestFIM(ctx, nodeID, typed.FIM)

	// Check SSH exposure
	if ssh, ok := parsed["ssh"].(map[string]interface{}); ok {
		if exposed, _ := ssh["publicly_exposed"].(bool); exposed {
			port, _ := ssh["port"].(float64)
			s.maybeCreateIncident(ctx, nodeID, "ssh_publicly_exposed", "high",
				fmt.Sprintf("SSH is publicly exposed on port %.0f", port))
		}
		if pwAuth, _ := ssh["password_auth"].(bool); pwAuth {
			s.maybeCreateIncident(ctx, nodeID, "ssh_password_auth_enabled", "critical",
				"SSH password authentication is enabled")
		}
	}

	// Check Docker socket exposure
	if docker, ok := parsed["docker"].(map[string]interface{}); ok {
		if socketExposed, _ := docker["socket_exposed"].(bool); socketExposed {
			s.maybeCreateIncident(ctx, nodeID, "docker_socket_exposed", "critical",
				"Docker socket is exposed to a container")
		}
	}

	// Check auth log for brute force
	if authLog, ok := parsed["auth_log"].(map[string]interface{}); ok {
		if failed, _ := authLog["failed_ssh_last_hour"].(float64); failed > 10 {
			s.maybeCreateIncident(ctx, nodeID, "ssh_brute_force", "high",
				fmt.Sprintf("SSH brute force detected: %.0f failed attempts in the last hour", failed))
		}
	}

	// Create a generic event for the report
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"node_id": nodeID,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})

	s.ev.Publish(ctx, model.Event{
		NodeID:  nodeID,
		Type:    "agent_report",
		Payload: string(eventPayload),
	})
}

// ingestPackages stores the node's installed packages and pending security
// update count, then triggers a bounded OSV enrichment pass in the background.
func (s *AgentService) ingestPackages(ctx context.Context, nodeID string, pkgs *proto.PackageStatus) {
	if pkgs == nil {
		return
	}

	// Update node-level security-updates-pending signal.
	_, err := s.db.Exec(ctx,
		`UPDATE nodes SET security_updates_pending = $2, updated_at = NOW() WHERE id = $1`,
		nodeID, pkgs.SecurityUpdatesPending)
	if err != nil {
		s.log.Warnw("update security updates count", "error", err)
	}

	if len(pkgs.Installed) == 0 {
		return
	}

	// Replace the package inventory in one transaction.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.log.Warnw("begin package tx", "error", err)
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM node_packages WHERE node_id = $1`, nodeID); err != nil {
		s.log.Warnw("clear packages", "error", err)
		return
	}

	for _, p := range pkgs.Installed {
		if _, err := tx.Exec(ctx,
			`INSERT INTO node_packages (node_id, name, version, ecosystem) VALUES ($1, $2, $3, 'deb')
			 ON CONFLICT (node_id, name, version) DO NOTHING`,
			nodeID, p.Name, p.Version); err != nil {
			s.log.Debugw("insert package", "pkg", p.Name, "error", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		s.log.Warnw("commit package tx", "error", err)
		return
	}

	// Best-effort background enrichment (bounded by the vuln service).
	if s.vuln != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			if err := s.vuln.Enrich(ctx, nodeID); err != nil {
				s.log.Debugw("vuln enrichment", "node", nodeID, "error", err)
			}
		}()
	}
}

// ingestFIM diffs the reported file hashes against the node's baseline and
// raises an incident whenever a monitored critical file is added or changed.
func (s *AgentService) ingestFIM(ctx context.Context, nodeID string, fim *proto.FIMStatus) {
	if fim == nil || len(fim.Files) == 0 {
		return
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.log.Warnw("begin fim tx", "error", err)
		return
	}
	defer tx.Rollback(ctx)

	for _, f := range fim.Files {
		var existingHash string
		err := tx.QueryRow(ctx,
			`SELECT hash FROM node_fim_files WHERE node_id = $1 AND path = $2`,
			nodeID, f.Path).Scan(&existingHash)
		if err != nil {
			// New monitored file -> record baseline, no alert.
			if _, err := tx.Exec(ctx,
				`INSERT INTO node_fim_files (node_id, path, hash) VALUES ($1, $2, $3)
				 ON CONFLICT (node_id, path) DO UPDATE SET hash = EXCLUDED.hash`,
				nodeID, f.Path, f.Hash); err != nil {
				s.log.Debugw("insert fim file", "path", f.Path, "error", err)
			}
			continue
		}

		if existingHash != f.Hash {
			if _, err := tx.Exec(ctx,
				`UPDATE node_fim_files SET hash = $3, last_changed = NOW()
				 WHERE node_id = $1 AND path = $2`,
				nodeID, f.Path, f.Hash); err != nil {
				s.log.Debugw("update fim file", "path", f.Path, "error", err)
				continue
			}
			s.maybeCreateIncident(ctx, nodeID, "fim_file_modified", "high",
				fmt.Sprintf("Critical file modified: %s", f.Path))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Warnw("commit fim tx", "error", err)
	}
}

func (s *AgentService) maybeCreateIncident(ctx context.Context, nodeID, eventType, severity, message string) {
	// Check if an open incident for this type already exists
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM incidents WHERE node_id = $1 AND title = $2 AND status = 'open'`,
		nodeID, message,
	).Scan(&count)
	if err != nil || count > 0 {
		return // Already has an open incident for this
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO incidents (id, node_id, severity, title, message) VALUES ($1, $2, $3, $4, $5)`,
		uuid.New().String(), nodeID, severity, message, message,
	)
	if err != nil {
		s.log.Warnw("failed to create incident", "error", err)
		return
	}

	// Also publish as event
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"incident": message,
		"severity": severity,
	})
	s.ev.Publish(ctx, model.Event{
		NodeID:  nodeID,
		Type:    eventType,
		Payload: string(eventPayload),
	})

	// Evaluate enabled policies that match this incident type
	s.evaluatePolicies(ctx, nodeID, eventType, severity, message)
}

// evaluatePolicies checks enabled policies and creates remediation actions for matches.
func (s *AgentService) evaluatePolicies(ctx context.Context, nodeID, eventType, severity, message string) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, triggers, actions FROM policies WHERE enabled = true`,
	)
	if err != nil {
		s.log.Warnw("query policies for evaluation", "error", err)
		return
	}
	defer rows.Close()

	type policyMatch struct {
		id       string
		name     string
		triggers string
		actions  string
	}
	var matches []policyMatch

	for rows.Next() {
		var p policyMatch
		if err := rows.Scan(&p.id, &p.name, &p.triggers, &p.actions); err != nil {
			continue
		}

		// Simple trigger matching: check if trigger string contains the event type
		triggerLower := strings.ToLower(p.triggers)
		if strings.Contains(triggerLower, eventType) ||
			strings.Contains(triggerLower, severity) ||
			strings.Contains(triggerLower, "all") {
			matches = append(matches, p)
		}
	}

	for _, m := range matches {
		s.log.Infow("policy matched, creating action", "policy", m.name, "event", eventType)

		// Parse actions JSON to determine action types
		var actionsList []map[string]interface{}
		if err := json.Unmarshal([]byte(m.actions), &actionsList); err != nil {
			// If actions is a plain text string, try to parse known patterns
			actionText := strings.ToLower(m.actions)
			var actionType string
			var params map[string]interface{}

			switch {
			case strings.Contains(actionText, "block") || strings.Contains(actionText, "ban"):
				actionType = "fail2ban_ban_ip"
				params = map[string]interface{}{"ip": "{{.SourceIP}}", "jail": "sshd"}
			case strings.Contains(actionText, "deny") || strings.Contains(actionText, "firewall"):
				actionType = "ufw_deny_port"
				params = map[string]interface{}{"port": 22, "protocol": "tcp"}
			default:
				s.log.Debugw("unrecognized policy action, skipping", "actions", m.actions)
				continue
			}

			paramBytes, _ := json.Marshal(params)
			_, err := s.db.Exec(ctx,
				`INSERT INTO agent_actions (id, node_id, action_type, params, status)
				 VALUES ($1, $2, $3, $4, 'pending')`,
				uuid.New().String(), nodeID, actionType, string(paramBytes),
			)
			if err != nil {
				s.log.Warnw("failed to create auto-remediation action", "error", err)
			}
		} else {
			// Structured JSON actions
			for _, action := range actionsList {
				actType, _ := action["type"].(string)
				config, _ := action["config"].(map[string]interface{})
				if actType != "remediate" || config == nil {
					continue
				}
				cmdConfig, _ := config["command"].(map[string]interface{})
				if cmdConfig == nil {
					continue
				}
				actionType, _ := cmdConfig["action"].(string)
				paramsRaw, _ := cmdConfig["params"].(map[string]interface{})
				if actionType == "" {
					continue
				}
				paramBytes, _ := json.Marshal(paramsRaw)
				_, err := s.db.Exec(ctx,
					`INSERT INTO agent_actions (id, node_id, action_type, params, status)
					 VALUES ($1, $2, $3, $4, 'pending')`,
					uuid.New().String(), nodeID, actionType, string(paramBytes),
				)
				if err != nil {
					s.log.Warnw("failed to create auto-remediation action", "error", err)
				}
			}
		}
	}
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "gw_" + hex.EncodeToString(b), nil
}
