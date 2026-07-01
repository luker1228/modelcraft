package http

import (
	"encoding/json"
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

func (s *Server) Whoami(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandlePATWhoami(w, r)
}

// GetOpenAPISpec serves the embedded OpenAPI specification.
func GetOpenAPISpec() ([]byte, error) {
	spec, err := generated.GetSwagger()
	if err != nil {
		return nil, err
	}
	return json.Marshal(spec)
}
