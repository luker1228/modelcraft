package rbac

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"

	"github.com/google/uuid"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// EndUserPermissionAppService 权限点应用服务
type EndUserPermissionAppService struct {
	rbacRepo  rbacdomain.EndUserPermissionRepository
	modelRepo modeldesign.ModelRepository // 跨域，用于 rowScope 字段前提校验
}

// NewEndUserPermissionAppService creates a new EndUserPermissionAppService.
func NewEndUserPermissionAppService(
	rbacRepo rbacdomain.EndUserPermissionRepository,
	modelRepo modeldesign.ModelRepository,
) *EndUserPermissionAppService {
	return &EndUserPermissionAppService{
		rbacRepo:  rbacRepo,
		modelRepo: modelRepo,
	}
}

// CreatePermission 创建权限点（含 rowScope 字段前提校验）
func (s *EndUserPermissionAppService) CreatePermission(
	ctx context.Context,
	cmd CreatePermissionCommand,
) (*rbacdomain.EndUserPermission, error) {
	// 1. rowScope 字段前提校验
	if err := s.validateRowScopePrerequisite(ctx, cmd.ModelID, cmd.RowScope); err != nil {
		return nil, err
	}

	// 2. 构建实体
	perm := &rbacdomain.EndUserPermission{
		ID:           uuid.NewString(),
		OrgName:      cmd.OrgName,
		ProjectSlug:  cmd.ProjectSlug,
		ModelID:      cmd.ModelID,
		Name:         cmd.Name,
		Description:  cmd.Description,
		Action:       cmd.Action,
		ColumnPolicy: cmd.ColumnPolicy,
		RowScope:     cmd.RowScope,
	}

	// 3. 领域校验
	if err := perm.Validate(); err != nil {
		return nil, err
	}

	// 4. 持久化
	if err := s.rbacRepo.CreatePermission(ctx, perm); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perm, nil
}

// UpdatePermission 更新权限点（只允许更新 name/description/columnPolicy；action 和 rowScope 不可变）
func (s *EndUserPermissionAppService) UpdatePermission(
	ctx context.Context,
	cmd UpdatePermissionCommand,
) (*rbacdomain.EndUserPermission, error) {
	// 1. 获取现有权限点
	existing, err := s.rbacRepo.GetPermissionByID(ctx, cmd.OrgName, cmd.ID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 2. 更新字段（action 和 modelID 不可变）
	existing.Name = cmd.Name
	existing.Description = cmd.Description
	existing.ColumnPolicy = cmd.ColumnPolicy

	// 3. 持久化
	if err := s.rbacRepo.UpdatePermission(ctx, existing); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return existing, nil
}

// DeletePermission 删除权限点
func (s *EndUserPermissionAppService) DeletePermission(
	ctx context.Context,
	cmd DeletePermissionCommand,
) error {
	if err := s.rbacRepo.DeletePermission(ctx, cmd.OrgName, cmd.ID); err != nil {
		if shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// ListPermissionsByProject 列出项目下所有权限点
func (s *EndUserPermissionAppService) ListPermissionsByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbacdomain.EndUserPermission, error) {
	perms, err := s.rbacRepo.ListPermissionsByProject(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perms, nil
}

// GetPermissionByID 获取权限点
func (s *EndUserPermissionAppService) GetPermissionByID(
	ctx context.Context,
	orgName, id string,
) (*rbacdomain.EndUserPermission, error) {
	perm, err := s.rbacRepo.GetPermissionByID(ctx, orgName, id)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, id)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perm, nil
}

// validateRowScopePrerequisite 校验 rowScope 字段前提（App 层，跨域访问 ModelRepository）
// SELF         → 要求 owner 字段（END_USER_REF 类型）
// DEPT / DEPT_AND_CHILDREN → 要求 dept_id 字段
// ALL          → 无要求
func (s *EndUserPermissionAppService) validateRowScopePrerequisite(
	ctx context.Context,
	modelID string,
	rowScope rbacdomain.RowScope,
) error {
	switch rowScope {
	case rbacdomain.RowScopeAll:
		return nil // ALL 无字段前提
	case rbacdomain.RowScopeSelf:
		model, err := s.modelRepo.GetByID(ctx, modelID)
		if err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if model == nil || model.GetOwnerField() == nil {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserRowScopeFieldMissing, "SELF", "owner")
		}
	case rbacdomain.RowScopeDept, rbacdomain.RowScopeDeptAndChildren:
		model, err := s.modelRepo.GetByID(ctx, modelID)
		if err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if model == nil {
			return bizerrors.NewErrorFromContext(
				ctx, bizerrors.EndUserRowScopeFieldMissing, string(rowScope), "dept_id")
		}
		if !model.HasField("dept_id") {
			return bizerrors.NewErrorFromContext(
				ctx, bizerrors.EndUserRowScopeFieldMissing, string(rowScope), "dept_id")
		}
	}
	return nil
}
