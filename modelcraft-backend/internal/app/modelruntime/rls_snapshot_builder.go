package modelruntime

import (
	"context"
	"errors"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"

	"github.com/google/cel-go/cel"
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

// Build constructs the RLSPolicySnapshot for the given model.
// Returns DenyAll=true when no matching policy exists.
func (b *RLSSnapshotBuilder) Build(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	action modelruntime.Action,
	userCtx *rls.UserContext,
	perms *modelruntime.ResolvedModelPermissions,
) (*modelruntime.RLSPolicySnapshot, error) {
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}

	// Merge permissions by OR: if no policy grants this action, deny all.
	if perms != nil && !perms.Get(action).Allowed {
		return &modelruntime.RLSPolicySnapshot{DenyAll: true}, nil
	}

	auth := buildAuthMap(userCtx)
	snap := &modelruntime.RLSPolicySnapshot{Auth: auth}

	switch action {
	case modelruntime.ActionSelect:
		selectUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionRead, userCtx)
		if err != nil {
			return nil, err
		}
		if selectUSING == nil {
			snap.DenyAll = true
			return snap, nil
		}
		snap.SelectUSING = selectUSING

	case modelruntime.ActionInsert:
		insertCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionCreate, userCtx)
		if err != nil {
			return nil, err
		}
		if insertCHECK == nil {
			snap.DenyAll = true
			return snap, nil
		}
		snap.InsertCHECK = insertCHECK

	case modelruntime.ActionUpdate:
		updateUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
		if err != nil {
			return nil, err
		}
		updateCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
		if err != nil {
			return nil, err
		}
		if updateUSING == nil && updateCHECK == nil {
			snap.DenyAll = true
			return snap, nil
		}
		snap.UpdateUSING = updateUSING
		snap.UpdateCHECK = updateCHECK

	case modelruntime.ActionDelete:
		deleteUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionDelete, userCtx)
		if err != nil {
			return nil, err
		}
		if deleteUSING == nil {
			snap.DenyAll = true
			return snap, nil
		}
		snap.DeleteUSING = deleteUSING
	}

	return snap, nil
}

// resolveUSING resolves the USING expression for a single action.
// Returns nil if no USING filter is needed (no policies match).
func (b *RLSSnapshotBuilder) resolveUSING(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, error) {
	sql, params, err := b.policySvc.ResolveUsing(ctx, orgName, projectSlug, modelID, action, userCtx)
	if err != nil {
		// No matching policy for this action — no filter needed
		if errors.Is(err, rls.ErrNoMatchingPolicy) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	if sql == "" || sql == "1=1" {
		return nil, nil //nolint:nilnil
	}
	return &modelruntime.RawSQLFilter{SQL: sql, Params: params}, nil
}

// compileCHECK compiles the CHECK expression into a CheckProgram.
// Returns nil if no CHECK expression exists.
func (b *RLSSnapshotBuilder) compileCHECK(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (*modelruntime.CheckProgram, error) {
	expr, err := b.policySvc.GetCheckExpr(ctx, orgName, projectSlug, modelID, action, userCtx)
	if err != nil {
		return nil, nil //nolint:nilnil
	}
	if expr == "" {
		return nil, nil //nolint:nilnil
	}

	// Compile the CEL expression
	ast, issues := b.celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile CHECK expression: %w", issues.Err())
	}
	program, err := b.celEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program CHECK expression: %w", err)
	}
	return modelruntime.NewCheckProgram(program), nil
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
		"userid":   userCtx.UserID,
		"username": userCtx.UserName,
		"roles":    userCtx.Roles,
	}
}
