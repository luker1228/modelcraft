package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"

	rbacdomain "modelcraft/internal/domain/rbac"

	"github.com/google/uuid"
)

type permissionRepository interface {
	CreatePermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	GetPermissionByID(ctx context.Context, orgName, id string) (*rbacdomain.EndUserPermission, error)
	ListPermissionsByProject(ctx context.Context, orgName, projectSlug string) ([]*rbacdomain.EndUserPermission, error)
	ListPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*rbacdomain.EndUserPermission, error)
	ListPresetPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*rbacdomain.EndUserPermission, error)
	UpdatePermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	DeletePermission(ctx context.Context, orgName, id string) error
	DeletePresetPermissionsByModel(ctx context.Context, orgName, modelID string) error
	UpdatePresetPermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	IsPermissionReferencedByBundle(ctx context.Context, permissionID string) (bool, error)
}

type modelRepository interface {
	GetByID(ctx context.Context, id string, opts ...*modeldesign.ModelQueryOptions) (*modeldesign.DataModel, error)
}

// EndUserPermissionAppService 权限点应用服务
type EndUserPermissionAppService struct {
	rbacRepo    permissionRepository
	modelRepo   modelRepository
	txManager   repository.TxManager
	repoFromTxQ func(q dbgen.Querier) permissionRepository
}

// NewEndUserPermissionAppService creates a new EndUserPermissionAppService.
func NewEndUserPermissionAppService(
	rbacRepo rbacdomain.EndUserPermissionRepository,
	modelRepo modeldesign.ModelRepository,
	txManagers ...repository.TxManager,
) *EndUserPermissionAppService {
	svc := &EndUserPermissionAppService{
		rbacRepo:  rbacRepo,
		modelRepo: modelRepo,
		repoFromTxQ: func(q dbgen.Querier) permissionRepository {
			return repository.NewSqlEndUserDataPermissionRepository(q)
		},
	}
	if len(txManagers) > 0 {
		svc.txManager = txManagers[0]
	}
	return svc
}

// CreatePermission 创建权限点（兼容旧输入 action/rowScope，内部转换为 rowPolicy）
func (s *EndUserPermissionAppService) CreatePermission(
	ctx context.Context,
	cmd CreatePermissionCommand,
) (*rbacdomain.EndUserPermission, error) {
	if err := s.validateRowScopePrerequisite(ctx, cmd.ModelID, cmd.RowScope); err != nil {
		return nil, err
	}

	databaseName, modelName, err := s.getPermissionModelNames(ctx, cmd.ModelID)
	if err != nil {
		return nil, err
	}

	rowPolicy := buildLegacyRowPolicy(cmd.Action, cmd.RowScope)
	perm := &rbacdomain.EndUserPermission{
		ID:           uuid.NewString(),
		OrgName:      cmd.OrgName,
		ProjectSlug:  cmd.ProjectSlug,
		ModelID:      cmd.ModelID,
		DatabaseName: databaseName,
		ModelName:    modelName,
		Name:         cmd.Name,
		Description:  cmd.Description,
		Type:         rbacdomain.PermissionTypeCustom,
		ColumnPolicy: cmd.ColumnPolicy,
		RowPolicy:    rowPolicy,
		Preset:       nil,
	}
	if err := perm.Validate(); err != nil {
		return nil, err
	}

	if err := s.rbacRepo.CreatePermission(ctx, perm); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perm, nil
}

// UpdatePermission 更新权限点（只允许更新 name/description/columnPolicy）
func (s *EndUserPermissionAppService) UpdatePermission(
	ctx context.Context,
	cmd UpdatePermissionCommand,
) (*rbacdomain.EndUserPermission, error) {
	existing, err := s.rbacRepo.GetPermissionByID(ctx, cmd.OrgName, cmd.ID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	existing.Name = cmd.Name
	existing.Description = cmd.Description
	existing.ColumnPolicy = cmd.ColumnPolicy

	if err := s.rbacRepo.UpdatePermission(ctx, existing); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return existing, nil
}

// DeletePermission 删除权限点（删除前校验是否被权限包引用）
func (s *EndUserPermissionAppService) DeletePermission(
	ctx context.Context,
	cmd DeletePermissionCommand,
) error {
	referenced, err := s.rbacRepo.IsPermissionReferencedByBundle(ctx, cmd.ID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if referenced {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionInUse, cmd.ID)
	}

	if err := s.rbacRepo.DeletePermission(ctx, cmd.OrgName, cmd.ID); err != nil {
		if shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionNotFound, cmd.ID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// ApplyPresetPolicy 对模型执行 PRESET 权限点差异同步（reconcile）。
func (s *EndUserPermissionAppService) ApplyPresetPolicy(
	ctx context.Context,
	cmd ApplyPresetPolicyCommand,
) ([]*rbacdomain.EndUserPermission, error) {
	desiredPresets, ownerField, databaseName, modelName, err := s.resolveDesiredPresetsWithModel(ctx, cmd)
	if err != nil {
		return nil, err
	}

	reconcile := func(execRepo permissionRepository) error {
		existingByPreset, listErr := s.listExistingPresetByModel(ctx, execRepo, cmd.OrgName, cmd.ModelID)
		if listErr != nil {
			return listErr
		}

		desiredSet := make(map[rbacdomain.PermissionPreset]struct{}, len(desiredPresets))
		for _, preset := range desiredPresets {
			desiredSet[preset] = struct{}{}
		}

		if deleteCheckErr := s.ensureDeletePresetSafe(
			ctx,
			execRepo,
			existingByPreset,
			desiredSet,
		); deleteCheckErr != nil {
			return deleteCheckErr
		}
		if upsertErr := s.upsertDesiredPresets(
			ctx,
			execRepo,
			cmd,
			desiredPresets,
			ownerField,
			databaseName,
			modelName,
			existingByPreset,
		); upsertErr != nil {
			return upsertErr
		}
		if deleteErr := s.deleteStalePresets(
			ctx,
			execRepo,
			cmd.OrgName,
			existingByPreset,
			desiredSet,
		); deleteErr != nil {
			return deleteErr
		}
		return nil
	}

	if s.txManager != nil && s.repoFromTxQ != nil {
		err = s.txManager.WithTx(ctx, func(txCtx context.Context, q dbgen.Querier) error {
			txRepo := s.repoFromTxQ(q)
			return reconcile(txRepo)
		})
	} else {
		err = reconcile(s.rbacRepo)
	}
	if err != nil {
		return nil, err
	}

	perms, err := s.rbacRepo.ListPermissionsByModel(ctx, cmd.OrgName, cmd.ModelID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perms, nil
}

// ListVirtualPresetsByModel 只读计算模型可用预设集合（不落库）。
func (s *EndUserPermissionAppService) ListVirtualPresetsByModel(
	ctx context.Context,
	modelID string,
) ([]rbacdomain.PermissionPreset, error) {
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
	return resolveDesiredPresets(nil, ownerField)
}

func (s *EndUserPermissionAppService) getPermissionModelNames(
	ctx context.Context,
	modelID string,
) (*string, *string, error) {
	model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
		}
		return nil, nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if model == nil {
		return nil, nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
	}
	return optionalStringPtr(model.DatabaseName), optionalStringPtr(model.ModelName), nil
}

func (s *EndUserPermissionAppService) resolveDesiredPresetsWithModel(
	ctx context.Context,
	cmd ApplyPresetPolicyCommand,
) ([]rbacdomain.PermissionPreset, string, *string, *string, error) {
	model, err := s.modelRepo.GetByID(ctx, cmd.ModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, "", nil, nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, cmd.ModelID)
		}
		return nil, "", nil, nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if model == nil {
		return nil, "", nil, nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, cmd.ModelID)
	}

	ownerField := ""
	if model.GetOwnerField() != nil {
		ownerField = model.GetOwnerField().Name
	}
	desired, err := resolveDesiredPresets(cmd.Preset, ownerField)
	if err != nil {
		return nil, "", nil, nil, err
	}
	return desired, ownerField, optionalStringPtr(model.DatabaseName), optionalStringPtr(model.ModelName), nil
}

func (s *EndUserPermissionAppService) listExistingPresetByModel(
	ctx context.Context,
	repo permissionRepository,
	orgName, modelID string,
) (map[rbacdomain.PermissionPreset]*rbacdomain.EndUserPermission, error) {
	items, err := repo.ListPresetPermissionsByModel(ctx, orgName, modelID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	result := map[rbacdomain.PermissionPreset]*rbacdomain.EndUserPermission{}
	for _, p := range items {
		if p.Type != rbacdomain.PermissionTypePreset || p.Preset == nil {
			continue
		}
		result[*p.Preset] = p
	}
	return result, nil
}

func (s *EndUserPermissionAppService) ensureDeletePresetSafe(
	ctx context.Context,
	repo permissionRepository,
	existing map[rbacdomain.PermissionPreset]*rbacdomain.EndUserPermission,
	desiredSet map[rbacdomain.PermissionPreset]struct{},
) error {
	for preset, existingPerm := range existing {
		if _, ok := desiredSet[preset]; ok {
			continue
		}
		referenced, err := repo.IsPermissionReferencedByBundle(ctx, existingPerm.ID)
		if err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if referenced {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.PresetDeleteBlockedByBundle, existingPerm.Name)
		}
	}
	return nil
}

func (s *EndUserPermissionAppService) upsertDesiredPresets(
	ctx context.Context,
	repo permissionRepository,
	cmd ApplyPresetPolicyCommand,
	desired []rbacdomain.PermissionPreset,
	ownerField string,
	databaseName, modelName *string,
	existing map[rbacdomain.PermissionPreset]*rbacdomain.EndUserPermission,
) error {
	for _, preset := range desired {
		rowPolicy, err := expandPreset(preset, ownerField)
		if err != nil {
			return err
		}
		if oldPerm, ok := existing[preset]; ok {
			if !presetPermissionNeedsUpdate(oldPerm, rowPolicy, preset) {
				continue
			}
			oldPerm.Name = presetPermissionName(preset)
			oldPerm.Type = rbacdomain.PermissionTypePreset
			oldPerm.ColumnPolicy = nil
			oldPerm.RowPolicy = rowPolicy
			nextPreset := preset
			oldPerm.Preset = &nextPreset
			if err := oldPerm.Validate(); err != nil {
				return err
			}
			if err := repo.UpdatePresetPermission(ctx, oldPerm); err != nil {
				return bizerrors.ConvertRepositoryError(ctx, err)
			}
			continue
		}
		newPerm := &rbacdomain.EndUserPermission{
			ID:           uuid.NewString(),
			OrgName:      cmd.OrgName,
			ProjectSlug:  cmd.ProjectSlug,
			ModelID:      cmd.ModelID,
			DatabaseName: databaseName,
			ModelName:    modelName,
			Name:         presetPermissionName(preset),
			Type:         rbacdomain.PermissionTypePreset,
			ColumnPolicy: nil,
			RowPolicy:    rowPolicy,
		}
		nextPreset := preset
		newPerm.Preset = &nextPreset
		if err := newPerm.Validate(); err != nil {
			return err
		}
		if err := repo.CreatePermission(ctx, newPerm); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
	}
	return nil
}

func (s *EndUserPermissionAppService) deleteStalePresets(
	ctx context.Context,
	repo permissionRepository,
	orgName string,
	existing map[rbacdomain.PermissionPreset]*rbacdomain.EndUserPermission,
	desiredSet map[rbacdomain.PermissionPreset]struct{},
) error {
	for preset, existingPerm := range existing {
		if _, ok := desiredSet[preset]; ok {
			continue
		}
		if err := repo.DeletePermission(ctx, orgName, existingPerm.ID); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
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

func isOwnerPreset(preset rbacdomain.PermissionPreset) bool {
	return preset == rbacdomain.PresetReadWriteOwner || preset == rbacdomain.PresetReadAllWriteOwner
}

func resolveDesiredPresets(
	explicit *rbacdomain.PermissionPreset,
	ownerField string,
) ([]rbacdomain.PermissionPreset, error) {
	if explicit != nil {
		if !explicit.IsValid() {
			return nil, bizerrors.NewValidationError(
				"rbac.permission.invalid_preset: invalid preset: %s",
				string(*explicit),
			)
		}
		if isOwnerPreset(*explicit) && ownerField == "" {
			return nil, bizerrors.NewError(bizerrors.EndUserPresetRequiresOwnerField, string(*explicit))
		}
		return []rbacdomain.PermissionPreset{*explicit}, nil
	}

	desired := []rbacdomain.PermissionPreset{
		rbacdomain.PresetReadWriteAll,
		rbacdomain.PresetReadAll,
	}
	if ownerField != "" {
		desired = append(desired, rbacdomain.PresetReadWriteOwner, rbacdomain.PresetReadAllWriteOwner)
	}
	return desired, nil
}

func presetPermissionName(preset rbacdomain.PermissionPreset) string {
	return fmt.Sprintf("preset:%s", preset)
}

func presetPermissionNeedsUpdate(
	existing *rbacdomain.EndUserPermission,
	desiredPolicy *rbacdomain.RowPolicy,
	desiredPreset rbacdomain.PermissionPreset,
) bool {
	if existing == nil {
		return true
	}
	if existing.Type != rbacdomain.PermissionTypePreset {
		return true
	}
	if existing.Preset == nil || *existing.Preset != desiredPreset {
		return true
	}
	if existing.Name != presetPermissionName(desiredPreset) {
		return true
	}
	return normalizedRowPolicyJSON(existing.RowPolicy) != normalizedRowPolicyJSON(desiredPolicy)
}

func normalizedRowPolicyJSON(policy *rbacdomain.RowPolicy) string {
	if policy == nil {
		return ""
	}
	clone := *policy
	clone.Normalize()
	data, err := json.Marshal(&clone)
	if err != nil {
		return ""
	}
	return string(data)
}

func expandPreset(preset rbacdomain.PermissionPreset, ownerField string) (*rbacdomain.RowPolicy, error) {
	ownerPredicate := json.RawMessage(fmt.Sprintf(`{"%s":{"_eq":"$endUserId"}}`, ownerField))

	switch preset {
	case rbacdomain.PresetReadWriteAll:
		return &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
			Insert: rbacdomain.InsertPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
			Update: rbacdomain.UpdatePolicy{Allowed: true, Scope: rbacdomain.ScopeAll, CheckScope: rbacdomain.ScopeAll},
			Delete: rbacdomain.DeletePolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
		}, nil
	case rbacdomain.PresetReadAll:
		return &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
			Insert: rbacdomain.InsertPolicy{Allowed: false},
			Update: rbacdomain.UpdatePolicy{Allowed: false},
			Delete: rbacdomain.DeletePolicy{Allowed: false},
		}, nil
	case rbacdomain.PresetReadWriteOwner:
		if ownerField == "" {
			return nil, bizerrors.NewError(bizerrors.EndUserPresetRequiresOwnerField, string(preset))
		}
		return &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{
				Allowed:   true,
				Scope:     rbacdomain.ScopeCustom,
				Predicate: ownerPredicate,
			},
			Insert: rbacdomain.InsertPolicy{
				Allowed: true,
				Scope:   rbacdomain.ScopeCustom,
				Check:   ownerPredicate,
			},
			Update: rbacdomain.UpdatePolicy{
				Allowed:    true,
				Scope:      rbacdomain.ScopeCustom,
				Predicate:  ownerPredicate,
				CheckScope: rbacdomain.ScopeCustom,
				Check:      ownerPredicate,
			},
			Delete: rbacdomain.DeletePolicy{
				Allowed:   true,
				Scope:     rbacdomain.ScopeCustom,
				Predicate: ownerPredicate,
			},
		}, nil
	case rbacdomain.PresetReadAllWriteOwner:
		if ownerField == "" {
			return nil, bizerrors.NewError(bizerrors.EndUserPresetRequiresOwnerField, string(preset))
		}
		return &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{
				Allowed: true,
				Scope:   rbacdomain.ScopeAll,
			},
			Insert: rbacdomain.InsertPolicy{
				Allowed: true,
				Scope:   rbacdomain.ScopeCustom,
				Check:   ownerPredicate,
			},
			Update: rbacdomain.UpdatePolicy{
				Allowed:    true,
				Scope:      rbacdomain.ScopeCustom,
				Predicate:  ownerPredicate,
				CheckScope: rbacdomain.ScopeCustom,
				Check:      ownerPredicate,
			},
			Delete: rbacdomain.DeletePolicy{
				Allowed:   true,
				Scope:     rbacdomain.ScopeCustom,
				Predicate: ownerPredicate,
			},
		}, nil
	default:
		return nil, bizerrors.NewValidationError("rbac.permission.invalid_preset: invalid preset: %s", string(preset))
	}
}

func buildLegacyRowPolicy(action rbacdomain.Action, rowScope rbacdomain.RowScope) *rbacdomain.RowPolicy {
	policy := &rbacdomain.RowPolicy{
		Select: rbacdomain.SelectPolicy{Allowed: false},
		Insert: rbacdomain.InsertPolicy{Allowed: false},
		Update: rbacdomain.UpdatePolicy{Allowed: false},
		Delete: rbacdomain.DeletePolicy{Allowed: false},
	}

	scope, predicate, check := legacyScopeExpr(rowScope)

	switch action {
	case rbacdomain.ActionSelect, rbacdomain.ActionExport:
		policy.Select.Allowed = true
		policy.Select.Scope = scope
		policy.Select.Predicate = predicate
	case rbacdomain.ActionInsert:
		policy.Insert.Allowed = true
		policy.Insert.Scope = scope
		policy.Insert.Check = check
	case rbacdomain.ActionUpdate:
		policy.Update.Allowed = true
		policy.Update.Scope = scope
		policy.Update.Predicate = predicate
		policy.Update.CheckScope = scope
		policy.Update.Check = check
	case rbacdomain.ActionDelete:
		policy.Delete.Allowed = true
		policy.Delete.Scope = scope
		policy.Delete.Predicate = predicate
	}
	policy.Normalize()
	return policy
}

func legacyScopeExpr(rowScope rbacdomain.RowScope) (
	scope rbacdomain.PolicyScope,
	predicate json.RawMessage,
	check json.RawMessage,
) {
	scope = rbacdomain.ScopeAll
	predicate = nil
	check = nil

	switch rowScope {
	case rbacdomain.RowScopeAll:
		return scope, predicate, check
	case rbacdomain.RowScopeSelf:
		scope = rbacdomain.ScopeCustom
		predicate = json.RawMessage(`{"owner":{"_eq":"$endUserId"}}`)
		check = json.RawMessage(`{"owner":{"_eq":"$endUserId"}}`)
	case rbacdomain.RowScopeDept:
		scope = rbacdomain.ScopeCustom
		predicate = json.RawMessage(`{"dept_id":{"_eq":"$deptId"}}`)
		check = json.RawMessage(`{"dept_id":{"_eq":"$deptId"}}`)
	case rbacdomain.RowScopeDeptAndChildren:
		scope = rbacdomain.ScopeCustom
		predicate = json.RawMessage(`{"dept_id":{"_in":"$deptAndChildrenIds"}}`)
		check = json.RawMessage(`{"dept_id":{"_in":"$deptAndChildrenIds"}}`)
	default:
		return rbacdomain.ScopeAll, nil, nil
	}
	return scope, predicate, check
}

func optionalStringPtr(v string) *string {
	if v == "" {
		return nil
	}
	vCopy := v
	return &vCopy
}

// validateRowScopePrerequisite 校验 rowScope 字段前提（兼容旧 CreatePermission 输入）。
func (s *EndUserPermissionAppService) validateRowScopePrerequisite(
	ctx context.Context,
	modelID string,
	rowScope rbacdomain.RowScope,
) error {
	switch rowScope {
	case rbacdomain.RowScopeAll:
		return nil
	case rbacdomain.RowScopeSelf:
		model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
		if err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if model == nil || model.GetOwnerField() == nil {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserRowScopeFieldMissing, "SELF", "owner")
		}
	case rbacdomain.RowScopeDept, rbacdomain.RowScopeDeptAndChildren:
		model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
		if err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if model == nil || !model.HasField("dept_id") {
			return bizerrors.NewErrorFromContext(
				ctx,
				bizerrors.EndUserRowScopeFieldMissing,
				string(rowScope),
				"dept_id",
			)
		}
	}
	return nil
}
