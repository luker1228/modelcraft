package ctxutils

import (
	"context"
	"fmt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for the request ID (set by ChiLoggerMiddleware).
	RequestIDKey contextKey = "request_id"
	// LangKey is the context key for the client language preference (set by ChiLoggerMiddleware).
	LangKey contextKey = "lang"

	// User context keys - these are the standard keys for storing user data in context
	// Use the typed Set*/Get* functions below rather than these constants directly.
	ContextKeyUserID      contextKey = "user_id"
	ContextKeyEndUserID   contextKey = "end_user_id"
	ContextKeyOrgName     contextKey = "org_name"
	ContextKeyPermissions contextKey = "permissions"

	// ContextKeyUseCache controls whether the runtime query uses the schema cache.
	// Defaults to true; set to false via query param ?useCache=false.
	ContextKeyUseCache contextKey = "use_cache"

	// ContextKeyProjectSlug is the key for storing the project slug in context.
	ContextKeyProjectSlug contextKey = "project_slug"
)

// SetRequestID stores the request ID in context.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID extracts the request ID from context.
// Returns empty string if not set.
func GetRequestID(ctx context.Context) string {
	val, _ := ctx.Value(RequestIDKey).(string)
	return val
}

// SetLang stores the client language preference in context.
func SetLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, LangKey, lang)
}

// GetLang extracts the client language preference from context.
// Returns empty string if not set (callers should default to English).
func GetLang(ctx context.Context) string {
	val, _ := ctx.Value(LangKey).(string)
	return val
}

// SetContextValue sets a value in context using the standard context key
// This is the recommended way to store user context data
func SetContextValue(parent context.Context, key contextKey, value any) context.Context {
	return context.WithValue(parent, key, value)
}

// SetUserID stores the user ID in context.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// SetEndUserID stores the end-user ID in context.
func SetEndUserID(ctx context.Context, endUserID string) context.Context {
	return context.WithValue(ctx, ContextKeyEndUserID, endUserID)
}

// SetOrgName stores the organization name in context.
func SetOrgName(ctx context.Context, orgName string) context.Context {
	return context.WithValue(ctx, ContextKeyOrgName, orgName)
}

// SetProjectSlug stores the project slug in context.
func SetProjectSlug(ctx context.Context, projectSlug string) context.Context {
	return context.WithValue(ctx, ContextKeyProjectSlug, projectSlug)
}

// SetPermissions stores the permissions list in context.
func SetPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, ContextKeyPermissions, permissions)
}

// GetOrgNameFromContext extracts organization name from context.
// This is the recommended way to get orgName - use standard context.Context.
// The orgName is set by authentication/tenant middleware.
func GetOrgNameFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(ContextKeyOrgName)
	if val == nil {
		return "", fmt.Errorf("organization name not found in context")
	}
	orgName, ok := val.(string)
	if !ok || orgName == "" {
		return "", fmt.Errorf("organization name not found in context")
	}
	return orgName, nil
}

// GetProjectSlugFromContext extracts project slug from context.
// Returns error if not found or empty.
func GetProjectSlugFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(ContextKeyProjectSlug)
	if val == nil {
		return "", fmt.Errorf("project slug not found in context")
	}
	projectSlug, ok := val.(string)
	if !ok || projectSlug == "" {
		return "", fmt.Errorf("project slug not found in context")
	}
	return projectSlug, nil
}

// GetUserIDFromContext extracts user ID from context.
// This is the recommended way to get userID - use standard context.Context.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(ContextKeyUserID)
	if val == nil {
		return "", fmt.Errorf("user ID not found in context")
	}
	userID, ok := val.(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetEndUserIDFromContext extracts end-user ID from context.
func GetEndUserIDFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(ContextKeyEndUserID)
	if val == nil {
		return "", fmt.Errorf("end-user ID not found in context")
	}
	endUserID, ok := val.(string)
	if !ok || endUserID == "" {
		return "", fmt.Errorf("end-user ID not found in context")
	}
	return endUserID, nil
}

// GetPermissionsFromContext extracts permissions from context.
// This is the recommended way to get permissions - use standard context.Context.
func GetPermissionsFromContext(ctx context.Context) ([]string, error) {
	val := ctx.Value(ContextKeyPermissions)
	if val == nil {
		return nil, fmt.Errorf("permissions not found in context")
	}
	permissions, ok := val.([]string)
	if !ok {
		return nil, fmt.Errorf("permissions not found in context")
	}
	return permissions, nil
}

// SetUseCache stores the useCache flag in context.
// When false, the runtime GraphQL handler bypasses the schema cache.
func SetUseCache(ctx context.Context, useCache bool) context.Context {
	return context.WithValue(ctx, ContextKeyUseCache, useCache)
}

// GetUseCache extracts the useCache flag from context.
// Returns true by default if the value has not been set.
func GetUseCache(ctx context.Context) bool {
	val := ctx.Value(ContextKeyUseCache)
	if val == nil {
		return true
	}
	v, ok := val.(bool)
	if !ok {
		return true
	}
	return v
}
