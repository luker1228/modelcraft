package modelruntime

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"

	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"modelcraft/pkg/logfacade"
)

// PolicyResolver provides access to RLS policy expressions.
// Implemented by app/rls.PolicyMatchingService.
type PolicyResolver interface {
	ResolveUsing(
		ctx context.Context, orgName, projectSlug, modelID string,
		action rls.Action, userCtx *rls.UserContext,
	) (string, []interface{}, error)
	GetCheckExpr(
		ctx context.Context, orgName, projectSlug, modelID string,
		action rls.Action, userCtx *rls.UserContext,
	) (string, error)
	// CompileUsingExpr compiles a single USING expression (JSON or CEL) to parameterised SQL.
	CompileUsingExpr(ctx context.Context, usingExpr string, userCtx *rls.UserContext) (*rls.CompiledPolicy, error)
}

// RLSSnapshotBuilder builds RLSPolicySnapshot at request entry.
type RLSSnapshotBuilder struct {
	policySvc PolicyResolver
	celEnv    *cel.Env
}

// NewRLSSnapshotBuilder creates a new RLSSnapshotBuilder.
func NewRLSSnapshotBuilder(policySvc PolicyResolver) *RLSSnapshotBuilder {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create CEL environment: %v", err))
	}
	return &RLSSnapshotBuilder{
		policySvc: policySvc,
		celEnv:    env,
	}
}

// Build constructs a comprehensive RLSPolicySnapshot for the given model.
// It builds per-action USING filters (select/update/delete) independently
// and compiles CHECK programs separately for create and update actions.
func (b *RLSSnapshotBuilder) Build(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	userCtx *rls.UserContext,
	perms *modelruntime.ResolvedModelPermissions,
) (*modelruntime.RLSPolicySnapshot, error) {
	logger := logfacade.GetLogger(ctx)

	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}

	auth := buildAuthMap(userCtx)
	snap := &modelruntime.RLSPolicySnapshot{Auth: auth}

	// Build per-action USING filters independently.
	for _, cfg := range []struct {
		action   modelruntime.Action
		dest     **modelruntime.RawSQLFilter
		noPolicy *bool
	}{
		{modelruntime.ActionSelect, &snap.SelectFilter, &snap.NoSelectPolicy},
		{modelruntime.ActionUpdate, &snap.UpdateFilter, &snap.NoUpdatePolicy},
		{modelruntime.ActionDelete, &snap.DeleteFilter, &snap.NoDeletePolicy},
	} {
		f, hasPolicy, err := b.mergeUSING(ctx, perms, cfg.action, userCtx)
		if err != nil {
			return nil, err
		}
		*cfg.dest = f
		*cfg.noPolicy = !hasPolicy
	}

	// Compile CREATE checks.
	createChecks, hasCreatePolicy, err := b.compileCHECKs(perms, modelruntime.ActionInsert)
	if err != nil {
		return nil, err
	}
	snap.CreateChecks = createChecks
	snap.NoCreatePolicy = !hasCreatePolicy

	// Compile UPDATE checks.
	updateChecks, _, err := b.compileCHECKs(perms, modelruntime.ActionUpdate)
	if err != nil {
		return nil, err
	}
	snap.UpdateChecks = updateChecks

	logger.Infof(ctx, "RLS snapshot built: model=%s endUser=%s select=%v update=%v delete=%v create=%v",
		modelID, userCtx.UserIDValue(),
		!snap.NoSelectPolicy, !snap.NoUpdatePolicy, !snap.NoDeletePolicy, !snap.NoCreatePolicy)

	return snap, nil
}

// mergeUSING collects all UsingExpr from perms.Policies matching the given action,
// compiles each to SQL, and OR-merges the results.
// Returns (filter, hasPolicy, error).
// hasPolicy=false means no policy was configured for this action at all (default deny).
func (b *RLSSnapshotBuilder) mergeUSING(
	ctx context.Context,
	perms *modelruntime.ResolvedModelPermissions,
	action modelruntime.Action,
	userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, bool, error) {
	if perms == nil {
		return &modelruntime.RawSQLFilter{SQL: "1=0"}, false, nil
	}

	var orClauses []string
	var allParams []interface{}

	for _, pol := range perms.Policies {
		if pol.Action != action || pol.UsingExpr == "" {
			continue
		}
		compiled, err := b.policySvc.CompileUsingExpr(ctx, pol.UsingExpr, userCtx)
		if err != nil {
			return nil, false, err
		}
		orClauses = append(orClauses, "("+compiled.SQL+")")
		allParams = append(allParams, compiled.Params...)
	}

	if len(orClauses) == 0 {
		return &modelruntime.RawSQLFilter{SQL: "1=0"}, false, nil
	}

	return &modelruntime.RawSQLFilter{
		SQL:    strings.Join(orClauses, " OR "),
		Params: allParams,
	}, true, nil
}

// compileCHECKs compiles all WithCheckExpr from perms.Policies matching the given action.
// CHECK expressions use OR logic — any single one passing is sufficient.
// Returns (checks, hasPolicy, error) where hasPolicy=false means no policy configured for this action.
func (b *RLSSnapshotBuilder) compileCHECKs(
	perms *modelruntime.ResolvedModelPermissions,
	action modelruntime.Action,
) ([]*modelruntime.CheckProgram, bool, error) {
	if perms == nil {
		return nil, false, nil
	}

	var checks []*modelruntime.CheckProgram
	for _, pol := range perms.Policies {
		if pol.Action != action || pol.WithCheckExpr == "" {
			continue
		}
		ast, issues := b.celEnv.Compile(pol.WithCheckExpr)
		if issues != nil && issues.Err() != nil {
			return nil, false, fmt.Errorf("compile CHECK expression: %w", issues.Err())
		}
		program, err := b.celEnv.Program(ast)
		if err != nil {
			return nil, false, fmt.Errorf("program CHECK expression: %w", err)
		}
		checks = append(checks, modelruntime.NewCheckProgram(program))
	}

	return checks, len(checks) > 0, nil
}

// buildAuthMap builds the auth context map for CEL evaluation.
func buildAuthMap(userCtx *rls.UserContext) map[string]any {
	if userCtx == nil {
		return map[string]any{
			"userid":   "",
			"username": "",
			"roles":    []string{},
		}
	}
	return map[string]any{
		"userid":   userCtx.UserIDValue(),
		"username": userCtx.UserName,
		"roles":    userCtx.Roles,
	}
}
