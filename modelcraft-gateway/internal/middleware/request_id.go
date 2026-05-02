package middleware

import (
	"context"
	"net/http"
	"unicode"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	headerClientRequestID = "X-Client-Request-Id"
	headerRequestID       = "X-Request-Id"

	maxClientRequestIDLen = 128
)

// RequestID is the gateway's request-tracing middleware. It:
//  1. Reads X-Client-Request-Id from the browser, validates and logs it.
//  2. Always generates a fresh X-Request-Id (UUID v4) — never trusts the
//     client-supplied value for internal propagation.
//  3. Stores the generated ID in the chi RequestID context key so that
//     chiMiddleware.GetReqID(ctx) works for downstream code.
//  4. Writes X-Request-Id back as a response header.
//
// OpenTelemetry is responsible for traceId/spanId via the traceparent header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// -- 1. Validate and log X-Client-Request-Id --
		clientReqID := r.Header.Get(headerClientRequestID)
		if clientReqID != "" {
			if isValidClientRequestID(clientReqID) {
				zap.L().Info("client request id received",
					zap.String("client_request_id", clientReqID))
			} else {
				zap.L().Warn("X-Client-Request-Id rejected: invalid format",
					zap.String("raw_value", clientReqID))
				clientReqID = "" // discard
			}
		}

		// -- 2. Always generate a fresh internal request ID --
		reqID := uuid.New().String()

		// -- 3. Store in context: chi's key for GetReqID, custom key for client ID --
		ctx := context.WithValue(r.Context(), chiMiddleware.RequestIDKey, reqID)
		if clientReqID != "" {
			ctx = context.WithValue(ctx, clientRequestIDKey{}, clientReqID)
		}
		r = r.WithContext(ctx)

		// -- 4. Write X-Request-Id response header --
		w.Header().Set(headerRequestID, reqID)

		next.ServeHTTP(w, r)
	})
}

// GetClientRequestID retrieves the validated X-Client-Request-Id from context.
// Returns empty string if the client did not send one or it was invalid.
func GetClientRequestID(r *http.Request) string {
	v, _ := r.Context().Value(clientRequestIDKey{}).(string)
	return v
}

type clientRequestIDKey struct{}

// isValidClientRequestID checks that the value is within the length limit
// and contains only printable ASCII characters (no control chars).
func isValidClientRequestID(s string) bool {
	if len(s) == 0 || len(s) > maxClientRequestIDLen {
		return false
	}
	for _, c := range s {
		if c > unicode.MaxASCII || !unicode.IsPrint(c) {
			return false
		}
	}
	return true
}
