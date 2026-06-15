package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
)

// PolicyPermissionResolver converts V2 RLS policies into a lightweight action summary.
type PolicyPermissionResolver struct {
	policyRepo rls.PolicyRepositoryV2
}

func NewPolicyPermissionResolver(policyRepo rls.PolicyRepositoryV2) *PolicyPermissionResolver {
	return &PolicyPermissionResolver{policyRepo: policyRepo}
}

// ResolveFromV2Policy loads V2 policies for a model and returns matching policies as a list.
func (r *PolicyPermissionResolver) ResolveFromV2Policy(
	ctx context.Context, orgName, projectSlug, modelID string, endUserRoles []string,
	action modelruntime.Action,
) (*modelruntime.ResolvedModelPermissions, error) {
	policies, err := r.policyRepo.ListByModel(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, err
	}

	roleSet := make(map[string]struct{}, len(endUserRoles)+1)
	for _, role := range endUserRoles {
		roleSet[role] = struct{}{}
	}
	roleSet[""] = struct{}{}

	var resolved []modelruntime.ResolvedPolicy
	for _, p := range policies {
		if _, ok := roleSet[p.Role]; !ok {
			continue
		}
		mapped := mapAction(p.Action)
		if mapped != action {
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

func adminWildcardPermissions() *modelruntime.ResolvedModelPermissions {
	return &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect},
			{Action: modelruntime.ActionInsert},
			{Action: modelruntime.ActionUpdate},
			{Action: modelruntime.ActionDelete},
		},
	}
}
