package modelruntime

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"modelcraft/pkg/logfacade"
	"strings"

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

// Build constructs the RLSPolicySnapshot for the given model.
func (b *RLSSnapshotBuilder) Build(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	action modelruntime.Action,
	userCtx *rls.UserContext,
	perms *modelruntime.ResolvedModelPermissions,
) (*modelruntime.RLSPolicySnapshot, error) {
	logger := logfacade.GetLogger(ctx)

	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}

	auth := buildAuthMap(userCtx)
	snap := &modelruntime.RLSPolicySnapshot{Auth: auth}

	// Collect expressions from perms, merge USING (OR), compile CHECKs.
	switch action {
	case modelruntime.ActionSelect:
		filter, err := b.mergeUSING(ctx, perms, action, userCtx)
		if err != nil {
			return nil, err
		}
		snap.USING = filter
		logger.Debugf(ctx, "RLS snapshot: action=%s model=%s using=%v checks=0", action, modelID, filter != nil)

	case modelruntime.ActionInsert:
		checks, err := b.compileCHECKs(perms, action)
		if err != nil {
			return nil, err
		}
		snap.CHECKs = checks
		logger.Debugf(ctx, "RLS snapshot: action=%s model=%s using=false checks=%d", action, modelID, len(checks))

	case modelruntime.ActionUpdate:
		filter, err := b.mergeUSING(ctx, perms, action, userCtx)
		if err != nil {
			return nil, err
		}
		checks, err := b.compileCHECKs(perms, action)
		if err != nil {
			return nil, err
		}
		snap.USING = filter
		snap.CHECKs = checks
		logger.Debugf(ctx, "RLS snapshot: action=%s model=%s using=%v checks=%d", action, modelID, filter != nil, len(checks))

	case modelruntime.ActionDelete:
		filter, err := b.mergeUSING(ctx, perms, action, userCtx)
		if err != nil {
			return nil, err
		}
		snap.USING = filter
		logger.Debugf(ctx, "RLS snapshot: action=%s model=%s using=%v checks=0", action, modelID, filter != nil)
	}

	return snap, nil
}

// mergeUSING collects all UsingExpr from perms.Policies matching the action,
// compiles each to SQL, and OR-merges the results.
func (b *RLSSnapshotBuilder) mergeUSING(
	ctx context.Context,
	perms *modelruntime.ResolvedModelPermissions,
	action modelruntime.Action,
	userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, error) {
	if perms == nil {
		return nil, nil //nolint:nilnil
	}

	var orClauses []string
	var allParams []interface{}

	for _, pol := range perms.Policies {
		if pol.Action != action || pol.UsingExpr == "" {
			continue
		}
		compiled, err := b.policySvc.CompileUsingExpr(ctx, pol.UsingExpr, userCtx)
		if err != nil {
			return nil, err
		}
		orClauses = append(orClauses, "("+compiled.SQL+")")
		allParams = append(allParams, compiled.Params...)
	}

	if len(orClauses) == 0 {
		return nil, nil //nolint:nilnil
	}

	return &modelruntime.RawSQLFilter{
		SQL:    strings.Join(orClauses, " OR "),
		Params: allParams,
	}, nil
}

// compileCHECKs compiles all WithCheckExpr from perms.Policies matching the action.
// CHECK expressions use OR logic — any single one passing is sufficient.
func (b *RLSSnapshotBuilder) compileCHECKs(
	perms *modelruntime.ResolvedModelPermissions,
	action modelruntime.Action,
) ([]*modelruntime.CheckProgram, error) {
	if perms == nil {
		return nil, nil
	}

	var checks []*modelruntime.CheckProgram
	for _, pol := range perms.Policies {
		if pol.Action != action || pol.WithCheckExpr == "" {
			continue
		}
		ast, issues := b.celEnv.Compile(pol.WithCheckExpr)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("compile CHECK expression: %w", issues.Err())
		}
		program, err := b.celEnv.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("program CHECK expression: %w", err)
		}
		checks = append(checks, modelruntime.NewCheckProgram(program))
	}

	return checks, nil
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
