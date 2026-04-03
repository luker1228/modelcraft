package middleware

import (
	"net/http"
)

// JSONContentTypeMiddleware automatically sets Content-Type: application/json
// for JSON responses if not already set by the handler.
// This eliminates the need for each handler to manually set Content-Type.
func JSONContentTypeMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the response writer to intercept WriteHeader calls
			wrapped := &jsonResponseWriter{ResponseWriter: w}
			next.ServeHTTP(wrapped, r)
		})
	}
}

// jsonResponseWriter wraps http.ResponseWriter to auto-set Content-Type
type jsonResponseWriter struct {
	http.ResponseWriter
	headerWritten bool
}

// WriteHeader captures the header write and ensures Content-Type is set
func (w *jsonResponseWriter) WriteHeader(statusCode int) {
	if !w.headerWritten {
		// Only set Content-Type if not already set
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.headerWritten = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures headers are written first with Content-Type set
func (w *jsonResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		// Set default Content-Type before first write
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.headerWritten = true
	}
	return w.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter for interfaces like http.Hijacker
func (w *jsonResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
