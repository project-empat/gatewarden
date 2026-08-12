// Package rbac mirrors the exported API of the private
// github.com/project-empat/gatewarden-enterprise/enterprise/rbac package so
// applications can import a single module path and swap between the OSS stub
// and the real implementation without code changes.
//
// The OSS implementation is a no-op: reads return empty results and writes
// return a clear "unavailable" error. Deliberately no build tag — see
// package licensing for the rationale.
package rbac

import (
	"context"
	"errors"
)

// ErrEnterpriseOnly is returned by OSS no-op operations that require the
// enterprise module.
var ErrEnterpriseOnly = errors.New("enterprise feature: advanced RBAC unavailable in OSS build")

// Effect is the outcome of a policy rule evaluation.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ParentID    string   `json:"parent_id,omitempty"`
	System      bool     `json:"system"`
}

type Permission struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type PolicyRule struct {
	ID         string         `json:"id"`
	RoleID     string         `json:"role_id"`
	Effect     Effect         `json:"effect"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	Conditions map[string]any `json:"conditions,omitempty"`
}

type Assignment struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
	Scope  string `json:"scope,omitempty"`
}

// Provider is the interface implemented by both the OSS no-op (this module)
// and the enterprise RBAC provider (gatewarden-enterprise).
type Provider interface {
	CreateRole(ctx context.Context, r *Role) (*Role, error)
	GetRole(ctx context.Context, roleID string) (*Role, error)
	UpdateRole(ctx context.Context, r *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context) ([]*Role, error)
	AssignRole(ctx context.Context, a *Assignment) error
	UnassignRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
	Evaluate(ctx context.Context, userID, action, resource string) (Effect, error)
	AddPolicyRule(ctx context.Context, rule *PolicyRule) error
	RemovePolicyRule(ctx context.Context, ruleID string) error
}

// Store persists roles, assignments and policy rules. Applications implement
// this against their own database; the enterprise provider calls it for all
// reads/writes.
type Store interface {
	CreateRole(ctx context.Context, r *Role) error
	GetRole(ctx context.Context, id string) (*Role, error)
	UpdateRole(ctx context.Context, r *Role) error
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context) ([]*Role, error)
	CreateAssignment(ctx context.Context, a *Assignment) error
	DeleteAssignment(ctx context.Context, userID, roleID string) error
	GetAssignmentsByUser(ctx context.Context, userID string) ([]*Assignment, error)
	CreatePolicyRule(ctx context.Context, r *PolicyRule) error
	DeletePolicyRule(ctx context.Context, id string) error
	GetPolicyRulesByRole(ctx context.Context, roleID string) ([]*PolicyRule, error)
}

// NewProvider returns a no-op provider for OSS builds. The real provider
// from the enterprise module has the identical signature.
func NewProvider(store Store) Provider {
	return &noopProvider{}
}

type noopProvider struct{}

func (p *noopProvider) CreateRole(ctx context.Context, r *Role) (*Role, error) {
	return nil, ErrEnterpriseOnly
}
func (p *noopProvider) GetRole(ctx context.Context, roleID string) (*Role, error) {
	return nil, ErrEnterpriseOnly
}
func (p *noopProvider) UpdateRole(ctx context.Context, r *Role) error { return ErrEnterpriseOnly }
func (p *noopProvider) DeleteRole(ctx context.Context, roleID string) error {
	return ErrEnterpriseOnly
}
func (p *noopProvider) ListRoles(ctx context.Context) ([]*Role, error) { return nil, nil }
func (p *noopProvider) AssignRole(ctx context.Context, a *Assignment) error {
	return ErrEnterpriseOnly
}
func (p *noopProvider) UnassignRole(ctx context.Context, userID, roleID string) error {
	return ErrEnterpriseOnly
}
func (p *noopProvider) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	return nil, nil
}
func (p *noopProvider) Evaluate(ctx context.Context, userID, action, resource string) (Effect, error) {
	return EffectDeny, ErrEnterpriseOnly
}
func (p *noopProvider) AddPolicyRule(ctx context.Context, rule *PolicyRule) error {
	return ErrEnterpriseOnly
}
func (p *noopProvider) RemovePolicyRule(ctx context.Context, ruleID string) error {
	return ErrEnterpriseOnly
}
