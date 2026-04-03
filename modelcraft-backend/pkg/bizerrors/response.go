package bizerrors

import (
	"time"
)

// ErrorResponse 统一的错误响应结构，用于API返回错误信息和请求上下文
type ErrorResponse struct {
	Success   bool                   `json:"success"`            // 请求是否成功
	Code      string                 `json:"code"`               // 错误代码 (A.B格式)
	Message   string                 `json:"message"`            // 本地化错误消息
	Details   string                 `json:"details,omitempty"`  // 详细描述
	Context   map[string]interface{} `json:"context,omitempty"`  // 错误上下文
	Timestamp time.Time              `json:"timestamp"`          // 错误发生时间
	TraceID   string                 `json:"trace_id,omitempty"` // 链路追踪ID
	Path      string                 `json:"path,omitempty"`     // 请求路径
	Method    string                 `json:"method,omitempty"`   // 请求方法
}

// SuccessResponse 统一的成功响应结构，包含响应数据和元信息
type SuccessResponse struct {
	Success   bool        `json:"success"`              // 请求是否成功
	Message   string      `json:"message"`              // 成功消息
	Data      interface{} `json:"data,omitempty"`       // 响应数据
	Timestamp time.Time   `json:"timestamp"`            // 响应时间
	RequestId string      `json:"request_id,omitempty"` // 链路追踪ID
}

// ValidationErrorResponse 验证错误的特殊响应格式，包含字段级错误详情列表
type ValidationErrorResponse struct {
	*ErrorResponse
	Errors []FieldError `json:"errors,omitempty"` // 字段错误列表
}

// FieldError 表示单个字段的验证错误信息，包含字段名和错误消息
type FieldError struct {
	Field   string      `json:"field"`           // 字段名
	Value   interface{} `json:"value,omitempty"` // 字段值
	Code    string      `json:"code"`            // 错误代码
	Message string      `json:"message"`         // 错误消息
}
