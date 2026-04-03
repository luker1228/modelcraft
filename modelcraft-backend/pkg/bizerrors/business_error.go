package bizerrors

import (
	"context"
	"fmt"
	"modelcraft/pkg/ctxutils"
)

// BusinessError 表示业务错误实例，包含错误定义和上下文信息
type BusinessError struct {
	info         ErrorDefinition // 错误定义 (包含错误代码和消息模板)
	msg          string          // 完整的错误消息（创建时生成）
	detail       string          // 自定义详情，用于附加额外的错误信息
	requestId    string          // 链路追踪ID
	language     string          // 消息语言
	wrappedError error           // 包装的原始错误，不对外暴露
}

// Error 实现error接口，返回错误的字符串表示形式
func (e *BusinessError) Error() string {
	result := fmt.Sprintf("[%s] %s", e.info.GetCode(), e.msg)

	if e.wrappedError != nil {
		result = fmt.Sprintf("%s: detail:%v", result, e.wrappedError)
	}

	return result
}

// Info 返回错误定义
func (e *BusinessError) Info() ErrorDefinition {
	return e.info
}

// Msg 返回错误消息
func (e *BusinessError) Msg() string {
	return e.msg
}

// Detail 返回错误详情
func (e *BusinessError) Detail() string {
	return e.detail
}

// Language 返回消息语言
func (e *BusinessError) Language() string {
	return e.language
}

// RequestId 返回请求ID
func (e *BusinessError) RequestId() string {
	return e.requestId
}

// Unwrap 实现错误链解包接口，返回被包装的原始错误
func (e *BusinessError) Unwrap() error {
	return e.wrappedError
}

// WrapError 包装现有错误并保留堆栈信息，用于错误链传递
// params 用于填充错误消息模板中的占位符
func WrapError(err error, code ErrorDefinition, params ...any) *BusinessError {
	if err == nil {
		return nil
	}

	// 如果已经是BusinessError，避免重复包装
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr
	}

	bizErr := NewError(code, params...)
	bizErr.wrappedError = err
	return bizErr
}

// GetHTTPStatusCode 根据错误定义返回对应的HTTP状态码
func (e *BusinessError) GetHTTPStatusCode() int {
	errorType := e.info.GetErrorType()
	switch errorType {
	case ErrorTypeNotFound:
		return 404 // Not Found
	case ErrorTypeParamInvalid:
		return 400 // Bad Request
	case ErrorTypeOperationFailed:
		return 403 // Forbidden
	case ErrorTypeConflict:
		return 409 // Conflict
	case ErrorTypeSystemError:
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}

// NewError 创建新的业务错误实例
// params 用于填充错误消息模板中的占位符，消息在创建时生成
func NewError(code ErrorDefinition, params ...any) *BusinessError {
	lang := LangEN
	msg := GetMessageWithParams(code, lang, params)

	return &BusinessError{
		info:     code,
		msg:      msg,
		language: lang,
	}
}

// NewErrorFromContext 从上下文创建业务错误，自动提取请求ID和语言设置
// params 用于填充错误消息模板中的占位符，消息在创建时生成
func NewErrorFromContext(ctx context.Context, errorInfo ErrorDefinition, params ...any) *BusinessError {
	contextV := ctxutils.FromContext(ctx)

	// Handle case where context value is nil (e.g., in tests or non-HTTP contexts)
	var lang string
	var requestId string
	if contextV != nil {
		lang = contextV.Lang
		requestId = contextV.RequestId
	} else {
		lang = "" // Will default to English in GetMessageWithParams
		requestId = ""
	}

	msg := GetMessageWithParams(errorInfo, lang, params)

	return &BusinessError{
		info:      errorInfo,
		msg:       msg,
		language:  lang,
		requestId: requestId,
	}
}

// IsNotFoundError 判断业务错误是否为资源不存在类型
func (e *BusinessError) IsNotFoundError() bool {
	return e.info.IsNotFoundError()
}

// IsParamInvalidError 判断业务错误是否为参数无效类型
func (e *BusinessError) IsParamInvalidError() bool {
	return e.info.IsParamInvalidError()
}

// IsOperationDeniedError 判断业务错误是否为操作拒绝类型
func (e *BusinessError) IsOperationDeniedError() bool {
	return e.info.IsOperationDeniedError()
}

// IsConflictError 判断业务错误是否为资源冲突类型
func (e *BusinessError) IsConflictError() bool {
	return e.info.IsConflictError()
}

// IsSystemErrorNew 判断业务错误是否为系统内部错误类型
func (e *BusinessError) IsSystemErrorNew() bool {
	return e.info.IsSystemErrorNew()
}

// ConvertRepositoryError 将Repository层错误统一转换为系统错误，便于上层处理
func ConvertRepositoryError(ctx context.Context, repoErr error) *BusinessError {
	// 如果不是RepositoryError类型，直接包装为系统错误
	if repoErr == nil {
		return nil
	}

	// 构建详细的错误描述，包含技术信息
	errorDetail := fmt.Sprintf("Repository operation failed: %s", repoErr.Error())

	// 统一创建系统错误
	bizErr := NewErrorFromContext(ctx, SystemError, errorDetail)

	// 保留原始Repository错误链，便于调试和日志记录
	bizErr.wrappedError = repoErr

	return bizErr
}
