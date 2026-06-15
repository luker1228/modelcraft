package middleware

import (
	"context"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// ChiHttpContextMiddleware creates HttpRequestContext from Chi's request ID and other request data.
// This middleware should run early in the Chi middleware chain to make HttpRequestContext
// available to all downstream handlers (including GraphQL).
func ChiHttpContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from Chi's context (set by chimw.RequestID middleware)
			requestID := chimw.GetReqID(r.Context())
			if requestID == "" {
				// Fallback to X-Request-ID header if Chi middleware didn't set it
				requestID = r.Header.Get(httpheader.XRequestID)
			}
			if requestID == "" {
				// Final fallback: generate a random UUID
				requestID = uuid.NewString()
			}

			// Determine language from Accept-Language header; default to English.
			lang := bizerrors.ParseLanguage(r.Header.Get(httpheader.AcceptLanguage))
			if lang == "" {
				lang = bizerrors.LangEN
			}

			// Create HttpRequestContext
			httpReqCtx := &ctxutils.HttpRequestContext{
				RequestId: requestID,
				Method:    r.Method,
				Path:      r.URL.Path,
				ClientIP:  r.RemoteAddr, // Chi doesn't have ClientIP() like Gin, use RemoteAddr
				Lang:      lang,
			}

			// Store in request context
			ctx := context.WithValue(r.Context(), ctxutils.HttpRequestContextKey, httpReqCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
