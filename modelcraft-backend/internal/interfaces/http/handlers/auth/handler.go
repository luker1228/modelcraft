package auth

import (
	"encoding/json"
	"io"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"

	appAuth "modelcraft/internal/app/auth"
)

// Handler handles HTTP auth endpoints.
type Handler struct {
	tokenService *appAuth.TokenService
	logger       logfacade.Logger
}

// NewHandler creates a new auth Handler.
func NewHandler(tokenService *appAuth.TokenService, logger logfacade.Logger) *Handler {
	return &Handler{
		tokenService: tokenService,
		logger:       logger,
	}
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

	// Build response with optional userName and orgName
	resp := generated.LoginResponse{
		RequestId:    requestID,
		UserId:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	}
	if result.UserName != "" {
		resp.UserName = &result.UserName
	}
	if result.OrgName != "" {
		resp.OrgName = &result.OrgName
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleRefresh handles POST /api/auth/refresh — token rotation.
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	var req generated.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(r.Context(), "Invalid refresh request body", logfacade.Err(err))
		writeAuthError(w, http.StatusBadRequest, requestID, "PARAM_INVALID.AUTH", "Invalid request body")
		return
	}

	result, err := h.tokenService.Refresh(r.Context(), appAuth.RefreshCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Refresh failed")
		return
	}

	writeJSON(w, http.StatusOK, generated.RefreshResponse{
		RequestId:    requestID,
		UserId:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
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
