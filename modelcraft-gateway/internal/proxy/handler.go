package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"modelcraft-gateway/internal/auth"
)

// Handler is a reverse proxy that forwards requests to the Go backend,
// injecting the X-Internal-Token header and stripping the Authorization header
// (auth is validated at the gateway level, not forwarded).
type Handler struct {
	backendURL    *url.URL
	internalToken string
	authService   *auth.Service
	reverseProxy  *httputil.ReverseProxy
}

func NewHandler(backendURL, internalToken string, authService *auth.Service) (*Handler, error) {
	parsed, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		backendURL:    parsed,
		internalToken: internalToken,
		authService:   authService,
	}
	h.reverseProxy = &httputil.ReverseProxy{
		Director:       h.director,
		ModifyResponse: h.modifyResponse,
	}
	return h, nil
}

// director rewrites the incoming request to target the upstream backend.
func (h *Handler) director(req *http.Request) {
	req.URL.Scheme = h.backendURL.Scheme
	req.URL.Host = h.backendURL.Host

	// Inject internal auth header; never forward user Bearer token upstream.
	req.Header.Set("X-Internal-Token", h.internalToken)
	req.Header.Del("Authorization")

	// Preserve the original host for virtual-hosting setups.
	req.Host = h.backendURL.Host
}

func (h *Handler) modifyResponse(resp *http.Response) error {
	// Remove any internal headers from the upstream response before forwarding to client.
	resp.Header.Del("X-Internal-Token")
	return nil
}

// GraphQLOrgHandler validates the gateway JWT, then proxies to /graphql/org/{orgName}/.
func (h *Handler) GraphQLOrgHandler(w http.ResponseWriter, r *http.Request) {
	if !h.verifyBearerToken(w, r) {
		return
	}
	h.reverseProxy.ServeHTTP(w, r)
}

// GraphQLProjectHandler validates the gateway JWT, then proxies to
// /graphql/org/{orgName}/project/{projectSlug}/.
func (h *Handler) GraphQLProjectHandler(w http.ResponseWriter, r *http.Request) {
	if !h.verifyBearerToken(w, r) {
		return
	}
	h.reverseProxy.ServeHTTP(w, r)
}

// verifyBearerToken extracts and validates the Authorization: Bearer <token> header.
// Returns true if valid, false after writing an error response.
func (h *Handler) verifyBearerToken(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
		return false
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if _, err := h.authService.VerifyAccessToken(tokenStr); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired access token")
		return false
	}
	return true
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"code":"` + code + `","message":"` + message + `"}`))
}
