package modelruntime

import "context"

// GetGraphqlRequestContextForTest 仅供测试包使用，导出内部 context 访问器。
// 生产代码使用 getGraphqlRequestContext（未导出）。
func GetGraphqlRequestContextForTest(ctx context.Context) (*graphqlRequestContext, bool) {
	return getGraphqlRequestContext(ctx)
}
