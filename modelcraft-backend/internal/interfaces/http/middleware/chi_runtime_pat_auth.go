package middleware

import (
	"context"
	"modelcraft/internal/app/apitoken"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// writeRuntimeJSONError writes a JSON error response for runtime endpoints.
func writeRuntimeJSONError(w http.ResponseWriter, status int, errMsg, _ string) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// IsOrgAdminFn checks whether an end-user has org-admin status.
// Used by the PAT whoami handler to report admin status in the whoami response.
type IsOrgAdminFn func(ctx context.Context, orgName, userID string) (bool, error)

// ChiRuntimePATMiddleware handles mc_pat_* Bearer tokens for /end-user/ routes.
// On success it injects:
//   - ctxutils end-user fields (EndUserID, UserID, OrgName, UserType)
//   - ctxutils UseAdmin flag when the caller sets X-MC-Auth-Useadmin: true,
//     indicating the caller explicitly requests admin-level access
//
// Non-PAT requests pass through unchanged.
func ChiRuntimePATMiddleware(
	svc *apitoken.APITokenService,
	logger logfacade.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get(httpheader.Authorization)
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

			ctx := r.Context()

			// Inject both end-user and internal user identity so downstream
			// handlers (graphql_app, resolvers) can resolve the caller.
			ctx = ctxutils.SetEndUserID(ctx, token.EndUserID)
			ctx = ctxutils.SetUserID(ctx, token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)

			// UseAdmin is an explicit opt-in by the PAT caller via header.
			// It represents the caller's intent to use admin privileges.
			if strings.EqualFold(r.Header.Get(httpheader.XMCAuthUseAdmin), "true") {
				ctx = ctxutils.SetIsAdmin(ctx, true)
			}

			pathOrgName := chi.URLParam(r, "orgName")
			if pathOrgName != "" && pathOrgName != token.OrgName {
				if logger != nil {
					logger.Warnf(r.Context(),
						"runtime PAT org mismatch: token org=%s path org=%s endUserID=%s",
						token.OrgName, pathOrgName, token.EndUserID,
					)
				}
				writeRuntimeJSONError(w, http.StatusForbidden, "forbidden org scope", "FORBIDDEN")
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
