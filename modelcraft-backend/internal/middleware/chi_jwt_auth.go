package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"net/http"
	"strings"
)

// writeJSONError is a helper for writing JSON error responses.
func writeJSONError(w http.ResponseWriter, status int, errMsg, code string) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// ChiJWTAuthMiddleware authenticates design-time requests using the gateway-trusted identity contract.
//
// Authentication paths (in priority order):
//  1. Short-circuit: upstream middleware already authenticated this request.
//  2. X-User-ID: injected by the trusted gateway after validating the bearer token.
//     The backend trusts this header unconditionally; it is safe only because the backend
//     is not directly reachable from the public internet.
//
// On /end-user/* routes, X-User-ID carries an end-user ID and is stored as EndUserID.
// On all other routes, X-User-ID carries a tenant (developer) user ID and is stored as UserID.
// These are separate identities — userId ≠ endUserId.
//
// Direct bearer token validation has been removed. All auth is now handled exclusively
// by the gateway, which strips the Authorization header and injects X-User-ID before
// forwarding to the backend.
func ChiJWTAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Short-circuit: upstream middleware already authenticated this request.
			if uid, err := ctxutils.GetUserIDFromContext(r.Context()); err == nil && uid != "" {
				next.ServeHTTP(w, r)
				return
			}
			if endUserID, err := ctxutils.GetEndUserIDFromContext(r.Context()); err == nil && endUserID != "" {
				next.ServeHTTP(w, r)
				return
			}

			// Gateway-trusted identity: the gateway validates the bearer token,
			// strips the Authorization header, and injects:
			// - X-User-ID: user identity (tenant user ID or end-user ID depending on route)
			// - X-Is-Admin: "true" / "false" (from is_admin JWT claim)
			//
			// userId and endUserId are distinct identities:
			//   /end-user/*  → X-User-ID is an end-user ID → EndUserID
			//   /graphql/*   → X-User-ID is a tenant user ID  → UserID
			if userID := r.Header.Get(httpheader.XUserID); userID != "" {
				ctx := r.Context()
				if strings.HasPrefix(r.URL.Path, "/end-user/") {
					ctx = ctxutils.SetEndUserID(ctx, userID)
				} else {
					ctx = ctxutils.SetUserID(ctx, userID)
				}
				if r.Header.Get(httpheader.XIsAdmin) == "true" {
					ctx = ctxutils.SetIsAdmin(ctx, true)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No trusted identity found — reject.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
