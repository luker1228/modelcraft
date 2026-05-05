package config

import (
	"os"
	"time"
)

// Config holds all runtime configuration for the gateway.
type Config struct {
	// Server
	Port string

	// JWT — EC public key (PEM) for verifying ES256 access tokens issued by the backend.
	// JWT signing is owned by the backend auth service; the gateway only verifies.
	JWTPublicKey             string
	RefreshTokenTTL          time.Duration
	RefreshCookieName        string
	EndUserRefreshCookieName string

	// Deprecated: EndUserJWTSecret 曾用于 HMAC-SHA256 端用户 token 验证。
	// 端用户 token 已迁移至 ES256（mc-platform issuer），此配置不再使用。
	// 保留字段以维持向后兼容，将在阶段 3 Schema 清理时删除。
	EndUserJWTSecret string

	// Upstream (Go Backend)
	BackendURL    string
	InternalToken string

	// CORS
	AllowedOrigins []string

	// Observability
	OTLPEndpoint string // e.g. "localhost:4317"; empty disables tracing

	// Logging
	// LogOutputPath is the file path to write logs to.
	// Empty string means write to stderr (human-readable console format).
	LogOutputPath string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port: getEnv("GATEWAY_PORT", "8090"),

		JWTPublicKey:             getEnv("JWT_PUBLIC_KEY", ""),
		RefreshTokenTTL:          7 * 24 * time.Hour,
		RefreshCookieName:        "mc_refresh_token",
		EndUserRefreshCookieName: "mc_enduser_refresh_token",

		EndUserJWTSecret: getEnv("JWT_SECRET", ""),

		BackendURL:    getEnv("BACKEND_URL", "http://localhost:8080"),
		InternalToken: mustEnv("INTERNAL_TOKEN"),

		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),

		LogOutputPath: getEnv("LOG_OUTPUT_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}

	return v
}
