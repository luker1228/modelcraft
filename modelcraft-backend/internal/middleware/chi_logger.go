package middleware

import (
	"bytes"
	"io"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// chiResponseWriter wraps Chi's response writer to capture response body
type chiResponseWriter struct {
	chimw.WrapResponseWriter
	body *bytes.Buffer
}

func (w *chiResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.WrapResponseWriter.Write(b)
}

// ChiLoggerMiddleware returns a Chi-compatible structured logging middleware.
// It logs request start/end with duration, status code, and request ID.
// A request-scoped logger with only the request_id field is stored in the context
// so downstream handlers can retrieve it via logfacade.GetLogger(ctx).
func ChiLoggerMiddleware(logger logfacade.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Use Chi's built-in request ID (set by chimw.RequestID)
			requestID := chimw.GetReqID(r.Context())
			if requestID == "" {
				requestID = r.Header.Get("X-Request-ID")
			}

			// Set X-Request-ID response header
			w.Header().Set("X-Request-ID", requestID)

			// Create a request-scoped logger with only request_id field and store in context.
			// Downstream handlers retrieve it via logfacade.GetLogger(ctx).
			requestLogger := logger.With(logfacade.String("request_id", requestID))
			ctx := logfacade.WithLogger(r.Context(), requestLogger)
			r = r.WithContext(ctx)

			// Read and capture request body
			var requestBody string
			if shouldCaptureChiRequestBody(r.Header.Get("Content-Type")) {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					requestBody = string(body)
					// Restore body for downstream handlers
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}

			// Log request start with structured fields
			requestLogger.Info(r.Context(), "request_start",
				logfacade.String("method", r.Method),
				logfacade.String("url", r.URL.String()),
				logfacade.String("path", r.URL.Path),
				logfacade.String("remote_addr", r.RemoteAddr),
				logfacade.String("user_agent", r.UserAgent()),
				logfacade.String("request_body", requestBody),
			)

			// Wrap response writer to capture status code and response body
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			responseBuffer := &bytes.Buffer{}
			wrappedWriter := &chiResponseWriter{
				WrapResponseWriter: ww,
				body:               responseBuffer,
			}

			defer func() {
				duration := time.Since(start)

				// Get response body
				var responseBody string
				if shouldCaptureChiResponseBody(ww.Header().Get("Content-Type")) {
					responseBody = responseBuffer.String()
				}

				// Log request end with structured fields
				requestLogger.Info(r.Context(), "request_end",
					logfacade.String("method", r.Method),
					logfacade.String("path", r.URL.Path),
					logfacade.Int("status", ww.Status()),
					logfacade.Int("duration_ms", int(duration.Milliseconds())),
					logfacade.Int("size", ww.BytesWritten()),
					logfacade.String("response_body", responseBody),
				)
			}()

			next.ServeHTTP(wrappedWriter, r)
		})
	}
}

// shouldCaptureChiRequestBody checks if request body should be logged
func shouldCaptureChiRequestBody(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}

// shouldCaptureChiResponseBody checks if response body should be logged
func shouldCaptureChiResponseBody(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}
