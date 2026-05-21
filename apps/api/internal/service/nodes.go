package service

import (
	"context"
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
