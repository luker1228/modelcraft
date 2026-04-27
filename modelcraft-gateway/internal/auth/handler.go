package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Handler exposes auth endpoints: login, register, refresh, logout.
// Token signing is fully handled by the backend auth service (ES256).
// The gateway's only responsibilities here are:
//   - Proxying requests to the backend
//   - Managing the httpOnly refresh-token cookie (browsers cannot access it directly)
//
// All request validation is delegated to the backend.
type Handler struct {
	authService *Service
	backendURL  string
	httpClient  *http.Client
}

func NewHandler(authSvc *Service, backendURL string, httpClient *http.Client) *Handler {
	return &Handler{
		authService: authSvc,
		backendURL:  backendURL,
		httpClient:  httpClient,
	}
}

// ---- handlers ----

// Login proxies the request body to the backend, extracts refreshToken from the
// response into an httpOnly cookie, then forwards the remaining fields to the browser.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	raw, err := h.postBackendRaw(r.Context(), "/api/auth/login", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}

	h.extractRefreshAndProxy(w, http.StatusOK, raw)
}

// Register proxies the request body to the backend, then auto-logs-in to obtain tokens.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// Read body once — we need it for both register and the follow-up login.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "failed to read request body")
		return
	}

	if _, err = h.postBackendRaw(r.Context(), "/api/auth/register", bytes.NewReader(body)); err != nil {
		proxyBackendError(w, err)
		return
	}

	// Extract userName+password from the original body to perform auto-login.
	var req struct {
		UserName string `json:"userName"`
		Password string `json:"password"`
	}
	if err = json.Unmarshal(body, &req); err != nil || req.UserName == "" {
		// Registration succeeded but can't auto-login — let client retry.
		writeJSON(w, http.StatusCreated, map[string]string{"message": "registered, please login"})
		return
	}

	loginBody, _ := json.Marshal(map[string]string{
		"identifier":     req.UserName,
		"identifierType": "USERNAME",
		"password":       req.Password,
	})
	loginRaw, err := h.postBackendRaw(r.Context(), "/api/auth/login", bytes.NewReader(loginBody))
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]string{"message": "registered, please login"})
		return
	}

	h.extractRefreshAndProxy(w, http.StatusCreated, loginRaw)
}

// Refresh reads the httpOnly refresh cookie, injects it into the backend request,
// extracts the new refreshToken into the cookie, and proxies the rest to the browser.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.authService.GetRefreshCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_MISSING", "refresh token not found")
		return
	}

	reqBody, _ := json.Marshal(map[string]string{"refreshToken": refreshToken})
	raw, err := h.postBackendRaw(r.Context(), "/api/auth/refresh", bytes.NewReader(reqBody))
	if err != nil {
		h.authService.ClearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "REFRESH_INVALID", "invalid or expired refresh token")
		return
	}

	h.extractRefreshAndProxy(w, http.StatusOK, raw)
}

// Logout forwards the refresh token to the backend for revocation, then clears the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, _ := h.authService.GetRefreshCookie(r)
	if refreshToken != "" {
		reqBody, _ := json.Marshal(map[string]string{"refreshToken": refreshToken})
		_, _ = h.postBackendRaw(r.Context(), "/api/auth/logout", bytes.NewReader(reqBody))
	}

	h.authService.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// ---- proxy helpers ----

// extractRefreshAndProxy extracts "refreshToken" from the raw JSON body,
// stores it in the httpOnly cookie, then writes the remaining fields to the browser.
func (h *Handler) extractRefreshAndProxy(w http.ResponseWriter, status int, raw []byte) {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		writeError(w, http.StatusBadGateway, "INVALID_UPSTREAM", "upstream returned invalid JSON")
		return
	}

	if rt, ok := body["refreshToken"].(string); ok && rt != "" {
		h.authService.SetRefreshCookie(w, rt)
	}
	delete(body, "refreshToken")

	writeJSON(w, status, body)
}

// proxyBackendError forwards a backend error response to the browser.
// It preserves the original HTTP status code when available.
func proxyBackendError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
}

// ---- backend client ----

// postBackendRaw POSTs to the backend and returns the raw response body.
// The body parameter is an io.Reader so callers can pass either a struct-marshaled
// payload or a forwarded request body directly.
func (h *Handler) postBackendRaw(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.backendURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Propagate request ID for end-to-end tracing.
	if reqID := chiMiddleware.GetReqID(ctx); reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("backend request: %w", err)
	}

	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read response: %w", readErr)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("backend error %d: %s", resp.StatusCode, string(raw))
	}

	return raw, nil
}

// ---- response utilities ----

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}
