package rls

import (
	"context"
	"modelcraft/internal/domain/rls"
	"testing"
)

// mockPolicyRepo is a mock PolicyRepository for testing.
type mockPolicyRepo struct {
	policies []*rls.Policy
}

func (m *mockPolicyRepo) ListByAction(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, roles []string) ([]*rls.Policy, error) {
	var result []*rls.Policy
	for _, p := range m.policies {
		if p.OrgName != orgName || p.ProjectSlug != projectSlug || p.ModelID != modelID || p.Action != action {
			continue
		}
		userCtx := &rls.UserContext{Roles: roles}
		if userCtx.HasRole(p.Role) {
			result = append(result, p)
		}
	}
	return result, nil
}

func TestMatch_MultipleRoles_OrMerge(t *testing.T) {
	compiler := NewPolicyCompiler()
	repo := &mockPolicyRepo{policies: []*rls.Policy{
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "admin_all", Action: rls.ActionRead, Role: "admin",
			UsingExpr: `true`,
		},
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "user_own", Action: rls.ActionRead, Role: "user",
			UsingExpr: `{"tenant_id": {"equals": "{{user_id}}"}}`,
		},
	}}
	svc := NewPolicyMatchingService(repo, compiler)

	ctx := context.Background()
	userCtx := &rls.UserContext{UserID: "123", Roles: []string{"admin", "user"}}

	sql, params, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// OR merge: (1=1) OR (tenant_id = ?)
	expected := "(1=1) OR (tenant_id = ?)"
	if sql != expected {
		t.Errorf("expected SQL %q, got %q", expected, sql)
	}
	if len(params) != 1 || params[0] != "123" {
		t.Errorf("expected params [123], got %v", params)
	}
}

func TestMatch_NoMatchingPolicies_DenyAll(t *testing.T) {
	compiler := NewPolicyCompiler()
	repo := &mockPolicyRepo{policies: []*rls.Policy{
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "admin_only", Action: rls.ActionRead, Role: "admin",
			UsingExpr: `true`,
		},
	}}
	svc := NewPolicyMatchingService(repo, compiler)

	ctx := context.Background()
	userCtx := &rls.UserContext{Roles: []string{"user"}}

	_, _, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, userCtx)
	if err == nil {
		t.Fatal("expected deny-all error, got nil")
	}
}

func TestMatch_EmptyRole_MatchesDefault(t *testing.T) {
	compiler := NewPolicyCompiler()
	repo := &mockPolicyRepo{policies: []*rls.Policy{
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "default_policy", Action: rls.ActionRead, Role: "",
			UsingExpr: `{"public": {"equals": true}}`,
		},
	}}
	svc := NewPolicyMatchingService(repo, compiler)

	ctx := context.Background()
	// empty roles
	userCtx := &rls.UserContext{Roles: []string{}}

	sql, params, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(public = ?)"
	if sql != expected {
		t.Errorf("expected SQL %q, got %q", expected, sql)
	}
	if len(params) != 1 || params[0] != true {
		t.Errorf("expected params [true], got %v", params)
	}
}

func TestResolveCheck_WithCheck(t *testing.T) {
	compiler := NewPolicyCompiler()
	repo := &mockPolicyRepo{policies: []*rls.Policy{
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "user_create", Action: rls.ActionCreate, Role: "user",
			WithCheckExpr: `{"owner_id": {"equals": "{{user_id}}"}}`,
		},
	}}
	svc := NewPolicyMatchingService(repo, compiler)

	ctx := context.Background()
	userCtx := &rls.UserContext{UserID: "456", Roles: []string{"user"}}

	sql, params, err := svc.ResolveCheck(ctx, "my-org", "my-proj", "model-1", rls.ActionCreate, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(owner_id = ?)"
	if sql != expected {
		t.Errorf("expected SQL %q, got %q", expected, sql)
	}
	if len(params) != 1 || params[0] != "456" {
		t.Errorf("expected params [456], got %v", params)
	}
}

func TestResolveCheck_NoCheckExpr_Deny(t *testing.T) {
	compiler := NewPolicyCompiler()
	repo := &mockPolicyRepo{policies: []*rls.Policy{
		{
			OrgName: "my-org", ProjectSlug: "my-proj", ModelID: "model-1",
			PolicyName: "read_only", Action: rls.ActionCreate, Role: "user",
			WithCheckExpr: "", // empty
		},
	}}
	svc := NewPolicyMatchingService(repo, compiler)

	ctx := context.Background()
	userCtx := &rls.UserContext{Roles: []string{"user"}}

	sql, params, err := svc.ResolveCheck(ctx, "my-org", "my-proj", "model-1", rls.ActionCreate, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No check expressions → deny
	if sql != "1=0" {
		t.Errorf("expected deny SQL '1=0', got %q", sql)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}
