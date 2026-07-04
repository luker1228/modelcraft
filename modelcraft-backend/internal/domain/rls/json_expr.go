package rls

// JsonExpr RLS 表达式（CEL 字符串）
type JsonExpr string

// ExprType 表达式类型（用于校验时区分 PREDICATE vs CHECK）
type ExprType string

const (
	ExprTypeSelectPredicate ExprType = "SELECT_PREDICATE"
	ExprTypeInsertCheck     ExprType = "INSERT_CHECK"
	ExprTypeUpdatePredicate ExprType = "UPDATE_PREDICATE"
	ExprTypeUpdateCheck     ExprType = "UPDATE_CHECK"
	ExprTypeDeletePredicate ExprType = "DELETE_PREDICATE"
)

// IsPredicate 判断是否为 PREDICATE 类型（USING / row.* 表达式）
func (t ExprType) IsPredicate() bool {
	return t == ExprTypeSelectPredicate ||
		t == ExprTypeUpdatePredicate ||
		t == ExprTypeDeletePredicate
}

// IsCheck 判断是否为 CHECK 类型（input.* 表达式）
func (t ExprType) IsCheck() bool {
	return t == ExprTypeInsertCheck || t == ExprTypeUpdateCheck
}
