package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	Auth  *AuthService
	Node  *NodeService
	Event *EventService
}

func New(db *pgxpool.Pool, log *zap.SugaredLogger, jwtSecret string) *Service {
	return &Service{
		Auth:  NewAuthService(db, jwtSecret),
		Node:  NewNodeService(db),
		Event: NewEventService(db, log),
	}
}
