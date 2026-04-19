package rls

import "context"

// ModelSchema RLS 验证所需的模型信息接口
// 由 modeldesign.DataModel 实现，避免直接依赖
// 注：这个接口定义在 rls 包中，以解决循环依赖问题
type ModelSchema interface {
	// HasField 判断模型是否包含指定字段
	HasField(fieldName string) bool
	// GetFieldNames 获取模型所有字段名列表
	GetFieldNames() []string
}

// AuthSchemaProvider 认证变量提供接口
type AuthSchemaProvider interface {
	// IsValidRef 判断变量引用是否合法
	IsValidRef(name string) bool
}

// PolicyValidator RLS 表达式校验器
type PolicyValidator interface {
	// Validate 校验 JSON 表达式合法性
	// - JSON Schema 结构合法性
	// - 字段名白名单（对照 Model 字段列表）
	// - _auth 变量白名单（uid 内置 + auth_schema 声明）
	// - _exists.model 白名单（已知 Model 或系统表）
	// - CHECK 类型不含 _exists / _ref
	Validate(ctx context.Context, expr JsonExpr, exprType ExprType,
		modelSchema ModelSchema, authSchema AuthSchemaProvider) []ValidationError
}

// ValidationError 校验错误
type ValidationError struct {
	Path    string `json:"path"`    // 错误位置，如 "selectPredicate._and[0].owner"
	Message string `json:"message"` // 错误描述
	Code    string `json:"code"`    // 错误码
}
