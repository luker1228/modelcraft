package rls

import "encoding/json"

// JsonExpr RLS 表达式（JSON 字符串）
type JsonExpr string

// IsTrue 判断是否为 true 常量
func (e JsonExpr) IsTrue() bool {
	var v interface{}
	if err := json.Unmarshal([]byte(e), &v); err != nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	// 检查 {"_const": true}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err == nil {
		if v, ok := obj["_const"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}

// IsFalse 判断是否为 false 常量
func (e JsonExpr) IsFalse() bool {
	var v interface{}
	if err := json.Unmarshal([]byte(e), &v); err != nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return !b
	}
	// 检查 {"_const": false}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err == nil {
		if v, ok := obj["_const"]; ok {
			if b, ok := v.(bool); ok {
				return !b
			}
		}
	}
	return false
}

// IsOwnerEqualsUser 判断是否为 {"owner":{"_eq":{"_auth":"uid"}}}
func (e JsonExpr) IsOwnerEqualsUser() bool {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err != nil {
		return false
	}
	owner, ok := obj["owner"].(map[string]interface{})
	if !ok {
		return false
	}
	eq, ok := owner["_eq"].(map[string]interface{})
	if !ok {
		return false
	}
	auth, ok := eq["_auth"].(string)
	return ok && auth == "uid"
}

// ExprType 表达式类型（用于校验时区分 PREDICATE vs CHECK）
type ExprType string

const (
	ExprTypeSelectPredicate ExprType = "SELECT_PREDICATE"
	ExprTypeInsertCheck     ExprType = "INSERT_CHECK"
	ExprTypeUpdatePredicate ExprType = "UPDATE_PREDICATE"
	ExprTypeUpdateCheck     ExprType = "UPDATE_CHECK"
	ExprTypeDeletePredicate ExprType = "DELETE_PREDICATE"
)

// IsPredicate 判断是否为 PREDICATE 类型（允许 _exists, _ref）
func (t ExprType) IsPredicate() bool {
	return t == ExprTypeSelectPredicate ||
		t == ExprTypeUpdatePredicate ||
		t == ExprTypeDeletePredicate
}

// IsCheck 判断是否为 CHECK 类型（不允许 _exists, _ref）
func (t ExprType) IsCheck() bool {
	return t == ExprTypeInsertCheck || t == ExprTypeUpdateCheck
}
