package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	Auth   *AuthService
	Node   *NodeService
	Event  *EventService
	Agent  *AgentService
	Policy *PolicyService
}

func New(db *pgxpool.Pool, log *zap.SugaredLogger, jwtSecret string) *Service {
	ev := NewEventService(db, log)
	return &Service{
		Auth:   NewAuthService(db, jwtSecret),
		Node:   NewNodeService(db),
		Event:  ev,
		Agent:  NewAgentService(db, log, ev),
		Policy: NewPolicyService(db),
	}
}
