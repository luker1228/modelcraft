package middleware

import (
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RequestLogger logs request_start (before) and request_end (after) for every
// incoming request. It also injects a per-request *zap.Logger (pre-tagged with
// request_id) into the context so downstream handlers can call
// middleware.LoggerFromCtx(r.Context()).
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID := chiMiddleware.GetReqID(r.Context())
			clientReqID := GetClientRequestID(r)

			// Build a per-request child logger so every log line from this
			// request automatically carries request_id.
			reqLogger := logger.With(zap.String("request_id", reqID))
			if clientReqID != "" {
				reqLogger = reqLogger.With(zap.String("client_request_id", clientReqID))
			}
			ctx := WithLogger(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			reqLogger.Info("request_start",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			endFields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
			}
			if sc := trace.SpanFromContext(r.Context()).SpanContext(); sc.IsValid() {
				endFields = append(endFields,
					zap.String("trace_id", sc.TraceID().String()),
					zap.String("span_id", sc.SpanID().String()),
				)
			}
			reqLogger.Info("request_end", endFields...)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
