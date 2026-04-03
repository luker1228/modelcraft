package auth

import (
	"encoding/json"
	"io"
	appAuth "modelcraft/internal/app/auth"
	"modelcraft/pkg/logfacade"
	"net/http"
)

// Handler handles HTTP auth endpoints.
type Handler struct {
	casdoorURL   string
	clientID     string
	clientSecret string
	redirectURI  string
	tokenService *appAuth.TokenService
	logger       logfacade.Logger
}

// Config holds Casdoor OAuth configuration (used for GetLoginURL).
type Config struct {
	CasdoorURL   string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// NewHandler creates a new auth Handler.
func NewHandler(config Config, tokenService *appAuth.TokenService, logger logfacade.Logger) *Handler {
	return &Handler{
		casdoorURL:   config.CasdoorURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		redirectURI:  config.RedirectURI,
		tokenService: tokenService,
		logger:       logger,
	}
}

// HandleLogin handles POST /api/auth/login — called by BFF after Casdoor OAuth.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.tokenService.Login(r.Context(), appAuth.LoginCommand{
		ExternalID: req.ExternalID,
		Email:      req.Email,
		Name:       req.Name,
	})
	if err != nil {
		h.logger.Error(r.Context(), "Login failed", logfacade.Err(err), logfacade.Stack(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// HandleRefresh handles POST /api/auth/refresh — token rotation.
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.tokenService.Refresh(r.Context(), appAuth.RefreshCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.logger.Error(r.Context(), "Refresh failed", logfacade.Err(err), logfacade.Stack(err))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, refreshResponse{
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// HandleLogout handles POST /api/auth/logout — revokes refresh token.
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = h.tokenService.Logout(r.Context(), appAuth.LogoutCommand{
		RefreshToken: req.RefreshToken,
	})
	w.WriteHeader(http.StatusNoContent)
}

// GetLoginURL handles GET /api/auth/login-url — returns Casdoor OAuth login URL.
func (h *Handler) GetLoginURL(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	loginURL := h.casdoorURL + "/login/oauth/authorize?client_id=" + h.clientID +
		"&redirect_uri=" + h.redirectURI + "&response_type=code&scope=read&state=" + state
	writeJSON(w, http.StatusOK, map[string]string{"loginUrl": loginURL})
}

// ==================== Request/Response types ====================

type loginRequest struct {
	ExternalID string `json:"externalId"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

type loginResponse struct {
	UserID       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshResponse struct {
	UserID       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt"`
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
