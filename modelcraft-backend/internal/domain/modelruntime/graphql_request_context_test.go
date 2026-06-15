package modelruntime_test

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"testing"
)

func TestWithGraphqlRequestContext_EndUserPerms(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect, },
		},
	}
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, // clientRepo
		"org1", "proj1",
		&modelruntime.RLSContext{EndUserID: "user123", Permissions: perms},
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.RLS == nil || rctx.RLS.Permissions == nil {
		t.Fatal("expected EndUserPerms to be set")
	}
	if !rctx.RLS.Permissions.Get(modelruntime.ActionSelect).Allowed {
		t.Error("expected select allowed")
	}
}

func TestWithGraphqlRequestContext_NilPerms_TenantAdmin(t *testing.T) {
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, "org1", "proj1",
		&modelruntime.RLSContext{}, // tenant admin
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.RLS == nil || rctx.RLS.Permissions != nil {
		t.Error("tenant admin should have nil EndUserPerms")
	}
}
