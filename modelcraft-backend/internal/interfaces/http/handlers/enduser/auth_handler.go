package enduser

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/config"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainAuth "modelcraft/internal/domain/auth"

	appEnduser "modelcraft/internal/app/enduser"
)

const endUserRefreshCookieName = "mc_enduser_refresh_token"

// AuthHandler handles end-user authentication HTTP requests.
// All endpoints are registered via the OpenAPI-generated ServerInterface;
// no separate internal routes are needed.
type AuthHandler struct {
	authService *appEnduser.EndUserAuthAppService
	jwtSigner   *domainAuth.JWTSigner
	cookieCfg   config.CookieConfig
	logger      logfacade.Logger
}

// NewAuthHandler creates an AuthHandler.
// jwtSigner is the ES256 signer used to verify end-user Bearer tokens in /me.
func NewAuthHandler(
	authService *appEnduser.EndUserAuthAppService,
	jwtSigner *domainAuth.JWTSigner,
	cookieCfg config.CookieConfig,
	logger logfacade.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtSigner:   jwtSigner,
		cookieCfg:   cookieCfg,
		logger:      logger,
	}
}

func (h *AuthHandler) setEndUserRefreshCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteStrictMode
	switch h.cookieCfg.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     endUserRefreshCookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		HttpOnly: true,
		Secure:   h.cookieCfg.Secure,
		SameSite: sameSite,
		MaxAge:   30 * 24 * 60 * 60,
	})
}

func (h *AuthHandler) clearEndUserRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     endUserRefreshCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		HttpOnly: true,
		Secure:   h.cookieCfg.Secure,
		MaxAge:   -1,
	})
}

// ============================================================
// OpenAPI-generated ServerInterface methods
// ============================================================

// EndUserLogin handles POST /api/end-user/auth/login.
func (h *AuthHandler) EndUserLogin(w http.ResponseWriter, r *http.Request) {
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
		h.handleBizError(w, r, requestID, err, "end-user login failed")
		return
	}

	h.setEndUserRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
		result.AccessToken, "" /* stored in httpOnly cookie */, result.ExpiresAt,
		toProjectList(result.Projects), ""))
}

// EndUserRegister handles POST /api/end-user/auth/register.
func (h *AuthHandler) EndUserRegister(w http.ResponseWriter, r *http.Request) {
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
	if req.OrgName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName is required")
		return
	}

	result, err := h.authService.RegisterEndUser(ctx, appEnduser.RegisterCommand{
		OrgName:  req.OrgName,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "end-user register failed")
		return
	}

	h.setEndUserRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
		"", "" /* stored in httpOnly cookie */, result.ExpiresAt, nil, ""))
}

// EndUserLogout handles POST /api/end-user/auth/logout.
func (h *AuthHandler) EndUserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, _ := r.Cookie(endUserRefreshCookieName)
	if cookie != nil && cookie.Value != "" {
		var req struct {
			OrgName string `json:"orgName"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = h.authService.LogoutEndUser(ctx, appEnduser.LogoutCommand{
			RefreshToken: cookie.Value,
			OrgName:      req.OrgName,
		})
	}

	h.clearEndUserRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// EndUserRefreshToken handles POST /api/end-user/auth/refresh.
func (h *AuthHandler) EndUserRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	cookie, err := r.Cookie(endUserRefreshCookieName)
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

	result, err := h.authService.RefreshEndUserToken(ctx, appEnduser.RefreshCommand{
		OrgName:      req.OrgName,
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.clearEndUserRefreshCookie(w)
		h.handleBizError(w, r, requestID, err, "end-user refresh failed")
		return
	}

	h.setEndUserRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
		result.AccessToken, "" /* stored in httpOnly cookie */, result.ExpiresAt,
		toProjectList(result.Projects), ""))
}

// EndUserSelectProject handles POST /api/end-user/auth/select-project.
func (h *AuthHandler) EndUserSelectProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	cookie, err := r.Cookie(endUserRefreshCookieName)
	if err != nil || cookie.Value == "" {
		h.writeError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
		return
	}

	var req struct {
		OrgName     string `json:"orgName"`
		ProjectSlug string `json:"projectSlug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}
	if req.OrgName == "" || req.ProjectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "orgName and projectSlug are required")
		return
	}

	result, err := h.authService.SelectProjectContext(ctx, appEnduser.SelectProjectCommand{
		OrgName:      req.OrgName,
		ProjectSlug:  req.ProjectSlug,
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "end-user select-project failed")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"requestId":       requestID,
		"selectedProject": result.ProjectSlug,
	})
}

// EndUserMe handles GET /api/end-user/auth/me.
// Identity is resolved entirely from the Bearer JWT (ES256-verified).
// No external headers required — orgName and userID come from token claims.
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

	user, err := h.authService.GetEndUserMe(ctx, appEnduser.GetMeCommand{
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

// parseEndUserJWT 使用 ES256 公钥验证端用户 Bearer token，解析为 PlatformClaims。
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
// Response builders
// ============================================================

type projectItem struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

func toProjectList(items []appEnduser.AccessibleProject) []projectItem {
	out := make([]projectItem, 0, len(items))
	for _, p := range items {
		out = append(out, projectItem{Slug: p.Slug, Title: p.Title})
	}
	return out
}

func buildTokenResponse(
	requestID, userID, accessToken, refreshToken string,
	expiresAt time.Time,
	projects []projectItem,
	selectedProject string,
) map[string]any {
	m := map[string]any{
		"requestId": requestID,
		"userId":    userID,
	}
	if accessToken != "" {
		m["accessToken"] = accessToken
	}
	if refreshToken != "" {
		m["refreshToken"] = refreshToken
	}
	if !expiresAt.IsZero() {
		m["expiresAt"] = expiresAt.UTC().Format(time.RFC3339)
	}
	if len(projects) > 0 {
		m["projects"] = projects
	}
	if selectedProject != "" {
		m["selectedProject"] = selectedProject
	}
	return m
}

// ============================================================
// Error / JSON helpers
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
