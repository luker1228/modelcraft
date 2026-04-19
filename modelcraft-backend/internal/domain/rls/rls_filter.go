package rls

// RLSFilter RLS 运行时过滤器
type RLSFilter struct {
	SelectPredicate JsonExpr `json:"selectPredicate"`
	InsertCheck     JsonExpr `json:"insertCheck"`
	UpdatePredicate JsonExpr `json:"updatePredicate"`
	UpdateCheck     JsonExpr `json:"updateCheck"`
	DeletePredicate JsonExpr `json:"deletePredicate"`
	FieldName       string   `json:"fieldName"` // 固定为 "owner"
	EndUserID       string   `json:"endUserId"`
}

// IsDenyAll 判断是否 DENY ALL（所有谓词都是 false）
func (f *RLSFilter) IsDenyAll() bool {
	return f.SelectPredicate.IsFalse() &&
		f.UpdatePredicate.IsFalse() &&
		f.DeletePredicate.IsFalse()
}

// ShouldInjectWhere 判断是否需要注入 WHERE 条件
func (f *RLSFilter) ShouldInjectWhere() bool {
	// 不是全量 true 且不是全量 false
	return !f.SelectPredicate.IsTrue() && !f.SelectPredicate.IsFalse()
}
