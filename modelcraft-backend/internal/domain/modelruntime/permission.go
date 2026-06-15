package modelruntime

import (
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
}

// ResolvedPolicy 单条匹配的策略摘要。
type ResolvedPolicy struct {
	Name          string // 策略名称（对应 RLS 策略表中的 policyName）
	Action        Action
	UsingExpr     string // 原始 USING 表达式（JSON）
	WithCheckExpr string // 原始 WITH CHECK 表达式（JSON）
}

// ResolvedModelPermissions 单次请求的权限快照（策略列表）。
// 在 Execute() 入口解析一次，注入 graphqlRequestContext，resolver 只读。
// nil 表示 tenant admin 请求，跳过所有检查。
type ResolvedModelPermissions struct {
	Policies []ResolvedPolicy
}

// Get 合并指定 action 的所有策略，返回权限状态。
// nil receiver (tenant admin) returns Allowed=true for known actions, false for unknown.
func (p *ResolvedModelPermissions) Get(action Action) ActionPermission {
	if p == nil {
		return ActionPermission{Allowed: true} // tenant admin
	}
	perm := ActionPermission{}
	for _, pol := range p.Policies {
		if pol.Action == action {
			perm.Allowed = true
		}
	}
	return perm
}

// IsEmpty 返回 true 表示策略列表为空（非 admin 但无任何权限）。
// nil receiver (tenant admin) 返回 false。
func (p *ResolvedModelPermissions) IsEmpty() bool {
	return p != nil && len(p.Policies) == 0
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

// enforceOwnerOnCreate injects the current end-user ID into the END_USER_REF field
// of the create payload.
//
// Caller: executeCreateOne / executeCreateMany.
func enforceOwnerOnCreate(
	rctx *graphqlRequestContext,
	fields map[string]*RuntimeField,
	data map[string]any,
) error {
	ownerID := rctx.resolveEndUserOwnerID()
	if ownerID == "" {
		return nil // tenant admin without admin-claim: use payload as-is
	}
	for _, field := range fields {
		if field.Type == nil || field.Type.Format != modeldesign.FormatEndUserRef {
			continue
		}
		data[field.Name] = ownerID
		return nil
	}
	return nil
}

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
// 返回 nil 表示无需注入（ownerField/endUserID 为空）。
func BuildRowFilter(
	perms *ResolvedModelPermissions,
	action Action,
	ownerField string,
	endUserID string,
) map[string]any {
	if perms == nil || ownerField == "" || endUserID == "" {
		return nil
	}
	if !perms.Get(action).Allowed {
		return nil
	}
	return map[string]any{ownerField: endUserID}
}
