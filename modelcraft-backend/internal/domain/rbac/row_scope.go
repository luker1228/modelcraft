package rbac

// RowScope 行策略枚举
type RowScope string

const (
	RowScopeAll             RowScope = "ALL"
	RowScopeSelf            RowScope = "SELF"
	RowScopeDept            RowScope = "DEPT"
	RowScopeDeptAndChildren RowScope = "DEPT_AND_CHILDREN"
)

// rowScopeOrder 行策略范围宽泛度排序（值越大范围越宽）
var rowScopeOrder = map[RowScope]int{
	RowScopeSelf:            1,
	RowScopeDept:            2,
	RowScopeDeptAndChildren: 3,
	RowScopeAll:             4,
}

// IsValid 判断枚举值是否合法
func (r RowScope) IsValid() bool {
	_, ok := rowScopeOrder[r]
	return ok
}

// MergeRowScope 返回两个行策略中范围更宽泛的那个
// 规则：ALL > DEPT_AND_CHILDREN > DEPT > SELF
func MergeRowScope(a, b RowScope) RowScope {
	if rowScopeOrder[a] >= rowScopeOrder[b] {
		return a
	}
	return b
}
