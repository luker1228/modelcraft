package bizerrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode(t *testing.T) {
	// 测试错误代码的类型
	code := ParamInvalid

	assert.Equal(t, "PARAM_INVALID", code.GetErrorType(), "Expected error type 'PARAM_INVALID'")

	// 测试错误类型判断
	assert.True(t, code.IsParamInvalidError(), "Expected ParamInvalid to be a param invalid error")
	assert.False(t, code.IsSystemErrorNew(), "Expected ParamInvalid not to be a system error")
}

func TestErrorMsg(t *testing.T) {
	code := ParamInvalid

	bizerr := NewError(code, "value")
	fmt.Println(bizerr.Error())
}

func TestBusinessError(t *testing.T) {
	// 创建业务错误
	err := NewError(NotFound, "resource_type", "User")

	// 测试错误信息
	assert.Contains(t, err.Error(), "NOT_FOUND")
	assert.Contains(t, err.Error(), "Resource not found")
}

func TestBusinessErrorWithParams(t *testing.T) {
	// 测试带参数的错误
	err := NewError(ParamInvalid, "username")

	// 验证消息在创建时已生成
	assert.Contains(t, err.Msg(), "username")

	// 验证Error()方法返回的消息包含参数
	assert.Contains(t, err.Error(), "username")

	// 验证Msg()也返回正确的消息
	message := err.Msg()
	assert.Contains(t, message, "username")
}

func TestWrapError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapError(originalErr, SystemError, "wrapped message")

	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "System error")
	assert.Contains(t, wrappedErr.Error(), "original error")
}

func TestBusinessErrorLanguage(t *testing.T) {
	// 创建错误时使用默认语言（英文）
	err := NewError(ParamInvalid, "test")

	// 验证消息在创建时已生成为英文
	assert.Contains(t, err.Msg(), "Invalid parameter")
	assert.Contains(t, err.Msg(), "test")

	// Msg() 现在直接返回创建时生成的消息
	message := err.Msg()
	assert.Contains(t, message, "Invalid parameter")
}

func TestHTTPStatusCodes(t *testing.T) {
	// 测试各种错误类型的HTTP状态码
	testCases := []struct {
		code     ErrorDefinition
		expected int
	}{
		{ParamInvalid, 400},
		{NotFound, 404},
		{OperationDenied, 403},
		{Conflict, 409},
		{SystemError, 500},
	}

	for _, tc := range testCases {
		t.Run(tc.code.GetCode(), func(t *testing.T) {
			err := NewError(tc.code)
			statusCode := err.GetHTTPStatusCode()

			assert.Equal(t, tc.expected, statusCode,
				fmt.Sprintf("Expected status code %d for %s, got %d",
					tc.expected, tc.code.GetCode(), statusCode))
		})
	}
}

func TestErrorTypeJudgment(t *testing.T) {
	// 测试错误类型判断方法
	testCases := []struct {
		code     ErrorDefinition
		testFunc func(ErrorDefinition) bool
		expected bool
	}{
		{NotFound, func(e ErrorDefinition) bool { return e.IsNotFoundError() }, true},
		{ParamInvalid, func(e ErrorDefinition) bool { return e.IsParamInvalidError() }, true},
		{OperationDenied, func(e ErrorDefinition) bool { return e.IsOperationDeniedError() }, true},
		{Conflict, func(e ErrorDefinition) bool { return e.IsConflictError() }, true},
		{SystemError, func(e ErrorDefinition) bool { return e.IsSystemErrorNew() }, true},
	}

	for _, tc := range testCases {
		t.Run(tc.code.GetCode(), func(t *testing.T) {
			result := tc.testFunc(tc.code)
			assert.Equal(t, tc.expected, result,
				fmt.Sprintf("Error type judgment failed for %s", tc.code.GetCode()))
		})
	}
}
