package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/project-empat/gatewarden-enterprise/enterprise/licensing"
)

// ErrEnterpriseOnly is returned when a premium operation is attempted in an
// OSS build (or when the enterprise module is not linked).
var ErrEnterpriseOnly = errors.New("enterprise feature: not available in OSS build")

// defaultTenantID is the single-tenant identifier used by the self-hosted
// MVP. Multi-tenancy (MSP) will introduce real tenant scoping.
const defaultTenantID = "default"

// knownFeatures is the full feature catalog. Free/OSS builds enable the
// core set; the rest are unlocked by a license (see gatewarden-enterprise's
// plan definitions).
var knownFeatures = []string{
	"core_gateway", "basic_rbac", "basic_audit",
	"advanced_rbac", "sso_oidc", "sso_saml", "audit_export", "audit_stream",
	"policy_engine", "automation", "msp_multi_tenant", "msp_isolation",
}

// freeFeatures are enabled in the OSS build without any license.
var freeFeatures = map[string]bool{
	"core_gateway": true,
	"basic_rbac":   true,
	"basic_audit":  true,
}

// LicenseService exposes license state and activation. In OSS builds the
// underlying manager is the enterprise module's no-op; in enterprise builds
// it is the real licensing manager, backed by the same DB store below.
type LicenseService struct {
	db      *pgxpool.Pool
	log     *zap.SugaredLogger
	manager licensing.Manager
}

func NewLicenseService(db *pgxpool.Pool, log *zap.SugaredLogger) *LicenseService {
	return &LicenseService{
		db:      db,
		log:     log,
		manager: licensing.NewManager(&licenseStore{db: db}, &licenseEntitlements{db: db}),
	}
}

// BuildMode reports "oss" or "enterprise" based on the Go build tags.
func (s *LicenseService) BuildMode() string { return buildMode }

// Current returns the active license, falling back to a synthetic free
// license when none is installed — the product must keep working offline for
// core free features.
func (s *LicenseService) Current(ctx context.Context) (*licensing.License, error) {
	l, err := s.manager.GetLicense(ctx, defaultTenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || buildMode == "oss" {
			return freeLicense(), nil
		}
		return nil, fmt.Errorf("get license: %w", err)
	}
	return l, nil
}

// Activate validates and stores a license key. Returns ErrEnterpriseOnly in
// OSS builds.
func (s *LicenseService) Activate(ctx context.Context, key string) (*licensing.License, error) {
	if key == "" {
		return nil, errors.New("license key required")
	}
	l, err := s.manager.ActivateLicense(ctx, defaultTenantID, key)
	if err != nil {
		if buildMode == "oss" {
			return nil, ErrEnterpriseOnly
		}
		return nil, fmt.Errorf("activate license: %w", err)
	}
	return l, nil
}

// FeatureStatus describes a single feature's availability.
type FeatureStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Features returns the full feature catalog with enabled state for the
// current build/license.
func (s *LicenseService) Features(ctx context.Context) ([]FeatureStatus, error) {
	if buildMode == "oss" {
		out := make([]FeatureStatus, 0, len(knownFeatures))
		for _, f := range knownFeatures {
			out = append(out, FeatureStatus{Name: f, Enabled: freeFeatures[f]})
		}
		return out, nil
	}

	res, err := s.manager.CheckFeatures(ctx, &licensing.FeatureCheckRequest{
		TenantID: defaultTenantID,
		Features: knownFeatures,
	})
	if err != nil {
		return nil, fmt.Errorf("check features: %w", err)
	}

	missing := make(map[string]bool, len(res.Missing))
	for _, f := range res.Missing {
		missing[f] = true
	}
	out := make([]FeatureStatus, 0, len(knownFeatures))
	for _, f := range knownFeatures {
		out = append(out, FeatureStatus{Name: f, Enabled: !missing[f]})
	}
	return out, nil
}

// RequireFeature gates a premium operation behind the current license.
// Returns ErrEnterpriseOnly in OSS builds or when the license lacks the
// feature.
func (s *LicenseService) RequireFeature(ctx context.Context, feature string) error {
	if buildMode == "oss" {
		if freeFeatures[feature] {
			return nil
		}
		return ErrEnterpriseOnly
	}
	res, err := s.manager.CheckFeatures(ctx, &licensing.FeatureCheckRequest{
		TenantID: defaultTenantID,
		Features: []string{feature},
	})
	if err != nil {
		return fmt.Errorf("check feature %q: %w", feature, err)
	}
	if !res.Allowed {
		return fmt.Errorf("%w: %s requires an enterprise license", ErrEnterpriseOnly, feature)
	}
	return nil
}

// --- DB-backed store implementing licensing.Store ---

type licenseStore struct {
	db *pgxpool.Pool
}

func (s *licenseStore) GetLicense(ctx context.Context, tenantID string) (*licensing.License, error) {
	var l licensing.License
	var featuresJSON, metaJSON string
	err := s.db.QueryRow(ctx,
		`SELECT id, tenant_id, plan, status, seats, features, issued_at, expires_at,
		        grace_period_end, metadata
		 FROM licenses WHERE tenant_id = $1`, tenantID,
	).Scan(&l.ID, &l.TenantID, &l.Plan, &l.Status, &l.Seats, &featuresJSON,
		&l.IssuedAt, &l.ExpiresAt, &l.GracePeriodEnd, &metaJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(featuresJSON), &l.Features); err != nil {
		return nil, fmt.Errorf("decode features: %w", err)
	}
	if err := json.Unmarshal([]byte(metaJSON), &l.Metadata); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}
	return &l, nil
}

func (s *licenseStore) SaveLicense(ctx context.Context, l *licensing.License) error {
	featuresJSON, err := json.Marshal(l.Features)
	if err != nil {
		return fmt.Errorf("encode features: %w", err)
	}
	metaJSON, err := json.Marshal(l.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO licenses (id, tenant_id, plan, status, seats, features, issued_at, expires_at, grace_period_end, metadata, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		 ON CONFLICT (tenant_id) DO UPDATE SET
		   plan = EXCLUDED.plan, status = EXCLUDED.status, seats = EXCLUDED.seats,
		   features = EXCLUDED.features, issued_at = EXCLUDED.issued_at,
		   expires_at = EXCLUDED.expires_at, grace_period_end = EXCLUDED.grace_period_end,
		   metadata = EXCLUDED.metadata, updated_at = NOW()`,
		l.ID, l.TenantID, l.Plan, l.Status, l.Seats, string(featuresJSON),
		l.IssuedAt, l.ExpiresAt, l.GracePeriodEnd, string(metaJSON),
	)
	if err != nil {
		return fmt.Errorf("save license: %w", err)
	}
	return nil
}

func (s *licenseStore) UpdateLicense(ctx context.Context, l *licensing.License) error {
	featuresJSON, err := json.Marshal(l.Features)
	if err != nil {
		return fmt.Errorf("encode features: %w", err)
	}
	metaJSON, err := json.Marshal(l.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	_, err = s.db.Exec(ctx,
		`UPDATE licenses SET plan = $2, status = $3, seats = $4, features = $5,
		   issued_at = $6, expires_at = $7, grace_period_end = $8, metadata = $9, updated_at = NOW()
		 WHERE tenant_id = $1`,
		l.TenantID, l.Plan, l.Status, l.Seats, string(featuresJSON),
		l.IssuedAt, l.ExpiresAt, l.GracePeriodEnd, string(metaJSON),
	)
	if err != nil {
		return fmt.Errorf("update license: %w", err)
	}
	return nil
}

// --- DB-backed entitlement service ---

type licenseEntitlements struct {
	db *pgxpool.Pool
}

func (s *licenseEntitlements) GetEntitlements(ctx context.Context, tenantID string) ([]*licensing.Entitlement, error) {
	store := &licenseStore{db: s.db}
	l, err := store.GetLicense(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]*licensing.Entitlement, 0, len(l.Features))
	for _, f := range l.Features {
		out = append(out, &licensing.Entitlement{
			Feature: f,
			Enabled: l.Status == licensing.LicenseStatusActive,
		})
	}
	return out, nil
}

func (s *licenseEntitlements) IncrementUsage(ctx context.Context, tenantID, feature string) error {
	return nil // quota tracking is a future enterprise capability
}

func (s *licenseEntitlements) CheckQuota(ctx context.Context, tenantID, feature string) (bool, error) {
	return true, nil
}

// freeLicense is the synthetic default returned when no license is installed.
func freeLicense() *licensing.License {
	now := time.Now().UTC()
	return &licensing.License{
		ID:        "free",
		TenantID:  defaultTenantID,
		Plan:      "free",
		Status:    licensing.LicenseStatusActive,
		Features:  []string{"core_gateway", "basic_rbac", "basic_audit"},
		IssuedAt:  now,
		ExpiresAt: now.AddDate(100, 0, 0),
	}
}
