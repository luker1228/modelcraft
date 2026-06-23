package middleware

import (
	"encoding/json"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
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
					logger.With(
						logfacade.Any("panic", rec),
						logfacade.String(logfacade.StackFieldKey, string(debug.Stack())),
						logfacade.Duration(logfacade.DurationKey, time.Since(start)),
						logfacade.String(logfacade.RequestIDKey, requestID),
					).Errorf(r.Context(), nil, "request_panic")

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

// resolveRequestID resolves the request ID from the X-Request-ID header.
// Recovery runs as the outermost middleware, before ChiLoggerMiddleware sets
// the request_id in context, so we read the header directly with a UUID fallback.
func resolveRequestID(r *http.Request) string {
	if requestID := r.Header.Get(httpheader.XRequestID); requestID != "" {
		return requestID
	}
	return uuid.NewString()
}
