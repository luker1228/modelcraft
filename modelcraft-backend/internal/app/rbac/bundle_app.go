package rbac

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"

	rbacdomain "modelcraft/internal/domain/rbac"

	"github.com/google/uuid"
)

type bundleRepository interface {
	CreateBundle(ctx context.Context, b *rbacdomain.EndUserPermissionBundle) error
	GetBundleByID(ctx context.Context, orgName, id string) (*rbacdomain.EndUserPermissionBundle, error)
	ListBundlesByProject(
		ctx context.Context,
		orgName, projectSlug string,
	) ([]*rbacdomain.EndUserPermissionBundle, error)
	UpdateBundle(ctx context.Context, b *rbacdomain.EndUserPermissionBundle) error
	DeleteBundle(ctx context.Context, orgName, id string) error
	AddPermissionToBundle(ctx context.Context, bundleID, permissionID string, sortOrder int) error
	RemovePermissionFromBundle(ctx context.Context, bundleID, permissionID string) error
	ListPermissionsInBundle(ctx context.Context, bundleID string) ([]*rbacdomain.EndUserPermission, error)
	GrantBundleToUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error
	RevokeBundleFromUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error
	ListPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*rbacdomain.EndUserPermission, error)
	GetPermissionByModelTypeName(
		ctx context.Context,
		orgName, modelID string,
		permissionType rbacdomain.PermissionType,
		name string,
	) (*rbacdomain.EndUserPermission, error)
	CreatePermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	// Snapshot operations
	SaveBundleSnapshot(ctx context.Context, snapshot *rbacdomain.BundleSnapshot) error
	ListBundleSnapshots(ctx context.Context, bundleID string) ([]rbacdomain.BundleSnapshot, error)
	DeleteOldBundleSnapshots(ctx context.Context, bundleID string) error
	GetBundleCurrentVersion(ctx context.Context, bundleID string) (int, error)
	GetBundleSnapshotByVersion(ctx context.Context, bundleID string, version int) (*rbacdomain.BundleSnapshot, error)
	ClearBundlePermissions(ctx context.Context, bundleID string) error
}

// EndUserBundleAppService 权限包应用服务
type EndUserBundleAppService struct {
	rbacRepo  bundleRepository
	modelRepo modelRepository
}

// NewEndUserBundleAppService creates a new EndUserBundleAppService.
func NewEndUserBundleAppService(
	rbacRepo rbacdomain.EndUserPermissionRepository,
	modelRepo modeldesign.ModelRepository,
) *EndUserBundleAppService {
	return &EndUserBundleAppService{
		rbacRepo:  rbacRepo,
		modelRepo: modelRepo,
	}
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
	// 展开历史快照列表
	snapshots, err := s.rbacRepo.ListBundleSnapshots(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	bundle.Snapshots = snapshots
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
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	// 返回更新后的权限包（含权限点列表）
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// AddPresetToBundle 向权限包添加模型预设权限（自动 ensure 预设权限点，幂等）
func (s *EndUserBundleAppService) AddPresetToBundle(
	ctx context.Context,
	cmd AddPresetToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	bundle, err := s.rbacRepo.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.BundleID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	presetPerm, err := s.ensurePresetPermission(ctx, bundle, cmd.ModelID, cmd.Preset)
	if err != nil {
		return nil, err
	}

	if err := s.addPermissionToBundleIfAbsent(ctx, bundle.ID, presetPerm.ID, cmd.SortOrder); err != nil {
		return nil, err
	}

	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

func (s *EndUserBundleAppService) ensurePresetPermission(
	ctx context.Context,
	bundle *rbacdomain.EndUserPermissionBundle,
	modelID string,
	preset rbacdomain.PermissionPreset,
) (*rbacdomain.EndUserPermission, error) {
	if !preset.IsValid() {
		return nil, bizerrors.NewValidationError("rbac.permission.invalid_preset: invalid preset: %s", string(preset))
	}

	model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	ownerField := ""
	if model != nil && model.GetOwnerField() != nil {
		ownerField = model.GetOwnerField().Name
	}

	existing, err := s.rbacRepo.GetPermissionByModelTypeName(
		ctx,
		bundle.OrgName,
		modelID,
		rbacdomain.PermissionTypePreset,
		presetPermissionName(preset),
	)
	if err == nil {
		return existing, nil
	}
	if !shared.IsNotFoundError(err) {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	rowPolicy, err := expandPreset(preset, ownerField)
	if err != nil {
		return nil, err
	}

	toCreate := &rbacdomain.EndUserPermission{
		ID:           uuid.NewString(),
		OrgName:      bundle.OrgName,
		ProjectSlug:  bundle.ProjectSlug,
		ModelID:      modelID,
		Name:         presetPermissionName(preset),
		Type:         rbacdomain.PermissionTypePreset,
		ColumnPolicy: nil,
		RowPolicy:    rowPolicy,
	}
	presetCopy := preset
	toCreate.Preset = &presetCopy
	if err := toCreate.Validate(); err != nil {
		return nil, err
	}

	if err := s.rbacRepo.CreatePermission(ctx, toCreate); err != nil {
		if !shared.IsDuplicateKeyError(err) {
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}
		reloaded, getErr := s.rbacRepo.GetPermissionByModelTypeName(
			ctx,
			bundle.OrgName,
			modelID,
			rbacdomain.PermissionTypePreset,
			presetPermissionName(preset),
		)
		if getErr != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, getErr)
		}
		return reloaded, nil
	}

	return toCreate, nil
}

func (s *EndUserBundleAppService) addPermissionToBundleIfAbsent(
	ctx context.Context,
	bundleID, permissionID string,
	sortOrder int,
) error {
	perms, err := s.rbacRepo.ListPermissionsInBundle(ctx, bundleID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	for _, p := range perms {
		if p.ID == permissionID {
			return nil
		}
	}

	if err := s.rbacRepo.AddPermissionToBundle(ctx, bundleID, permissionID, sortOrder); err != nil {
		if shared.IsDuplicateKeyError(err) {
			return nil
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// RemovePermissionFromBundle 从权限包移除权限点
func (s *EndUserBundleAppService) RemovePermissionFromBundle(
	ctx context.Context,
	cmd RemovePermissionFromBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.rbacRepo.RemovePermissionFromBundle(ctx, cmd.BundleID, cmd.PermissionID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// saveBundleSnapshot 写入快照并执行滚动删除（在权限点列表变更后调用）
func (s *EndUserBundleAppService) saveBundleSnapshot(ctx context.Context, bundleID string) error {
	// 1. 获取当前版本号
	currentVersion, err := s.rbacRepo.GetBundleCurrentVersion(ctx, bundleID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	newVersion := currentVersion + 1

	// 2. 获取当前权限点列表
	perms, err := s.rbacRepo.ListPermissionsInBundle(ctx, bundleID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	entries := make([]rbacdomain.SnapshotPermissionEntry, 0, len(perms))
	for i, p := range perms {
		entries = append(entries, rbacdomain.SnapshotPermissionEntry{
			PermissionID: p.ID,
			SortOrder:    i,
		})
	}

	// 3. 写入快照
	snapshot := &rbacdomain.BundleSnapshot{
		ID:          uuid.NewString(),
		BundleID:    bundleID,
		Version:     newVersion,
		Permissions: entries,
	}
	if err := s.rbacRepo.SaveBundleSnapshot(ctx, snapshot); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 4. 滚动删除超出上限的旧快照
	if err := s.rbacRepo.DeleteOldBundleSnapshots(ctx, bundleID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
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

// RestoreBundle 将权限包回滚到历史快照版本
// 操作步骤：查询目标快照 → 清空当前权限点 → 按快照重建 → 写新快照（restored_from） → 滚动删除
func (s *EndUserBundleAppService) RestoreBundle(
	ctx context.Context,
	cmd RestoreBundleCommand,
) (*RestoreBundleResult, error) {
	// 1. 查询目标快照
	snapshot, err := s.rbacRepo.GetBundleSnapshotByVersion(ctx, cmd.BundleID, cmd.TargetVersion)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(
				ctx,
				bizerrors.EndUserPermissionBundleSnapshotNotFound,
				cmd.TargetVersion,
				cmd.BundleID,
			)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 2. 清空当前权限点关联
	if err := s.rbacRepo.ClearBundlePermissions(ctx, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 3. 按快照重建权限点关联
	for _, entry := range snapshot.Permissions {
		if addErr := s.rbacRepo.AddPermissionToBundle(
			ctx,
			cmd.BundleID,
			entry.PermissionID,
			entry.SortOrder,
		); addErr != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, addErr)
		}
	}

	// 4. 获取新版本号并写入快照（restored_from 指向目标版本）
	currentVersion, err := s.rbacRepo.GetBundleCurrentVersion(ctx, cmd.BundleID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	newVersion := currentVersion + 1
	targetVersion := cmd.TargetVersion
	restoreSnapshot := &rbacdomain.BundleSnapshot{
		ID:           uuid.NewString(),
		BundleID:     cmd.BundleID,
		Version:      newVersion,
		Permissions:  snapshot.Permissions,
		RestoredFrom: &targetVersion,
	}
	if err := s.rbacRepo.SaveBundleSnapshot(ctx, restoreSnapshot); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 5. 滚动删除超出上限的旧快照
	if err := s.rbacRepo.DeleteOldBundleSnapshots(ctx, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 6. 返回更新后的权限包
	bundle, err := s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
	if err != nil {
		return nil, err
	}
	return &RestoreBundleResult{Bundle: bundle, NewVersion: newVersion}, nil
}
