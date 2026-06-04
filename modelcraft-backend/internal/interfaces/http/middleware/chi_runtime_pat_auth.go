package middleware

import (
	"context"
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"
)

// writeRuntimeJSONError writes a JSON error response for runtime endpoints.
func writeRuntimeJSONError(w http.ResponseWriter, status int, errMsg, _ string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// ChiRuntimePATMiddleware handles mc_pat_* Bearer tokens for /end-user/ runtime routes.
// On success it injects:
//   - EndUserIdentity into context (read by rls_resolver.go via GetEndUserIdentity)
//   - ctxutils user fields (UserID, OrgName, UserType) so the JWT middleware short-circuits
//
// Non-PAT requests pass through unchanged (JWT middleware handles them).
func ChiRuntimePATMiddleware(
	svc *appEnduser.APITokenService,
	logger logfacade.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearer, "Bearer mc_pat_") {
				next.ServeHTTP(w, r)
				return
			}

			plaintext := strings.TrimPrefix(bearer, "Bearer ")
			token, err := svc.ValidateToken(r.Context(), plaintext)
			if err != nil || token == nil {
				if logger != nil {
					logger.Warnf(r.Context(), "runtime PAT validation failed: %v", err)
				}
				writeRuntimeJSONError(w, http.StatusUnauthorized, "invalid or expired API token", "UNAUTHENTICATED")
				return
			}

			// Fire-and-forget: update last_used_at
			go func() {
				if updateErr := svc.UpdateLastUsedAt(context.Background(), token.ID, time.Now()); updateErr != nil {
					if logger != nil {
						logger.Warnf(context.Background(), "update last_used_at failed: %v", updateErr)
					}
				}
			}()

			// Inject EndUserIdentity for RLS resolver (GetEndUserIdentity reads this)
			identity := &EndUserIdentity{
				EndUserID: token.EndUserID,
				Issuer:    issuerPlatform,
			}
			ctx := context.WithValue(r.Context(), endUserContextKey, identity)

			// Also inject ctxutils fields so ChiJWTAuthMiddleware short-circuits
			ctx = ctxutils.SetUserID(ctx, token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)
			ctx = ctxutils.SetUserType(ctx, ctxutils.UserTypeEndUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
