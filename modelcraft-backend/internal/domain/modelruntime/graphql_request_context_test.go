package modelruntime_test

import (
	"context"
	"testing"

	"modelcraft/internal/domain/modelruntime"
)

func TestWithGraphqlRequestContext_EndUserPerms(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
	}
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, // clientRepo
		"org1", "proj1", "user123",
		perms,
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.EndUserPerms == nil {
		t.Fatal("expected EndUserPerms to be set")
	}
	if !rctx.EndUserPerms.Select.IsSelf {
		t.Error("expected Select.IsSelf = true")
	}
}

func TestWithGraphqlRequestContext_NilPerms_TenantAdmin(t *testing.T) {
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, "org1", "proj1", "",
		nil, // tenant admin
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.EndUserPerms != nil {
		t.Error("tenant admin should have nil EndUserPerms")
	}
}
