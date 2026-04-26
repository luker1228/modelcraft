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

// Handler exposes auth endpoints: login, register, refresh, logout.
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

// loginRequest is the payload accepted by the gateway from the browser.
type loginRequest struct {
	Identifier     string `json:"identifier"`               // username or phone
	IdentifierType string `json:"identifierType,omitempty"` // "USERNAME" or empty
	Password       string `json:"password"`
}

// registerRequest is the payload accepted by the gateway from the browser.
type registerRequest struct {
	Phone    string `json:"phone"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// tokenResponse is what the gateway returns to the browser.
type tokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// ---- backend API types (matching modelcraft-backend generated structs) ----

type backendLoginRequest struct {
	Identifier     string `json:"identifier"`
	IdentifierType string `json:"identifierType,omitempty"`
	Password       string `json:"password"`
}

type backendLoginResponse struct {
	RequestID    string    `json:"requestId"`
	UserID       string    `json:"userId"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	UserName     string    `json:"userName,omitempty"`
	OrgName      string    `json:"orgName,omitempty"`
}

type backendRegisterRequest struct {
	Phone    string `json:"phone"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type backendRegisterResponse struct {
	RequestID string `json:"requestId"`
	UserID    string `json:"userId"`
	OrgName   string `json:"orgName"`
}

type backendRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type backendRefreshResponse struct {
	RequestID    string    `json:"requestId"`
	UserID       string    `json:"userId"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type backendLogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ---- handlers ----

// Login authenticates against the Go backend, issues a gateway JWT + sets refresh cookie.
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

	// Issue gateway short-lived access token.
	accessToken, err := h.authService.IssueAccessToken(backendResp.UserID, backendResp.UserName, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	// Store backend's opaque refresh token in httpOnly cookie.
	// Persistence is owned by the backend (DB); gateway only proxies the token via cookie.
	h.authService.SetRefreshCookie(w, backendResp.RefreshToken)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

// Register creates a new user via the Go backend, then issues tokens.
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

	// After register, perform a login to obtain a refresh token.
	loginResp, err := h.callBackendLogin(r.Context(), loginRequest{
		Identifier:     req.UserName,
		IdentifierType: "USERNAME",
		Password:       req.Password,
	})
	if err != nil {
		// Registration succeeded but auto-login failed — return userId so client can retry login.
		writeJSON(w, http.StatusCreated, map[string]string{
			"userId":  backendResp.UserID,
			"orgName": backendResp.OrgName,
		})
		return
	}

	accessToken, err := h.authService.IssueAccessToken(loginResp.UserID, loginResp.UserName, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	h.authService.SetRefreshCookie(w, loginResp.RefreshToken)
	writeJSON(w, http.StatusCreated, tokenResponse{AccessToken: accessToken})
}

// Refresh reads the httpOnly refresh cookie, forwards it to the backend for rotation,
// then issues a new gateway access token and updates the cookie.
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

	accessToken, err := h.authService.IssueAccessToken(backendResp.UserID, "", "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	// Update cookie with the rotated token from the backend.
	h.authService.SetRefreshCookie(w, backendResp.RefreshToken)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

// Logout forwards the refresh token to the backend for revocation, then clears the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, _ := h.authService.GetRefreshCookie(r)
	if refreshToken != "" {
		// Best-effort: tell the backend to revoke. Ignore errors — cookie is cleared regardless.
		_ = h.callBackendLogout(r.Context(), refreshToken)
	}
	h.authService.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// ---- backend client helpers ----

func (h *Handler) callBackendLogin(ctx context.Context, req loginRequest) (*backendLoginResponse, error) {
	body := backendLoginRequest{
		Identifier:     req.Identifier,
		IdentifierType: req.IdentifierType,
		Password:       req.Password,
	}
	var resp backendLoginResponse
	if err := h.postBackend(ctx, "/api/auth/login", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) callBackendRegister(ctx context.Context, req registerRequest) (*backendRegisterResponse, error) {
	body := backendRegisterRequest{
		Phone:    req.Phone,
		UserName: req.UserName,
		Password: req.Password,
	}
	var resp backendRegisterResponse
	if err := h.postBackend(ctx, "/api/auth/register", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) callBackendRefresh(ctx context.Context, token string) (*backendRefreshResponse, error) {
	body := backendRefreshRequest{RefreshToken: token}
	var resp backendRefreshResponse
	if err := h.postBackend(ctx, "/api/auth/refresh", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) callBackendLogout(ctx context.Context, token string) error {
	body := backendLogoutRequest{RefreshToken: token}
	return h.postBackend(ctx, "/api/auth/logout", body, nil)
}

// postBackend is a generic helper for posting JSON to the Go backend and decoding the response.
// Pass nil for out to ignore the response body (e.g. logout 204).
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
