package rbac

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	rbacdomain "modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"

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

	GetPermissionByID(ctx context.Context, orgName, id string) (*rbacdomain.EndUserPermission, error)

	UpsertBundleDataPermissionItem(ctx context.Context, item *rbacdomain.EndUserBundleDataPermissionItem) error
	RemoveBundleDataPermissionItem(ctx context.Context, bundleID, modelID string) error
	ListBundleDataPermissionItems(
		ctx context.Context,
		bundleID string,
	) ([]*rbacdomain.EndUserBundleDataPermissionItem, error)
	GetBundleDataPermissionItemByBundleAndModel(
		ctx context.Context,
		bundleID, modelID string,
	) (*rbacdomain.EndUserBundleDataPermissionItem, error)

	GrantBundleToUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error
	RevokeBundleFromUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error

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

// GetBundleByID 获取权限包（含 item 列表与快照）
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

	items, err := s.rbacRepo.ListBundleDataPermissionItems(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	bundle.Items = items

	// 兼容旧 DTO 字段 permissions
	legacyPermissions := make([]*rbacdomain.EndUserPermission, 0, len(items))
	for _, item := range items {
		switch item.GrantType {
		case rbacdomain.PermissionTypeCustom:
			if item.CustomPermissionID == nil || *item.CustomPermissionID == "" {
				continue
			}
			p, getErr := s.rbacRepo.GetPermissionByID(ctx, orgName, *item.CustomPermissionID)
			if getErr == nil {
				legacyPermissions = append(legacyPermissions, p)
			}
		case rbacdomain.PermissionTypePreset:
			if item.Preset == nil {
				continue
			}
			ownerField, _ := s.tryGetOwnerField(ctx, item.ModelID)
			rowPolicy, expandErr := expandPreset(*item.Preset, ownerField)
			if expandErr != nil {
				continue
			}
			presetCopy := *item.Preset
			legacyPermissions = append(legacyPermissions, &rbacdomain.EndUserPermission{
				ID:           fmt.Sprintf("preset:%s:%s:%s", bundle.ID, item.ModelID, presetCopy),
				OrgName:      bundle.OrgName,
				ProjectSlug:  bundle.ProjectSlug,
				ModelID:      item.ModelID,
				Name:         presetPermissionName(presetCopy),
				Type:         rbacdomain.PermissionTypePreset,
				Preset:       &presetCopy,
				ColumnPolicy: nil,
				RowPolicy:    rowPolicy,
			})
		}
	}
	bundle.Permissions = legacyPermissions

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

// AddPermissionToBundle 绑定 custom item（replace 语义）
func (s *EndUserBundleAppService) AddPermissionToBundle(
	ctx context.Context,
	cmd AddPermissionToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if _, err := s.rbacRepo.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.BundleID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	perm, err := s.rbacRepo.GetPermissionByID(ctx, cmd.OrgName, cmd.PermissionID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.PermissionID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	customID := perm.ID
	item := &rbacdomain.EndUserBundleDataPermissionItem{
		ID:                 uuid.NewString(),
		BundleID:           cmd.BundleID,
		ModelID:            perm.ModelID,
		GrantType:          rbacdomain.PermissionTypeCustom,
		CustomPermissionID: &customID,
		SortOrder:          cmd.SortOrder,
	}
	if err := s.rbacRepo.UpsertBundleDataPermissionItem(ctx, item); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// AddPresetToBundle 绑定 preset item（replace 语义）
func (s *EndUserBundleAppService) AddPresetToBundle(
	ctx context.Context,
	cmd AddPresetToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if _, err := s.rbacRepo.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, cmd.BundleID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	if !cmd.Preset.IsValid() {
		return nil, bizerrors.NewValidationError(
			"rbac.permission.invalid_preset: invalid preset: %s", string(cmd.Preset))
	}

	ownerField, err := s.requireModelAndGetOwnerField(ctx, cmd.ModelID)
	if err != nil {
		return nil, err
	}
	if _, err := expandPreset(cmd.Preset, ownerField); err != nil {
		return nil, err
	}

	preset := cmd.Preset
	item := &rbacdomain.EndUserBundleDataPermissionItem{
		ID:        uuid.NewString(),
		BundleID:  cmd.BundleID,
		ModelID:   cmd.ModelID,
		GrantType: rbacdomain.PermissionTypePreset,
		Preset:    &preset,
		SortOrder: cmd.SortOrder,
	}
	if err := s.rbacRepo.UpsertBundleDataPermissionItem(ctx, item); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// RemovePermissionFromBundle 从权限包移除 custom item（按 permission 对应 model 删除）
func (s *EndUserBundleAppService) RemovePermissionFromBundle(
	ctx context.Context,
	cmd RemovePermissionFromBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	perm, err := s.rbacRepo.GetPermissionByID(ctx, cmd.OrgName, cmd.PermissionID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.PermissionID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	if err := s.rbacRepo.RemoveBundleDataPermissionItem(ctx, cmd.BundleID, perm.ModelID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// BindPresetItem 绑定预设模板 item 到 bundle（replace 语义，显式 item-centric 入口）
func (s *EndUserBundleAppService) BindPresetItem(
	ctx context.Context,
	cmd BindPresetItemToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.verifyBundleScope(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.AddPresetToBundle(ctx, AddPresetToBundleCommand{
		OrgName:   cmd.OrgName,
		BundleID:  cmd.BundleID,
		ModelID:   cmd.ModelID,
		Preset:    cmd.Preset,
		SortOrder: cmd.SortOrder,
	})
}

// BindCustomItem 绑定自定义权限 item 到 bundle（replace 语义，显式 item-centric 入口）。
// 与旧 AddPermissionToBundle 的区别：直接使用 cmd.ModelID，不再依赖 permission 实体查询。
func (s *EndUserBundleAppService) BindCustomItem(
	ctx context.Context,
	cmd BindCustomItemToBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.verifyBundleScope(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.BundleID); err != nil {
		return nil, err
	}

	perm, err := s.rbacRepo.GetPermissionByID(ctx, cmd.OrgName, cmd.CustomPermissionID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.CustomPermissionID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	customID := perm.ID
	item := &rbacdomain.EndUserBundleDataPermissionItem{
		ID:                 uuid.NewString(),
		BundleID:           cmd.BundleID,
		ModelID:            cmd.ModelID,
		GrantType:          rbacdomain.PermissionTypeCustom,
		CustomPermissionID: &customID,
		SortOrder:          cmd.SortOrder,
	}
	if err := s.rbacRepo.UpsertBundleDataPermissionItem(ctx, item); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// RemoveDataPermissionItemFromBundle 从权限包移除指定模型的 data permission item（按 modelID）
func (s *EndUserBundleAppService) RemoveDataPermissionItemFromBundle(
	ctx context.Context,
	cmd RemoveDataPermissionItemFromBundleCommand,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if err := s.verifyBundleScope(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.BundleID); err != nil {
		return nil, err
	}

	if err := s.rbacRepo.RemoveBundleDataPermissionItem(ctx, cmd.BundleID, cmd.ModelID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := s.saveBundleSnapshot(ctx, cmd.BundleID); err != nil {
		return nil, err
	}
	return s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
}

// verifyBundleScope 校验 bundle 归属（org + project），防止跨 project 操作。
func (s *EndUserBundleAppService) verifyBundleScope(
	ctx context.Context,
	orgName, projectSlug, bundleID string,
) error {
	b, err := s.rbacRepo.GetBundleByID(ctx, orgName, bundleID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, bundleID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if b.ProjectSlug != projectSlug {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, bundleID)
	}
	return nil
}

// saveBundleSnapshot 写入 item 快照并执行滚动删除。
func (s *EndUserBundleAppService) saveBundleSnapshot(ctx context.Context, bundleID string) error {
	currentVersion, err := s.rbacRepo.GetBundleCurrentVersion(ctx, bundleID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	newVersion := currentVersion + 1

	items, err := s.rbacRepo.ListBundleDataPermissionItems(ctx, bundleID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	snapshotItems := make([]rbacdomain.SnapshotItemEntry, 0, len(items))
	legacyEntries := make([]rbacdomain.SnapshotPermissionEntry, 0, len(items))
	for _, item := range items {
		snapshotItems = append(snapshotItems, rbacdomain.SnapshotItemEntry{
			ModelID:            item.ModelID,
			GrantType:          item.GrantType,
			Preset:             item.Preset,
			CustomPermissionID: item.CustomPermissionID,
			SortOrder:          item.SortOrder,
		})

		permissionID := item.ModelID
		if item.CustomPermissionID != nil && *item.CustomPermissionID != "" {
			permissionID = *item.CustomPermissionID
		}
		legacyEntries = append(legacyEntries, rbacdomain.SnapshotPermissionEntry{
			PermissionID: permissionID,
			SortOrder:    item.SortOrder,
		})
	}

	snapshot := &rbacdomain.BundleSnapshot{
		ID:          uuid.NewString(),
		BundleID:    bundleID,
		Version:     newVersion,
		Permissions: legacyEntries,
		Items:       snapshotItems,
	}
	if err := s.rbacRepo.SaveBundleSnapshot(ctx, snapshot); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
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

// RestoreBundle 将权限包回滚到历史快照版本。
func (s *EndUserBundleAppService) RestoreBundle(
	ctx context.Context,
	cmd RestoreBundleCommand,
) (*RestoreBundleResult, error) {
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

	if err := s.rbacRepo.ClearBundlePermissions(ctx, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	if len(snapshot.Items) > 0 {
		for _, entry := range snapshot.Items {
			item := &rbacdomain.EndUserBundleDataPermissionItem{
				ID:                 uuid.NewString(),
				BundleID:           cmd.BundleID,
				ModelID:            entry.ModelID,
				GrantType:          entry.GrantType,
				Preset:             entry.Preset,
				CustomPermissionID: entry.CustomPermissionID,
				SortOrder:          entry.SortOrder,
			}
			if upsertErr := s.rbacRepo.UpsertBundleDataPermissionItem(ctx, item); upsertErr != nil {
				return nil, bizerrors.ConvertRepositoryError(ctx, upsertErr)
			}
		}
	}

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
		Items:        snapshot.Items,
		Permissions:  snapshot.Permissions,
		RestoredFrom: &targetVersion,
	}
	if err := s.rbacRepo.SaveBundleSnapshot(ctx, restoreSnapshot); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	if err := s.rbacRepo.DeleteOldBundleSnapshots(ctx, cmd.BundleID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	bundle, err := s.GetBundleByID(ctx, cmd.OrgName, cmd.BundleID)
	if err != nil {
		return nil, err
	}
	return &RestoreBundleResult{Bundle: bundle, NewVersion: newVersion}, nil
}

func (s *EndUserBundleAppService) requireModelAndGetOwnerField(ctx context.Context, modelID string) (string, error) {
	model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		if shared.IsNotFoundError(err) {
			return "", bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
		}
		return "", bizerrors.ConvertRepositoryError(ctx, err)
	}
	ownerField := ""
	if model != nil && model.GetOwnerField() != nil {
		ownerField = model.GetOwnerField().Name
	}
	return ownerField, nil
}

func (s *EndUserBundleAppService) tryGetOwnerField(ctx context.Context, modelID string) (string, error) {
	model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil || model == nil {
		return "", err
	}
	if model.GetOwnerField() == nil {
		return "", nil
	}
	return model.GetOwnerField().Name, nil
}
