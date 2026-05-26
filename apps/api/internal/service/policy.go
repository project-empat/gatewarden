package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatewarden/api/internal/model"
)

type PolicyService struct {
	db *pgxpool.Pool
}

func NewPolicyService(db *pgxpool.Pool) *PolicyService {
	return &PolicyService{db: db}
}

func (s *PolicyService) List(ctx context.Context) ([]model.Policy, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, description, enabled, severity, triggers, actions, created_at, updated_at
		 FROM policies ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	var policies []model.Policy
	for rows.Next() {
		var p model.Policy
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Enabled, &p.Severity, &p.Triggers, &p.Actions, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, p)
	}
	if policies == nil {
		policies = []model.Policy{}
	}
	return policies, rows.Err()
}

func (s *PolicyService) Get(ctx context.Context, id string) (*model.Policy, error) {
	var p model.Policy
	err := s.db.QueryRow(ctx,
		`SELECT id, name, description, enabled, severity, triggers, actions, created_at, updated_at
		 FROM policies WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Enabled, &p.Severity, &p.Triggers, &p.Actions, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get policy: %w", err)
	}
	return &p, nil
}

func (s *PolicyService) Create(ctx context.Context, p model.Policy) (*model.Policy, error) {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt

	_, err := s.db.Exec(ctx,
		`INSERT INTO policies (id, name, description, enabled, severity, triggers, actions, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, p.Description, p.Enabled, p.Severity, p.Triggers, p.Actions, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}

	return &p, nil
}

func (s *PolicyService) Update(ctx context.Context, id string, p model.Policy) (*model.Policy, error) {
	p.UpdatedAt = time.Now().UTC()

	_, err := s.db.Exec(ctx,
		`UPDATE policies SET name=$1, description=$2, enabled=$3, severity=$4, triggers=$5, actions=$6, updated_at=$7
		 WHERE id=$8`,
		p.Name, p.Description, p.Enabled, p.Severity, p.Triggers, p.Actions, p.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update policy: %w", err)
	}

	p.ID = id
	return &p, nil
}

func (s *PolicyService) Delete(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM policies WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	return nil
}

func (s *PolicyService) Toggle(ctx context.Context, id string) (*model.Policy, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Enabled = !p.Enabled
	p.UpdatedAt = time.Now().UTC()

	_, err = s.db.Exec(ctx,
		`UPDATE policies SET enabled=$1, updated_at=$2 WHERE id=$3`,
		p.Enabled, p.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("toggle policy: %w", err)
	}

	return p, nil
}
