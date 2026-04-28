package rbac

// EffectivePermission 有效权限单条记录（Step 5 判定后的产物）
type EffectivePermission struct {
	ModelID      string
	Action       Action
	ColumnPolicy *ColumnPolicy // nil = 全列默认
	RowScope     RowScope      // 兼容旧链路：ALL / SELF
}

// EffectivePermissionSet 用户在某 Project 下的有效权限集合（key = "modelID:action"）
type EffectivePermissionSet map[string]*EffectivePermission

// Merge 将一批权限点合并进有效权限集（Step 5 核心逻辑）。
// rowPolicy 会按 action 展开为 grant，columnPolicy 取更宽泛并集。
func (eps EffectivePermissionSet) Merge(permissions []*EndUserPermission) EffectivePermissionSet {
	for _, p := range permissions {
		if p == nil || p.RowPolicy == nil {
			continue
		}

		p.RowPolicy.Normalize()
		for _, grant := range expandRowPolicyToGrants(p.RowPolicy) {
			key := p.ModelID + ":" + string(grant.Action)
			if existing, ok := eps[key]; ok {
				existing.RowScope = MergeRowScope(existing.RowScope, grant.RowScope)
				existing.ColumnPolicy = mergeColumnPolicy(existing.ColumnPolicy, p.ColumnPolicy)
				continue
			}

			eps[key] = &EffectivePermission{
				ModelID:      p.ModelID,
				Action:       grant.Action,
				ColumnPolicy: p.ColumnPolicy,
				RowScope:     grant.RowScope,
			}
		}
	}
	return eps
}

type actionGrant struct {
	Action   Action
	RowScope RowScope
}

func expandRowPolicyToGrants(policy *RowPolicy) []actionGrant {
	grants := make([]actionGrant, 0, 4)

	if policy.Select.Allowed {
		grants = append(grants, actionGrant{Action: ActionSelect, RowScope: scopeToRowScope(policy.Select.Scope)})
	}
	if policy.Insert.Allowed {
		grants = append(grants, actionGrant{Action: ActionInsert, RowScope: scopeToRowScope(policy.Insert.Scope)})
	}
	if policy.Update.Allowed {
		grants = append(grants, actionGrant{Action: ActionUpdate, RowScope: scopeToRowScope(policy.Update.Scope)})
	}
	if policy.Delete.Allowed {
		grants = append(grants, actionGrant{Action: ActionDelete, RowScope: scopeToRowScope(policy.Delete.Scope)})
	}
	return grants
}

func scopeToRowScope(scope PolicyScope) RowScope {
	if scope == ScopeCustom {
		return RowScopeSelf
	}
	return RowScopeAll
}

// HasPermission 默认拒绝原则：只有命中 allow 才返回 true
func (eps EffectivePermissionSet) HasPermission(modelID string, action Action) bool {
	_, ok := eps[modelID+":"+string(action)]
	return ok
}

// GetPermission 获取有效权限（用于提取 rowScope 传给 RLS 引擎），不存在返回 nil
func (eps EffectivePermissionSet) GetPermission(modelID string, action Action) *EffectivePermission {
	return eps[modelID+":"+string(action)]
}
