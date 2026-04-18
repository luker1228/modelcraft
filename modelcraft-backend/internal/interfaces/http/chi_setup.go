package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/internal/middleware"
	"modelcraft/pkg/config"
	"modelcraft/pkg/logfacade"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	orgHandlers "modelcraft/internal/interfaces/http/handlers/org"
	userHandlers "modelcraft/internal/interfaces/http/handlers/user"
)

// ChiRouterConfig holds all dependencies needed to set up the Chi router.
type ChiRouterConfig struct {
	Logger logfacade.Logger
	Config *config.Config
	DB     *sql.DB // For health check

	// Handlers for OpenAPI routes (tenant-management only)
	AuthHandler    *authHandlers.Handler
	OrgHandler     *orgHandlers.CreateHandler
	UserHandler    *userHandlers.Handler

	// Design handlers for GraphQL routes
	DesignHandlers *DesignHandlers

	// Runtime handlers for dynamic model GraphQL routes
	RuntimeHandlers *RuntimeHandlers

	// JWT configuration for protected routes
	JWTConfig *middleware.JWTAuthConfig
}

// SetupChiRouter creates and configures the main Chi router as the primary HTTP mux.
//
// Architecture:
//
//	Chi Router (primary)
//	  ├── Global: RequestID, RealIP, CORS, Logger, Recoverer
//	  ├── /health, /test                          → net/http (no auth)
//	  ├── /api/openapi.json                       → net/http (no auth)
//	  ├── /org/*                                  → Gin mount (Gin handles JWT)
//	  ├── /graphql/*                              → Gin mount (no auth currently)
//	  └── /api/* (generated OpenAPI handler)      → Chi with conditional auth
//	        ├── /api/auth/register, /login, /logout     → public (no auth)
//	        └── everything else                       → JWT + Tenant middleware
func SetupChiRouter(cfg *ChiRouterConfig) chi.Router {
	r := chi.NewRouter()

	// ============================================================
	// Global Middleware (applies to ALL routes)
	// ============================================================
	r.Use(chimw.RealIP)
	r.Use(chimw.StripSlashes)
	r.Use(middleware.ChiCORS())
	// Auto-set Content-Type: application/json (must be before other middleware that write responses)
	r.Use(middleware.JSONContentTypeMiddleware())
	r.Use(middleware.ChiLoggerMiddleware(cfg.Logger))
	// Initialize HttpRequestContext for all requests
	r.Use(middleware.ChiHttpContextMiddleware())
	// Request timeout: 60 seconds
	r.Use(chimw.Timeout(60 * time.Second))
	// Custom recovery with structured logging
	r.Use(middleware.ChiRecoveryMiddleware(cfg.Logger))

	// ============================================================
	// Health & Debug Endpoints (no auth, pure net/http)
	// ============================================================
	r.Get("/health", healthHandler(cfg))

	if cfg.Config.Server.Mode == "debug" {
		r.Get("/test", debugHandler())
	}

	// ============================================================
	// OpenAPI Spec Endpoint (no auth, pure net/http)
	// ============================================================
	r.Get("/api/openapi.json", openAPISpecHandler())

	// ============================================================
	// GraphQL Routes - Org API
	// ============================================================
	if cfg.DesignHandlers != nil {
		SetupOrgGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config)
	}

	// ============================================================
	// GraphQL Routes - Project API
	// ============================================================
	if cfg.DesignHandlers != nil {
		SetupProjectGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config)
	}

	// ============================================================
	// GraphQL Routes - Runtime API
	// ============================================================
	if cfg.RuntimeHandlers != nil {
		SetupRuntimeGraphQLRoutesOnChi(r, cfg.RuntimeHandlers, cfg.Config)
	}

	// ============================================================
	// Internal End-User HTTP Routes
	// ============================================================
	if cfg.DesignHandlers != nil {
		SetupEndUserRoutesOnChi(r, cfg.DesignHandlers, cfg.Config)
	}

	// ============================================================
	// OpenAPI Routes via Generated Chi Handler
	// ============================================================

	// Create the Server that implements generated.ServerInterface
	// Only tenant-management (Auth, Org, User) handlers.
	// Business domain APIs are served via GraphQL.
	server := NewServer(
		cfg.AuthHandler,
		cfg.OrgHandler,
		cfg.UserHandler,
	)

	// Create generated Chi handler with NO built-in middleware.
	// Use generated.ChiServerOptions.Middlewares to apply our middleware to all routes.
	authMiddleware := conditionalAuthMiddleware(cfg.JWTConfig)

	// Register generated OpenAPI routes with middleware
	_ = generated.HandlerWithOptions(server, generated.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []generated.MiddlewareFunc{
			authMiddleware,
			requestIDInjectorMiddleware,
		},
	})

	return r
}

// conditionalAuthMiddleware applies JWT middleware to all routes
// EXCEPT known public paths that should not require authentication.
// These routes (Auth, Org, User) never require organization context.
func conditionalAuthMiddleware(jwtConfig *middleware.JWTAuthConfig) func(http.Handler) http.Handler {
	jwtMW := middleware.ChiJWTAuthMiddleware(jwtConfig)

	// Paths that are public and should NOT require authentication
	publicPaths := map[string]bool{
		"/api/auth/register":   true,
		"/api/auth/login":      true,
		"/api/auth/logout":  true,
		"/api/auth/refresh": true,
	}

	return func(next http.Handler) http.Handler {
		jwtOnly := jwtMW(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Public paths: no authentication required
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// All other paths: JWT authentication only, no organization context required
			jwtOnly.ServeHTTP(w, r)
		})
	}
}

// healthHandler returns a pure net/http health check handler.
func healthHandler(cfg *ChiRouterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if cfg.DB != nil {
			if err := cfg.DB.Ping(); err != nil {
				dbStatus = "error"
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"service":  "modelcraft-unified-service",
			"version":  "1.0.0",
			"database": dbStatus,
			"config": map[string]string{
				"mode": cfg.Config.Server.Mode,
				"port": cfg.Config.Server.Port,
			},
			"message": "ModelCraft Unified Service is running! Ready to build amazing models!",
		})
	}
}

// debugHandler returns a pure net/http debug endpoint handler.
func debugHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Test endpoint (GET /test) - debug mode only"}`)
	}
}

// openAPISpecHandler returns a pure net/http handler that serves the embedded OpenAPI spec.
func openAPISpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec, err := generated.GetSwagger()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"Failed to load OpenAPI spec"}`)
			return
		}
		// Clear servers so clients use relative URLs
		spec.Servers = nil

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(spec)
	}
}
