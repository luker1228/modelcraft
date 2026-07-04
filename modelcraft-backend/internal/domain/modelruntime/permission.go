package modelruntime

import (
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
