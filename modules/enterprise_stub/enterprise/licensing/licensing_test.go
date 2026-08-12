package licensing

import (
	"context"
	"errors"
	"testing"
)

// noopManagerBehavior verifies the OSS no-op contract: reads degrade to
// empty/false, writes return a clear enterprise-only error.
func TestNoopManager(t *testing.T) {
	m := NewManager(nil, nil)
	ctx := context.Background()

	// Reads degrade gracefully.
	if _, err := m.GetLicense(ctx, "default"); !errors.Is(err, ErrNoLicense) {
		t.Errorf("GetLicense error = %v, want ErrNoLicense", err)
	}
	if _, err := m.ValidateLicense(ctx, "lic_x"); !errors.Is(err, ErrNoLicense) {
		t.Errorf("ValidateLicense error = %v, want ErrNoLicense", err)
	}

	res, err := m.CheckFeatures(ctx, &FeatureCheckRequest{
		TenantID: "default",
		Features: []string{"sso_oidc", "audit_export"},
	})
	if err != nil {
		t.Fatalf("CheckFeatures error: %v", err)
	}
	if res.Allowed {
		t.Error("CheckFeatures.Allowed = true, want false in OSS build")
	}
	if len(res.Missing) != 2 {
		t.Errorf("CheckFeatures.Missing = %v, want all features missing", res.Missing)
	}

	ents, err := m.ListEntitlements(ctx, "default")
	if err != nil {
		t.Fatalf("ListEntitlements error: %v", err)
	}
	if len(ents) != 0 {
		t.Errorf("ListEntitlements = %v, want empty", ents)
	}

	// Writes fail loudly.
	if _, err := m.ActivateLicense(ctx, "default", "enterprise-key"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("ActivateLicense error = %v, want ErrEnterpriseOnly", err)
	}
	if err := m.RevokeLicense(ctx, "lic_x"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("RevokeLicense error = %v, want ErrEnterpriseOnly", err)
	}
}
