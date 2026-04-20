package enduser

import (
	"encoding/json"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"

	appEnduser "modelcraft/internal/app/enduser"
)

// AuthHandler 处理终端用户认证 HTTP 请求。
type AuthHandler struct {
	authService *appEnduser.EndUserAuthAppService
	logger      logfacade.Logger
}

// NewAuthHandler 创建认证 Handler。
func NewAuthHandler(
	authService *appEnduser.EndUserAuthAppService,
	logger logfacade.Logger,
) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

type authCredentialRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshLogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type authTokenResponse struct {
	RequestID    string `json:"requestId"`
	UserID       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt"`
}

type meResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
}

const (
	missingOrgProjectHeaders = "X-Org-Name and X-Project-Slug headers are required"
	missingEndUserHeaders    = "X-Org-Name, X-Project-Slug and X-End-User-Id headers are required"
)

// Register 处理 POST /internal/end-user/auth/register。
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req authCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgProjectHeaders)
		return
	}

	result, err := h.authService.RegisterEndUser(ctx, appEnduser.RegisterCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Username:    req.Username,
		Password:    req.Password,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user register failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Login 处理 POST /internal/end-user/auth/login。
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req authCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgProjectHeaders)
		return
	}

	result, err := h.authService.LoginEndUser(ctx, appEnduser.LoginCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Username:    req.Username,
		Password:    req.Password,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user login failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Logout 处理 POST /internal/end-user/auth/logout。
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req refreshLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgProjectHeaders)
		return
	}

	err := h.authService.LogoutEndUser(ctx, appEnduser.LogoutCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Refresh 处理 POST /internal/end-user/auth/refresh。
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req refreshLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgProjectHeaders)
		return
	}

	result, err := h.authService.RefreshEndUserToken(ctx, appEnduser.RefreshCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user refresh failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Me 处理 GET /internal/end-user/auth/me。
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	endUserID := r.Header.Get("X-End-User-Id")
	if orgName == "" || projectSlug == "" || endUserID == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingEndUserHeaders)
		return
	}

	user, err := h.authService.GetEndUserMe(ctx, appEnduser.GetMeCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		UserID:      endUserID,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user me failed")
		return
	}

	createdAt := user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	updatedAt := user.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")

	h.writeJSON(w, http.StatusOK, meResponse{
		RequestID: requestID,
		EndUser: &EndUserJSON{
			ID:          user.ID,
			Username:    user.Username,
			IsForbidden: user.IsForbidden,
			CreatedBy:   user.CreatedBy,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		},
	})
}

func (h *AuthHandler) handleBusinessError(
	w http.ResponseWriter,
	r *http.Request,
	requestID string,
	err error,
	logMsg string,
) {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
		h.writeError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
	h.writeError(w, bizErr.GetHTTPStatusCode(), requestID, bizErr.Info().GetCode(), bizErr.Msg())
}

func (h *AuthHandler) writeError(w http.ResponseWriter, statusCode int, requestID, code, message string) {
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

func (h *AuthHandler) writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
