package rls

import "context"

// PolicyCompiler RLS 表达式编译器
type PolicyCompiler interface {
	// Compile 将 JSON 表达式编译为 CompiledPolicy
	// 递归解析 JSON 为参数化 SQL 片段
	// {{user_id}} / {{user_name}} → 从 UserContext 查值 → ? 占位符
	// _auth.uid → ? 占位符（兼容旧写法）
	// _ref → 跨表字段引用（仅 PREDICATE 允许）
	Compile(ctx context.Context, expr JsonExpr, userCtx *UserContext) (*CompiledPolicy, error)
}

// CompiledPolicy 编译后的策略
type CompiledPolicy struct {
	SQL    string        // 参数化 SQL 片段
	Params []interface{} // 参数占位符说明（如 {"_auth": "uid"}）
}
