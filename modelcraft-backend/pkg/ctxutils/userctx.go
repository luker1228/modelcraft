package ctxutils

import (
	"context"
	"fmt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// HttpRequestContextKey is the key for storing HTTP request context
	HttpRequestContextKey contextKey = "http_request_context"

	// User context keys - these are the standard keys for storing user data in context
	// Use the typed Set*/Get* functions below rather than these constants directly.
	ContextKeyUserID      contextKey = "user_id"
	ContextKeyOrgName     contextKey = "org_name"
	ContextKeyPermissions contextKey = "permissions"

	// ContextKeyUseCache controls whether the runtime query uses the schema cache.
	// Defaults to true; set to false via query param ?useCache=false.
	ContextKeyUseCache contextKey = "use_cache"

	// ContextKeyProjectSlug is the key for storing the project slug in context.
	ContextKeyProjectSlug contextKey = "project_slug"

	// ContextKeyUserType distinguishes "end_user" from "tenant" (developer) callers.
	ContextKeyUserType contextKey = "user_type"

	// ContextKeyIsAdmin stores whether the end-user is an org admin, derived from the
	// is_admin JWT claim injected by APISIX as X-Is-Admin header.
	ContextKeyIsAdmin contextKey = "is_admin"

	// ContextKeyTenantUserID stores the tenant admin's user ID.
	// Set from X-Tenant-User-Id header (injected by APISIX for tenant tokens).
	ContextKeyTenantUserID contextKey = "tenant_user_id"
)

const UserTypeEndUser = "end_user"

// HttpRequestContext encapsulates HTTP request-related data
// This is for HTTP layer concerns, not business logic
type HttpRequestContext struct {
	RequestId string // Unique request identifier for tracing
	Method    string // HTTP method (GET, POST, etc.)
	Path      string // Request path
	ClientIP  string // Client IP address
	Lang      string // Client language preference
}

// NewHttpContext creates a new context with HttpRequestContext
func NewHttpContext(parent context.Context, hrc *HttpRequestContext) context.Context {
	return context.WithValue(parent, HttpRequestContextKey, hrc)
}

// FromContext extracts HttpRequestContext from context
func FromContext(ctx context.Context) *HttpRequestContext {
	val, ok := ctx.Value(HttpRequestContextKey).(*HttpRequestContext)
	if !ok {
		return nil
	}
	return val
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if hrc := FromContext(ctx); hrc != nil {
		return hrc.RequestId
	}
	return ""
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

// SetUserType stores the user type ("end_user" or "tenant") in context.
func SetUserType(ctx context.Context, userType string) context.Context {
	return context.WithValue(ctx, ContextKeyUserType, userType)
}

// SetIsAdmin stores the end-user admin flag in context.
// This is populated from the X-Is-Admin header injected by APISIX.
func SetIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, ContextKeyIsAdmin, isAdmin)
}

// GetIsAdminFromContext returns whether the current end-user is an org admin.
// Returns false if the flag is not set in context.
func GetIsAdminFromContext(ctx context.Context) bool {
	val, _ := ctx.Value(ContextKeyIsAdmin).(bool)
	return val
}

// IsEndUser returns true if the request is from an EndUser caller.
func IsEndUser(ctx context.Context) bool {
	val, _ := ctx.Value(ContextKeyUserType).(string)
	return val == UserTypeEndUser
}

// SetUseCache stores the useCache flag in context.
// When false, the runtime GraphQL handler bypasses the schema cache.
func SetUseCache(ctx context.Context, useCache bool) context.Context {
	return context.WithValue(ctx, ContextKeyUseCache, useCache)
}

// SetTenantUserID stores the tenant admin's user ID in context.
func SetTenantUserID(ctx context.Context, tenantUserID string) context.Context {
	return context.WithValue(ctx, ContextKeyTenantUserID, tenantUserID)
}

// GetTenantUserIDFromContext extracts the tenant admin's user ID from context.
// Returns ("", error) if not set — only present for tenant (developer/admin) callers.
func GetTenantUserIDFromContext(ctx context.Context) (string, error) {
	val, _ := ctx.Value(ContextKeyTenantUserID).(string)
	if val == "" {
		return "", fmt.Errorf("tenant user ID not found in context")
	}
	return val, nil
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
