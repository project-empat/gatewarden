package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatewarden/api/internal/model"
)

type SettingsService struct {
	db *pgxpool.Pool
}

func NewSettingsService(db *pgxpool.Pool) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) Get(ctx context.Context) (*model.Settings, error) {
	var st model.Settings
	err := s.db.QueryRow(ctx,
		`SELECT agent_auto_approve, heartbeat_interval, log_retention_days,
		        cloudflare_api_token, tailscale_api_key, tailscale_tailnet, updated_at
		 FROM settings WHERE id = 1`,
	).Scan(&st.AgentAutoApprove, &st.HeartbeatInterval, &st.LogRetentionDays,
		&st.CloudflareAPIToken, &st.TailscaleAPIKey, &st.TailscaleTailnet, &st.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return &st, nil
}

func (s *SettingsService) Update(ctx context.Context, st model.Settings) (*model.Settings, error) {
	st.UpdatedAt = time.Now().UTC()
	_, err := s.db.Exec(ctx,
		`UPDATE settings SET
			agent_auto_approve = $1,
			heartbeat_interval = $2,
			log_retention_days = $3,
			cloudflare_api_token = $4,
			tailscale_api_key = $5,
			tailscale_tailnet = $6,
			updated_at = $7
		 WHERE id = 1`,
		st.AgentAutoApprove, st.HeartbeatInterval, st.LogRetentionDays,
		st.CloudflareAPIToken, st.TailscaleAPIKey, st.TailscaleTailnet, st.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}
	return &st, nil
}
