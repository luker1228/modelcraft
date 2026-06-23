package middleware

import (
	"bytes"
	"io"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
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
//
// It is also the single source of request_id: Chi ctx → X-Request-ID header → UUID fallback.
// The resolved request_id is written into the HttpRequestContext (for downstream business code)
// and into a request-scoped logger (stored via logfacade.WithLogger) so both share one id.
func ChiLoggerMiddleware(logger logfacade.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// request_id — single source: Chi ctx → X-Request-ID header → UUID fallback
			requestID := chimw.GetReqID(r.Context())
			if requestID == "" {
				requestID = r.Header.Get(httpheader.XRequestID)
			}
			if requestID == "" {
				requestID = uuid.NewString()
			}

			// Set X-Request-ID response header
			w.Header().Set(httpheader.XRequestID, requestID)

			// Parse language from Accept-Language header; default to English.
			lang := bizerrors.ParseLanguage(r.Header.Get(httpheader.AcceptLanguage))
			if lang == "" {
				lang = bizerrors.LangEN
			}

			// Build scoped logger fields: always include request_id; add
			// client_request_id, trace_id, span_id when present.
			logFields := []logfacade.Field{logfacade.String(logfacade.RequestIDKey, requestID)}

			if clientReqID := r.Header.Get(httpheader.XClientRequestID); clientReqID != "" {
				logFields = append(logFields, logfacade.String(logfacade.ClientRequestIDKey, clientReqID))
			}

			if traceID, spanID, ok := parseTraceparent(r.Header.Get(httpheader.Traceparent)); ok {
				logFields = append(logFields,
					logfacade.String(logfacade.TraceIDKey, traceID),
					logfacade.String(logfacade.SpanIDKey, spanID),
				)
			}

			requestLogger := logger.With(logFields...)
			ctx := logfacade.WithLogger(r.Context(), requestLogger)
			ctx = ctxutils.SetRequestID(ctx, requestID)
			ctx = ctxutils.SetLang(ctx, lang)
			r = r.WithContext(ctx)

			// Read and capture request body
			var requestBody string
			if shouldCaptureChiRequestBody(r.Header.Get(httpheader.ContentType)) {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					requestBody = string(body)
					r.Body = io.NopCloser(bytes.NewBuffer(body))
				}
			}

			requestLogger.With(
				logfacade.String(logfacade.MethodKey, r.Method),
				logfacade.String(logfacade.URLKey, r.URL.String()),
				logfacade.String(logfacade.PathKey, r.URL.Path),
				logfacade.String(logfacade.RemoteAddrKey, r.RemoteAddr),
				logfacade.String(logfacade.UserAgentKey, r.UserAgent()),
				logfacade.String(logfacade.ActionKey, r.Header.Get(httpheader.XTcAction)),
				logfacade.String(logfacade.RequestBodyKey, requestBody),
			).Infof(r.Context(), "request_start")

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			responseBuffer := &bytes.Buffer{}
			wrappedWriter := &chiResponseWriter{
				WrapResponseWriter: ww,
				body:               responseBuffer,
			}

			defer func() {
				duration := time.Since(start)

				var responseBody string
				if shouldCaptureChiResponseBody(ww.Header().Get(httpheader.ContentType)) {
					responseBody = responseBuffer.String()
				}

				requestLogger.With(
					logfacade.String(logfacade.MethodKey, r.Method),
					logfacade.String(logfacade.PathKey, r.URL.Path),
					logfacade.Int(logfacade.StatusKey, ww.Status()),
					logfacade.Int(logfacade.DurationKey, int(duration.Milliseconds())),
					logfacade.Int(logfacade.NetOutKey, ww.BytesWritten()),
					logfacade.String(logfacade.ResponseBodyKey, responseBody),
				).Infof(r.Context(), "request_end")
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
