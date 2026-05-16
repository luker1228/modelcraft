package auth

import (
	"encoding/json"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/config"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"

	appAuth "modelcraft/internal/app/auth"
)

// Handler handles HTTP auth endpoints.
type Handler struct {
	tokenService *appAuth.TokenService
	cookieCfg    config.CookieConfig
	logger       logfacade.Logger
}

// NewHandler creates a new auth Handler.
func NewHandler(tokenService *appAuth.TokenService, cookieCfg config.CookieConfig, logger logfacade.Logger) *Handler {
	return &Handler{
		tokenService: tokenService,
		cookieCfg:    cookieCfg,
		logger:       logger,
	}
}

const tenantRefreshCookieName = "mc_refresh_token"

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteStrictMode
	switch h.cookieCfg.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     tenantRefreshCookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		HttpOnly: true,
		Secure:   h.cookieCfg.Secure,
		SameSite: sameSite,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     tenantRefreshCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		HttpOnly: true,
		Secure:   h.cookieCfg.Secure,
		MaxAge:   -1,
	})
}

// HandleRegister handles POST /api/auth/register — phone+userName+password registration.
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	var req generated.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(r.Context(), "Invalid register request body", logfacade.Err(err))
		writeAuthError(w, http.StatusBadRequest, requestID, "PARAM_INVALID.AUTH", "Invalid request body")
		return
	}

	result, err := h.tokenService.Register(r.Context(), appAuth.RegisterCommand{
		Phone:    req.Phone,
		Password: req.Password,
		UserName: req.UserName,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Register failed")
		return
	}

	writeJSON(w, http.StatusCreated, generated.RegisterResponse{
		RequestId: requestID,
		UserId:    result.UserID,
		OrgName:   result.OrgName,
		Profile: generated.RegisterProfileSnapshot{
			Id:        result.Profile.ID,
			UserId:    result.Profile.UserID,
			Nickname:  result.Profile.Nickname,
			AvatarUrl: result.Profile.AvatarURL,
			Bio:       result.Profile.Bio,
		},
	})
}

// HandleLogin handles POST /api/auth/login — supports phone or username login.
// The refresh token is stored in an httpOnly cookie and NOT returned in the response body.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	var req generated.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(r.Context(), "Invalid login request body", logfacade.Err(err))
		writeAuthError(w, http.StatusBadRequest, requestID, "PARAM_INVALID.AUTH", "Invalid request body")
		return
	}

	// Build LoginCommand with new identifier-based fields
	cmd := appAuth.LoginCommand{
		Identifier: req.Identifier,
		Password:   req.Password,
	}

	// Map identifierType from generated enum to app enum
	if req.IdentifierType != nil {
		switch *req.IdentifierType {
		case generated.USERNAME:
			cmd.IdentifierType = appAuth.IdentifierTypeUsername
		default:
			cmd.IdentifierType = appAuth.IdentifierTypePhone
		}
	} else {
		cmd.IdentifierType = appAuth.IdentifierTypePhone
	}

	// Backward compat: if identifier is empty but phone is provided, use phone
	if cmd.Identifier == "" && req.Phone != nil {
		cmd.Phone = *req.Phone
	}

	result, err := h.tokenService.Login(r.Context(), cmd)
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Login failed")
		return
	}

	var userName *string
	if result.UserName != "" {
		s := result.UserName
		userName = &s
	}

	var orgName *string
	if result.OrgName != "" {
		s := result.OrgName
		orgName = &s
	}

	h.setRefreshCookie(w, result.RefreshToken)

	writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"userId":      result.UserID,
		"userName":    userName,
		"orgName":     orgName,
		"accessToken": result.AccessToken,
		"expiresIn":   result.ExpiresIn,
		// refreshToken intentionally omitted — stored in httpOnly cookie
	})
}

// HandleRefresh handles POST /api/auth/refresh — token rotation.
// The refresh token is read from the httpOnly cookie set at login.
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	cookie, err := r.Cookie(tenantRefreshCookieName)
	if err != nil || cookie.Value == "" {
		writeAuthError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
		return
	}

	result, err := h.tokenService.Refresh(r.Context(), appAuth.RefreshCommand{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.clearRefreshCookie(w)
		h.handleBusinessError(w, r, requestID, err, "Refresh failed")
		return
	}

	h.setRefreshCookie(w, result.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"accessToken": result.AccessToken,
		"expiresIn":   result.ExpiresIn,
	})
}

// HandleLogout handles POST /api/auth/logout — revokes refresh token.
// The refresh token is read from the httpOnly cookie and then the cookie is cleared.
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(tenantRefreshCookieName)
	if cookie != nil && cookie.Value != "" {
		_ = h.tokenService.Logout(r.Context(), appAuth.LogoutCommand{
			RefreshToken: cookie.Value,
		})
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// handleBusinessError maps a BusinessError to the appropriate HTTP error response.
// This is the error conversion point — logfacade.Stack(err) is logged here per architecture rules.
func (h *Handler) handleBusinessError(
	w http.ResponseWriter,
	r *http.Request,
	requestID string,
	err error,
	logMsg string,
) {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		// Not a BusinessError — unexpected system error
		h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
		writeAuthError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	// Log with stack at the Interfaces layer error conversion point
	h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))

	statusCode := bizErr.GetHTTPStatusCode()
	code := bizErr.Info().GetCode()
	msg := bizErr.Msg()

	writeAuthError(w, statusCode, requestID, code, msg)
}

// writeAuthError writes a structured error response matching the OpenAPI error schemas.
func writeAuthError(w http.ResponseWriter, statusCode int, requestID, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requestId": requestID,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
