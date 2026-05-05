package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPlatformClaims_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)

	t.Run("valid claims", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "acme",
			Scope:   TokenScopeOrg,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		if err := c.Validate(); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("empty user_id", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "",
			OrgName: "acme",
			Scope:   TokenScopeOrg,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		err := c.Validate()
		if err == nil {
			t.Error("expected error for empty user_id")
		}
	})

	t.Run("empty org_name", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "",
			Scope:   TokenScopeOrg,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		err := c.Validate()
		if err == nil {
			t.Error("expected error for empty org_name")
		}
	})

	t.Run("invalid scope", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "acme",
			Scope:   "invalid",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		err := c.Validate()
		if err == nil {
			t.Error("expected error for invalid scope")
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "acme",
			Scope:   TokenScopeOrg,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "mc-developer",
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		err := c.Validate()
		if err == nil {
			t.Error("expected error for wrong issuer")
		}
	})

	t.Run("expired", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "acme",
			Scope:   TokenScopeOrg,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
			},
		}
		err := c.Validate()
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("scope constants", func(t *testing.T) {
		if TokenScopeOrg != "org" {
			t.Errorf("TokenScopeOrg = %q, want %q", TokenScopeOrg, "org")
		}
		if TokenScopeProject != "project" {
			t.Errorf("TokenScopeProject = %q, want %q", TokenScopeProject, "project")
		}
		if TokenScopeServiceKey != "service_key" {
			t.Errorf("TokenScopeServiceKey = %q, want %q", TokenScopeServiceKey, "service_key")
		}
	})
}

func TestPlatformClaims_ProjectScope(t *testing.T) {
	future := time.Now().Add(time.Hour)
	c := &PlatformClaims{
		UserID:  "user-1",
		OrgName: "acme",
		Scope:   TokenScopeProject,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(IssuerPlatform),
			ExpiresAt: jwt.NewNumericDate(future),
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("expected nil for project scope, got %v", err)
	}
}
