package repository

import (
	"context"
	"modelcraft/pkg/bizerrors"
)

// NewError 创建一个包装的仓储错误，转换为业务错误
// 这是一个辅助函数，用于在仓储层统一处理数据库错误
func NewError(ctx context.Context, message string, err error) *bizerrors.BusinessError {
	// 使用 bizerrors 的 WrapError 函数统一处理数据库错误
	// 所有数据库错误都转换为系统错误
	if err != nil {
		return bizerrors.WrapError(err, bizerrors.SystemError, message)
	}

	// 如果没有原始错误，直接创建系统错误
	return bizerrors.NewError(bizerrors.SystemError, message)
}
