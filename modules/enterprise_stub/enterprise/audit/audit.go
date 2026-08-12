// Package audit mirrors the exported API of the private
// github.com/project-empat/gatewarden-enterprise/enterprise/audit package so
// applications can import a single module path and swap between the OSS stub
// and the real implementation without code changes.
//
// The OSS implementation is a no-op: reads return empty results and writes
// return a clear "unavailable" error. Deliberately no build tag — see
// package licensing for the rationale.
package audit

import (
	"context"
	"errors"
	"time"
)

// ErrEnterpriseOnly is returned by OSS no-op operations that require the
// enterprise module.
var ErrEnterpriseOnly = errors.New("enterprise feature: advanced audit unavailable in OSS build")

// EventSeverity represents the importance level of an audit event.
type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

type Event struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	UserID     string         `json:"user_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id,omitempty"`
	Severity   EventSeverity  `json:"severity"`
	SourceIP   string         `json:"source_ip,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Changes    []Change       `json:"changes,omitempty"`
	Outcome    string         `json:"outcome"`
	Error      string         `json:"error,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

type Change struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

type Query struct {
	TenantID  string        `json:"tenant_id,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	Action    string        `json:"action,omitempty"`
	Resource  string        `json:"resource,omitempty"`
	Severity  EventSeverity `json:"severity,omitempty"`
	StartTime time.Time     `json:"start_time,omitempty"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
}

type RetentionPolicy struct {
	MaxAge    time.Duration `json:"max_age"`
	MaxSizeGB int           `json:"max_size_gb"`
	ArchiveS3 string        `json:"archive_s3,omitempty"`
}

// Provider is the interface implemented by both the OSS no-op (this module)
// and the enterprise audit provider (gatewarden-enterprise).
type Provider interface {
	Record(ctx context.Context, event *Event) error
	Query(ctx context.Context, q *Query) ([]*Event, error)
	GetByID(ctx context.Context, eventID string) (*Event, error)
	Export(ctx context.Context, q *Query, format string) ([]byte, error)
	SetRetention(ctx context.Context, policy *RetentionPolicy) error
	GetRetention(ctx context.Context) (*RetentionPolicy, error)
	Stream(ctx context.Context, tenantID string) (<-chan *Event, error)
}

// Store persists audit events. Applications implement this against their
// own database; the enterprise provider calls it for all reads/writes.
type Store interface {
	Insert(ctx context.Context, e *Event) error
	Search(ctx context.Context, q *Query) ([]*Event, error)
	Get(ctx context.Context, id string) (*Event, error)
}

// NewProvider returns a no-op provider for OSS builds. The real provider
// from the enterprise module has the identical signature.
func NewProvider(store Store) Provider {
	return &noopProvider{}
}

type noopProvider struct{}

func (p *noopProvider) Record(ctx context.Context, event *Event) error { return ErrEnterpriseOnly }
func (p *noopProvider) Query(ctx context.Context, q *Query) ([]*Event, error) {
	return nil, nil
}
func (p *noopProvider) GetByID(ctx context.Context, eventID string) (*Event, error) {
	return nil, ErrEnterpriseOnly
}
func (p *noopProvider) Export(ctx context.Context, q *Query, format string) ([]byte, error) {
	return nil, ErrEnterpriseOnly
}
func (p *noopProvider) SetRetention(ctx context.Context, policy *RetentionPolicy) error {
	return ErrEnterpriseOnly
}
func (p *noopProvider) GetRetention(ctx context.Context) (*RetentionPolicy, error) {
	return nil, nil
}
func (p *noopProvider) Stream(ctx context.Context, tenantID string) (<-chan *Event, error) {
	return nil, ErrEnterpriseOnly
}
