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
			Key:     ApisixConsumerKey,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    string(IssuerPlatform),
				ExpiresAt: jwt.NewNumericDate(future),
			},
		}
		if err := c.Validate(); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if c.Key != ApisixConsumerKey {
			t.Errorf("Key = %q, want %q", c.Key, ApisixConsumerKey)
		}
	})

	t.Run("empty user_id", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "",
			OrgName: "acme",
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

	t.Run("wrong issuer", func(t *testing.T) {
		c := &PlatformClaims{
			UserID:  "user-1",
			OrgName: "acme",
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

	t.Run("audience constants", func(t *testing.T) {
		if AudienceTenant != "tenant" {
			t.Errorf("AudienceTenant = %q, want %q", AudienceTenant, "tenant")
		}
		if AudienceEndUser != "end_user" {
			t.Errorf("AudienceEndUser = %q, want %q", AudienceEndUser, "end_user")
		}
	})
}

func TestPlatformClaims_EndUserToken(t *testing.T) {
	future := time.Now().Add(time.Hour)
	c := &PlatformClaims{
		UserID:  "user-1",
		OrgName: "acme",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(IssuerPlatform),
			Audience:  jwt.ClaimStrings{AudienceEndUser},
			ExpiresAt: jwt.NewNumericDate(future),
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("expected nil for end_user token, got %v", err)
	}
}
