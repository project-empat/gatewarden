package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/gatewarden/api/internal/model"
)

type EventService struct {
	db     *pgxpool.Pool
	log    *zap.SugaredLogger
	subs   map[string]chan model.Event
	mu     sync.RWMutex
}

func NewEventService(db *pgxpool.Pool, log *zap.SugaredLogger) *EventService {
	return &EventService{
		db:   db,
		log:  log,
		subs: make(map[string]chan model.Event),
	}
}

func (s *EventService) Subscribe() (string, <-chan model.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New().String()
	ch := make(chan model.Event, 64)
	s.subs[id] = ch
	return id, ch
}

func (s *EventService) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.subs[id]; ok {
		close(ch)
		delete(s.subs, id)
	}
}

func (s *EventService) Publish(ctx context.Context, event model.Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subs {
		select {
		case ch <- event:
		default:
			s.log.Warnw("event channel full, dropping", "type", event.Type)
		}
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO events (id, node_id, type, payload) VALUES ($1, $2, $3, $4)`,
		uuid.New().String(), event.NodeID, event.Type, event.Payload,
	)
	return err
}

func (s *EventService) ListEvents(ctx context.Context, limit int) ([]model.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, node_id, type, payload, created_at FROM events ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.NodeID, &e.Type, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
