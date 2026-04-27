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
	JWTPublicKey      string
	RefreshTokenTTL   time.Duration
	RefreshCookieName string

	// Upstream (Go Backend)
	BackendURL    string
	InternalToken string

	// CORS
	AllowedOrigins []string

	// Observability
	OTLPEndpoint string // e.g. "localhost:4317"; empty disables tracing
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port: getEnv("GATEWAY_PORT", "8090"),

		JWTPublicKey:      getEnv("JWT_PUBLIC_KEY", ""),
		RefreshTokenTTL:   7 * 24 * time.Hour,
		RefreshCookieName: "mc_refresh_token",

		BackendURL:    getEnv("BACKEND_URL", "http://localhost:8080"),
		InternalToken: mustEnv("INTERNAL_TOKEN"),

		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
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
