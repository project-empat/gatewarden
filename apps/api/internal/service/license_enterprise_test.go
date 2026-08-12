//go:build enterprise

package service

import (
	"context"
	"testing"

	"github.com/project-empat/gatewarden-enterprise/enterprise/licensing"
)

// fakeManager implements licensing.Manager for exercising the enterprise
// build-mode paths of LicenseService. Runs only when built with the
// enterprise tag (GOWORK=go.work.enterprise go test -tags enterprise).
type fakeManager struct {
	license  *licensing.License
	checkRes *licensing.FeatureCheckResult
	activate *licensing.License
	err      error
}

func (f *fakeManager) ValidateLicense(ctx context.Context, licenseID string) (*licensing.License, error) {
	return f.license, f.err
}
func (f *fakeManager) CheckFeatures(ctx context.Context, req *licensing.FeatureCheckRequest) (*licensing.FeatureCheckResult, error) {
	return f.checkRes, f.err
}
func (f *fakeManager) GetLicense(ctx context.Context, tenantID string) (*licensing.License, error) {
	return f.license, f.err
}
func (f *fakeManager) ActivateLicense(ctx context.Context, tenantID, licenseKey string) (*licensing.License, error) {
	return f.activate, f.err
}
func (f *fakeManager) RevokeLicense(ctx context.Context, licenseID string) error {
	return f.err
}
func (f *fakeManager) ListEntitlements(ctx context.Context, tenantID string) ([]*licensing.Entitlement, error) {
	return nil, nil
}

func TestLicenseServiceEnterprisePaths(t *testing.T) {
	ctx := context.Background()
	active := &licensing.License{
		ID: "lic_1", TenantID: defaultTenantID, Plan: "pro",
		Status: licensing.LicenseStatusActive, Seats: 50,
		Features: []string{"core_gateway", "advanced_rbac", "sso_oidc"},
	}

	svc := &LicenseService{db: nil, log: nil, manager: &fakeManager{license: active}}

	if got := svc.BuildMode(); got != "enterprise" {
		t.Fatalf("BuildMode = %q, want enterprise", got)
	}

	// Current returns the installed license.
	l, err := svc.Current(ctx)
	if err != nil {
		t.Fatalf("Current error: %v", err)
	}
	if l.Plan != "pro" {
		t.Errorf("Current plan = %q, want pro", l.Plan)
	}

	// Activate delegates to the manager.
	svc.manager = &fakeManager{activate: active}
	got, err := svc.Activate(ctx, "pro-key")
	if err != nil {
		t.Fatalf("Activate error: %v", err)
	}
	if got.Plan != "pro" {
		t.Errorf("Activate plan = %q, want pro", got.Plan)
	}

	// Features map CheckFeatures missing-set to enabled flags.
	svc.manager = &fakeManager{checkRes: &licensing.FeatureCheckResult{
		Allowed: true,
		Missing: []string{"sso_saml", "msp_multi_tenant"},
	}}
	fs, err := svc.Features(ctx)
	if err != nil {
		t.Fatalf("Features error: %v", err)
	}
	enabled := map[string]bool{}
	for _, f := range fs {
		enabled[f.Name] = f.Enabled
	}
	if !enabled["core_gateway"] || !enabled["sso_oidc"] {
		t.Error("core_gateway/sso_oidc should be enabled")
	}
	if enabled["sso_saml"] || enabled["msp_multi_tenant"] {
		t.Error("sso_saml/msp_multi_tenant should be disabled")
	}

	// RequireFeature passes for allowed features.
	if err := svc.RequireFeature(ctx, "core_gateway"); err != nil {
		t.Errorf("RequireFeature(core_gateway) error: %v", err)
	}

	// Revoke path.
	svc.manager = &fakeManager{}
	if err := svc.manager.RevokeLicense(ctx, "lic_1"); err != nil {
		t.Errorf("RevokeLicense error: %v", err)
	}
}
