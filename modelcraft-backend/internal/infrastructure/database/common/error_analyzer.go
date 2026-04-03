package repository

import (
	"database/sql"
	"errors"
	"modelcraft/internal/domain/shared"
	"strings"
)

// AnalyzeError 分析sql错误并创建RepositoryError
func AnalyzeError(err error) *shared.RepositoryError {
	if err == nil {
		return nil
	}

	var repoErr *shared.RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr
	}

	baseError := &shared.RepositoryError{
		Message: err.Error(),
		Cause:   err,
	}

	// 记录不存在 - 包裹 ErrRecordNotFound，调用方可通过 errors.Is(err, shared.ErrRecordNotFound) 判断
	if errors.Is(err, sql.ErrNoRows) {
		baseError.Type = shared.ErrTypeNotFound
		baseError.Cause = shared.ErrRecordNotFound
		return baseError
	}

	// 分析错误字符串，分类记录
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "connection"):
		baseError.Type = shared.ErrTypeConnection

	case strings.Contains(errMsg, "timeout"):
		baseError.Type = shared.ErrTypeTimeout

	case strings.Contains(errMsg, "duplicate entry"):
		baseError.Type = shared.ErrTypeConstraint

	case strings.Contains(errMsg, "foreign key"):
		baseError.Type = shared.ErrTypeConstraint

	case strings.Contains(errMsg, "permission denied"):
		baseError.Type = shared.ErrTypePermission

	default:
		baseError.Type = shared.ErrTypeUnknown
	}

	return baseError
}

// WrapDatabaseError 包装数据库错误
func WrapDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// 分析并创建Repository错误
	repoErr := AnalyzeError(err)

	return repoErr
}

// CreateConversionError 创建数据转换错误
func CreateConversionError(message string) *shared.RepositoryError {
	return shared.NewRepositoryError(
		shared.ErrTypeConversion,
		message,
	)
}
