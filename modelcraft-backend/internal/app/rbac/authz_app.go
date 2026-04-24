package rbac

import (
	"context"
	"fmt"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// EndUserAuthzService 鉴权应用服务（编排 prd/rbac/03-auth-flow.md 中的 5 步鉴权流程）
type EndUserAuthzService struct {
	rbacRepo rbacdomain.EndUserPermissionRepository
}

func NewEndUserAuthzService(rbacRepo rbacdomain.EndUserPermissionRepository) *EndUserAuthzService {
	return &EndUserAuthzService{rbacRepo: rbacRepo}
}

// GetEffectivePermissions 获取用户在指定 Project 下的有效权限集
//
// 5 步鉴权（来自 prd/rbac/03-auth-flow.md）：
//
//	Step 1: 查 end_user_user_bundles（直接授权）
//	Step 2: 查 end_user_role_users → end_user_role_bundles（显式角色，单次 JOIN）
//	Step 3: 查 end_user_roles WHERE is_implicit=true → end_user_role_bundles（隐式角色，对所有认证用户执行）
//	Step 4: 展开所有权限点（GetPermissionsByBundleIDs，动态 IN）
//	Step 5: 合并取并集（rowScope 取最宽泛范围）
//
// 返回：
//   - 有效权限集，key = "modelID:action"；空集合（无任何授权）不报错
//   - error 仅在 DB 查询失败时返回
func (s *EndUserAuthzService) GetEffectivePermissions(
	ctx context.Context,
	q GetEffectivePermissionsQuery,
) (rbacdomain.EffectivePermissionSet, error) {
	// Step 1
	directIDs, err := s.rbacRepo.GetBundleIDsByUserDirect(ctx, q.UserID, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step1 failed for user %s: %w", q.UserID, err)
	}

	// Step 2
	explicitIDs, err := s.rbacRepo.GetBundleIDsByUserExplicitRoles(ctx, q.UserID, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step2 failed for user %s: %w", q.UserID, err)
	}

	// Step 3（对所有认证用户执行，无需 userID）
	implicitIDs, err := s.rbacRepo.GetBundleIDsByImplicitRoles(ctx, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step3 failed for project %s/%s: %w", q.OrgName, q.ProjectSlug, err)
	}

	// 合并去重 bundle IDs
	allBundleIDs := deduplicateStrings(append(append(directIDs, explicitIDs...), implicitIDs...))
	if len(allBundleIDs) == 0 {
		return rbacdomain.EffectivePermissionSet{}, nil // 快速返回空集，无需查 Step 4
	}

	// Step 4: 展开权限点
	permissions, err := s.rbacRepo.GetPermissionsByBundleIDs(ctx, q.OrgName, allBundleIDs)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step4 failed: %w", err)
	}

	// Step 5: 合并取并集（rowScope 取最宽泛范围）
	eps := rbacdomain.EffectivePermissionSet{}
	return eps.Merge(permissions), nil
}

// deduplicateStrings 简单去重工具函数
func deduplicateStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
