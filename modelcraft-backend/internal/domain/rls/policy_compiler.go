package rls

// CompiledPolicy 编译后的策略
type CompiledPolicy struct {
	SQL    string        // 参数化 SQL 片段
	Params []interface{} // 参数占位符
}
