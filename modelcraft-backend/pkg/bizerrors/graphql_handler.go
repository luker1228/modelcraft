package bizerrors

import (
	"context"
	"modelcraft/pkg/logfacade"
)

// WithGraphqlErrorHandler 统一处理 GraphQL 错误的包装函数
// 它接受一个返回 error 的函数，并统一处理其中的业务错误
// 返回处理后的 GraphQLError 和原始错误
func WithGraphqlErrorHandler(ctx context.Context, fn func() error) error {
	logger := logfacade.GetLogger(ctx)
	err := fn()
	if err != nil {
		// 记录错误日志
		logger.Errorf(ctx, err, "GraphQL operation error")

		// 如果是业务错误，转换为 GraphQL 错误格式
		if bizErr, ok := err.(*BusinessError); ok {
			return NewGraphqlErr(bizErr, bizErr.Info().Code)
		}

		// 对于非业务错误，直接返回原始错误
		return NewGraphqlErr(err, SystemError.Code)
	}
	return nil
}
