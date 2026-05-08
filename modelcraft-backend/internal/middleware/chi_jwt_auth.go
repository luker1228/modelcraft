package middleware

import (
	"modelcraft/pkg/ctxutils"
	"net/http"
)

// JWTAuthConfig holds configuration for the JWT authentication middleware.
type JWTAuthConfig struct {
	// ModelCraftSecret is retained for compatibility but no longer used for bearer token validation.
	// Backend design-time endpoints authenticate exclusively via gateway-injected headers.
	ModelCraftSecret []byte
	SkipValidation   bool
	// InternalToken allows BFF server-side callers to authenticate via X-Internal-Token header,
	// bypassing the requirement for a user JWT. When set, requests carrying a matching
	// X-Internal-Token are granted access without a userID in context.
	InternalToken string
}

// writeJSONError is a helper for writing JSON error responses.
func writeJSONError(w http.ResponseWriter, status int, errMsg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// ChiJWTAuthMiddleware authenticates design-time requests using the gateway-trusted identity contract.
//
// Authentication paths (in priority order):
//  1. SkipValidation: bypass all checks (dev/test only).
//  2. X-Internal-Token: BFF server-side callers authenticate without a user identity.
//  3. X-User-ID: injected by the trusted gateway after validating the developer bearer token.
//     The backend trusts this header unconditionally; it is safe only because the backend
//     is not directly reachable from the public internet.
//
// Direct developer bearer token validation has been removed. All developer auth is
// now handled exclusively by the gateway, which strips the Authorization header and
// injects X-User-ID before forwarding to the backend.
func ChiJWTAuthMiddleware(config *JWTAuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.SkipValidation {
				next.ServeHTTP(w, r)
				return
			}

			if tryInternalTokenAuth(config, w, r, next) {
				return
			}

			// Gateway-trusted identity: the gateway validates the bearer token,
			// strips the Authorization header, and injects X-User-ID (and optionally
			// X-User-Type for EndUser callers) into the forwarded request.
			if userID := r.Header.Get("X-User-ID"); userID != "" {
				ctx := ctxutils.SetUserID(r.Context(), userID)
				if userType := r.Header.Get("X-User-Type"); userType != "" {
					ctx = ctxutils.SetUserType(ctx, userType)
				}
				if adminID := r.Header.Get("X-End-User-Admin-ID"); adminID != "" {
					ctx = ctxutils.SetEndUserAdminID(ctx, adminID)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No trusted identity found — reject. Direct bearer token submission is no longer
			// an accepted authentication path for design-time endpoints.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// tryInternalTokenAuth checks for X-Internal-Token header and handles the request if present.
// Returns true if the request was handled (either accepted or rejected), false if not an internal token request.
// When the internal token is valid, the request is forwarded with a superuser permission set so that
// @hasPermission directives pass — internal token callers are trusted backend services.
// If X-User-ID is also present, it is injected into the context to satisfy user identity checks.
func tryInternalTokenAuth(config *JWTAuthConfig, w http.ResponseWriter, r *http.Request, next http.Handler) bool {
	if config.InternalToken == "" {
		return false
	}
	provided := r.Header.Get("X-Internal-Token")
	if provided == "" {
		return false
	}
	if !matchInternalToken(config.InternalToken, provided) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return true
	}
	// Internal token authenticated. Grant wildcard permissions so that @hasPermission
	// directives pass without a database lookup — internal callers are fully trusted.
	ctx := ctxutils.SetPermissions(r.Context(), []string{"*"})
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		ctx = ctxutils.SetUserID(ctx, userID)
	}
	next.ServeHTTP(w, r.WithContext(ctx))
	return true
}
