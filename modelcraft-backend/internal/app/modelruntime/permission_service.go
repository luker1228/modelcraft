package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rbac"
)

// endUserPermissionServiceImpl 实现 modelruntime.EndUserPermissionService。
// 依赖 rbac.EndUserPermissionRepository，per-request 按 model 粒度查权限。
type endUserPermissionServiceImpl struct {
	rbacRepo rbac.EndUserPermissionRepository
}

// NewEndUserPermissionService 创建 EndUserPermissionService 实例。
func NewEndUserPermissionService(rbacRepo rbac.EndUserPermissionRepository) modelruntime.EndUserPermissionService {
	return &endUserPermissionServiceImpl{rbacRepo: rbacRepo}
}

// Resolve 查询 end-user 在指定 model 上的有效权限，返回权限快照。
// endUserID 为空时（tenant admin）直接返回 nil, nil。
func (s *endUserPermissionServiceImpl) Resolve(
	ctx context.Context, orgName, projectSlug, endUserID, modelID string,
) (*modelruntime.ResolvedModelPermissions, error) {
	if endUserID == "" {
		return nil, nil //nolint:nilnil // nil ResolvedModelPermissions is the tenant-admin sentinel (skip all checks)
	}

	permissions, err := s.rbacRepo.FindPermissionsByEndUserAndModel(ctx, orgName, projectSlug, endUserID, modelID)
	if err != nil {
		return nil, err
	}

	// 合并所有权限点（取各 action 最宽 rowScope）
	eps := rbac.EffectivePermissionSet{}.Merge(permissions)
	return toResolvedModelPermissions(eps, modelID), nil
}

// toResolvedModelPermissions 将 EffectivePermissionSet 映射为 ResolvedModelPermissions。
// rowScope 映射规则：
//   - RowScopeAll  → IsSelf=false（不注入 WHERE）
//   - RowScopeSelf → IsSelf=true（注入 WHERE <EndUserRef> = $endUserID）
func toResolvedModelPermissions(
	eps rbac.EffectivePermissionSet, modelID string,
) *modelruntime.ResolvedModelPermissions {
	return &modelruntime.ResolvedModelPermissions{
		Select: toActionPermission(eps.GetPermission(modelID, rbac.ActionSelect)),
		Insert: toActionPermission(eps.GetPermission(modelID, rbac.ActionInsert)),
		Update: toActionPermission(eps.GetPermission(modelID, rbac.ActionUpdate)),
		Delete: toActionPermission(eps.GetPermission(modelID, rbac.ActionDelete)),
	}
}

func toActionPermission(ep *rbac.EffectivePermission) modelruntime.ActionPermission {
	if ep == nil {
		return modelruntime.ActionPermission{Allowed: false}
	}
	return modelruntime.ActionPermission{
		Allowed: true,
		IsSelf:  ep.RowScope == rbac.RowScopeSelf,
	}
}
