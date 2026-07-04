package middleware

import (
	"modelcraft/pkg/httpheader"
	"net/http"
	"strings"
)

// ChiCORS returns a Chi-compatible CORS middleware that mirrors
// the existing Gin CORS middleware behavior.
func ChiCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.AccessControlAllowOrigin, "*")
			w.Header().Set(httpheader.AccessControlAllowCredentials, "true")
			w.Header().Set(httpheader.AccessControlAllowHeaders,
				strings.Join(httpheader.CORSAllowRequestHeaders, ", "))
			w.Header().Set(httpheader.AccessControlAllowMethods, "POST, OPTIONS, GET, PUT, DELETE")
			w.Header().Set(httpheader.AccessControlExposeHeaders,
				strings.Join(httpheader.CORSExposeResponseHeaders, ", "))

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
