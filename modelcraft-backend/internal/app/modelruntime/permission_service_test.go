package modelruntime_test

import (
	"context"
	appruntimeimport "modelcraft/internal/app/modelruntime"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"testing"
)

type stubPolicyRepo struct {
	policies []*rls.Policy
}

func (s *stubPolicyRepo) ListByModel(
	_ context.Context, _, _, _ string,
) ([]*rls.Policy, error) {
	return s.policies, nil
}

func (s *stubPolicyRepo) Upsert(_ context.Context, _, _ string, _ *rls.Policy) error {
	return nil
}

func (s *stubPolicyRepo) Delete(_ context.Context, _, _ string, _ int64) error {
	return nil
}

func (s *stubPolicyRepo) DeleteByModel(_ context.Context, _, _, _ string) error {
	return nil
}

func TestPolicyPermissionResolver_ResolveFromV2Policy_NoPermissions(t *testing.T) {
	resolver := appruntimeimport.NewPolicyPermissionResolver(&stubPolicyRepo{})
	perms, err := resolver.ResolveFromV2Policy(context.Background(), "org1", "proj1", "model-id", []string{"viewer"}, modelruntime.ActionSelect)
	if err != nil {
		t.Fatal(err)
	}
	if perms == nil {
		t.Fatal("expected non-nil perms")
	}
	if !perms.IsEmpty() {
		t.Fatal("expected IsEmpty when no v2 policy matches")
	}
}

func TestPolicyPermissionResolver_ResolveFromV2Policy_SelectAll(t *testing.T) {
	resolver := appruntimeimport.NewPolicyPermissionResolver(&stubPolicyRepo{
		policies: []*rls.Policy{
			{ModelID: "model-id", Action: rls.ActionRead, Role: "viewer", UsingExpr: rls.JsonExpr(`{"status":{"_eq":"open"}}`)},
		},
	})

	perms, err := resolver.ResolveFromV2Policy(context.Background(), "org1", "proj1", "model-id", []string{"viewer"}, modelruntime.ActionSelect)
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Get(modelruntime.ActionSelect).Allowed {
		t.Fatal("expected select allowed")
	}
	if perms.Get(modelruntime.ActionInsert).Allowed ||
		perms.Get(modelruntime.ActionUpdate).Allowed ||
		perms.Get(modelruntime.ActionDelete).Allowed {
		t.Fatal("expected non-select actions denied")
	}
}

func TestPolicyPermissionResolver_ResolveFromV2Policy_SelfScoped(t *testing.T) {
	resolver := appruntimeimport.NewPolicyPermissionResolver(&stubPolicyRepo{
		policies: []*rls.Policy{
			{ModelID: "model-id", Action: rls.ActionRead, Role: "member", UsingExpr: rls.JsonExpr(`{"owner":{"_eq":"$endUserId"}}`)},
			{ModelID: "model-id", Action: rls.ActionCreate, Role: "member", WithCheckExpr: rls.JsonExpr(`{"owner":{"_eq":"$endUserId"}}`)},
		},
	})

	// Resolve for select — only read policy matches
	perms, err := resolver.ResolveFromV2Policy(context.Background(), "org1", "proj1", "model-id", []string{"member"}, modelruntime.ActionSelect)
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Get(modelruntime.ActionSelect).Allowed {
		t.Fatal("expected select allowed")
	}

	// Resolve for insert — only create policy matches
	perms, err = resolver.ResolveFromV2Policy(context.Background(), "org1", "proj1", "model-id", []string{"member"}, modelruntime.ActionInsert)
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Get(modelruntime.ActionInsert).Allowed {
		t.Fatal("expected insert allowed")
	}
}

func TestPolicyPermissionResolver_ResolveFromV2Policy_DefaultRole(t *testing.T) {
	resolver := appruntimeimport.NewPolicyPermissionResolver(&stubPolicyRepo{
		policies: []*rls.Policy{
			{ModelID: "model-id", Action: rls.ActionDelete, Role: "", UsingExpr: rls.JsonExpr(`true`)},
		},
	})

	perms, err := resolver.ResolveFromV2Policy(context.Background(), "org1", "proj1", "model-id", []string{"unknown"}, modelruntime.ActionDelete)
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Get(modelruntime.ActionDelete).Allowed {
		t.Fatal("expected default role policy to allow delete")
	}
}
