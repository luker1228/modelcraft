package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"modelcraft/pkg/logfacade"
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
// It logs request start/end with duration, status code, and all tracing identifiers:
// request_id, client_request_id, trace_id, span_id (parsed from W3C traceparent header).
// A request-scoped logger carrying these fields is stored in context so downstream
// handlers can retrieve it via logfacade.GetLogger(ctx).
func ChiLoggerMiddleware(logger logfacade.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// request_id — forwarded by the Gateway (always present in production)
			requestID := chimw.GetReqID(r.Context())
			if requestID == "" {
				requestID = r.Header.Get("X-Request-Id")
			}

			// Set X-Request-ID response header
			w.Header().Set("X-Request-Id", requestID)

			// Build scoped logger fields: always include request_id; add
			// client_request_id, trace_id, span_id when present.
			logFields := []logfacade.Field{logfacade.String("request_id", requestID)}

			if clientReqID := r.Header.Get("X-Client-Request-Id"); clientReqID != "" {
				logFields = append(logFields, logfacade.String("client_request_id", clientReqID))
			}

			if traceID, spanID, ok := parseTraceparent(r.Header.Get("Traceparent")); ok {
				logFields = append(logFields,
					logfacade.String("trace_id", traceID),
					logfacade.String("span_id", spanID),
				)
			}

			requestLogger := logger.With(logFields...)
			ctx := logfacade.WithLogger(r.Context(), requestLogger)
			r = r.WithContext(ctx)

			// Read and capture request body
			var requestBody string
			if shouldCaptureChiRequestBody(r.Header.Get("Content-Type")) {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					requestBody = string(body)
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}

			requestLogger.Info(r.Context(), "request_start",
				logfacade.String("method", r.Method),
				logfacade.String("url", r.URL.String()),
				logfacade.String("path", r.URL.Path),
				logfacade.String("remote_addr", r.RemoteAddr),
				logfacade.String("user_agent", r.UserAgent()),
				logfacade.String("request_body", requestBody),
			)

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			responseBuffer := &bytes.Buffer{}
			wrappedWriter := &chiResponseWriter{
				WrapResponseWriter: ww,
				body:               responseBuffer,
			}

			defer func() {
				duration := time.Since(start)

				var responseBody string
				if shouldCaptureChiResponseBody(ww.Header().Get("Content-Type")) {
					responseBody = responseBuffer.String()
				}

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

// parseTraceparent parses a W3C traceparent header of the form:
// "00-{traceId(32hex)}-{spanId(16hex)}-{flags(2hex)}"
// Returns traceID, spanID, and true on success.
func parseTraceparent(header string) (traceID, spanID string, ok bool) {
	parts := strings.Split(header, "-")
	if len(parts) != 4 || parts[0] != "00" {
		return "", "", false
	}
	if len(parts[1]) != 32 || len(parts[2]) != 16 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func shouldCaptureChiRequestBody(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}

func shouldCaptureChiResponseBody(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}
