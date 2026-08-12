package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/gatewarden/api/internal/model"
)

// AuditService records and queries the audit trail. Basic audit logging is
// part of the free feature set; advanced querying/export/streaming are
// premium (enterprise/audit).
type AuditService struct {
	db  *pgxpool.Pool
	log *zap.SugaredLogger
}

func NewAuditService(db *pgxpool.Pool, log *zap.SugaredLogger) *AuditService {
	return &AuditService{db: db, log: log}
}

// Record appends an event to the audit trail. An empty UserID denotes a
// system-originated event. Recording failures are logged, not returned —
// auditing must never break the audited operation.
func (s *AuditService) Record(ctx context.Context, e model.AuditEvent) error {
	if e.Outcome == "" {
		e.Outcome = "success"
	}
	metaJSON, err := json.Marshal(e.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO audit_events (user_id, action, resource, resource_id, severity, source_ip, metadata, outcome, error)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.UserID, e.Action, e.Resource, e.ResourceID, e.Severity, e.SourceIP,
		string(metaJSON), e.Outcome, e.Error,
	)
	if err != nil {
		s.log.Warnw("audit record failed", "action", e.Action, "error", err)
		return err
	}
	return nil
}

// List returns audit events matching the query, newest first.
func (s *AuditService) List(ctx context.Context, q model.AuditQuery) ([]model.AuditEvent, error) {
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	where := "WHERE 1=1"
	args := []any{}
	if q.Action != "" {
		args = append(args, q.Action)
		where += fmt.Sprintf(" AND action = $%d", len(args))
	}
	if q.UserID != "" {
		args = append(args, q.UserID)
		where += fmt.Sprintf(" AND user_id = $%d", len(args))
	}
	if q.Resource != "" {
		args = append(args, q.Resource)
		where += fmt.Sprintf(" AND resource = $%d", len(args))
	}
	args = append(args, limit, q.Offset)
	where += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, action, resource, resource_id, severity, source_ip, metadata, outcome, error, timestamp
		 FROM audit_events `+where, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit: %w", err)
	}
	defer rows.Close()

	var events []model.AuditEvent
	for rows.Next() {
		var e model.AuditEvent
		var metaJSON string
		var userID *string
		if err := rows.Scan(&e.ID, &userID, &e.Action, &e.Resource, &e.ResourceID,
			&e.Severity, &e.SourceIP, &metaJSON, &e.Outcome, &e.Error, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		if userID != nil {
			e.UserID = *userID
		}
		if err := json.Unmarshal([]byte(metaJSON), &e.Metadata); err != nil {
			e.Metadata = map[string]any{}
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit rows: %w", err)
	}
	return events, nil
}
