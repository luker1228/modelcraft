package modelruntime

// ExecuteGraphQLCommand 执行GraphQL查询命令
type ExecuteGraphQLCommand struct {
	Query         string                 // GraphQL查询字符串
	Variables     map[string]interface{} // 查询变量
	OperationName string                 // 操作名称
}
