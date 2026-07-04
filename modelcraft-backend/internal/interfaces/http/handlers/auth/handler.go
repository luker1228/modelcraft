package auth

import (
	"encoding/json"
	"modelcraft/internal/app/apitoken"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/config"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	httpmiddleware "modelcraft/internal/interfaces/http/middleware"

	appAuth "modelcraft/internal/app/auth"
)

// Handler handles HTTP auth endpoints.
type Handler struct {
	tokenService *appAuth.TokenService
	apiTokenSvc  *apitoken.APITokenService
	cookieCfg    config.CookieConfig
	isOrgAdminFn httpmiddleware.IsOrgAdminFn
}

// NewHandler creates a new auth Handler.
func NewHandler(
	tokenService *appAuth.TokenService,
	apiTokenSvc *apitoken.APITokenService,
	cookieCfg config.CookieConfig,
	isOrgAdminFn httpmiddleware.IsOrgAdminFn,
) *Handler {
	return &Handler{
		tokenService: tokenService,
		apiTokenSvc:  apiTokenSvc,
		cookieCfg:    cookieCfg,
		isOrgAdminFn: isOrgAdminFn,
	}
}

// RefreshCookieName is the unified httpOnly cookie name for refresh tokens.
// Both tenant and end-user auth share this cookie so the browser stores only one.
const RefreshCookieName = "mc_refresh_token"

const tenantRefreshCookieName = RefreshCookieName

// SetRefreshCookie writes the mc_refresh_token httpOnly cookie.
// Exported so that other handlers (e.g. enduser) can share the same cookie name and settings.
func (h *Handler) SetRefreshCookie(w http.ResponseWriter, token string) {
	h.setRefreshCookie(w, token)
}

// ClearRefreshCookie clears the mc_refresh_token httpOnly cookie.
// Exported so that other handlers (e.g. enduser) can share the same cookie name and settings.
func (h *Handler) ClearRefreshCookie(w http.ResponseWriter) {
	h.clearRefreshCookie(w)
}

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
		logfacade.GetLogger(r.Context()).With(logfacade.Err(err)).Warnf(r.Context(), "Invalid register request body")
		writeAuthError(w, http.StatusBadRequest, requestID, "PARAM_INVALID.AUTH", "Invalid request body")
		return
	}

	result, err := h.tokenService.Register(r.Context(), appAuth.RegisterCommand{
		Phone:            req.Phone,
		Password:         req.Password,
		UserName:         req.UserName,
		OrgDisplayName:   req.OrgDisplayName,
		OrganizationName: derefString(req.OrganizationName),
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
		logfacade.GetLogger(r.Context()).With(logfacade.Err(err)).Warnf(r.Context(), "Invalid login request body")
		writeAuthError(w, http.StatusBadRequest, requestID, "PARAM_INVALID.AUTH", "Invalid request body")
		return
	}

	// Build LoginCommand — userName + password
	cmd := appAuth.LoginCommand{
		UserName: req.UserName,
		Password: req.Password,
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

// HandlePATWhoami resolves identity for a validated PAT request.
// The PAT middleware authenticates the token and injects user/org fields into context.
func (h *Handler) HandlePATWhoami(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	authHeader := r.Header.Get(httpheader.Authorization)
	if strings.HasPrefix(authHeader, "Bearer mc_pat_") {
		h.handlePATWhoami(w, r, requestID, authHeader)
		return
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		h.handlePlatformWhoami(w, r, requestID, strings.TrimPrefix(authHeader, "Bearer "))
		return
	}

	logfacade.GetLogger(r.Context()).
		Warnf(r.Context(), "whoami rejected: invalid auth header requestId=%s", requestID)
	writeAuthError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "valid bearer token required")
}

func (h *Handler) handlePATWhoami(w http.ResponseWriter, r *http.Request, requestID, authHeader string) {
	if h.apiTokenSvc == nil {
		logfacade.GetLogger(r.Context()).Errorf(r.Context(), nil, "PAT whoami unavailable: api token service is nil")
		writeAuthError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	plaintext := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := h.apiTokenSvc.ValidateToken(r.Context(), plaintext)
	if err != nil || token == nil {
		msg := "valid PAT token required"
		if err != nil {
			msg = err.Error()
		}
		logfacade.GetLogger(r.Context()).
			Warnf(r.Context(), "PAT whoami validation failed requestId=%s err=%v", requestID, err)
		writeAuthError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", msg)
		return
	}

	go func() {
		if updateErr := h.apiTokenSvc.UpdateLastUsedAt(r.Context(), token.ID, time.Now()); updateErr != nil {
			logfacade.GetLogger(r.Context()).Warnf(r.Context(), "PAT whoami update last_used_at failed: %v", updateErr)
		}
	}()

	isAdmin := false
	if h.isOrgAdminFn != nil {
		if ok, adminErr := h.isOrgAdminFn(r.Context(), token.OrgName, token.EndUserID); adminErr == nil {
			isAdmin = ok
		}
	}

	// Issue a short-lived JWT so the gateway can replace the raw PAT bearer
	// token before forwarding to downstream GraphQL endpoints.
	var accessToken string
	if h.tokenService != nil {
		var jwtErr error
		accessToken, jwtErr = h.tokenService.IssueToken(token.EndUserID, token.OrgName, isAdmin)
		if jwtErr != nil {
			logfacade.GetLogger(r.Context()).Warnf(r.Context(),
				"PAT whoami jwt issue failed requestId=%s err=%v", requestID, jwtErr)
		}
	}

	memberships, membershipsErr := h.tokenService.GetUserMembershipSnapshots(r.Context(), token.EndUserID)
	if membershipsErr != nil {
		logfacade.GetLogger(r.Context()).Warnf(r.Context(),
			"PAT whoami memberships load failed requestId=%s err=%v", requestID, membershipsErr)
	}

	resp := map[string]any{
		"requestId":   requestID,
		"userId":      token.EndUserID,
		"orgName":     token.OrgName,
		"isAdmin":     isAdmin,
		"memberships": buildWhoamiMemberships(memberships),
	}
	if accessToken != "" {
		resp["accessToken"] = accessToken
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handlePlatformWhoami(w http.ResponseWriter, r *http.Request, requestID, token string) {
	claims, err := h.tokenService.ParsePlatformToken(token)
	if err != nil {
		logfacade.GetLogger(r.Context()).
			Warnf(r.Context(), "platform whoami validation failed requestId=%s err=%v", requestID, err)
		writeAuthError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "valid bearer token required")
		return
	}

	memberships, membershipsErr := h.tokenService.GetUserMembershipSnapshots(r.Context(), claims.UserID)
	if membershipsErr != nil {
		logfacade.GetLogger(r.Context()).
			Warnf(r.Context(), "platform whoami memberships load failed requestId=%s err=%v", requestID, membershipsErr)
	}

	resp := map[string]any{
		"requestId":   requestID,
		"userId":      claims.UserID,
		"orgName":     claims.OrgName,
		"isAdmin":     claims.IsAdmin,
		"memberships": buildWhoamiMemberships(memberships),
	}
	writeJSON(w, http.StatusOK, resp)
}

func buildWhoamiMemberships(items []appAuth.UserMembershipSnapshot) []map[string]any {
	memberships := make([]map[string]any, 0, len(items))
	for _, item := range items {
		memberships = append(memberships, map[string]any{
			"orgId":       item.OrgID,
			"orgName":     item.OrgName,
			"displayName": item.DisplayName,
			"role":        item.Role,
			"joinedAt":    item.JoinedAt.UnixMilli(),
		})
	}
	return memberships
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
		logfacade.GetLogger(r.Context()).Errorf(r.Context(), err, logMsg)
		writeAuthError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	// Log with stack at the Interfaces layer error conversion point
	logfacade.GetLogger(r.Context()).Errorf(r.Context(), err, logMsg)

	statusCode := bizErr.GetHTTPStatusCode()
	code := bizErr.Info().GetCode()
	msg := bizErr.Msg()

	writeAuthError(w, statusCode, requestID, code, msg)
}

// writeAuthError writes a structured error response matching the OpenAPI error schemas.
func writeAuthError(w http.ResponseWriter, statusCode int, requestID, code, message string) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
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
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// derefString dereferences a string pointer, returning "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// HandleDemoLogin handles POST /api/tenant/auth/demo — issues a guest JWT.
func (h *Handler) HandleDemoLogin(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	result, err := h.tokenService.DemoLogin(r.Context())
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Demo login failed")
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

	writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"userId":      result.UserID,
		"userName":    userName,
		"orgName":     orgName,
		"accessToken": result.AccessToken,
		"expiresIn":   result.ExpiresIn,
	})
}
