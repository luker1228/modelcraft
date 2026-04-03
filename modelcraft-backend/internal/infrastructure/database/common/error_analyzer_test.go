package repository

import (
	"database/sql"
	"errors"
	"modelcraft/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeError_Nil(t *testing.T) {
	result := AnalyzeError(nil)
	assert.Nil(t, result)
}

func TestAnalyzeError_AlreadyRepositoryError(t *testing.T) {
	original := shared.NewRepositoryError(shared.ErrTypeNotFound, "already wrapped")
	result := AnalyzeError(original)
	assert.Equal(t, original, result)
}

func TestAnalyzeError_SqlErrNoRows_Type(t *testing.T) {
	result := AnalyzeError(sql.ErrNoRows)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeNotFound, result.Type)
}

func TestAnalyzeError_SqlErrNoRows_CauseIsErrRecordNotFound(t *testing.T) {
	result := AnalyzeError(sql.ErrNoRows)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrRecordNotFound, result.Cause)
}

// 核心测试：验证 errors.Is 可以通过 Unwrap 链找到 ErrRecordNotFound
func TestAnalyzeError_SqlErrNoRows_ErrorsIs(t *testing.T) {
	result := AnalyzeError(sql.ErrNoRows)
	require.NotNil(t, result)
	assert.True(t, errors.Is(result, shared.ErrRecordNotFound),
		"errors.Is(result, shared.ErrRecordNotFound) 应为 true，但实际为 false")
}

func TestWrapDatabaseError_Nil(t *testing.T) {
	result := WrapDatabaseError(nil)
	assert.Nil(t, result)
}

func TestWrapDatabaseError_SqlErrNoRows_ErrorsIs(t *testing.T) {
	result := WrapDatabaseError(sql.ErrNoRows)
	require.NotNil(t, result)
	assert.True(t, errors.Is(result, shared.ErrRecordNotFound),
		"errors.Is(result, shared.ErrRecordNotFound) 应为 true，但实际为 false")
}

func TestWrapDatabaseError_SqlErrNoRows_IsNotFoundHelper(t *testing.T) {
	result := WrapDatabaseError(sql.ErrNoRows)
	require.NotNil(t, result)
	assert.True(t, shared.IsNotFoundError(result),
		"shared.IsNotFoundError 应返回 true")
}

func TestAnalyzeError_ConnectionError(t *testing.T) {
	err := errors.New("connection refused")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeConnection, result.Type)
}

func TestAnalyzeError_TimeoutError(t *testing.T) {
	err := errors.New("operation timeout exceeded")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeTimeout, result.Type)
}

func TestAnalyzeError_DuplicateEntryError(t *testing.T) {
	err := errors.New("Duplicate entry 'foo' for key 'PRIMARY'")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeConstraint, result.Type)
}

func TestAnalyzeError_ForeignKeyError(t *testing.T) {
	err := errors.New("foreign key constraint fails")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeConstraint, result.Type)
}

func TestAnalyzeError_PermissionDenied(t *testing.T) {
	err := errors.New("permission denied for table users")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypePermission, result.Type)
}

func TestAnalyzeError_UnknownError(t *testing.T) {
	err := errors.New("something completely unexpected happened")
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, shared.ErrTypeUnknown, result.Type)
}

func TestAnalyzeError_UnknownError_OriginalErrPreservedInMessage(t *testing.T) {
	msg := "something completely unexpected happened"
	err := errors.New(msg)
	result := AnalyzeError(err)
	require.NotNil(t, result)
	assert.Equal(t, msg, result.Message)
}
