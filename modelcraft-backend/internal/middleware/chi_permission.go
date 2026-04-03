package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
)

// CheckPermission checks if a permission list grants the required permission.
// Supports three matching modes:
//  1. Global wildcard: "*" matches any permission
//  2. Resource wildcard: "resource:*" matches any action on that resource
//  3. Exact match: "resource:action" matches exactly
func CheckPermission(userPermissions []string, required string) bool {
	for _, perm := range userPermissions {
		if perm == "*" {
			return true
		}
		if perm == required {
			return true
		}
		if strings.HasSuffix(perm, ":*") {
			permResource := strings.TrimSuffix(perm, ":*")
			requiredParts := strings.SplitN(required, ":", 2)
			if len(requiredParts) == 2 && requiredParts[0] == permResource {
				return true
			}
		}
	}
	return false
}

// ChiRequirePermission returns a Chi-compatible middleware that checks if the user has the required permission.
// It must run after ChiJWTAuthMiddleware so that permissions are available in the request context.
func ChiRequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logfacade.GetLogger(r.Context())
			permissions, _ := ctxutils.GetPermissionsFromContext(r.Context())
			if permissions == nil {
				logger.Errorf(r.Context(), "No permissions found in context for required permission: %s", permission)
				writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
				return
			}

			if !CheckPermission(permissions, permission) {
				userID, _ := ctxutils.GetUserIDFromContext(r.Context())
				logger.Errorf(r.Context(), "User %s lacks required permission: %s", userID, permission)
				writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ChiRequireAnyPermission returns a Chi-compatible middleware that checks if the user has at least one
// of the required permissions.
func ChiRequireAnyPermission(requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logfacade.GetLogger(r.Context())
			permissions, _ := ctxutils.GetPermissionsFromContext(r.Context())
			if permissions == nil {
				logger.Errorf(r.Context(), "No permissions found in context, required any of: %v", requiredPermissions)
				writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
				return
			}

			for _, p := range requiredPermissions {
				if CheckPermission(permissions, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			userID, _ := ctxutils.GetUserIDFromContext(r.Context())
			logger.Errorf(r.Context(), "User %s lacks any of required permissions: %v", userID, requiredPermissions)
			writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
		})
	}
}

// ChiRequireAllPermissions returns a Chi-compatible middleware that checks if the user has all of the
// required permissions.
func ChiRequireAllPermissions(requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logfacade.GetLogger(r.Context())
			permissions, _ := ctxutils.GetPermissionsFromContext(r.Context())
			if permissions == nil {
				logger.Errorf(r.Context(), "No permissions found in context, required all of: %v", requiredPermissions)
				writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
				return
			}

			for _, p := range requiredPermissions {
				if !CheckPermission(permissions, p) {
					userID, _ := ctxutils.GetUserIDFromContext(r.Context())
					logger.Errorf(r.Context(), "User %s lacks required permission: %s", userID, p)
					writeJSONError(w, http.StatusForbidden, "permission denied", "AUTH_PERMISSION_DENIED")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
