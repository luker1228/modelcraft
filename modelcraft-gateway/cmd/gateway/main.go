package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"modelcraft-gateway/internal/auth"
	"modelcraft-gateway/internal/config"
	"modelcraft-gateway/internal/middleware"
	"modelcraft-gateway/internal/proxy"
)

func main() {
	// Human-readable logging in development.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg := config.Load()

	// Initialise auth service (ES256 public-key verifier + cookie manager).
	authSvc, err := auth.NewService(
		cfg.JWTPublicKey,
		cfg.RefreshTokenTTL,
		cfg.RefreshCookieName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise auth service")
	}

	// Initialise proxy handler.
	proxyHandler, err := proxy.NewHandler(cfg.BackendURL, cfg.InternalToken, authSvc)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create proxy handler")
	}

	// Initialise REST proxy handler.
	restHandler, err := proxy.NewRESTHandler(cfg.BackendURL, authSvc)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rest handler")
	}

	// Initialise auth handler.
	authHandler := auth.NewHandler(authSvc, cfg.BackendURL)

	// Build router.
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestLogger)

	// Auth endpoints — no JWT required.
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	// GraphQL proxy endpoints — JWT required (validated inside handler).
	r.Post("/graphql/org/{orgName}", proxyHandler.GraphQLOrgHandler)
	r.Post("/graphql/org/{orgName}/", proxyHandler.GraphQLOrgHandler)
	r.Post("/graphql/org/{orgName}/project/{projectSlug}", proxyHandler.GraphQLProjectHandler)
	r.Post("/graphql/org/{orgName}/project/{projectSlug}/", proxyHandler.GraphQLProjectHandler)

	// Health check.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Protected REST endpoints — JWT validated, userID injected as X-User-ID.
	r.Route("/api/user", func(r chi.Router) {
		r.Handle("/*", http.HandlerFunc(restHandler.Handle))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("modelcraft-gateway starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("gateway stopped")
}
