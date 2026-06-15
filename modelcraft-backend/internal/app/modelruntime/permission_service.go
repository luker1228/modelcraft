package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"strings"
)

// PolicyPermissionResolver converts V2 RLS policies into a lightweight action summary.
type PolicyPermissionResolver struct {
	policyRepo rls.PolicyRepositoryV2
}

func NewPolicyPermissionResolver(policyRepo rls.PolicyRepositoryV2) *PolicyPermissionResolver {
	return &PolicyPermissionResolver{policyRepo: policyRepo}
}

// ResolveFromV2Policy loads V2 policies for a model and summarizes action availability.
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
	roleSet[""] = struct{}{}

	perms := &modelruntime.ResolvedModelPermissions{}
	for _, p := range policies {
		if _, ok := roleSet[p.Role]; !ok {
			continue
		}
		switch p.Action {
		case rls.ActionRead:
			perms.Select = mergeActionPermission(perms.Select, p)
		case rls.ActionCreate:
			perms.Insert = mergeActionPermission(perms.Insert, p)
		case rls.ActionUpdate:
			perms.Update = mergeActionPermission(perms.Update, p)
		case rls.ActionDelete:
			perms.Delete = mergeActionPermission(perms.Delete, p)
		}
	}
	return perms, nil
}

func mergeActionPermission(curr modelruntime.ActionPermission, policy *rls.Policy) modelruntime.ActionPermission {
	curr.Allowed = true
	if looksLikeSelfScoped(policy) {
		curr.IsSelf = true
	}
	return curr
}

func looksLikeSelfScoped(policy *rls.Policy) bool {
	if policy == nil {
		return false
	}
	return strings.Contains(string(policy.UsingExpr), "$endUserId") ||
		strings.Contains(string(policy.WithCheckExpr), "$endUserId")
}

func adminWildcardPermissions() *modelruntime.ResolvedModelPermissions {
	all := modelruntime.ActionPermission{Allowed: true, IsSelf: false}
	return &modelruntime.ResolvedModelPermissions{
		Select: all,
		Insert: all,
		Update: all,
		Delete: all,
	}
}
