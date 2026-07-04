package http

import (
	"modelcraft/pkg/config"
	"testing"
)

func TestNewChiRouterConfigIncludesUserAPITokenService(t *testing.T) {
	handlers := &DesignHandlers{
		UserAPITokenService: nil,
	}
	runtimeHandlers := &RuntimeHandlers{}
	cfg := &config.Config{}

	chiConfig := NewChiRouterConfig(nil, cfg, handlers, runtimeHandlers, nil)

	if chiConfig.DesignHandlers != handlers {
		t.Fatal("expected design handlers to be forwarded unchanged")
	}
	if chiConfig.RuntimeHandlers != runtimeHandlers {
		t.Fatal("expected runtime handlers to be forwarded unchanged")
	}
	if chiConfig.APITokenService != handlers.UserAPITokenService {
		t.Fatal("expected APITokenService to be sourced from design handlers")
	}
}
