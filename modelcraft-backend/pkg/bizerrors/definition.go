package bizerrors

// 错误性质常量定义
const (
	// 资源不存在
	ErrorTypeNotFound = "NOT_FOUND"

	// 参数无效（输入验证失败）
	ErrorTypeParamInvalid = "PARAM_INVALID"

	// 操作拒绝（权限不足、状态不允许、不可编辑、不可删除等）
	ErrorTypeOperationFailed = "OPERATION_FAILED"

	// 资源冲突（重复创建、并发冲突等）
	ErrorTypeConflict = "CONFLICT"

	// 认证失败（登录失败，token过期等)
	ErrorTypeAuthentication = "AUTHENTICATION_FAILED"

	// 未授权访问（缺少必要的认证信息或上下文）
	ErrorTypeUnauthorized = "UNAUTHORIZED"

	// 系统错误（技术问题）
	ErrorTypeSystemError = "SYSTEM_ERROR"
)

// ErrorDefinition 定义错误的基本信息，包括错误码和多语言消息模板
type ErrorDefinition struct {
	Code      string // 错误码
	EnMessage string // 英文消息模板
	ZhMessage string // 中文消息模板
}

// GetCode 返回错误定义的错误代码字符串
func (e ErrorDefinition) GetCode() string {
	return e.Code
}

// GetMessageTemplate 返回指定语言的错误消息模板，如果语言不存在则返回默认语言模板
func (e ErrorDefinition) GetMessageTemplate(lang string) string {
	if lang == LangEN {
		if e.EnMessage != "" {
			return e.EnMessage
		}
		// 如果英文消息为空，返回错误代码本身
		return e.Code
	} else {
		// 默认返回中文
		if e.ZhMessage != "" {
			return e.ZhMessage
		}
		// 如果中文消息为空，返回英文
		return e.EnMessage
	}
}

// String 实现fmt.Stringer接口，返回错误定义的字符串表示形式
func (e ErrorDefinition) String() string {
	return e.Code
}

// GetErrorType 返回错误的类型分类，用于判断错误的性质
func (e ErrorDefinition) GetErrorType() string {
	code := e.Code
	for i, char := range code {
		if char == '.' {
			return code[:i]
		}
	}
	return code
}

// IsNotFoundError 判断错误定义是否为资源不存在类型的错误
func (e ErrorDefinition) IsNotFoundError() bool {
	return e.GetErrorType() == ErrorTypeNotFound
}

// IsParamInvalidError 判断错误定义是否为参数无效类型的错误
func (e ErrorDefinition) IsParamInvalidError() bool {
	return e.GetErrorType() == ErrorTypeParamInvalid
}

// IsOperationDeniedError 判断错误定义是否为操作拒绝类型的错误
func (e ErrorDefinition) IsOperationDeniedError() bool {
	return e.GetErrorType() == ErrorTypeOperationFailed
}

// IsConflictError 判断错误定义是否为资源冲突类型的错误
func (e ErrorDefinition) IsConflictError() bool {
	return e.GetErrorType() == ErrorTypeConflict
}

// IsSystemErrorNew 判断错误定义是否为系统内部错误类型
func (e ErrorDefinition) IsSystemErrorNew() bool {
	return e.GetErrorType() == ErrorTypeSystemError
}
