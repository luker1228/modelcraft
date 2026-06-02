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

	appEnduser "modelcraft/internal/app/enduser"
	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	userHandlers "modelcraft/internal/interfaces/http/handlers/user"
)

// ChiRouterConfig holds all dependencies needed to set up the Chi router.
type ChiRouterConfig struct {
	Logger logfacade.Logger
	Config *config.Config
	DB     *sql.DB // For health check

	// Handlers for OpenAPI routes (tenant-management only)
	AuthHandler *authHandlers.Handler
	UserHandler *userHandlers.Handler

	// Design handlers for GraphQL routes
	DesignHandlers *DesignHandlers

	// Runtime handlers for dynamic model GraphQL routes
	RuntimeHandlers *RuntimeHandlers

	// JWT configuration for protected routes
	JWTConfig *middleware.JWTAuthConfig

	// APITokenService for PAT Bearer token validation (nil = disabled)
	APITokenService *appEnduser.APITokenService
}

// SetupChiRouter creates and configures the main Chi router as the primary HTTP mux.
//
// Architecture:
//
//	Chi Router (primary)
//	  ├── Global: RequestID, RealIP, CORS, Logger, Recoverer
//	  ├── /health, /test                               → net/http (no auth)
//	  ├── /api/openapi.json                            → net/http (no auth)
//	  ├── /graphql/org/*                               → tenant-only GraphQL (X-Tenant-User-Id)
//	  ├── /end-user/graphql/org/*                      → end-user GraphQL (X-User-ID)
//	  ├── /graphql/org/*/project/*/db/*/model/*        → runtime GraphQL
//	  └── /api/* (generated OpenAPI handler)           → Chi with conditional auth
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
	// GraphQL Routes - Tenant (Org + Project)
	// ============================================================
	if cfg.DesignHandlers != nil {
		SetupOrgGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config, cfg.JWTConfig)
		SetupProjectGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config, cfg.JWTConfig)
	}

	// ============================================================
	// GraphQL Routes - End-User (Org + Project, same resolvers)
	// ============================================================
	if cfg.DesignHandlers != nil {
		SetupEndUserOrgGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config, cfg.JWTConfig)
		SetupEndUserProjectGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config, cfg.JWTConfig)
	}

	// ============================================================
	// GraphQL Routes - Runtime API
	// ============================================================
	if cfg.RuntimeHandlers != nil {
		SetupRuntimeGraphQLRoutesOnChi(r, cfg.RuntimeHandlers, cfg.Config)
	}

	// ============================================================
	// CLI Auth Routes (no httpOnly cookie — token in body)
	// ============================================================
	if cfg.DesignHandlers != nil && cfg.DesignHandlers.EndUserAuthHandler != nil {
		h := cfg.DesignHandlers.EndUserAuthHandler
		r.Post("/api/cli/end-user/auth/login", h.CLILogin)
		r.Post("/api/cli/end-user/auth/refresh", h.CLIRefresh)
		r.Post("/api/cli/end-user/auth/logout", h.CLILogout)
	}

	// ============================================================
	// PAT Auth Middleware
	// ============================================================
	// ============================================================
	// CLI PAT-authenticated Routes (require PAT middleware above)
	// ============================================================
	if cfg.DesignHandlers != nil && cfg.DesignHandlers.EndUserAuthHandler != nil && cfg.APITokenService != nil {
		h := cfg.DesignHandlers.EndUserAuthHandler
		r.Group(func(gr chi.Router) {
			gr.Use(middleware.ChiPATAuthMiddleware(cfg.APITokenService, cfg.Logger))
			gr.Get("/api/cli/end-user/auth/whoami", h.CLIWhoami)
		})
	}

	// ============================================================
	// OpenAPI Routes via Generated Chi Handler
	// ============================================================

	// Create the Server that implements generated.ServerInterface
	// Only tenant-management (Auth, Org, User) handlers.
	// Business domain APIs are served via GraphQL.
	server := NewServer(
		cfg.AuthHandler,
		cfg.UserHandler,
		cfg.DesignHandlers.EndUserAuthHandler,
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

	// Paths that are public and should NOT require authentication.
	// All /api/end-user/auth/* routes bypass the design-time JWT middleware because
	// they carry end-user HMAC JWTs (endUserClaims) which are validated inside each
	// handler. Applying the design-time ChiJWTAuthMiddleware here would always fail.
	publicPaths := map[string]bool{
		"/api/tenant/auth/register": true,
		"/api/tenant/auth/login":    true,
		"/api/tenant/auth/logout":   true,
		"/api/tenant/auth/refresh":  true,
		// End-user auth: all routes use their own in-handler JWT validation
		"/api/end-user/auth/login":   true,
		"/api/end-user/auth/refresh": true,
		"/api/end-user/auth/logout":  true,
		"/api/end-user/auth/me":      true,
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
		_, _ = fmt.Fprintf(w, `{"message":"Test endpoint (GET /test) - debug mode only"}`)
	}
}

// openAPISpecHandler returns a pure net/http handler that serves the embedded OpenAPI spec.
func openAPISpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec, err := generated.GetSwagger()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"error":"Failed to load OpenAPI spec"}`)
			return
		}
		// Clear servers so clients use relative URLs
		spec.Servers = nil

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(spec)
	}
}
