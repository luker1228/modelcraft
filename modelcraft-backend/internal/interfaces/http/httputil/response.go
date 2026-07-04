package httputil

import (
	"encoding/json"
	"modelcraft/pkg/httpheader"
	"net/http"
)

// WriteJSON writes a JSON response with the given status code and body.
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response with {"error": message}.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]string{"error": message})
}
