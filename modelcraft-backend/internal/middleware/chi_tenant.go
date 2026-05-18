package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ChiGraphQLOrgMiddleware returns a Chi middleware for GraphQL routes that require organization context.
// It extracts organization name from the URL path parameter {orgName} and injects it
// into the request context via requestContextKeyOrgName. Must run after ChiJWTAuthMiddleware.
func ChiGraphQLOrgMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logfacade.GetLogger(r.Context())

			orgName := chi.URLParam(r, "orgName")
			if orgName == "" {
				logger.Errorf(r.Context(), "No organization specified in URL path")
				writeJSONError(w, http.StatusBadRequest,
					"organization not specified in URL path", "TENANT_ORG_REQUIRED")
				return
			}

			logger.Infof(r.Context(), "Organization context set from URL path: %s", orgName)

			ctx := ctxutils.SetOrgName(r.Context(), orgName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ChiGraphQLProjectMiddleware extracts projectSlug from URL and sets it in context.
func ChiGraphQLProjectMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectSlug := chi.URLParam(r, "projectSlug")
			if projectSlug == "" {
				http.Error(w, "projectSlug not found in URL", http.StatusBadRequest)
				return
			}
			ctx := ctxutils.SetProjectSlug(r.Context(), projectSlug)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
