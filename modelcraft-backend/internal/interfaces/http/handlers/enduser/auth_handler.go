package enduser

import (
	"encoding/json"
	domainAuth "modelcraft/internal/domain/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"

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

// Whoami handles GET /api/tenant/auth/whoami.
// This endpoint sits after the PAT middleware, which injects UserID and OrgName
// into context when a valid mc_pat_xxx Bearer token is present.
// It returns the identity and accessible projects for PAT-based callers.
func (h *AuthHandler) Whoami(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	userID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil || userID == "" {
		h.logger.Warnf(ctx, "whoami rejected: missing userID requestId=%s", requestID)
		h.writeError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "valid PAT token required")
		return
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil || orgName == "" {
		h.logger.Warnf(ctx, "whoami rejected: missing orgName requestId=%s userId=%s", requestID, userID)
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

	h.logger.Infof(ctx, "whoami resolved requestId=%s userId=%s orgName=%s isAdmin=%v projects=%d", requestID, userID, orgName, isAdmin, len(projects))

	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId": requestID,
		"userId":    userID,
		"orgName":   orgName,
		"isAdmin":   isAdmin,
		"projects":  projects,
	})
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
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
