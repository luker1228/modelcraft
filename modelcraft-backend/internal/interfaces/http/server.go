package http

import (
	"modelcraft/internal/interfaces/http/generated"
	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	"net/http"
)

// Server implements the generated.ServerInterface using standard net/http handlers.
//
// Covers tenant-management auth endpoints.
// Business domain APIs (Projects, Models, Clusters, Enums) are served exclusively via GraphQL.
type Server struct {
	authHandler *authHandlers.Handler
}

// Ensure compile-time interface compliance.
var _ generated.ServerInterface = (*Server)(nil)

// NewServer creates a new Server that implements the generated.ServerInterface.
func NewServer(
	authHandler *authHandlers.Handler,
) *Server {
	return &Server{
		authHandler: authHandler,
	}
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

func (s *Server) Whoami(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandlePATWhoami(w, r)
}

func (s *Server) DemoLogin(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleDemoLogin(w, r)
}

// ========================
// End-User Auth Endpoints (not yet wired — routes served via end-user GraphQL)
// ========================

func (s *Server) EndUserLogin(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s *Server) EndUserLogout(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s *Server) EndUserRefreshToken(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s *Server) EndUserMe(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// ========================
// User Endpoints
// ========================

func (s *Server) GetUserMemberships(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
