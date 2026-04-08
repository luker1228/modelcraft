package sqlerr

import (
	"database/sql"
	"errors"
	"modelcraft/internal/domain/shared"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Error Analysis & Classification
// ----------------------------------------------------------------------------

// errorPattern defines a SQL error message pattern and its corresponding error type.
type errorPattern struct {
	pattern string
	errType shared.RepositoryErrorType
}

// errorPatterns defines SQL error message patterns and their corresponding error types.
// Order matters: more specific patterns should come first.
var errorPatterns = []errorPattern{
	// MySQL specific error codes (more specific, check first)
	{"error 1062", shared.ErrTypeDuplicatedKey}, // Duplicate entry
	{"error 1451", shared.ErrTypeConstraint},    // FK constraint (delete/update parent)
	{"error 1452", shared.ErrTypeConstraint},    // FK constraint (insert/update child)
	{"error 1064", shared.ErrTypeDML},           // SQL syntax error
	{"error 1146", shared.ErrTypeNotFound},      // Table doesn't exist

	// Generic patterns (less specific, check later)
	{"duplicate entry", shared.ErrTypeDuplicatedKey},
	{"unique constraint", shared.ErrTypeDuplicatedKey},
	{"foreign key", shared.ErrTypeConstraint},
	{"connection refused", shared.ErrTypeConnection},
	{"connection reset", shared.ErrTypeConnection},
	{"connection", shared.ErrTypeConnection},
	{"timeout", shared.ErrTypeTimeout},
	{"context deadline exceeded", shared.ErrTypeTimeout},
	{"deadlock", shared.ErrTypeTransaction},
	{"lock wait timeout", shared.ErrTypeTransaction},
	{"permission denied", shared.ErrTypePermission},
	{"access denied", shared.ErrTypePermission},
}

// IsNotFoundError reports whether err represents a "record not found" error.
// This is a convenience wrapper around shared.IsNotFoundError.
func IsNotFoundError(err error) bool {
	return shared.IsNotFoundError(err)
}

// AnalyzeSQLError analyzes database/sql errors and returns a typed RepositoryError.
// If err is already a RepositoryError, it returns unchanged.
func AnalyzeSQLError(err error) error {
	// Fast path: already wrapped
	var repoErr *shared.RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr
	}

	// Handle sql.ErrNoRows specifically (most common case)
	if errors.Is(err, sql.ErrNoRows) {
		return shared.NewNotFoundError(err.Error())
	}

	// Pattern matching for error classification
	errType := classifyError(err)
	return shared.NewRepositoryError(errType, err.Error()).WithCause(err)
}

// classifyError determines the error type based on error message patterns.
func classifyError(err error) shared.RepositoryErrorType {
	lowerMsg := strings.ToLower(err.Error())
	for _, p := range errorPatterns {
		if strings.Contains(lowerMsg, p.pattern) {
			return p.errType
		}
	}
	return shared.ErrTypeUnknown
}

// ----------------------------------------------------------------------------
// Error Wrapping Helpers
// ----------------------------------------------------------------------------

// WrapSQLError wraps a database/sql error as a RepositoryError.
// Returns nil if err is nil (idiomatic Go nil propagation).
func WrapSQLError(err error) error {
	if err == nil {
		return nil
	}
	return AnalyzeSQLError(err)
}

// WrapSQLErrorInPlace wraps err in-place and is suitable for generated wrappers
// that use named return values.
func WrapSQLErrorInPlace(err *error) {
	if err == nil {
		return
	}
	*err = WrapSQLError(*err)
}

// ----------------------------------------------------------------------------
// sql.Null* ↔ Go Type Converters
// ----------------------------------------------------------------------------

// NullStrToPtr converts sql.NullString to *string.
func NullStrToPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

// PtrToNullStr converts *string to sql.NullString.
func PtrToNullStr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullTimeToPtr converts sql.NullTime to *time.Time.
func NullTimeToPtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	return &n.Time
}

// PtrToNullTime converts *time.Time to sql.NullTime.
func PtrToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// NullInt64ToPtr converts sql.NullInt64 to *int64.
func NullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

// PtrToNullInt64 converts *int64 to sql.NullInt64.
func PtrToNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// NullInt32ToPtr converts sql.NullInt32 to *int32.
func NullInt32ToPtr(n sql.NullInt32) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}

// PtrToNullInt32 converts *int32 to sql.NullInt32.
func PtrToNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

// NullBoolToPtr converts sql.NullBool to *bool.
func NullBoolToPtr(n sql.NullBool) *bool {
	if !n.Valid {
		return nil
	}
	return &n.Bool
}

// NullBoolToBool converts sql.NullBool to bool, returning false for NULL.
func NullBoolToBool(n sql.NullBool) bool {
	return n.Valid && n.Bool
}

// BoolToNullBool converts bool to sql.NullBool (always valid).
func BoolToNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}
