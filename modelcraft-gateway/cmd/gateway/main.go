package main

import (
	"context"
	"net/http"
	"path/filepath"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"modelcraft-gateway/internal/auth"
	"modelcraft-gateway/internal/config"
	"modelcraft-gateway/internal/middleware"
	"modelcraft-gateway/internal/proxy"
	"modelcraft-gateway/internal/telemetry"
)

func main() {
	cfg := config.Load()

	logger := buildLogger(cfg.LogOutputPath)
	defer func() { _ = logger.Sync() }()

	// Register as global so zap.L() works everywhere (e.g. middleware.RequestID).
	zap.ReplaceGlobals(logger)

	ctx := context.Background()
	shutdownTracer, err := telemetry.InitTracerProvider(ctx, "modelcraft-gateway", cfg.OTLPEndpoint)
	if err != nil {
		logger.Fatal("failed to initialise tracer provider", zap.Error(err))
	}
	defer shutdownTracer()

	authSvc, err := auth.NewService(
		cfg.JWTPublicKey,
		cfg.RefreshTokenTTL,
		cfg.RefreshCookieName,
		cfg.EndUserRefreshCookieName,
		cfg.EndUserJWTSecret, // Deprecated: 端用户 token 已迁移 ES256，此参数将在阶段 3 移除
	)
	if err != nil {
		logger.Fatal("failed to initialise auth service", zap.Error(err))
	}

	tracedHTTPClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	proxyHandler, err := proxy.NewHandler(cfg.BackendURL, cfg.InternalToken, authSvc)
	if err != nil {
		logger.Fatal("failed to create proxy handler", zap.Error(err))
	}

	restHandler, err := proxy.NewRESTHandler(cfg.BackendURL, authSvc)
	if err != nil {
		logger.Fatal("failed to create rest handler", zap.Error(err))
	}

	authHandler := auth.NewHandler(authSvc, cfg.BackendURL, tracedHTTPClient, cfg.InternalToken)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id", "X-Client-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// RequestID must run before OTel so that GetReqID works in all spans.
	r.Use(middleware.RequestID)
	r.Use(otelchi.Middleware("modelcraft-gateway", otelchi.WithChiRoutes(r)))
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestLogger(logger))

	// Auth endpoints — no JWT required.
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	// Public end-user auth endpoints — no JWT required.
	r.Route("/api/end-user/auth", func(r chi.Router) {
		r.Post("/login", authHandler.EndUserLogin)
		r.Post("/register", authHandler.EndUserRegister)
		r.Post("/refresh", authHandler.EndUserRefresh)
		r.Post("/logout", authHandler.EndUserLogout)
		r.Post("/select-project", authHandler.EndUserSelectProject)
		r.Get("/me", authHandler.EndUserMe)
	})

	// GraphQL proxy endpoints — JWT required (validated inside handler).
	r.Post("/graphql/org/{orgName}", proxyHandler.GraphQLOrgHandler)
	r.Post("/graphql/org/{orgName}/", proxyHandler.GraphQLOrgHandler)
	r.Post("/graphql/org/{orgName}/project/{projectSlug}", proxyHandler.GraphQLProjectHandler)
	r.Post("/graphql/org/{orgName}/project/{projectSlug}/", proxyHandler.GraphQLProjectHandler)

	// End-User GraphQL — end-user HMAC JWT required.
	r.Post("/graphql/end-user/org/{orgName}/project/{projectSlug}", proxyHandler.EndUserGraphQLHandler)
	r.Post("/graphql/end-user/org/{orgName}/project/{projectSlug}/", proxyHandler.EndUserGraphQLHandler)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("modelcraft-gateway starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("gateway stopped")
}

// buildLogger constructs a zap.Logger.
//
//   - logOutputPath == ""  → human-readable colored output to stderr (dev mode)
//   - logOutputPath != ""  → JSON lines written to the file via lumberjack (rotation enabled)
func buildLogger(logOutputPath string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	var core zapcore.Core
	if logOutputPath == "" {
		// Dev: colored console output to stderr.
		consoleCfg := zap.NewDevelopmentEncoderConfig()
		consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleCfg),
			zapcore.AddSync(os.Stderr),
			zapcore.DebugLevel,
		)
	} else {
		// Ensure the parent directory exists so file logging does not fail when
		// the log directory has not been created yet.
		if dir := filepath.Dir(logOutputPath); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}

		// Prod: JSON lines to a rotating file.
		rotator := &lumberjack.Logger{
			Filename:   logOutputPath,
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(rotator),
			zapcore.InfoLevel,
		)
	}

	return zap.New(core, zap.AddCaller())
}
