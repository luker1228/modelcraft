package auth

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTSigner_IssueAccessToken(t *testing.T) {
	signer, err := GenerateDevSigner()
	if err != nil {
		t.Fatalf("GenerateDevSigner: %v", err)
	}

	t.Run("valid token with org scope", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", TokenScopeOrg)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		if tokenStr == "" {
			t.Error("expected non-empty token")
		}

		// 验证 token 内容
		pubPEM, err := signer.PublicKeyPEM()
		if err != nil {
			t.Fatalf("PublicKeyPEM: %v", err)
		}
		pubKey, err := jwt.ParseECPublicKeyFromPEM([]byte(pubPEM))
		if err != nil {
			t.Fatalf("ParseECPublicKeyFromPEM: %v", err)
		}

		claims := &PlatformClaims{}
		_, err = jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
			return pubKey, nil
		})
		if err != nil {
			t.Fatalf("ParseWithClaims: %v", err)
		}

		if claims.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
		}
		if claims.OrgName != "acme" {
			t.Errorf("OrgName = %q, want %q", claims.OrgName, "acme")
		}
		if claims.Scope != TokenScopeOrg {
			t.Errorf("Scope = %q, want %q", claims.Scope, TokenScopeOrg)
		}
		if claims.Issuer != string(IssuerPlatform) {
			t.Errorf("Issuer = %q, want %q", claims.Issuer, IssuerPlatform)
		}
	})

	t.Run("empty orgName returns error", func(t *testing.T) {
		_, err := signer.IssueAccessToken("user-1", "", TokenScopeOrg)
		if err == nil {
			t.Error("expected error for empty orgName")
		}
		if !strings.Contains(err.Error(), "orgName") {
			t.Errorf("error should mention orgName, got: %v", err)
		}
	})

	t.Run("invalid scope returns error", func(t *testing.T) {
		_, err := signer.IssueAccessToken("user-1", "acme", "invalid-scope")
		if err == nil {
			t.Error("expected error for invalid scope")
		}
		if !strings.Contains(err.Error(), "scope") {
			t.Errorf("error should mention scope, got: %v", err)
		}
	})

	t.Run("project scope token", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-2", "acme", TokenScopeProject)
		if err != nil {
			t.Fatalf("IssueAccessToken for project scope: %v", err)
		}
		if tokenStr == "" {
			t.Error("expected non-empty token for project scope")
		}
	})
}

func TestIssuerIsValid(t *testing.T) {
	if !IssuerPlatform.IsValid() {
		t.Error("IssuerPlatform should be valid")
	}
	if IssuerDeveloper.IsValid() {
		t.Error("IssuerDeveloper should be invalid (deprecated)")
	}
	if IssuerEndUser.IsValid() {
		t.Error("IssuerEndUser should be invalid (deprecated)")
	}
}

func TestJWTSigner_ParsePlatformClaims(t *testing.T) {
	signer, err := GenerateDevSigner()
	if err != nil {
		t.Fatalf("GenerateDevSigner: %v", err)
	}

	t.Run("valid org token roundtrip", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", TokenScopeOrg)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		claims, err := signer.ParsePlatformClaims(tokenStr)
		if err != nil {
			t.Fatalf("ParsePlatformClaims: %v", err)
		}
		if claims.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
		}
		if claims.OrgName != "acme" {
			t.Errorf("OrgName = %q, want %q", claims.OrgName, "acme")
		}
		if claims.Scope != TokenScopeOrg {
			t.Errorf("Scope = %q, want %q", claims.Scope, TokenScopeOrg)
		}
	})

	t.Run("valid project token roundtrip", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-2", "acme", TokenScopeProject)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		claims, err := signer.ParsePlatformClaims(tokenStr)
		if err != nil {
			t.Fatalf("ParsePlatformClaims: %v", err)
		}
		if claims.Scope != TokenScopeProject {
			t.Errorf("Scope = %q, want %q", claims.Scope, TokenScopeProject)
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		_, err := signer.ParsePlatformClaims("not-a-valid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("tampered token returns error", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", TokenScopeOrg)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		// 修改 token 内容（破坏签名）
		tampered := tokenStr[:len(tokenStr)-5] + "XXXXX"
		_, err = signer.ParsePlatformClaims(tampered)
		if err == nil {
			t.Error("expected error for tampered token")
		}
	})
}
