package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ---- browser-facing response types ----

// loginTokenResponse is what the gateway returns to the browser after login/register.
// It omits refreshToken — that lives in the httpOnly cookie only.
type loginTokenResponse struct {
	RequestID   string `json:"requestId,omitempty"`
	UserID      string `json:"userId"`
	UserName    string `json:"userName,omitempty"`
	OrgName     string `json:"orgName,omitempty"`
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

// refreshTokenResponse is what the gateway returns to the browser after token refresh.
type refreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

// Handler exposes auth endpoints: login, register, refresh, logout.
// Token signing is fully handled by the backend auth service (ES256).
// The gateway's only responsibilities here are:
//   - Proxying requests to the backend
//   - Managing the httpOnly refresh-token cookie (browsers cannot access it directly)
type Handler struct {
	authService *Service
	backendURL  string
	httpClient  *http.Client
}

func NewHandler(authService *Service, backendURL string) *Handler {
	return &Handler{
		authService: authService,
		backendURL:  backendURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// ---- gateway request / response types ----

type loginRequest struct {
	Identifier     string `json:"identifier"`
	IdentifierType string `json:"identifierType,omitempty"`
	Password       string `json:"password"`
}

type registerRequest struct {
	Phone    string `json:"phone"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// ---- backend API response types ----

// backendLoginResponse mirrors the backend's login/register response body.
// refreshToken is present here so the gateway can extract it for the cookie;
// it is never forwarded to the browser.
type backendLoginResponse struct {
	RequestID    string `json:"requestId"`
	UserID       string `json:"userId"`
	UserName     string `json:"userName,omitempty"`
	OrgName      string `json:"orgName,omitempty"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type backendRegisterResponse struct {
	RequestID string `json:"requestId"`
	UserID    string `json:"userId"`
	OrgName   string `json:"orgName"`
}

// backendRefreshResponse mirrors the backend's refresh response body.
// refreshToken is extracted for cookie rotation; not forwarded to browser.
type backendRefreshResponse struct {
	RequestID    string `json:"requestId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type backendLogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ---- handlers ----

// Login proxies to the backend, stores the refreshToken in an httpOnly cookie,
// and returns the full login payload (userId, userName, orgName, accessToken, expiresIn)
// to the browser — refreshToken is stripped.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.Identifier == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "PARAM_INVALID", "identifier and password are required")
		return
	}

	backendResp, err := h.callBackendLogin(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "AUTH_FAILED", err.Error())
		return
	}

	h.authService.SetRefreshCookie(w, backendResp.RefreshToken)
	writeJSON(w, http.StatusOK, loginTokenResponse{
		RequestID:   backendResp.RequestID,
		UserID:      backendResp.UserID,
		UserName:    backendResp.UserName,
		OrgName:     backendResp.OrgName,
		AccessToken: backendResp.AccessToken,
		ExpiresIn:   backendResp.ExpiresIn,
	})
}

// Register creates a new user via the backend, then auto-logs-in to obtain tokens.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.UserName == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "PARAM_INVALID", "userName and password are required")
		return
	}

	backendResp, err := h.callBackendRegister(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REGISTER_FAILED", err.Error())
		return
	}

	// After register, perform a login to obtain tokens.
	loginResp, err := h.callBackendLogin(r.Context(), loginRequest{
		Identifier:     req.UserName,
		IdentifierType: "USERNAME",
		Password:       req.Password,
	})
	if err != nil {
		// Registration succeeded but auto-login failed — return userId so client can retry.
		writeJSON(w, http.StatusCreated, map[string]string{
			"userId":  backendResp.UserID,
			"orgName": backendResp.OrgName,
		})
		return
	}

	h.authService.SetRefreshCookie(w, loginResp.RefreshToken)
	writeJSON(w, http.StatusCreated, loginTokenResponse{
		RequestID:   loginResp.RequestID,
		UserID:      loginResp.UserID,
		UserName:    loginResp.UserName,
		OrgName:     loginResp.OrgName,
		AccessToken: loginResp.AccessToken,
		ExpiresIn:   loginResp.ExpiresIn,
	})
}

// Refresh reads the httpOnly refresh cookie, forwards it to the backend for rotation,
// then returns the new accessToken and updates the cookie.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.authService.GetRefreshCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_MISSING", "refresh token not found")
		return
	}

	backendResp, err := h.callBackendRefresh(r.Context(), refreshToken)
	if err != nil {
		h.authService.ClearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "REFRESH_INVALID", "invalid or expired refresh token")
		return
	}

	h.authService.SetRefreshCookie(w, backendResp.RefreshToken)
	writeJSON(w, http.StatusOK, refreshTokenResponse{
		AccessToken: backendResp.AccessToken,
		ExpiresIn:   backendResp.ExpiresIn,
	})
}

// Logout forwards the refresh token to the backend for revocation, then clears the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, _ := h.authService.GetRefreshCookie(r)
	if refreshToken != "" {
		_ = h.callBackendLogout(r.Context(), refreshToken)
	}

	h.authService.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// ---- backend client helpers ----

func (h *Handler) callBackendLogin(ctx context.Context, req loginRequest) (*backendLoginResponse, error) {
	var resp backendLoginResponse
	if err := h.postBackend(ctx, "/api/auth/login", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (h *Handler) callBackendRegister(ctx context.Context, req registerRequest) (*backendRegisterResponse, error) {
	var resp backendRegisterResponse
	if err := h.postBackend(ctx, "/api/auth/register", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (h *Handler) callBackendRefresh(ctx context.Context, token string) (*backendRefreshResponse, error) {
	body := map[string]string{"refreshToken": token}

	var resp backendRefreshResponse
	if err := h.postBackend(ctx, "/api/auth/refresh", body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (h *Handler) callBackendLogout(ctx context.Context, token string) error {
	return h.postBackend(ctx, "/api/auth/logout", backendLogoutRequest{RefreshToken: token}, nil)
}

func (h *Handler) postBackend(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.backendURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend error %d: %s", resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
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
