package http

import (
	"encoding/json"
	"errors"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"

	appOrg "modelcraft/internal/app/organization"
	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	enduserHandlers "modelcraft/internal/interfaces/http/handlers/enduser"
	userHandlers "modelcraft/internal/interfaces/http/handlers/user"
)

// Server implements the generated.ServerInterface using standard net/http handlers.
//
// Covers tenant-management (Auth, User) and end-user auth endpoints.
// Business domain APIs (Projects, Models, Clusters, Enums) are served exclusively via GraphQL.
type Server struct {
	authHandler        *authHandlers.Handler
	userHandler        *userHandlers.Handler
	endUserAuthHandler *enduserHandlers.AuthHandler
	createOrgService   *appOrg.CreateOrganizationService
}

// Ensure compile-time interface compliance.
var _ generated.ServerInterface = (*Server)(nil)

// NewServer creates a new Server that implements the generated.ServerInterface.
func NewServer(
	authHandler *authHandlers.Handler,
	userHandler *userHandlers.Handler,
	endUserAuthHandler *enduserHandlers.AuthHandler,
	createOrgService *appOrg.CreateOrganizationService,
) *Server {
	return &Server{
		authHandler:        authHandler,
		userHandler:        userHandler,
		endUserAuthHandler: endUserAuthHandler,
		createOrgService:   createOrgService,
	}
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ========================
// Auth Endpoints
// ========================

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleLogin(w, r)
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleRegister(w, r)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleLogout(w, r)
}

func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleRefresh(w, r)
}

// ========================
// End-User Auth Endpoints
// ========================

func (s *Server) EndUserLogin(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserLogin(w, r)
}

func (s *Server) EndUserRegister(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserRegister(w, r)
}

func (s *Server) EndUserLogout(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserLogout(w, r)
}

func (s *Server) EndUserRefreshToken(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserRefreshToken(w, r)
}

func (s *Server) EndUserSelectProject(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserSelectProject(w, r)
}

func (s *Server) EndUserMe(w http.ResponseWriter, r *http.Request) {
	s.endUserAuthHandler.EndUserMe(w, r)
}

// ========================
// User Endpoints
// ========================

func (s *Server) GetUserMemberships(w http.ResponseWriter, r *http.Request) {
	logger := logfacade.GetLogger(r.Context())

	userID, err := ctxutils.GetUserIDFromContext(r.Context())
	if err != nil {
		logger.Error(r.Context(), "User ID not found in request context", logfacade.Err(err), logfacade.Stack(err))
		writeJSON(w, http.StatusUnauthorized, generated.UnauthorizedError{
			Error: struct {
				Code    generated.UnauthorizedErrorErrorCode `json:"code"`
				Message string                               `json:"message"`
			}{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in request context",
			},
		})
		return
	}

	resp, err := s.userHandler.GetUserMemberships(r.Context(), userID)
	if err != nil {
		logger.Error(r.Context(), "Failed to get user memberships", logfacade.Err(err), logfacade.Stack(err))
		writeJSON(w, http.StatusInternalServerError, generated.SystemError{
			Error: struct {
				Code    generated.SystemErrorErrorCode `json:"code"`
				Details *map[string]any                `json:"details,omitempty"`
				Message string                         `json:"message"`
			}{
				Code:    "SYSTEM_ERROR",
				Message: "Failed to get user memberships",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetOpenAPISpec serves the embedded OpenAPI specification.
func GetOpenAPISpec() ([]byte, error) {
	spec, err := generated.GetSwagger()
	if err != nil {
		return nil, err
	}
	return json.Marshal(spec)
}

// ========================
// Organization Endpoints
// ========================

// InitOrg handles POST /api/org/init — creates an organization for the authenticated user.
func (s *Server) InitOrg(w http.ResponseWriter, r *http.Request) {
	requestID := ctxutils.GetRequestID(r.Context())

	var req struct {
		DisplayName          string `json:"displayName"`
		OrganizationName     string `json:"organizationName"`
		EndUserAdminPassword string `json:"endUserAdminPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body", "requestId": requestID,
		})
		return
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "displayName is required", "requestId": requestID,
		})
		return
	}
	if strings.TrimSpace(req.EndUserAdminPassword) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "endUserAdminPassword is required", "requestId": requestID,
		})
		return
	}

	userID, err := ctxutils.GetUserIDFromContext(r.Context())
	if err != nil || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error": "user not authenticated", "requestId": requestID,
		})
		return
	}

	result, err := s.createOrgService.Execute(r.Context(), &appOrg.CreateOrganizationInput{
		DisplayName:          req.DisplayName,
		OrganizationName:     req.OrganizationName,
		OwnerUserID:          userID,
		EndUserAdminPassword: req.EndUserAdminPassword,
	})
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			writeJSON(w, bizErr.GetHTTPStatusCode(), map[string]any{
				"error": bizErr.Msg(), "code": bizErr.Info().GetCode(), "requestId": requestID,
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "org init failed", "requestId": requestID,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"requestId":        requestID,
		"organizationName": result.OrganizationName,
		"displayName":      result.DisplayName,
		"alreadyExisted":   result.AlreadyExisted,
	})
}
