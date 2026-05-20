package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/rbac"

	rbacapp "modelcraft/internal/app/rbac"
)

// endUserPermissionServiceImpl 实现 modelruntime.EndUserPermissionService。
// 依赖 rbacapp.EndUserAuthzService，支持 CUSTOM 和 PRESET 两种 grant_type。
type endUserPermissionServiceImpl struct {
	authzSvc *rbacapp.EndUserAuthzService
}

// NewEndUserPermissionService 创建 EndUserPermissionService 实例。
// modelRepo 可选：传入后支持 PRESET=READ_WRITE_OWNER 等需要 owner 字段的预设展开；
// 传 nil 时 PRESET owner-scoped 类型会被跳过（不影响 READ_ALL / READ_WRITE_ALL）。
func NewEndUserPermissionService(
	rbacRepo rbac.EndUserPermissionRepository,
	modelRepo modeldesign.ModelRepository,
) modelruntime.EndUserPermissionService {
	svc := rbacapp.NewEndUserAuthzService(rbacRepo, modelRepo)
	return &endUserPermissionServiceImpl{authzSvc: svc}
}

// Resolve 查询 end-user 在指定 model 上的有效权限，返回权限快照。
// endUserID 为空时（tenant admin）直接返回 nil, nil。
func (s *endUserPermissionServiceImpl) Resolve(
	ctx context.Context, orgName, projectSlug, endUserID, modelID string,
) (*modelruntime.ResolvedModelPermissions, error) {
	if endUserID == "" {
		return nil, nil //nolint:nilnil // nil ResolvedModelPermissions is the tenant-admin sentinel (skip all checks)
	}

	eps, err := s.authzSvc.GetEffectivePermissions(ctx, rbacapp.GetEffectivePermissionsQuery{
		ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
		UserID:       endUserID,
	})
	if err != nil {
		return nil, err
	}

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
