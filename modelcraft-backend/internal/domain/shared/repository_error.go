package shared

import (
	"errors"
	"fmt"
)

// ============================================================================
// Sentinel Errors - 简单、符合 Go 惯例，支持 errors.Is() 检查
// ============================================================================

var (
	// ErrRecordNotFound 记录未找到的 sentinel error
	// 使用方式: errors.Is(err, shared.ErrRecordNotFound)
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateKey 重复键的 sentinel error
	// 使用方式: errors.Is(err, shared.ErrDuplicateKey)
	ErrDuplicateKey = errors.New("duplicate key")
)

// ============================================================================
// Repository Error Types - 用于结构化错误分类
// ============================================================================

// RepositoryErrorType Repository层技术错误类型（仅用于记录和分类）
type RepositoryErrorType string

const (
	// 数据库连接相关
	ErrTypeConnection  RepositoryErrorType = "CONNECTION"  // 连接失败
	ErrTypeTimeout     RepositoryErrorType = "TIMEOUT"     // 操作超时
	ErrTypeTransaction RepositoryErrorType = "TRANSACTION" // 事务错误

	// 数据操作相关
	ErrTypeConstraint     RepositoryErrorType = "CONSTRAINT"       // 约束违反
	ErrTypeConversion     RepositoryErrorType = "CONVERSION"       // 数据转换失败
	ErrTypeSQLConvertion  RepositoryErrorType = "SQL_CONVERSION"   // SQL转换失败
	ErrTypeDDL            RepositoryErrorType = "DDL_SYNTAX_ERR"   // DDL语法错误
	ErrTypeNoRowsAffected RepositoryErrorType = "NO_ROWS_AFFECTED" // 影响数据为0
	ErrTypeDML            RepositoryErrorType = "DML_SYNTAX_ERR"   // DML语法错误
	ErrTypeNotFound       RepositoryErrorType = "NOT_FOUND"
	ErrTypeDuplicatedKey  RepositoryErrorType = "DUPLICATED_KEY"

	// 系统相关
	ErrTypePermission RepositoryErrorType = "PERMISSION" // 权限不足
	ErrTypeUnknown    RepositoryErrorType = "UNKNOWN"    // 未知错误

)

// RepositoryError Repository层错误，专注于信息记录
type RepositoryError struct {
	Type    RepositoryErrorType `json:"type"`    // 技术错误类型
	Message string              `json:"message"` // 技术错误描述
	Cause   error               `json:"-"`       // 原始错误（不序列化）
}

// IsRepoError 判断错误是否为指定类型的Repository错误
func IsRepoError(err error, errorType RepositoryErrorType) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Type == errorType
	}
	return false
}

// Error 实现error接口
func (e *RepositoryError) Error() string {
	return fmt.Sprintf("[REPO_%s] %s", e.Type, e.Message)
}

// Unwrap 实现错误链
func (e *RepositoryError) Unwrap() error {
	return e.Cause
}

// NewRepositoryError 创建新的Repository错误
func NewRepositoryError(errType RepositoryErrorType, message string) *RepositoryError {
	return &RepositoryError{
		Type:    errType,
		Message: message,
	}
}

// WithCause 设置原始错误
func (e *RepositoryError) WithCause(cause error) *RepositoryError {
	e.Cause = cause
	return e
}

// ============================================================================
// Convenience Constructors - 常用错误的便捷构造函数
// ============================================================================

// NewNotFoundError 创建 NotFound 类型的 Repository 错误
// 自动 wrap ErrRecordNotFound sentinel error，支持 errors.Is() 检查
func NewNotFoundError(message string) *RepositoryError {
	return &RepositoryError{
		Type:    ErrTypeNotFound,
		Message: message,
		Cause:   ErrRecordNotFound,
	}
}

// NewDuplicateKeyError 创建 DuplicatedKey 类型的 Repository 错误
// 自动 wrap ErrDuplicateKey sentinel error，支持 errors.Is() 检查
func NewDuplicateKeyError(message string) *RepositoryError {
	return &RepositoryError{
		Type:    ErrTypeDuplicatedKey,
		Message: message,
		Cause:   ErrDuplicateKey,
	}
}

// IsNotFoundError 便捷函数：检查错误是否为 NotFound 类型
// 支持两种检查方式：sentinel error 和 RepositoryError 类型
func IsNotFoundError(err error) bool {
	if errors.Is(err, ErrRecordNotFound) {
		return true
	}
	return IsRepoError(err, ErrTypeNotFound)
}

// IsDuplicateKeyError 便捷函数：检查错误是否为 DuplicateKey 类型
// 支持两种检查方式：sentinel error 和 RepositoryError 类型
func IsDuplicateKeyError(err error) bool {
	if errors.Is(err, ErrDuplicateKey) {
		return true
	}
	return IsRepoError(err, ErrTypeDuplicatedKey)
}
