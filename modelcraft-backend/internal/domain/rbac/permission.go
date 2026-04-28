package rbac

import "modelcraft/pkg/bizerrors"

// Action 操作动作枚举（用于鉴权结果模型）
type Action string

const (
	ActionSelect Action = "select"
	ActionInsert Action = "insert"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionExport Action = "export"
)

// validActions 合法 Action 集合（用于 IsValid 校验）
var validActions = map[Action]struct{}{
	ActionSelect: {},
	ActionInsert: {},
	ActionUpdate: {},
	ActionDelete: {},
	ActionExport: {},
}

// IsValid 判断 Action 枚举值是否合法
func (a Action) IsValid() bool {
	_, ok := validActions[a]
	return ok
}

// PermissionType 权限点来源类型
type PermissionType string

const (
	PermissionTypePreset PermissionType = "PRESET"
	PermissionTypeCustom PermissionType = "CUSTOM"
)

var validPermissionTypes = map[PermissionType]struct{}{
	PermissionTypePreset: {},
	PermissionTypeCustom: {},
}

// IsValid 判断 PermissionType 是否合法
func (t PermissionType) IsValid() bool {
	_, ok := validPermissionTypes[t]
	return ok
}

// PermissionPreset 预设策略枚举
type PermissionPreset string

const (
	PresetReadWriteAll      PermissionPreset = "READ_WRITE_ALL"
	PresetReadAll           PermissionPreset = "READ_ALL"
	PresetReadWriteOwner    PermissionPreset = "READ_WRITE_OWNER"
	PresetReadAllWriteOwner PermissionPreset = "READ_ALL_WRITE_OWNER"
)

var validPermissionPresets = map[PermissionPreset]struct{}{
	PresetReadWriteAll:      {},
	PresetReadAll:           {},
	PresetReadWriteOwner:    {},
	PresetReadAllWriteOwner: {},
}

// IsValid 判断预设值是否合法
func (p PermissionPreset) IsValid() bool {
	_, ok := validPermissionPresets[p]
	return ok
}

// EndUserPermission 权限点（最小权限定义单元）
// 以 rowPolicy 表达四类动作的允许性与数据范围。
type EndUserPermission struct {
	OrgName     string
	ProjectSlug string
	ID          string
	ModelID     string
	Name        string
	Description *string

	Type PermissionType
	// ColumnPolicy 为 nil 表示全列默认（DefaultMode=VISIBLE，无 Rules）
	ColumnPolicy *ColumnPolicy
	RowPolicy    *RowPolicy
	Preset       *PermissionPreset
}

// Validate 校验权限点合法性（纯域内校验，不查 DB）
func (p *EndUserPermission) Validate() error {
	if p.ModelID == "" {
		return bizerrors.NewValidationError("rbac.permission.modelID_required: modelID is required")
	}
	if p.Name == "" {
		return bizerrors.NewValidationError("rbac.permission.name_required: name is required")
	}
	if !p.Type.IsValid() {
		return bizerrors.NewValidationError("rbac.permission.invalid_type: invalid type: %s", string(p.Type))
	}
	if p.Preset != nil && !p.Preset.IsValid() {
		return bizerrors.NewValidationError("rbac.permission.invalid_preset: invalid preset: %s", string(*p.Preset))
	}
	if p.Type == PermissionTypePreset && p.Preset == nil {
		return bizerrors.NewValidationError("rbac.permission.preset_required: preset is required when type is PRESET")
	}
	if p.Type == PermissionTypeCustom && p.Preset != nil {
		return bizerrors.NewValidationError("rbac.permission.preset_forbidden: preset must be nil when type is CUSTOM")
	}
	if p.ColumnPolicy != nil && !p.ColumnPolicy.DefaultMode.IsValid() {
		return bizerrors.NewValidationError(
			"rbac.permission.invalid_column_default_mode: invalid default column access mode")
	}
	if p.RowPolicy == nil {
		return bizerrors.NewValidationError("rbac.permission.row_policy_required: rowPolicy is required")
	}
	p.RowPolicy.Normalize()
	if err := p.RowPolicy.Validate(); err != nil {
		return err
	}
	return nil
}
