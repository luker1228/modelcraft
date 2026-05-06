package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
)

// Action 操作类型。在 domain/modelruntime 内独立定义，避免循环依赖 domain/rbac。
// 值与 rbac.Action 对应（均为小写），app 层负责转换。
type Action string

const (
	ActionSelect Action = "select"
	ActionInsert Action = "insert"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// ActionPermission 单个操作的权限状态。
type ActionPermission struct {
	Allowed bool
	// IsSelf true 表示 rowScope=SELF：查询/更新/删除时需注入
	// WHERE <EndUserRef字段> = $currentEndUserID
	IsSelf bool
}

// ResolvedModelPermissions 单次请求的权限快照。
// 在 Execute() 入口解析一次，注入 graphqlRequestContext，resolver 只读。
// nil 表示 tenant admin 请求，跳过所有检查。
type ResolvedModelPermissions struct {
	Select ActionPermission
	Insert ActionPermission
	Update ActionPermission
	Delete ActionPermission
}

// Get 返回指定 action 的权限状态。
// nil receiver (tenant admin) returns Allowed=true for known actions, false for unknown.
func (p *ResolvedModelPermissions) Get(action Action) ActionPermission {
	switch action {
	case ActionSelect:
		if p != nil {
			return p.Select
		}
	case ActionInsert:
		if p != nil {
			return p.Insert
		}
	case ActionUpdate:
		if p != nil {
			return p.Update
		}
	case ActionDelete:
		if p != nil {
			return p.Delete
		}
	default:
		return ActionPermission{Allowed: false} // unknown action always denied
	}
	return ActionPermission{Allowed: true} // nil receiver (tenant admin) + known action
}

// CheckAction 默认拒绝原则。nil receiver = tenant admin，直接放行。
func (p *ResolvedModelPermissions) CheckAction(action Action) error {
	if p == nil {
		return nil // tenant admin: skip all checks
	}
	if !p.Get(action).Allowed {
		return bizerrors.NewError(bizerrors.PermissionDenied, string(action))
	}
	return nil
}

// EndUserPermissionService 依赖倒置接口，app 层实现。
// domain/modelruntime 只依赖此接口，不感知 domain/rbac 包细节。
type EndUserPermissionService interface {
	// Resolve 查询并合并 end-user 在指定 model 上的有效权限。
	// endUserID 为空（tenant admin）时返回 nil, nil。
	Resolve(ctx context.Context, orgName, projectSlug, endUserID, modelID string) (*ResolvedModelPermissions, error)
}

// FindEndUserRefFieldName 在 RuntimeModel Fields 中找到 FormatEndUserRef 类型字段的名称。
// 若无此字段，返回空字符串（SELF scope 降级为 ALL，不注入 WHERE）。
func FindEndUserRefFieldName(fields map[string]*RuntimeField) string {
	for _, f := range fields {
		if f.Type != nil && f.Type.Format == modeldesign.FormatEndUserRef {
			return f.Name
		}
	}
	return ""
}

// BuildRowFilter 根据权限和 action 构造 WHERE 注入 map。
// 返回 nil 表示无需注入（IsSelf=false 或 ownerField/endUserID 为空）。
func BuildRowFilter(
	perms *ResolvedModelPermissions,
	action Action,
	ownerField string,
	endUserID string,
) map[string]any {
	if perms == nil || ownerField == "" || endUserID == "" {
		return nil
	}
	if !perms.Get(action).IsSelf {
		return nil
	}
	return map[string]any{ownerField: endUserID}
}
