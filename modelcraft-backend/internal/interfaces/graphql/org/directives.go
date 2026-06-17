package orggraphql

import (
	"context"
	"fmt"
	"modelcraft/internal/middleware"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	appPermission "modelcraft/internal/app/permission"
)

// HasPermissionDirective implements the @hasPermission directive for operation-level authorization
type HasPermissionDirective struct {
	userRoleService    *appPermission.UserRoleService
	enforceEndUserGate bool // true for EndUser link handler, false for internal link handler
}

// NewHasPermissionDirective creates a directive handler for the internal GraphQL link.
// Internal link ignores allowEndUser — all operations go through RBAC.
func NewHasPermissionDirective(userRoleService *appPermission.UserRoleService) *HasPermissionDirective {
	return &HasPermissionDirective{
		userRoleService:    userRoleService,
		enforceEndUserGate: false,
	}
}

// NewEndUserHasPermissionDirective creates a directive handler for the EndUser link.
// EndUser link enforces allowEndUser: only operations explicitly marked allowEndUser=true are accessible.
func NewEndUserHasPermissionDirective(userRoleService *appPermission.UserRoleService) *HasPermissionDirective {
	return &HasPermissionDirective{
		userRoleService:    userRoleService,
		enforceEndUserGate: true,
	}
}

// HasPermission checks if the user has the required permission before executing the field resolver
// Optimized to prioritize JWT claims (context permissions) over database queries for better performance.
// allowEndUser: when true, EndUser callers bypass RBAC checks entirely for this operation.
func (d *HasPermissionDirective) HasPermission(
	ctx context.Context,
	obj interface{},
	next graphql.Resolver,
	action string,
	allowEndUser bool,
) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)

	// Validate action format
	if action == "" {
		logger.Errorf(ctx, "@hasPermission directive: empty action parameter")
		return nil, newGQLError("Invalid permission directive configuration", "INVALID_DIRECTIVE")
	}

	// FAST PATH: If wildcard permissions are already set in context,
	// skip user identity validation and grant access immediately.
	if permissions, pErr := ctxutils.GetPermissionsFromContext(ctx); pErr == nil && len(permissions) > 0 {
		if middleware.CheckPermission(permissions, "*") || middleware.CheckPermission(permissions, action) {
			logger.Infof(ctx, "@hasPermission directive: access granted via context permissions (source: context)")
			return next(ctx)
		}
	}

	// Validate authentication and organization context
	userID, orgName, err := d.validateContext(ctx, logger)
	if err != nil {
		return nil, err
	}

	// End-user link: default-deny. Only operations explicitly marked allowEndUser=true
	// are accessible. This gate only applies to the EndUser link handler —
	// the internal link handler has enforceEndUserGate=false and skips this check.
	if d.enforceEndUserGate {
		if !allowEndUser {
			logger.Infof(ctx, "@hasPermission directive: end-user link access denied (allowEndUser=false) user=%s action=%s",
				userID, action)
			return nil, newPermissionDeniedError(action, userID, orgName, "end-user-link-default-deny")
		}
		logger.Infof(ctx, "@hasPermission directive: end-user link access granted (allowEndUser=true) user=%s action=%s",
			userID, action)
		return next(ctx)
	}

	// OPTIMIZATION: Try to get permissions from JWT context first (no DB query)
	if permissions, err := ctxutils.GetPermissionsFromContext(ctx); err == nil && permissions != nil {
		logger.Infof(
			ctx, "@hasPermission directive: using permissions from JWT context (source: jwt, count: %d)",
			len(permissions),
		)
		return d.checkContextPermission(ctx, next, logger, userID, orgName, action, permissions)
	}

	// FALLBACK: Query database if permissions not in context
	return d.checkDatabasePermission(ctx, next, logger, userID, orgName, action)
}

// validateContext extracts and validates user ID and organization from context
func (d *HasPermissionDirective) validateContext(
	ctx context.Context,
	logger logfacade.Logger,
) (userID, orgName string, err error) {
	// All authenticated callers use the standard user ID in context.
	userID, err = ctxutils.GetUserIDFromContext(ctx)
	if err != nil || userID == "" {
		logger.Errorf(ctx, "@hasPermission directive: missing user authentication: %v", err)
		return "", "", &gqlerror.Error{
			Message:    "Authentication required",
			Extensions: map[string]interface{}{"code": "UNAUTHENTICATED"},
		}
	}

	orgName, err = ctxutils.GetOrgNameFromContext(ctx)
	if err != nil || orgName == "" {
		logger.Errorf(ctx, "@hasPermission directive: missing organization context for user %s: %v", userID, err)
		return "", "", &gqlerror.Error{
			Message:    "Organization context required",
			Extensions: map[string]interface{}{"code": "MISSING_ORGANIZATION"},
		}
	}

	return userID, orgName, nil
}

// checkContextPermission validates permission using JWT context (fast path)
func (d *HasPermissionDirective) checkContextPermission(
	ctx context.Context,
	next graphql.Resolver,
	logger logfacade.Logger,
	userID, orgName, action string,
	permissions []string,
) (interface{}, error) {
	if middleware.CheckPermission(permissions, action) {
		logger.Infof(
			ctx, "@hasPermission directive: access granted for user %s, org %s, action %s (source: jwt)",
			userID, orgName, action,
		)
		return next(ctx)
	}

	logger.Infof(
		ctx, "@hasPermission directive: access denied for user %s, org %s, action %s (source: jwt)",
		userID, orgName, action,
	)
	return nil, newPermissionDeniedError(action, userID, orgName, "jwt")
}

// checkDatabasePermission validates permission by querying database (fallback path)
func (d *HasPermissionDirective) checkDatabasePermission(
	ctx context.Context,
	next graphql.Resolver,
	logger logfacade.Logger,
	userID, orgName, action string,
) (interface{}, error) {
	logger.Infof(
		ctx,
		"@hasPermission directive: permissions not in JWT context, "+
			"falling back to database query (user: %s, org: %s)",
		userID, orgName,
	)

	allowed, err := d.userRoleService.CheckPermission(ctx, userID, orgName, action)
	if err != nil {
		logger.Errorf(
			ctx, "@hasPermission directive: permission check failed for user %s, action %s: %v",
			userID, action, err,
		)
		return nil, newGQLError("Permission check failed", "PERMISSION_CHECK_ERROR")
	}

	if !allowed {
		logger.Infof(
			ctx, "@hasPermission directive: access denied for user %s, org %s, action %s (source: database)",
			userID, orgName, action,
		)
		return nil, newPermissionDeniedError(action, userID, orgName, "database")
	}

	logger.Infof(
		ctx, "@hasPermission directive: access granted for user %s, org %s, action %s (source: database)",
		userID, orgName, action,
	)
	return next(ctx)
}

// newGQLError creates a GraphQL error with code extension
func newGQLError(message, code string) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    message,
		Extensions: map[string]interface{}{"code": code},
	}
}

// newPermissionDeniedError creates a permission denied error with details
func newPermissionDeniedError(action, userID, orgName, source string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: fmt.Sprintf("Permission denied: requires '%s' permission", action),
		Extensions: map[string]interface{}{
			"code":           "PERMISSION_DENIED",
			"requiredAction": action,
			"userId":         userID,
			"organizationId": orgName,
			"source":         source,
		},
	}
}
