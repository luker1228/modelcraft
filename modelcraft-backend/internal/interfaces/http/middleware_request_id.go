package http

import (
	"bytes"
	"encoding/json"
	"modelcraft/pkg/ctxutils"
	"net/http"
)

// responseRecorder buffers the response body so the middleware can inject fields.
type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	buf         bytes.Buffer
	wroteHeader bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.wroteHeader = true
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	return rr.buf.Write(b)
}

// requestIDInjectorMiddleware intercepts all JSON responses and injects
// the requestId field from context into the response body if it is absent.
// Assumes all responses are JSON.
func requestIDInjectorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rr, r)

		body := rr.buf.Bytes()

		// Inject requestId
		injected := injectRequestID(r, body)

		if rr.wroteHeader {
			w.WriteHeader(rr.statusCode)
		}
		_, _ = w.Write(injected)
	})
}

// injectRequestID inserts requestId into a JSON object if the field is absent or empty.
func injectRequestID(r *http.Request, body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}

	existing, hasKey := obj["requestId"]
	if hasKey {
		var s string
		if err := json.Unmarshal(existing, &s); err == nil && s != "" {
			return body
		}
	}

	requestID := ctxutils.GetRequestID(r.Context())
	if requestID == "" {
		return body
	}

	encoded, err := json.Marshal(requestID)
	if err != nil {
		return body
	}
	obj["requestId"] = encoded

	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}
