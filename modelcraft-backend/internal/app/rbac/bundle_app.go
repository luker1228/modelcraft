package rbac

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"

	"github.com/google/uuid"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// EndUserBundleAppService 权限包应用服务
type EndUserBundleAppService struct {
	rbacRepo rbacdomain.EndUserPermissionRepository
}

// NewEndUserBundleAppService creates a new EndUserBundleAppService.
func NewEndUserBundleAppService(rbacRepo rbacdomain.EndUserPermissionRepository) *EndUserBundleAppService {
	return &EndUserBundleAppService{rbacRepo: rbacRepo}
}

// CreateBundle 创建权限包
func (s *EndUserBundleAppService) CreateBundle(
	ctx context.Context,
	cmd CreateBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	bundle := &rbacdomain.EndUserPermissionBundle{
		ID:          uuid.NewString(),
		OrgName:     cmd.OrgName,
		ProjectSlug: cmd.ProjectSlug,
		Name:        cmd.Name,
		Description: cmd.Description,
	}
	if err := bundle.Validate(); err != nil {
		return nil, err
	}
	if err := s.rbacRepo.CreateBundle(ctx, bundle); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return bundle, nil
}

// UpdateBundle 更新权限包 name/description
func (s *EndUserBundleAppService) UpdateBundle(
	ctx context.Context,
	cmd UpdateBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	existing, err := s.rbacRepo.GetBundleByID(ctx, cmd.OrgName, cmd.ID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	existing.Name = cmd.Name
	existing.Description = cmd.Description
	if err := s.rbacRepo.UpdateBundle(ctx, existing); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return existing, nil
}

// DeleteBundle 删除权限包
func (s *EndUserBundleAppService) DeleteBundle(
	ctx context.Context,
	cmd DeleteBundleCommand,
) error {
	if err := s.rbacRepo.DeleteBundle(ctx, cmd.OrgName, cmd.ID); err != nil {
		if shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.ID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// GetBundleByID 获取权限包（含权限点列表）
func (s *EndUserBundleAppService) GetBundleByID(
	ctx context.Context,
	orgName, id string,
) (*rbacdomain.EndUserPermissionBundle, error) {
	bundle, err := s.rbacRepo.GetBundleByID(ctx, orgName, id)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, id)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	// 展开权限点列表
	perms, err := s.rbacRepo.ListPermissionsInBundle(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	bundle.Permissions = perms
	return bundle, nil
}

// ListBundlesByProject 列出项目下所有权限包
func (s *EndUserBundleAppService) ListBundlesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbacdomain.EndUserPermissionBundle, error) {
	bundles, err := s.rbacRepo.ListBundlesByProject(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return bundles, nil
}

// AddPermissionToBundle 向权限包添加权限点
func (s *EndUserBundleAppService) AddPermissionToBundle(
	ctx context.Context,
	cmd AddPermissionToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.rbacRepo.AddPermissionToBundle(ctx, cmd.BundleID, cmd.PermissionID, cmd.SortOrder); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	// 返回更新后的权限包（含权限点列表）
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// RemovePermissionFromBundle 从权限包移除权限点
func (s *EndUserBundleAppService) RemovePermissionFromBundle(
	ctx context.Context,
	cmd RemovePermissionFromBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.rbacRepo.RemovePermissionFromBundle(ctx, cmd.BundleID, cmd.PermissionID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// GrantBundleToUser 直接将权限包授予用户（通道 1）
func (s *EndUserBundleAppService) GrantBundleToUser(
	ctx context.Context,
	cmd GrantBundleToUserCommand,
) error {
	if err := s.rbacRepo.GrantBundleToUser(ctx, cmd.UserID, cmd.OrgName, cmd.ProjectSlug, cmd.BundleID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// RevokeBundleFromUser 撤销用户权限包
func (s *EndUserBundleAppService) RevokeBundleFromUser(
	ctx context.Context,
	cmd RevokeBundleFromUserCommand,
) error {
	if err := s.rbacRepo.RevokeBundleFromUser(ctx, cmd.UserID, cmd.OrgName, cmd.ProjectSlug, cmd.BundleID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}
