package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"modelcraft-gateway/internal/auth"
)

// RESTHandler proxies protected REST endpoints to the backend.
// It validates the Gateway-issued JWT, extracts the userID from claims,
// injects it as X-User-ID, and forwards the request via reverse proxy.
// The backend trusts X-User-ID because it is only reachable via the Gateway.
type RESTHandler struct {
	proxy   *httputil.ReverseProxy
	authSvc *auth.Service
}

func NewRESTHandler(backendURL string, authSvc *auth.Service) (*RESTHandler, error) {
	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}
	return &RESTHandler{
		proxy:   httputil.NewSingleHostReverseProxy(target),
		authSvc: authSvc,
	}, nil
}

func (h *RESTHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		restWriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing access token")
		return
	}

	claims, err := h.authSvc.VerifyAccessToken(token)
	if err != nil {
		restWriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
		return
	}

	// Strip Gateway JWT; backend identifies the caller via X-User-ID only.
	r.Header.Del("Authorization")
	r.Header.Set("X-User-ID", claims.UserID)

	h.proxy.ServeHTTP(w, r)
}

type restErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func restWriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(restErrorResponse{Code: code, Message: message})
}
