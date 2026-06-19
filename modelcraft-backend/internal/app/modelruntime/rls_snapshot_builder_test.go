package modelruntime_test

import (
	"context"
	"errors"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"strings"
	"testing"

	appruntime "modelcraft/internal/app/modelruntime"
)

// stubPolicyResolver implements appruntime.PolicyResolver.
type stubPolicyResolver struct {
	compileUsingExpr func(ctx context.Context, usingExpr string, userCtx *rls.UserContext) (*rls.CompiledPolicy, error)
}

func (s *stubPolicyResolver) ResolveUsing(
	_ context.Context, _, _, _ string, _ rls.Action, _ *rls.UserContext,
) (string, []interface{}, error) {
	return "", nil, nil
}

func (s *stubPolicyResolver) GetCheckExpr(
	_ context.Context, _, _, _ string, _ rls.Action, _ *rls.UserContext,
) (string, error) {
	return "", nil
}

func (s *stubPolicyResolver) CompileUsingExpr(
	ctx context.Context, usingExpr string, userCtx *rls.UserContext,
) (*rls.CompiledPolicy, error) {
	if s.compileUsingExpr != nil {
		return s.compileUsingExpr(ctx, usingExpr, userCtx)
	}
	return &rls.CompiledPolicy{SQL: usingExpr}, nil
}

func newTestBuilder(svc *stubPolicyResolver) *appruntime.RLSSnapshotBuilder {
	return appruntime.NewRLSSnapshotBuilder(svc)
}

func perms(policies ...modelruntime.ResolvedPolicy) *modelruntime.ResolvedModelPermissions {
	return &modelruntime.ResolvedModelPermissions{Policies: policies}
}

// ─── nil userCtx ──────────────────────────────────────────────────────────────

func TestBuild_NilUserCtx_UsesDefaults(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: "status = 'open'"}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		nil, // userCtx
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionSelect,
			UsingExpr: `{"status":{"_eq":"open"}}`,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Auth == nil {
		t.Fatal("expected Auth map")
	}
	if snap.Auth["userid"] != "" || snap.Auth["username"] != "" {
		t.Errorf("expected empty auth defaults, got userid=%q username=%q", snap.Auth["userid"], snap.Auth["username"])
	}
}

// ─── USING (read/write actions) ──────────────────────────────────────────────

func TestBuild_Select_SingleUSING(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: "status = 'open'"}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionSelect,
			UsingExpr: `{"status":{"_eq":"open"}}`,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.SelectFilter == nil {
		t.Fatal("expected SelectFilter")
	}
	if snap.SelectFilter.SQL != "(status = 'open')" {
		t.Errorf("expected SQL='(status = \\'open\\')', got %q", snap.SelectFilter.SQL)
	}
	if snap.NoSelectPolicy {
		t.Error("expected NoSelectPolicy=false")
	}
}

func TestBuild_Select_MultipleUSING_ORMerged(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: expr, Params: []interface{}{expr}}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(
			modelruntime.ResolvedPolicy{
				Action:    modelruntime.ActionSelect,
				UsingExpr: "filter_a",
			},
			modelruntime.ResolvedPolicy{
				Action:    modelruntime.ActionSelect,
				UsingExpr: "filter_b",
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.SelectFilter == nil {
		t.Fatal("expected SelectFilter")
	}
	if !strings.Contains(snap.SelectFilter.SQL, " OR ") {
		t.Errorf("expected OR-merged SQL, got %q", snap.SelectFilter.SQL)
	}
	if len(snap.SelectFilter.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(snap.SelectFilter.Params))
	}
}

func TestBuild_Select_NoPolicies_DenyFilter(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(), // no policies at all
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.SelectFilter == nil || snap.SelectFilter.SQL != "1=0" {
		t.Errorf("expected deny filter SQL='1=0', got %v", snap.SelectFilter)
	}
	if !snap.NoSelectPolicy {
		t.Error("expected NoSelectPolicy=true")
	}
}

// ─── CHECKs (insert/update actions) ──────────────────────────────────────────

func TestBuild_Insert_WithCHECK(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:        modelruntime.ActionInsert,
			WithCheckExpr: "input.title != ''",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.CreateChecks == nil {
		t.Fatal("expected CreateChecks")
	}
	if snap.NoCreatePolicy {
		t.Error("expected NoCreatePolicy=false")
	}
}

func TestBuild_Insert_NoPolicies_NoInsertPolicy(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(), // no policies
	)
	if err != nil {
		t.Fatal(err)
	}
	if !snap.NoCreatePolicy {
		t.Error("expected NoCreatePolicy=true")
	}
}

func TestBuild_Insert_MultipleCHECKs_AllCompiled(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(
			modelruntime.ResolvedPolicy{
				Action:        modelruntime.ActionInsert,
				WithCheckExpr: "input.title != ''",
			},
			modelruntime.ResolvedPolicy{
				Action:        modelruntime.ActionInsert,
				WithCheckExpr: "input.status == 'active'",
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.CreateChecks) != 2 {
		t.Fatalf("expected 2 CreateChecks, got %d", len(snap.CreateChecks))
	}
}

func TestBuild_Insert_InvalidCEL_Error(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	_, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:        modelruntime.ActionInsert,
			WithCheckExpr: "!!! not valid CEL !!!",
		}),
	)
	if err == nil {
		t.Fatal("expected CEL compilation error")
	}
}

// ─── Update policies (USING + CHECK) ─────────────────────────────────────────

func TestBuild_Update_BothUSINGAndCHECK(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: expr}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(
			modelruntime.ResolvedPolicy{
				Action:        modelruntime.ActionUpdate,
				UsingExpr:     "owner = $endUserId",
				WithCheckExpr: "input.title != ''",
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.UpdateFilter == nil || snap.NoUpdatePolicy {
		t.Fatal("expected UpdateFilter with policy")
	}
	if snap.UpdateChecks == nil {
		t.Fatal("expected UpdateChecks")
	}
}

func TestBuild_Update_OnlyUSING(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: expr}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionUpdate,
			UsingExpr: "owner = $endUserId",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.UpdateFilter == nil || snap.NoUpdatePolicy {
		t.Fatal("expected UpdateFilter with policy")
	}
	if snap.UpdateChecks != nil {
		t.Fatal("expected no UpdateChecks")
	}
}

func TestBuild_Update_OnlyCHECK(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:        modelruntime.ActionUpdate,
			WithCheckExpr: "input.title != ''",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !snap.NoUpdatePolicy {
		t.Fatal("expected NoUpdatePolicy=true (no usingExpr on update)")
	}
	if snap.UpdateChecks == nil {
		t.Fatal("expected UpdateChecks")
	}
}

// ─── Delete policies (USING) ─────────────────────────────────────────────────

func TestBuild_Delete_WithUSING(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: expr}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionDelete,
			UsingExpr: "owner = $endUserId",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.DeleteFilter == nil || snap.NoDeletePolicy {
		t.Fatal("expected DeleteFilter with policy")
	}
}

// ─── error propagation ──────────────────────────────────────────────────────

func TestBuild_CompileUsingExprError(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, _ string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return nil, errors.New("compile failed")
		},
	})

	_, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{UserIDStr: "u1"},
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionSelect,
			UsingExpr: "bad-expr",
		}),
	)
	if err == nil {
		t.Fatal("expected compilation error")
	}
}

// ─── userCtx propagation ─────────────────────────────────────────────────────

func TestBuild_UserCtxInAuth(t *testing.T) {
	b := newTestBuilder(&stubPolicyResolver{
		compileUsingExpr: func(_ context.Context, expr string, _ *rls.UserContext) (*rls.CompiledPolicy, error) {
			return &rls.CompiledPolicy{SQL: expr}, nil
		},
	})

	snap, err := b.Build(
		context.Background(),
		"org1", "proj1", "model-1",
		&rls.UserContext{
			UserIDStr: "user-abc",
			UserName:  "Alice",
			Roles:     []string{"admin", "member"},
		},
		perms(modelruntime.ResolvedPolicy{
			Action:    modelruntime.ActionSelect,
			UsingExpr: "filter",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Auth["userid"] != "user-abc" {
		t.Errorf("expected userid='user-abc', got %q", snap.Auth["userid"])
	}
	if snap.Auth["username"] != "Alice" {
		t.Errorf("expected username='Alice', got %q", snap.Auth["username"])
	}
}
