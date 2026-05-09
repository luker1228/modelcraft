package rbac

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"

	"github.com/google/uuid"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// EndUserRoleAppService 角色应用服务（含隐式角色保护）
type EndUserRoleAppService struct {
	rbacRepo rbacdomain.EndUserPermissionRepository
}

// NewEndUserRoleAppService creates a new EndUserRoleAppService.
func NewEndUserRoleAppService(rbacRepo rbacdomain.EndUserPermissionRepository) *EndUserRoleAppService {
	return &EndUserRoleAppService{rbacRepo: rbacRepo}
}

// CreateRole 创建普通 RBAC 角色（isImplicit 固定 false，内置角色由系统初始化）
func (s *EndUserRoleAppService) CreateRole(
	ctx context.Context,
	cmd CreateRoleCommand,
) (*rbacdomain.EndUserRole, error) {
	role := &rbacdomain.EndUserRole{
		ID:          uuid.NewString(),
		OrgName:     cmd.OrgName,
		ProjectSlug: cmd.ProjectSlug,
		Name:        cmd.Name,
		Description: cmd.Description,
		IsImplicit:  false, // 通过 API 创建的角色永远不是隐式角色
	}
	if err := s.rbacRepo.CreateRole(ctx, role); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return role, nil
}

// UpdateRole 更新角色 name/description（隐式角色 GuardUpdate() 阻断）
func (s *EndUserRoleAppService) UpdateRole(
	ctx context.Context,
	cmd UpdateRoleCommand,
) (*rbacdomain.EndUserRole, error) {
	existing, err := s.getRoleOrNotFound(ctx, cmd.OrgName, cmd.ID)
	if err != nil {
		return nil, err
	}
	// 隐式角色保护：name 不可更新
	if err := existing.GuardUpdate(); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserImplicitRoleCannotBeModified, existing.Name)
	}
	existing.Name = cmd.Name
	existing.Description = cmd.Description
	if err := s.rbacRepo.UpdateRole(ctx, existing); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserRoleNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return existing, nil
}

// DeleteRole 删除角色（隐式角色 GuardDelete() 阻断）
func (s *EndUserRoleAppService) DeleteRole(
	ctx context.Context,
	cmd DeleteRoleCommand,
) error {
	existing, err := s.getRoleOrNotFound(ctx, cmd.OrgName, cmd.ID)
	if err != nil {
		return err
	}
	// 隐式角色保护：不可删除
	if err := existing.GuardDelete(); err != nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserImplicitRoleCannotBeModified, existing.Name)
	}
	if err := s.rbacRepo.DeleteRole(ctx, cmd.OrgName, cmd.ID); err != nil {
		if shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserRoleNotFound, cmd.ID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// GetRoleByID 获取角色（含权限包列表）
func (s *EndUserRoleAppService) GetRoleByID(
	ctx context.Context,
	orgName, id string,
) (*rbacdomain.EndUserRole, error) {
	role, err := s.getRoleOrNotFound(ctx, orgName, id)
	if err != nil {
		return nil, err
	}
	// 展开关联权限包
	bundles, err := s.rbacRepo.ListBundlesByRole(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	role.Bundles = bundles
	return role, nil
}

// ListRolesByProject 列出项目下所有角色（隐式角色在前）
func (s *EndUserRoleAppService) ListRolesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbacdomain.EndUserRole, error) {
	roles, err := s.rbacRepo.ListRolesByProject(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return roles, nil
}

// AssignBundleToRole 将权限包关联到角色（受保护角色不可修改权限包）
func (s *EndUserRoleAppService) AssignBundleToRole(
	ctx context.Context,
	cmd AssignBundleToRoleCommand,
) (*rbacdomain.EndUserRole, error) {
	role, err := s.getRoleOrNotFound(ctx, cmd.OrgName, cmd.RoleID)
	if err != nil {
		return nil, err
	}
	if err := role.GuardBundleModify(); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserProtectedRoleCannotBeModified, role.Name)
	}
	if err := s.rbacRepo.AssignBundleToRole(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.RoleID, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return s.GetRoleByID(ctx, cmd.OrgName, cmd.RoleID)
}

// RevokeBundleFromRole 从角色解除权限包关联（受保护角色不可修改权限包）
func (s *EndUserRoleAppService) RevokeBundleFromRole(
	ctx context.Context,
	cmd RevokeBundleFromRoleCommand,
) (*rbacdomain.EndUserRole, error) {
	role, err := s.getRoleOrNotFound(ctx, cmd.OrgName, cmd.RoleID)
	if err != nil {
		return nil, err
	}
	if err := role.GuardBundleModify(); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserProtectedRoleCannotBeModified, role.Name)
	}
	if err := s.rbacRepo.RevokeBundleFromRole(ctx, cmd.RoleID, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return s.GetRoleByID(ctx, cmd.OrgName, cmd.RoleID)
}

// AssignRoleToUser 将显式角色分配给用户（通道 2；隐式角色不可手动分配）
// 存在性校验：若 EndUser 不属于当前 Org（外键约束失败），返回 EndUserNotFoundInProject。
func (s *EndUserRoleAppService) AssignRoleToUser(
	ctx context.Context,
	cmd AssignRoleToUserCommand,
) error {
	// 检查角色是否为隐式角色
	role, err := s.getRoleOrNotFound(ctx, cmd.OrgName, cmd.RoleID)
	if err != nil {
		return err
	}
	if role.IsImplicit {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserCannotAssignImplicitRole)
	}
	if err := s.rbacRepo.AssignRoleToUser(ctx, cmd.UserID, cmd.OrgName, cmd.ProjectSlug, cmd.RoleID); err != nil {
		// 外键约束失败 = EndUser 不属于该 Org（end_user_users.org_name 不匹配）
		if shared.IsRepoError(err, shared.ErrTypeConstraint) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFoundInProject, cmd.UserID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// RevokeRoleFromUser 撤销用户角色分配
func (s *EndUserRoleAppService) RevokeRoleFromUser(
	ctx context.Context,
	cmd RevokeRoleFromUserCommand,
) error {
	if err := s.rbacRepo.RevokeRoleFromUser(ctx, cmd.UserID, cmd.OrgName, cmd.ProjectSlug, cmd.RoleID); err != nil {
		// 外键约束失败 = EndUser 不属于该 Org
		if shared.IsRepoError(err, shared.ErrTypeConstraint) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFoundInProject, cmd.UserID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// ListProjectEndUserRoleUsers 列出 Project 下所有有角色分配的用户
func (s *EndUserRoleAppService) ListProjectEndUserRoleUsers(
	ctx context.Context,
	cmd ListProjectEndUserRoleUsersQuery,
) ([]*rbacdomain.ProjectEndUserRoleUser, int64, error) {
	items, total, err := s.rbacRepo.ListProjectEndUserRoleUsers(ctx, rbacdomain.ListProjectEndUserRoleUsersQuery{
		OrgName:     cmd.OrgName,
		ProjectSlug: cmd.ProjectSlug,
		Search:      cmd.Search,
		RoleID:      cmd.RoleID,
		First:       cmd.First,
		After:       cmd.After,
	})
	if err != nil {
		return nil, 0, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return items, total, nil
}

// getRoleOrNotFound 内部辅助：获取角色，不存在时返回 NotFound 业务错误
func (s *EndUserRoleAppService) getRoleOrNotFound(
	ctx context.Context,
	orgName, id string,
) (*rbacdomain.EndUserRole, error) {
	role, err := s.rbacRepo.GetRoleByID(ctx, orgName, id)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserRoleNotFound, id)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return role, nil
}
