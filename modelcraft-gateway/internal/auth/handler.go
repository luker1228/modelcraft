package auth

import (
	"encoding/json"
	"net/http"
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
		httpClient:  &http.Client{},
	}
}

// ---- request / response types ----

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// ---- handlers ----

// Login authenticates against the Go backend, issues a gateway JWT + sets refresh cookie.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "PARAM_INVALID", "username and password are required")
		return
	}

	// Forward credentials to Go backend's Casdoor auth endpoint.
	backendResp, err := h.callBackendLogin(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "AUTH_FAILED", err.Error())
		return
	}

	// Issue gateway access token.
	accessToken, err := h.authService.IssueAccessToken(backendResp.UserID, backendResp.Username, backendResp.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	// Generate opaque refresh token and set httpOnly cookie.
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to generate refresh token")
		return
	}

	// TODO: persist refreshToken → userID mapping in a store (Redis / DB).
	// For now the refresh token is stored as-is; replace with a proper store in Task #2.
	_ = refreshToken

	h.authService.SetRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

// Register creates a new user via the Go backend, then issues tokens.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	backendResp, err := h.callBackendRegister(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REGISTER_FAILED", err.Error())
		return
	}

	accessToken, err := h.authService.IssueAccessToken(backendResp.UserID, backendResp.Username, backendResp.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to generate refresh token")
		return
	}
	_ = refreshToken // TODO: persist

	h.authService.SetRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusCreated, tokenResponse{AccessToken: accessToken})
}

// Refresh reads the httpOnly refresh cookie and issues a new access token.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.authService.GetRefreshCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_MISSING", "refresh token not found")
		return
	}

	// TODO: validate refreshToken against store, look up user, rotate token.
	// Placeholder: decode as opaque token and re-issue for the same user.
	userID, username, email, err := h.lookupRefreshToken(refreshToken)
	if err != nil {
		h.authService.ClearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "REFRESH_INVALID", "invalid or expired refresh token")
		return
	}

	accessToken, err := h.authService.IssueAccessToken(userID, username, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to issue access token")
		return
	}

	// Rotate refresh token.
	newRefresh, err := GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "failed to rotate refresh token")
		return
	}
	_ = newRefresh // TODO: persist rotation

	h.authService.SetRefreshCookie(w, newRefresh)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

// Logout clears the refresh cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authService.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

type backendUserInfo struct {
	UserID   string
	Username string
	Email    string
}

// callBackendLogin forwards credentials to the Go backend.
// The actual implementation will call /api/auth/token on the Go backend.
func (h *Handler) callBackendLogin(_ interface{}, _, _ string) (*backendUserInfo, error) {
	// TODO: implement actual backend call in Task #2
	// This is a stub so the gateway compiles as a skeleton.
	return nil, http.ErrNoCookie // placeholder error
}

// callBackendRegister forwards registration to the Go backend.
func (h *Handler) callBackendRegister(_ interface{}, _ registerRequest) (*backendUserInfo, error) {
	// TODO: implement in Task #2
	return nil, http.ErrNoCookie // placeholder
}

// lookupRefreshToken validates an opaque refresh token and returns its associated user.
func (h *Handler) lookupRefreshToken(_ string) (userID, username, email string, err error) {
	// TODO: implement token store lookup in Task #2
	return "", "", "", http.ErrNoCookie // placeholder
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
