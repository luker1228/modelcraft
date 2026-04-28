package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	rbacdomain "modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"

	"github.com/google/uuid"
)

type permissionRepository interface {
	CreatePermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	GetPermissionByID(ctx context.Context, orgName, id string) (*rbacdomain.EndUserPermission, error)
	ListPermissionsByProject(ctx context.Context, orgName, projectSlug string) ([]*rbacdomain.EndUserPermission, error)
	ListPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*rbacdomain.EndUserPermission, error)
	UpdatePermission(ctx context.Context, p *rbacdomain.EndUserPermission) error
	DeletePermission(ctx context.Context, orgName, id string) error
	DeletePresetPermissionsByModel(ctx context.Context, orgName, modelID string) error
}

type modelRepository interface {
	GetByID(ctx context.Context, id string, opts ...*modeldesign.ModelQueryOptions) (*modeldesign.DataModel, error)
}

// EndUserPermissionAppService 权限点应用服务
type EndUserPermissionAppService struct {
	rbacRepo  permissionRepository
	modelRepo modelRepository
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

// CreatePermission 创建权限点（兼容旧输入 action/rowScope，内部转换为 rowPolicy）
func (s *EndUserPermissionAppService) CreatePermission(
	ctx context.Context,
	cmd CreatePermissionCommand,
) (*rbacdomain.EndUserPermission, error) {
	if err := s.validateRowScopePrerequisite(ctx, cmd.ModelID, cmd.RowScope); err != nil {
		return nil, err
	}

	rowPolicy := buildLegacyRowPolicy(cmd.Action, cmd.RowScope)
	perm := &rbacdomain.EndUserPermission{
		ID:           uuid.NewString(),
		OrgName:      cmd.OrgName,
		ProjectSlug:  cmd.ProjectSlug,
		ModelID:      cmd.ModelID,
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

// ApplyPresetPolicy 应用预设策略（先删旧 PRESET，再插入新 PRESET）
func (s *EndUserPermissionAppService) ApplyPresetPolicy(
	ctx context.Context,
	cmd ApplyPresetPolicyCommand,
) ([]*rbacdomain.EndUserPermission, error) {
	if !cmd.Preset.IsValid() {
		return nil, bizerrors.NewValidationError(
			"rbac.permission.invalid_preset: invalid preset: %s",
			string(cmd.Preset),
		)
	}

	model, err := s.modelRepo.GetByID(ctx, cmd.ModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, cmd.ModelID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	ownerField := ""
	if model != nil && model.GetOwnerField() != nil {
		ownerField = model.GetOwnerField().Name
	}

	if isOwnerPreset(cmd.Preset) && ownerField == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPresetRequiresOwnerField, string(cmd.Preset))
	}

	if err := s.rbacRepo.DeletePresetPermissionsByModel(ctx, cmd.OrgName, cmd.ModelID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	logfacade.GetLogger(ctx).Warnf(
		ctx,
		"ApplyPresetPolicy: deleted PRESET permissions; FK CASCADE may remove bundle mappings (org=%s model=%s)",
		cmd.OrgName,
		cmd.ModelID,
	)

	rowPolicy, err := expandPreset(cmd.Preset, ownerField)
	if err != nil {
		return nil, err
	}

	preset := cmd.Preset
	perm := &rbacdomain.EndUserPermission{
		ID:           uuid.NewString(),
		OrgName:      cmd.OrgName,
		ProjectSlug:  cmd.ProjectSlug,
		ModelID:      cmd.ModelID,
		Name:         fmt.Sprintf("preset:%s", cmd.Preset),
		Type:         rbacdomain.PermissionTypePreset,
		ColumnPolicy: nil,
		RowPolicy:    rowPolicy,
		Preset:       &preset,
	}
	if err := perm.Validate(); err != nil {
		return nil, err
	}
	if err := s.rbacRepo.CreatePermission(ctx, perm); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	perms, err := s.rbacRepo.ListPermissionsByModel(ctx, cmd.OrgName, cmd.ModelID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return perms, nil
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
