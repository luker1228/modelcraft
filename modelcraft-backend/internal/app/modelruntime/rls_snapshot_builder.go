package modelruntime

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/http/middleware"
	"modelcraft/pkg/logfacade"

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
	logger    logfacade.Logger
	policySvc PolicyResolver
	celEnv    *cel.Env
}

// NewRLSSnapshotBuilder creates a new RLSSnapshotBuilder.
func NewRLSSnapshotBuilder(logger logfacade.Logger, policySvc PolicyResolver) *RLSSnapshotBuilder {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create CEL environment: %v", err))
	}
	return &RLSSnapshotBuilder{
		logger:    logger,
		policySvc: policySvc,
		celEnv:    env,
	}
}

// Build constructs the RLSPolicySnapshot for the given model.
// Returns nil for developer access (no RLS applied).
// Returns DenyAll=true when no matching policy exists.
func (b *RLSSnapshotBuilder) Build(
	ctx context.Context,
	orgName, projectSlug, modelID string,
) (*modelruntime.RLSPolicySnapshot, error) {
	// Developer JWT — no RLS
	identity := middleware.GetEndUserIdentity(ctx)
	if identity == nil || identity.IsDeveloper() {
		return nil, nil //nolint:nilnil
	}

	userCtx := middleware.GetUserContext(ctx)
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}

	auth := buildAuthMap(userCtx)

	// Resolve USING for each read/write action
	selectUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionRead, userCtx)
	if err != nil {
		return nil, err
	}
	updateUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
	if err != nil {
		return nil, err
	}
	deleteUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionDelete, userCtx)
	if err != nil {
		return nil, err
	}

	// Compile CHECK for insert and update
	insertCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionCreate, userCtx)
	if err != nil {
		return nil, err
	}
	updateCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
	if err != nil {
		return nil, err
	}

	// DenyAll: no USING and no CHECK for any action
	if selectUSING == nil && updateUSING == nil && deleteUSING == nil &&
		insertCHECK == nil && updateCHECK == nil {
		return &modelruntime.RLSPolicySnapshot{
			DenyAll: true,
		}, nil
	}

	return &modelruntime.RLSPolicySnapshot{
		SelectUSING: selectUSING,
		UpdateUSING: updateUSING,
		DeleteUSING: deleteUSING,
		InsertCHECK: insertCHECK,
		UpdateCHECK: updateCHECK,
		Auth:        auth,
	}, nil
}

// resolveUSING resolves the USING expression for a single action.
// Returns nil if no USING filter is needed (no policies match).
func (b *RLSSnapshotBuilder) resolveUSING(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, error) {
	sql, params, err := b.policySvc.ResolveUsing(ctx, orgName, projectSlug, modelID, action, userCtx)
	if err != nil {
		// No matching policy for this action — that's ok, just no filter
		return nil, nil //nolint:nilnil
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
		return nil, nil //nolint:nilnil // no matching policy — no CHECK needed
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
