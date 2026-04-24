package rbac

// EffectivePermission 有效权限单条记录（Step 5 判定后的产物）
type EffectivePermission struct {
	ModelID      string
	Action       Action
	ColumnPolicy *ColumnPolicy // nil = 全列默认
	RowScope     RowScope      // 已取最宽泛范围（多来源合并后）
}

// EffectivePermissionSet 用户在某 Project 下的有效权限集合（key = "modelID:action"）
type EffectivePermissionSet map[string]*EffectivePermission

// Merge 将一批权限点合并进有效权限集（Step 5 核心逻辑）
// rowScope 取并集最大范围，columnPolicy 取更宽泛的模式
func (eps EffectivePermissionSet) Merge(permissions []*EndUserPermission) EffectivePermissionSet {
	for _, p := range permissions {
		key := p.ModelID + ":" + string(p.Action)
		if existing, ok := eps[key]; ok {
			// 行策略取最宽泛（ALL > DEPT_AND_CHILDREN > DEPT > SELF）
			existing.RowScope = MergeRowScope(existing.RowScope, p.RowScope)
			// 列策略取并集（DefaultMode 取更宽泛，Rules 中同字段取更宽泛）
			existing.ColumnPolicy = mergeColumnPolicy(existing.ColumnPolicy, p.ColumnPolicy)
			continue
		}

		eps[key] = &EffectivePermission{
			ModelID:      p.ModelID,
			Action:       p.Action,
			ColumnPolicy: p.ColumnPolicy,
			RowScope:     p.RowScope,
		}
	}
	return eps
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
