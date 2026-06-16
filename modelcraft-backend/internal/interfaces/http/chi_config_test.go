package http

import (
	"testing"

	"modelcraft/internal/middleware"
	"modelcraft/pkg/config"
)

func TestNewChiRouterConfigIncludesUserAPITokenService(t *testing.T) {
	handlers := &DesignHandlers{
		UserAPITokenService: nil,
	}
	runtimeHandlers := &RuntimeHandlers{}
	jwtConfig := &middleware.JWTAuthConfig{}
	cfg := &config.Config{}

	chiConfig := NewChiRouterConfig(nil, cfg, handlers, runtimeHandlers, jwtConfig, nil)

	if chiConfig.DesignHandlers != handlers {
		t.Fatal("expected design handlers to be forwarded unchanged")
	}
	if chiConfig.RuntimeHandlers != runtimeHandlers {
		t.Fatal("expected runtime handlers to be forwarded unchanged")
	}
	if chiConfig.JWTConfig != jwtConfig {
		t.Fatal("expected jwt config to be forwarded unchanged")
	}
	if chiConfig.APITokenService != handlers.UserAPITokenService {
		t.Fatal("expected APITokenService to be sourced from design handlers")
	}
}
