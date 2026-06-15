package middleware

import (
	"encoding/json"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"runtime/debug"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// ChiRecoveryMiddleware is a Chi-compatible recovery middleware that
// logs panics with stack trace using the provided logger and returns
// a JSON error response containing the request_id for tracing.
func ChiRecoveryMiddleware(logger logfacade.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			defer func() {
				if rec := recover(); rec != nil {
					requestID := resolveRequestID(r)

					// Log panic with stack trace and duration
					logger.Error(r.Context(), "request_panic",
						logfacade.Any("panic", rec),
						logfacade.String("stack", string(debug.Stack())),
						logfacade.Duration("duration", time.Since(start)),
						logfacade.String("request_id", requestID),
					)

					// Return JSON error response similar to Gin middleware
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error":      "Internal Server Error",
						"request_id": requestID,
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func resolveRequestID(r *http.Request) string {
	if hrc := ctxutils.FromContext(r.Context()); hrc != nil && hrc.RequestId != "" {
		return hrc.RequestId
	}

	if requestID := chimw.GetReqID(r.Context()); requestID != "" {
		return requestID
	}

	return r.Header.Get(httpheader.XRequestID)
}
