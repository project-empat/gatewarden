package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/gatewarden/api/internal/model"
)

type ActionService struct {
	db  *pgxpool.Pool
	log *zap.SugaredLogger
}

func NewActionService(db *pgxpool.Pool, log *zap.SugaredLogger) *ActionService {
	return &ActionService{db: db, log: log}
}

// Create inserts a new pending action for a node.
func (s *ActionService) Create(ctx context.Context, nodeID, actionType string, params json.RawMessage) (*model.AgentAction, error) {
	a := &model.AgentAction{
		ID:         uuid.New().String(),
		NodeID:     nodeID,
		ActionType: actionType,
		Params:     string(params),
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO agent_actions (id, node_id, action_type, params, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.ID, a.NodeID, a.ActionType, a.Params, a.Status, a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create action: %w", err)
	}

	s.log.Infow("action created", "action_id", a.ID, "node_id", nodeID, "type", actionType)
	return a, nil
}

// PollPending returns pending actions for a specific node and marks them as "delivered".
func (s *ActionService) PollPending(ctx context.Context, nodeID string) ([]model.AgentAction, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, node_id, action_type, params, status, created_at, completed_at
		 FROM agent_actions
		 WHERE node_id = $1 AND status = 'pending'
		 ORDER BY created_at ASC
		 LIMIT 10`,
		nodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("poll actions: %w", err)
	}
	defer rows.Close()

	var actions []model.AgentAction
	for rows.Next() {
		var a model.AgentAction
		if err := rows.Scan(&a.ID, &a.NodeID, &a.ActionType, &a.Params, &a.Status, &a.CreatedAt, &a.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		actions = append(actions, a)
	}

	// Mark as delivered
	if len(actions) > 0 {
		ids := make([]string, len(actions))
		for i, a := range actions {
			ids[i] = a.ID
		}
		for _, id := range ids {
			_, _ = s.db.Exec(ctx, `UPDATE agent_actions SET status = 'delivered' WHERE id = $1 AND status = 'pending'`, id)
		}
	}

	if actions == nil {
		actions = []model.AgentAction{}
	}
	return actions, rows.Err()
}

// Complete marks an action as completed (success or failure).
func (s *ActionService) Complete(ctx context.Context, actionID, status string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(ctx,
		`UPDATE agent_actions SET status = $1, completed_at = $2 WHERE id = $3`,
		status, now, actionID,
	)
	return err
}
