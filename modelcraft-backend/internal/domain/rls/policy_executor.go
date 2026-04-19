package rls

import "context"

// PolicyExecutor RLS 策略执行器
type PolicyExecutor interface {
	// ToSQL 将编译后的策略绑定运行时 authCtx 生成最终 SQL
	// 返回参数化查询 + 绑定参数数组
	ToSQL(ctx context.Context, compiled *CompiledPolicy, authCtx *AuthContext) (string, []interface{}, error)

	// ValidateCheck 应用层校验 CHECK 约束（用于 insertCheck / updateCheck）
	// 返回 RLSCheckViolation 错误（如有）
	ValidateCheck(ctx context.Context, expr JsonExpr, rowData map[string]interface{},
		authCtx *AuthContext) error
}

// AuthContext 运行时认证上下文
type AuthContext struct {
	EndUserID string                 `json:"endUserId"`
	Variables map[string]interface{} `json:"variables"` // auth_schema 声明的扩展变量
}
