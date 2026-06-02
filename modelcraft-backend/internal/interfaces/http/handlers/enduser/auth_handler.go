package enduser

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainAuth "modelcraft/internal/domain/auth"

	appAuth "modelcraft/internal/app/auth"
	appEnduser "modelcraft/internal/app/enduser"
	authhandler "modelcraft/internal/interfaces/http/handlers/auth"
)

// AuthHandler handles end-user authentication HTTP requests.
// Delegates all business logic to the unified TokenService (appAuth.TokenService).
type AuthHandler struct {
	authService *appAuth.TokenService
	endUserSvc  *appEnduser.EndUserManagementAppService
	jwtSigner   *domainAuth.JWTSigner
	shared      *authhandler.Handler // cookie set/clear delegated here
	logger      logfacade.Logger
}

// NewAuthHandler creates an AuthHandler.
// shared is the unified auth handler used only for cookie management.
func NewAuthHandler(
	authService *appAuth.TokenService,
	endUserSvc *appEnduser.EndUserManagementAppService,
	jwtSigner *domainAuth.JWTSigner,
	shared *authhandler.Handler,
	logger logfacade.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		endUserSvc:  endUserSvc,
		jwtSigner:   jwtSigner,
		shared:      shared,
		logger:      logger,
	}
}

// ============================================================
// OpenAPI-generated ServerInterface methods
// ============================================================

// EndUserLogin handles POST /api/end-user/auth/login.
func (h *AuthHandler) EndUserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName        string `json:"orgName"`
		Username       string `json:"username"`       // 保留向后兼容
		Identifier     string `json:"identifier"`     // 新增：手机号或用户名
		IdentifierType string `json:"identifierType"` // 新增：USERNAME | PHONE
		Password       string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}
	result, err := h.authService.LoginEndUser(ctx, appAuth.LoginEndUserCommand{
		OrgName:        req.OrgName,
		Username:       req.Username,
		Identifier:     req.Identifier,
		IdentifierType: appAuth.IdentifierType(req.IdentifierType),
		Password:       req.Password,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "end-user login failed")
		return
	}

	h.shared.SetRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"userId":      result.UserID,
		"orgName":     result.OrgName,
		"accessToken": result.AccessToken,
		"expiresAt":   result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// EndUserLogout handles POST /api/end-user/auth/logout.
func (h *AuthHandler) EndUserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, _ := r.Cookie(authhandler.RefreshCookieName)
	if cookie != nil && cookie.Value != "" {
		_ = h.authService.Logout(ctx, appAuth.LogoutCommand{
			RefreshToken: cookie.Value,
		})
	}

	h.shared.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// EndUserRefreshToken handles POST /api/end-user/auth/refresh.
func (h *AuthHandler) EndUserRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	cookie, err := r.Cookie(authhandler.RefreshCookieName)
	if err != nil || cookie.Value == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
		return
	}

	var req struct {
		OrgName string `json:"orgName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}

	result, err := h.authService.RefreshEndUserToken(ctx, appAuth.RefreshEndUserCommand{
		OrgName:      req.OrgName,
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.shared.ClearRefreshCookie(w)
		h.handleBizError(w, r, requestID, err, "end-user refresh failed")
		return
	}

	h.shared.SetRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"userId":      result.UserID,
		"orgName":     result.OrgName,
		"accessToken": result.AccessToken,
		"expiresAt":   result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// EndUserMe handles GET /api/end-user/auth/me.
// Identity is resolved entirely from the Bearer JWT (ES256-verified).
func (h *AuthHandler) EndUserMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	tokenStr := extractBearer(r)
	if tokenStr == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "UNAUTHORIZED", "Authorization header required")
		return
	}

	claims, err := h.parseEndUserJWT(tokenStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, requestID, "UNAUTHORIZED", "invalid or expired token")
		return
	}

	user, err := h.authService.GetEndUserMe(ctx, appAuth.GetEndUserMeCommand{
		OrgName: claims.OrgName,
		UserID:  claims.Subject,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "end-user me failed")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"requestId": requestID,
		"endUser": map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"isForbidden": user.IsForbidden,
			"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

// ============================================================
// JWT helper
// ============================================================

// parseEndUserJWT uses the ES256 public key to verify an end-user Bearer token and parse PlatformClaims.
func (h *AuthHandler) parseEndUserJWT(tokenStr string) (*domainAuth.PlatformClaims, error) {
	pubKeyPEM, err := h.jwtSigner.PublicKeyPEM()
	if err != nil {
		return nil, fmt.Errorf("get public key: %w", err)
	}
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	token, err := jwt.ParseWithClaims(tokenStr, &domainAuth.PlatformClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pub, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*domainAuth.PlatformClaims)
	if !ok || claims.OrgName == "" || claims.Subject == "" {
		return nil, fmt.Errorf("missing required claims")
	}
	return claims, nil
}

func extractBearer(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if strings.HasPrefix(v, "Bearer ") {
		return strings.TrimPrefix(v, "Bearer ")
	}
	return ""
}

// ============================================================
// CLI-specific handlers (no httpOnly cookie — token in body)
// ============================================================

// CLILogin handles POST /api/cli/end-user/auth/login.
// Unlike the browser login endpoint, this returns refreshToken in the JSON body
// (no httpOnly cookie) so that the CLI can persist it locally.
// It also returns the list of accessible projects.
func (h *AuthHandler) CLILogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName  string `json:"orgName"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}

	result, err := h.authService.LoginEndUser(ctx, appAuth.LoginEndUserCommand{
		OrgName:  req.OrgName,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "cli end-user login failed")
		return
	}

	// Fetch accessible projects so the CLI can display them after login.
	var projects []map[string]any
	if h.endUserSvc != nil {
		items, projErr := h.endUserSvc.ListAccessibleProjects(ctx, result.OrgName, result.UserID)
		if projErr == nil {
			for _, p := range items {
				projects = append(projects, map[string]any{
					"slug":  p.Slug,
					"title": p.Title,
				})
			}
		}
	}
	if projects == nil {
		projects = []map[string]any{}
	}

	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId":    requestID,
		"userId":       result.UserID,
		"orgName":      result.OrgName,
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
		"expiresAt":    result.ExpiresAt.UTC().Format(time.RFC3339),
		"projects":     projects,
	})
}

// CLIWhoami handles GET /api/cli/end-user/auth/whoami.
// This endpoint sits after the PAT middleware, which injects UserID and OrgName
// into context when a valid mc_pat_xxx Bearer token is present.
// It returns the identity and accessible projects — allowing the CLI to perform
// PAT-based login without a username/password exchange.
func (h *AuthHandler) CLIWhoami(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	userID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil || userID == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "valid PAT token required")
		return
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil || orgName == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "valid PAT token required")
		return
	}

	// Resolve isAdmin: check user_orgs.is_admin so APISIX can set X-Is-Admin correctly.
	// Admins (org owners / admins) can see all projects; regular end-users see only
	// explicitly granted projects.
	isAdmin := false
	if h.endUserSvc != nil {
		if u, uErr := h.endUserSvc.GetEndUser(ctx, appEnduser.GetEndUserCommand{
			OrgName: orgName,
			UserID:  userID,
		}); uErr == nil && u != nil {
			isAdmin = u.IsAdmin
		}
	}

	var projects []map[string]any
	if h.endUserSvc != nil {
		// Inject isAdmin into context so ListAccessibleProjects can use the fast path.
		adminCtx := ctxutils.SetIsAdmin(ctx, isAdmin)
		items, projErr := h.endUserSvc.ListAccessibleProjects(adminCtx, orgName, userID)
		if projErr == nil {
			for _, p := range items {
				projects = append(projects, map[string]any{
					"slug":  p.Slug,
					"title": p.Title,
				})
			}
		}
	}
	if projects == nil {
		projects = []map[string]any{}
	}

	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId": requestID,
		"userId":    userID,
		"orgName":   orgName,
		"isAdmin":   isAdmin,
		"projects":  projects,
	})
}

// CLIRefresh handles POST /api/cli/end-user/auth/refresh.
// Reads the refresh token from the request body instead of an httpOnly cookie.
func (h *AuthHandler) CLIRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName      string `json:"orgName"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}
	if req.RefreshToken == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refreshToken is required")
		return
	}

	result, err := h.authService.RefreshEndUserToken(ctx, appAuth.RefreshEndUserCommand{
		OrgName:      req.OrgName,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "cli end-user refresh failed")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId":    requestID,
		"userId":       result.UserID,
		"orgName":      result.OrgName,
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
		"expiresAt":    result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// CLILogout handles POST /api/cli/end-user/auth/logout.
// Reads the refresh token from the request body instead of an httpOnly cookie.
func (h *AuthHandler) CLILogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		OrgName      string `json:"orgName"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
		_ = h.authService.Logout(ctx, appAuth.LogoutCommand{
			RefreshToken: req.RefreshToken,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================

func (h *AuthHandler) handleBizError(
	w http.ResponseWriter, r *http.Request, requestID string, err error, logMsg string,
) {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
		h.writeError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "internal server error")
		return
	}
	h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
	h.writeError(w, bizErr.GetHTTPStatusCode(), requestID, bizErr.Info().GetCode(), bizErr.Msg())
}

func (h *AuthHandler) writeError(w http.ResponseWriter, status int, requestID, code, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"requestId": requestID,
		"error":     map[string]string{"code": code, "message": message},
	})
}

func (h *AuthHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
