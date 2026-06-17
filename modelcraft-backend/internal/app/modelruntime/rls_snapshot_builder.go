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

// Build constructs a comprehensive RLSPolicySnapshot for the given model.
// It merges USING filters from all read/write actions (select, update, delete)
// and compiles CHECK programs from insert/update actions.
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

	// Merge USING from select, update, delete policies (OR logic)
	filter, err := b.mergeUSING(ctx, perms, []modelruntime.Action{
		modelruntime.ActionSelect,
		modelruntime.ActionUpdate,
		modelruntime.ActionDelete,
	}, userCtx)
	if err != nil {
		return nil, err
	}
	snap.USING = filter

	// Compile CHECKs from insert, update policies
	checks, err := b.compileCHECKs(perms, []modelruntime.Action{
		modelruntime.ActionInsert,
		modelruntime.ActionUpdate,
	})
	if err != nil {
		return nil, err
	}
	snap.CHECKs = checks

	logger.Debugf(ctx, "RLS snapshot: model=%s using=%v checks=%d", modelID, filter != nil, len(checks))

	return snap, nil
}

// mergeUSING collects all UsingExpr from perms.Policies matching any of the given actions,
// compiles each to SQL, and OR-merges the results.
func (b *RLSSnapshotBuilder) mergeUSING(
	ctx context.Context,
	perms *modelruntime.ResolvedModelPermissions,
	actions []modelruntime.Action,
	userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, error) {
	if perms == nil {
		return nil, nil //nolint:nilnil
	}

	actionSet := make(map[modelruntime.Action]struct{}, len(actions))
	for _, a := range actions {
		actionSet[a] = struct{}{}
	}

	var orClauses []string
	var allParams []interface{}

	for _, pol := range perms.Policies {
		if _, ok := actionSet[pol.Action]; !ok || pol.UsingExpr == "" {
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
		// No USING policy matched — default deny: inject 1=0 to block all rows.
		return &modelruntime.RawSQLFilter{SQL: "1=0", Params: nil}, nil
	}

	return &modelruntime.RawSQLFilter{
		SQL:    strings.Join(orClauses, " OR "),
		Params: allParams,
	}, nil
}

// compileCHECKs compiles all WithCheckExpr from perms.Policies matching any of the given actions.
// CHECK expressions use OR logic — any single one passing is sufficient.
func (b *RLSSnapshotBuilder) compileCHECKs(
	perms *modelruntime.ResolvedModelPermissions,
	actions []modelruntime.Action,
) ([]*modelruntime.CheckProgram, error) {
	if perms == nil {
		return nil, nil
	}

	actionSet := make(map[modelruntime.Action]struct{}, len(actions))
	for _, a := range actions {
		actionSet[a] = struct{}{}
	}

	var checks []*modelruntime.CheckProgram
	for _, pol := range perms.Policies {
		if _, ok := actionSet[pol.Action]; !ok || pol.WithCheckExpr == "" {
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
		"userid":   userCtx.UserIDValue(),
		"username": userCtx.UserName,
		"roles":    userCtx.Roles,
	}
}
