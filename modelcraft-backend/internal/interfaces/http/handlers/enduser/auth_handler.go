package enduser

import (
	"encoding/json"
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
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

// publicAuthCredentialRequest is used by public endpoints where orgName comes from the request body.
type publicAuthCredentialRequest struct {
	OrgName  string `json:"orgName"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshLogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// publicRefreshLogoutRequest is used by public endpoints where orgName comes from the request body.
type publicRefreshLogoutRequest struct {
	OrgName      string `json:"orgName"`
	RefreshToken string `json:"refreshToken"`
}

type selectProjectRequest struct {
	RefreshToken string `json:"refreshToken"`
	ProjectSlug  string `json:"projectSlug"`
}

type accessibleProjectResponse struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type authTokenResponse struct {
	RequestID       string                      `json:"requestId"`
	UserID          string                      `json:"userId"`
	AccessToken     string                      `json:"accessToken,omitempty"`
	Projects        []accessibleProjectResponse `json:"projects,omitempty"`
	SelectedProject string                      `json:"selectedProject,omitempty"`
	RefreshToken    string                      `json:"refreshToken,omitempty"`
	ExpiresAt       string                      `json:"expiresAt,omitempty"`
}

type meResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
}

const (
	missingOrgProjectHeaders = "X-Org-Name and X-Project-Slug headers are required"
	missingOrgHeader         = "X-Org-Name header is required"
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
	if orgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgHeader)
		return
	}

	result, err := h.authService.LoginEndUser(ctx, appEnduser.LoginCommand{
		OrgName:  orgName,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user login failed")
		return
	}

	projects := make([]accessibleProjectResponse, 0, len(result.Projects))
	for _, item := range result.Projects {
		projects = append(projects, accessibleProjectResponse{Slug: item.Slug, Title: item.Title})
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		Projects:     projects,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// SelectProject 处理 POST /internal/end-user/auth/select-project。
func (h *AuthHandler) SelectProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req selectProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.RefreshToken == "" || req.ProjectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "refreshToken and projectSlug are required")
		return
	}

	orgName := r.Header.Get("X-Org-Name")
	if orgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgHeader)
		return
	}

	result, err := h.authService.SelectProjectContext(ctx, appEnduser.SelectProjectCommand{
		OrgName:      orgName,
		ProjectSlug:  req.ProjectSlug,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user select project failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:       requestID,
		UserID:          result.UserID,
		SelectedProject: result.ProjectSlug,
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
	if orgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgHeader)
		return
	}

	err := h.authService.LogoutEndUser(ctx, appEnduser.LogoutCommand{
		OrgName:      orgName,
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
	if orgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", missingOrgHeader)
		return
	}

	result, err := h.authService.RefreshEndUserToken(ctx, appEnduser.RefreshCommand{
		OrgName:      orgName,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user refresh failed")
		return
	}

	projects := make([]accessibleProjectResponse, 0, len(result.Projects))
	for _, item := range result.Projects {
		projects = append(projects, accessibleProjectResponse{Slug: item.Slug, Title: item.Title})
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		Projects:     projects,
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

// ============================================================
// Public EndUser Auth Endpoints (no X-Internal-Token, orgName in body)
// ============================================================

// PublicLogin 处理 POST /api/end-user/auth/login。
// orgName 从请求 body 读取，无需 X-Internal-Token。
func (h *AuthHandler) PublicLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req publicAuthCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}

	result, err := h.authService.LoginEndUser(ctx, appEnduser.LoginCommand{
		OrgName:  req.OrgName,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user public login failed")
		return
	}

	projects := make([]accessibleProjectResponse, 0, len(result.Projects))
	for _, item := range result.Projects {
		projects = append(projects, accessibleProjectResponse{Slug: item.Slug, Title: item.Title})
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		Projects:     projects,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// PublicRegister 处理 POST /api/end-user/auth/register。
// orgName 和 projectSlug 从请求 body 读取，无需 X-Internal-Token。
func (h *AuthHandler) PublicRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName     string `json:"orgName"`
		ProjectSlug string `json:"projectSlug"`
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.OrgName == "" || req.ProjectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName and projectSlug are required")
		return
	}

	result, err := h.authService.RegisterEndUser(ctx, appEnduser.RegisterCommand{
		OrgName:     req.OrgName,
		ProjectSlug: req.ProjectSlug,
		Username:    req.Username,
		Password:    req.Password,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user public register failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// PublicLogout 处理 POST /api/end-user/auth/logout。
// orgName 从请求 body 读取，无需 X-Internal-Token。
func (h *AuthHandler) PublicLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req publicRefreshLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}

	err := h.authService.LogoutEndUser(ctx, appEnduser.LogoutCommand{
		OrgName:      req.OrgName,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user public logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PublicRefresh 处理 POST /api/end-user/auth/refresh。
// orgName 从请求 body 读取，无需 X-Internal-Token。
func (h *AuthHandler) PublicRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req publicRefreshLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}

	result, err := h.authService.RefreshEndUserToken(ctx, appEnduser.RefreshCommand{
		OrgName:      req.OrgName,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user public refresh failed")
		return
	}

	projects := make([]accessibleProjectResponse, 0, len(result.Projects))
	for _, item := range result.Projects {
		projects = append(projects, accessibleProjectResponse{Slug: item.Slug, Title: item.Title})
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:    requestID,
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		Projects:     projects,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// PublicSelectProject 处理 POST /api/end-user/auth/select-project。
// orgName 从请求 body 读取，无需 X-Internal-Token。
func (h *AuthHandler) PublicSelectProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName      string `json:"orgName"`
		RefreshToken string `json:"refreshToken"`
		ProjectSlug  string `json:"projectSlug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}
	if req.OrgName == "" || req.RefreshToken == "" || req.ProjectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID",
			"orgName, refreshToken and projectSlug are required")
		return
	}

	result, err := h.authService.SelectProjectContext(ctx, appEnduser.SelectProjectCommand{
		OrgName:      req.OrgName,
		ProjectSlug:  req.ProjectSlug,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user public select-project failed")
		return
	}

	h.writeJSON(w, http.StatusOK, authTokenResponse{
		RequestID:       requestID,
		UserID:          result.UserID,
		SelectedProject: result.ProjectSlug,
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
