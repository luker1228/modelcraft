package config

import (
	"os"
	"time"
)

// Config holds all runtime configuration for the gateway.
type Config struct {
	// Server
	Port string

	// JWT
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	RefreshCookieName  string

	// Upstream (Go Backend)
	BackendURL    string
	InternalToken string

	// CORS
	AllowedOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port: getEnv("GATEWAY_PORT", "8090"),

		JWTSecret:         getEnv("JWT_SECRET", "dev-jwt-secret-change-me"),
		AccessTokenTTL:    time.Hour,
		RefreshTokenTTL:   7 * 24 * time.Hour,
		RefreshCookieName: "mc_refresh_token",

		BackendURL:    getEnv("BACKEND_URL", "http://localhost:8080"),
		InternalToken: mustEnv("INTERNAL_TOKEN"),

		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
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
