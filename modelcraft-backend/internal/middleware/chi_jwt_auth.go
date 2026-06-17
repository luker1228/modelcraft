package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"net/http"
)

// writeJSONError is a helper for writing JSON error responses.
func writeJSONError(w http.ResponseWriter, status int, errMsg, code string) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// ChiJWTAuthMiddleware authenticates requests using the gateway-trusted identity contract.
//
// X-User-ID is injected by the trusted gateway after validating the bearer token.
// The backend trusts this header unconditionally; it is safe only because the backend
// is not directly reachable from the public internet.
//
// The X-User-ID value is stored as both EndUserID and UserID in context.
// Downstream code selects the appropriate identity based on the route context.
//
// Direct bearer token validation has been removed. All auth is now handled exclusively
// by the gateway, which strips the Authorization header and injects X-User-ID before
// forwarding to the backend.
func ChiJWTAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userID := r.Header.Get(httpheader.XUserID); userID != "" {
				ctx := r.Context()
				ctx = ctxutils.SetEndUserID(ctx, userID)
				ctx = ctxutils.SetUserID(ctx, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No trusted identity found — reject.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
