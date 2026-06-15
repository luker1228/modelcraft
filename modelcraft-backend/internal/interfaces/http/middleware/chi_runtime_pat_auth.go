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

// IsOrgAdminFn is a function that checks whether a user has org-admin status.
// Injected into ChiRuntimePATMiddleware so it can set the IsAdmin context flag
// without importing the full repository package.
type IsOrgAdminFn func(ctx context.Context, orgName, userID string) (bool, error)

// ChiRuntimePATMiddleware handles mc_pat_* Bearer tokens for /end-user/ runtime routes.
// On success it injects:
//   - ctxutils end-user fields (EndUserID, OrgName, UserType) so the JWT middleware short-circuits
//   - ctxutils IsAdmin flag when the user's user_orgs.is_admin=true, mirroring the
//     APISIX/JWT path so graphql_app.go can skip permission checks for org admins
//
// Non-PAT requests pass through unchanged (JWT middleware handles them).
func ChiRuntimePATMiddleware(
	svc *apitoken.APITokenService,
	logger logfacade.Logger,
	isOrgAdminFn IsOrgAdminFn,
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

			// Inject ctxutils fields so downstream handlers can distinguish internal user
			// identity from end-user identity without overloading ContextKeyUserID.
			ctx = ctxutils.SetEndUserID(ctx, token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)
			ctx = ctxutils.SetUserType(ctx, ctxutils.UserTypeEndUser)

			// Check org-admin status and inject IsAdmin flag, mirroring the
			// APISIX X-Is-Admin header path for JWT callers.
			if isOrgAdminFn != nil {
				if admin, adminErr := isOrgAdminFn(ctx, token.OrgName, token.EndUserID); adminErr == nil && admin {
					ctx = ctxutils.SetIsAdmin(ctx, true)
				}
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
