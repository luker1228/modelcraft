package middleware

import (
	"modelcraft/internal/app/enduser"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"
)

const patPrefix = "mc_pat_"

// ChiPATAuthMiddleware identifies Bearer mc_pat_xxx tokens, validates them,
// and injects EndUser identity into context. Non-PAT requests pass through unchanged.
func ChiPATAuthMiddleware(
	svc *enduser.APITokenService, logger logfacade.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get(httpheader.Authorization)
			if !strings.HasPrefix(bearer, "Bearer "+patPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			plaintext := strings.TrimPrefix(bearer, "Bearer ")
			token, err := svc.ValidateToken(r.Context(), plaintext)
			if err != nil {
				logger.Warnf(r.Context(), "PAT validation failed: %v", err)
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired token", "UNAUTHENTICATED")
				return
			}

			go func() {
				if updateErr := svc.UpdateLastUsedAt(r.Context(), token.ID, time.Now()); updateErr != nil {
					logger.Warnf(r.Context(), "update last_used_at failed: %v", updateErr)
				}
			}()

			ctx := ctxutils.SetUserID(r.Context(), token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)
			ctx = ctxutils.SetUserType(ctx, "end_user")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
