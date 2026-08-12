//go:build !enterprise

package service

import (
	"context"
	"errors"
	"testing"
)

// OSS build contract: no DB is required because every path short-circuits on
// the OSS build mode before touching the manager or the store.
func TestLicenseServiceOSSBuild(t *testing.T) {
	svc := NewLicenseService(nil, nil) // no db needed in OSS short-circuit paths
	ctx := context.Background()

	if got := svc.BuildMode(); got != "oss" {
		t.Errorf("BuildMode = %q, want oss", got)
	}

	// Current always returns a usable free license offline.
	l, err := svc.Current(ctx)
	if err != nil {
		t.Fatalf("Current error: %v", err)
	}
	if l.Plan != "free" || l.Status != "active" {
		t.Errorf("Current license = %s/%s, want free/active", l.Plan, l.Status)
	}

	// Activation is refused in OSS builds.
	if _, err := svc.Activate(ctx, "enterprise-key"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("Activate error = %v, want ErrEnterpriseOnly", err)
	}
	if _, err := svc.Activate(ctx, ""); err == nil {
		t.Error("Activate with empty key should error")
	}

	// Feature catalog: free features on, premium off.
	fs, err := svc.Features(ctx)
	if err != nil {
		t.Fatalf("Features error: %v", err)
	}
	byName := map[string]bool{}
	for _, f := range fs {
		byName[f.Name] = f.Enabled
	}
	for _, core := range []string{"core_gateway", "basic_rbac", "basic_audit"} {
		if !byName[core] {
			t.Errorf("feature %s should be enabled in OSS build", core)
		}
	}
	for _, premium := range []string{"sso_oidc", "advanced_rbac", "audit_export", "policy_engine", "msp_multi_tenant"} {
		if byName[premium] {
			t.Errorf("feature %s should be disabled in OSS build", premium)
		}
	}

	// Feature gate: free features pass, premium refused.
	if err := svc.RequireFeature(ctx, "core_gateway"); err != nil {
		t.Errorf("RequireFeature(core_gateway) error: %v", err)
	}
	if err := svc.RequireFeature(ctx, "sso_oidc"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("RequireFeature(sso_oidc) error = %v, want ErrEnterpriseOnly", err)
	}
}
