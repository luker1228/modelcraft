package rbac

import "modelcraft/pkg/bizerrors"

// Action 操作动作枚举
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

// EndUserPermission 权限点（最小权限定义单元）
// 包含四个维度：资源(ModelID) × 动作(Action) × 列策略(ColumnPolicy) × 行策略(RowScope)
type EndUserPermission struct {
	OrgName     string
	ProjectSlug string
	ID          string
	ModelID     string
	Name        string
	Description *string
	Action      Action
	// ColumnPolicy 为 nil 表示全列默认（DefaultMode=VISIBLE，无 Rules）
	ColumnPolicy *ColumnPolicy
	RowScope     RowScope
}

// Validate 校验权限点合法性（纯域内校验，不查 DB）
func (p *EndUserPermission) Validate() error {
	if p.ModelID == "" {
		return bizerrors.NewValidationError("rbac.permission.modelID_required: modelID is required")
	}
	if p.Name == "" {
		return bizerrors.NewValidationError("rbac.permission.name_required: name is required")
	}
	if !p.Action.IsValid() {
		return bizerrors.NewValidationError("rbac.permission.invalid_action: invalid action: %s", string(p.Action))
	}
	if !p.RowScope.IsValid() {
		return bizerrors.NewValidationError("rbac.permission.invalid_row_scope: invalid rowScope: %s", string(p.RowScope))
	}
	if p.ColumnPolicy != nil && !p.ColumnPolicy.DefaultMode.IsValid() {
		return bizerrors.NewValidationError("rbac.permission.invalid_column_default_mode: invalid default column access mode")
	}
	return nil
}
