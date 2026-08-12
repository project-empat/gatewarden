// Package licensing mirrors the exported API of the private
// github.com/project-empat/gatewarden-enterprise/enterprise/licensing
// package so applications can import a single module path and swap between
// the OSS stub (this module) and the real implementation (enterprise
// workspace) without code changes.
//
// The OSS implementation is a no-op: reads return empty/false results and
// writes return an error explaining the feature is unavailable in the OSS
// build. Deliberately no build tag here — the stub module is only ever
// resolved in the OSS workspace, and leaving the files tagless keeps
// `go build -tags enterprise` on an OSS workspace compiling (the no-op
// surfaces the "not linked" state at runtime instead of failing to build).
package licensing

import (
	"context"
	"errors"
	"time"
)

// ErrEnterpriseOnly is returned by OSS no-op operations that require the
// enterprise module.
var ErrEnterpriseOnly = errors.New("enterprise feature: licensing unavailable in OSS build")

// ErrNoLicense is returned when no license exists for the tenant.
var ErrNoLicense = errors.New("licensing: no license")

// LicenseStatus represents the current state of a license.
type LicenseStatus string

const (
	LicenseStatusActive      LicenseStatus = "active"
	LicenseStatusExpired     LicenseStatus = "expired"
	LicenseStatusRevoked     LicenseStatus = "revoked"
	LicenseStatusGracePeriod LicenseStatus = "grace_period"
)

type License struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenant_id"`
	Plan           string         `json:"plan"`
	Status         LicenseStatus  `json:"status"`
	Seats          int            `json:"seats"`
	Features       []string       `json:"features"`
	IssuedAt       time.Time      `json:"issued_at"`
	ExpiresAt      time.Time      `json:"expires_at"`
	GracePeriodEnd *time.Time     `json:"grace_period_end,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type Entitlement struct {
	Feature string `json:"feature"`
	Enabled bool   `json:"enabled"`
	Limit   int    `json:"limit,omitempty"`
	Usage   int    `json:"usage,omitempty"`
}

type FeatureCheckRequest struct {
	TenantID string   `json:"tenant_id"`
	Features []string `json:"features"`
}

type FeatureCheckResult struct {
	Allowed      bool           `json:"allowed"`
	Missing      []string       `json:"missing,omitempty"`
	Entitlements []*Entitlement `json:"entitlements,omitempty"`
}

// Manager is the interface implemented by both the OSS no-op (this module)
// and the enterprise licensing manager (gatewarden-enterprise).
type Manager interface {
	ValidateLicense(ctx context.Context, licenseID string) (*License, error)
	CheckFeatures(ctx context.Context, req *FeatureCheckRequest) (*FeatureCheckResult, error)
	GetLicense(ctx context.Context, tenantID string) (*License, error)
	ActivateLicense(ctx context.Context, tenantID, licenseKey string) (*License, error)
	RevokeLicense(ctx context.Context, licenseID string) error
	ListEntitlements(ctx context.Context, tenantID string) ([]*Entitlement, error)
}

// Store persists licenses. Applications implement this against their own
// database; the enterprise manager calls it for all license reads/writes.
type Store interface {
	GetLicense(ctx context.Context, tenantID string) (*License, error)
	SaveLicense(ctx context.Context, l *License) error
	UpdateLicense(ctx context.Context, l *License) error
}

type PlanFeatures struct {
	Plan     string
	Features []string
}

// EntitlementService tracks per-tenant feature usage and quota. The OSS
// no-op ignores it; the enterprise implementation enforces quotas.
type EntitlementService interface {
	GetEntitlements(ctx context.Context, tenantID string) ([]*Entitlement, error)
	IncrementUsage(ctx context.Context, tenantID, feature string) error
	CheckQuota(ctx context.Context, tenantID, feature string) (bool, error)
}

// NewManager returns a no-op manager for OSS builds. The real manager from
// the enterprise module has the identical signature, so call sites do not
// change between build modes.
func NewManager(store Store, svc EntitlementService) Manager {
	return &noopManager{}
}

type noopManager struct{}

func (m *noopManager) ValidateLicense(ctx context.Context, licenseID string) (*License, error) {
	return nil, ErrNoLicense
}

func (m *noopManager) CheckFeatures(ctx context.Context, req *FeatureCheckRequest) (*FeatureCheckResult, error) {
	missing := make([]string, len(req.Features))
	copy(missing, req.Features)
	return &FeatureCheckResult{Allowed: false, Missing: missing}, nil
}

func (m *noopManager) GetLicense(ctx context.Context, tenantID string) (*License, error) {
	return nil, ErrNoLicense
}

func (m *noopManager) ActivateLicense(ctx context.Context, tenantID, licenseKey string) (*License, error) {
	return nil, ErrEnterpriseOnly
}

func (m *noopManager) RevokeLicense(ctx context.Context, licenseID string) error {
	return ErrEnterpriseOnly
}

func (m *noopManager) ListEntitlements(ctx context.Context, tenantID string) ([]*Entitlement, error) {
	return nil, nil
}
