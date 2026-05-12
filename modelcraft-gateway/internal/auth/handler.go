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
	authService   *Service
	backendURL    string
	httpClient    *http.Client
	internalToken string
}

func NewHandler(authSvc *Service, backendURL string, httpClient *http.Client, internalToken string) *Handler {
	return &Handler{
		authService:   authSvc,
		backendURL:    backendURL,
		httpClient:    httpClient,
		internalToken: internalToken,
	}
}

// ---- handlers ----

// Login proxies the request body to the backend, extracts refreshToken from the
// response into an httpOnly cookie, then forwards the remaining fields to the browser.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/tenant/auth/login", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
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

	regRaw, regStatus, regErr := h.postBackendRaw(r.Context(), "/api/tenant/auth/register", bytes.NewReader(body))
	if regErr != nil {
		proxyBackendError(w, regErr)
		return
	}
	if regStatus >= 400 {
		writeRaw(w, regStatus, regRaw)
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
	loginRaw, loginStatus, loginErr := h.postBackendRaw(r.Context(), "/api/tenant/auth/login", bytes.NewReader(loginBody))
	if loginErr != nil || loginStatus >= 400 {
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
	raw, status, err := h.postBackendRaw(r.Context(), "/api/tenant/auth/refresh", bytes.NewReader(reqBody))
	if err != nil || status >= 400 {
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
		_, _, _ = h.postBackendRaw(r.Context(), "/api/tenant/auth/logout", bytes.NewReader(reqBody))
	}

	h.authService.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// EndUserLogin proxies to backend /api/end-user/auth/login and manages end-user refresh cookie.
func (h *Handler) EndUserLogin(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/login", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	h.extractEndUserRefreshAndProxy(w, http.StatusOK, raw)
}

// EndUserRegister proxies to backend /api/end-user/auth/register and manages end-user refresh cookie.
func (h *Handler) EndUserRegister(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/register", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	h.extractEndUserRefreshAndProxy(w, http.StatusOK, raw)
}

// EndUserRefresh reads end-user refresh cookie and proxies to backend /api/end-user/auth/refresh.
func (h *Handler) EndUserRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.authService.GetEndUserRefreshCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_MISSING", "refresh token not found")
		return
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}
	req["refreshToken"] = refreshToken

	body, _ := json.Marshal(req)
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/refresh", bytes.NewReader(body))
	if err != nil || status >= 400 {
		h.authService.ClearEndUserRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "REFRESH_INVALID", "invalid or expired refresh token")
		return
	}

	h.extractEndUserRefreshAndProxy(w, http.StatusOK, raw)
}

// EndUserLogout reads end-user refresh cookie, proxies to backend /api/end-user/auth/logout, then clears cookie.
func (h *Handler) EndUserLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken, _ := h.authService.GetEndUserRefreshCookie(r)
	if refreshToken != "" {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req = map[string]any{}
		}
		req["refreshToken"] = refreshToken

		body, _ := json.Marshal(req)
		_, _, _ = h.postBackendRaw(r.Context(), "/api/end-user/auth/logout", bytes.NewReader(body))
	}

	h.authService.ClearEndUserRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// EndUserSelectProject reads end-user refresh cookie and proxies to backend /api/end-user/auth/select-project.
func (h *Handler) EndUserSelectProject(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.authService.GetEndUserRefreshCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_MISSING", "refresh token not found")
		return
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}
	req["refreshToken"] = refreshToken

	body, _ := json.Marshal(req)
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/select-project", bytes.NewReader(body))
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// EndUserMe proxies GET /api/end-user/auth/me to the backend.
// The backend resolves identity entirely from the Bearer JWT — no extra headers needed.
func (h *Handler) EndUserMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
		return
	}

	raw, err := h.getBackendRaw(r.Context(), "/api/end-user/auth/me", func(req *http.Request) {
		req.Header.Set("Authorization", authHeader)
	})
	if err != nil {
		proxyBackendError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// CLIEndUserLogin proxies to backend /api/end-user/auth/login and returns raw JSON token payload.
// No cookies are set for CLI routes.
func (h *Handler) CLIEndUserLogin(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/login", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// CLIEndUserRefresh proxies to backend /api/end-user/auth/refresh and returns raw JSON token payload.
func (h *Handler) CLIEndUserRefresh(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/refresh", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// CLIEndUserLogout proxies to backend /api/end-user/auth/logout. No cookie operations are performed.
func (h *Handler) CLIEndUserLogout(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/logout", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CLIEndUserSelectProject proxies to backend /api/end-user/auth/select-project and returns raw JSON payload.
func (h *Handler) CLIEndUserSelectProject(w http.ResponseWriter, r *http.Request) {
	raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/select-project", r.Body)
	if err != nil {
		proxyBackendError(w, err)
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// CLIEndUserMe proxies GET /api/end-user/auth/me with Bearer token forwarding.
func (h *Handler) CLIEndUserMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
		return
	}

	raw, err := h.getBackendRaw(r.Context(), "/api/end-user/auth/me", func(req *http.Request) {
		req.Header.Set("Authorization", authHeader)
	})
	if err != nil {
		proxyBackendError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

// ---- proxy helpers ----

// extractRefreshAndProxy extracts "refreshToken" from the raw JSON body,
// stores it in the httpOnly cookie, then writes the remaining fields to the browser.
func (h *Handler) extractRefreshAndProxy(w http.ResponseWriter, status int, raw []byte) {
	h.extractRefreshAndProxyWithCookieSetter(w, status, raw, h.authService.SetRefreshCookie)
}

func (h *Handler) extractEndUserRefreshAndProxy(w http.ResponseWriter, status int, raw []byte) {
	h.extractRefreshAndProxyWithCookieSetter(w, status, raw, h.authService.SetEndUserRefreshCookie)
}

func (h *Handler) extractRefreshAndProxyWithCookieSetter(
	w http.ResponseWriter,
	status int,
	raw []byte,
	setCookie func(http.ResponseWriter, string),
) {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		writeError(w, http.StatusBadGateway, "INVALID_UPSTREAM", "upstream returned invalid JSON")
		return
	}

	if rt, ok := body["refreshToken"].(string); ok && rt != "" {
		setCookie(w, rt)
	}
	delete(body, "refreshToken")

	writeJSON(w, status, body)
}

// proxyBackendError writes a 502 for network/infra failures.
func proxyBackendError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
}

// ---- backend client ----

// postBackendRaw POSTs to the backend and returns (body, statusCode, error).
// Business errors (4xx) are returned as (body, statusCode, nil) so callers can
// decide whether to proxy them directly. Only network/infra failures return a non-nil error.
func (h *Handler) postBackendRaw(ctx context.Context, path string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.backendURL+path, body)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if reqID := chiMiddleware.GetReqID(ctx); reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("backend request: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", readErr)
	}

	return raw, resp.StatusCode, nil
}

// getBackendRaw performs a GET to the backend and returns the raw response body.
// The setHeaders callback allows callers to inject additional request headers.
func (h *Handler) getBackendRaw(ctx context.Context, path string, setHeaders func(*http.Request)) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.backendURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if setHeaders != nil {
		setHeaders(req)
	}

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

// writeRaw writes raw bytes directly to the response with the given status.
func writeRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
