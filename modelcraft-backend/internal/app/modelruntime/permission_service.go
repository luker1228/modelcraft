package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"strings"

	apprls "modelcraft/internal/app/rls"
)

// PolicyPermissionResolver converts V2 RLS policies into a lightweight action summary.
type PolicyPermissionResolver struct {
	policyRepo rls.PolicyRepositoryV2
}

func NewPolicyPermissionResolver(policyRepo rls.PolicyRepositoryV2) *PolicyPermissionResolver {
	return &PolicyPermissionResolver{policyRepo: policyRepo}
}

// ResolveFromV2Policy loads V2 policies for a model and returns ALL matching policies
// (across all actions). The resolver-level CheckAction calls filter by action at check time.
func (r *PolicyPermissionResolver) ResolveFromV2Policy(
	ctx context.Context, orgName, projectSlug, modelID string, endUserRoles []string,
) (*modelruntime.ResolvedModelPermissions, error) {
	policies, err := r.policyRepo.ListByModel(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, err
	}

	roleSet := make(map[string]struct{}, len(endUserRoles)+1)
	for _, role := range endUserRoles {
		roleSet[role] = struct{}{}
	}
	roleSet["*"] = struct{}{} // wildcard: matches all end-users regardless of role

	resolved := make([]modelruntime.ResolvedPolicy, 0, len(policies))
	for _, p := range policies {
		if _, ok := roleSet[p.Role]; !ok {
			continue
		}
		mapped := mapAction(p.Action)
		if mapped == "" {
			continue
		}
		resolved = append(resolved, modelruntime.ResolvedPolicy{
			Action:        mapped,
			UsingExpr:     string(p.UsingExpr),
			WithCheckExpr: string(p.WithCheckExpr),
		})
	}
	return &modelruntime.ResolvedModelPermissions{Policies: resolved}, nil
}

func mapAction(a rls.Action) modelruntime.Action {
	switch a {
	case rls.ActionRead:
		return modelruntime.ActionSelect
	case rls.ActionCreate:
		return modelruntime.ActionInsert
	case rls.ActionUpdate:
		return modelruntime.ActionUpdate
	case rls.ActionDelete:
		return modelruntime.ActionDelete
	default:
		return ""
	}
}

// CompileUsingExpr compiles a single USING expression (CEL or JSON) to parameterised SQL.
// Implements the PolicyResolver interface.
func (r *PolicyPermissionResolver) CompileUsingExpr(
	ctx context.Context, usingExpr string, userCtx *rls.UserContext,
) (*rls.CompiledPolicy, error) {
	return apprls.NewPolicyExpressionSQLCompiler().CompileUsing(ctx, usingExpr, userCtx)
}

// ResolveUsing resolves USING policies for a specific action from V2 policies.
// It loads all V2 policies for the model, filters by the given action and user roles,
// compiles each USING expression, and OR-merges the results.
// Implements the PolicyResolver interface.
func (r *PolicyPermissionResolver) ResolveUsing(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, []interface{}, error) {
	var roles []string
	if userCtx != nil {
		roles = userCtx.Roles
	}

	perms, err := r.ResolveFromV2Policy(ctx, orgName, projectSlug, modelID, roles)
	if err != nil {
		return "", nil, err
	}

	mappedAction := mapAction(action)
	if mappedAction == "" {
		return "", nil, nil
	}

	orClauses := make([]string, 0, len(perms.Policies))
	var allParams []interface{}
	for _, pol := range perms.Policies {
		if pol.Action != mappedAction || pol.UsingExpr == "" {
			continue
		}
		compiled, err := r.CompileUsingExpr(ctx, pol.UsingExpr, userCtx)
		if err != nil {
			return "", nil, err
		}
		orClauses = append(orClauses, "("+compiled.SQL+")")
		allParams = append(allParams, compiled.Params...)
	}

	if len(orClauses) == 0 {
		return "1=1", nil, nil
	}
	return strings.Join(orClauses, " OR "), allParams, nil
}

// GetCheckExpr returns the first CHECK expression for the given action from V2 policies.
// Implements the PolicyResolver interface.
func (r *PolicyPermissionResolver) GetCheckExpr(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, error) {
	var roles []string
	if userCtx != nil {
		roles = userCtx.Roles
	}

	perms, err := r.ResolveFromV2Policy(ctx, orgName, projectSlug, modelID, roles)
	if err != nil {
		return "", err
	}

	mappedAction := mapAction(action)
	for _, pol := range perms.Policies {
		if pol.Action == mappedAction && pol.WithCheckExpr != "" {
			return pol.WithCheckExpr, nil
		}
	}
	return "", nil
}
