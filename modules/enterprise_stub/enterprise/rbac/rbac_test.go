package rbac

import (
	"context"
	"errors"
	"testing"
)

func TestNoopProvider(t *testing.T) {
	p := NewProvider(nil)
	ctx := context.Background()

	// Reads degrade gracefully.
	roles, err := p.ListRoles(ctx)
	if err != nil {
		t.Fatalf("ListRoles error: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("ListRoles = %v, want empty", roles)
	}
	userRoles, err := p.GetUserRoles(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetUserRoles error: %v", err)
	}
	if len(userRoles) != 0 {
		t.Errorf("GetUserRoles = %v, want empty", userRoles)
	}

	// Writes fail loudly.
	if _, err := p.CreateRole(ctx, &Role{Name: "x"}); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("CreateRole error = %v, want ErrEnterpriseOnly", err)
	}
	if err := p.AssignRole(ctx, &Assignment{UserID: "u", RoleID: "r"}); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("AssignRole error = %v, want ErrEnterpriseOnly", err)
	}
	if err := p.AddPolicyRule(ctx, &PolicyRule{Action: "*", Resource: "*"}); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("AddPolicyRule error = %v, want ErrEnterpriseOnly", err)
	}
	if _, err := p.Evaluate(ctx, "user-1", "nodes:read", "nodes"); !errors.Is(err, ErrEnterpriseOnly) {
		t.Errorf("Evaluate error = %v, want ErrEnterpriseOnly", err)
	}
}
