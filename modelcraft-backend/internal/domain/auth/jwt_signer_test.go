package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTSigner_IssueAccessToken(t *testing.T) {
	signer, err := GenerateDevSigner()
	if err != nil {
		t.Fatalf("GenerateDevSigner: %v", err)
	}

	t.Run("valid platform token", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", false)
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
		}, jwt.WithAudience(AudiencePlatform))
		if err != nil {
			t.Fatalf("ParseWithClaims: %v", err)
		}

		if claims.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
		}
		if claims.OrgName != "acme" {
			t.Errorf("OrgName = %q, want %q", claims.OrgName, "acme")
		}
		if len(claims.Audience) == 0 || claims.Audience[0] != AudiencePlatform {
			t.Errorf("Audience = %v, want [%q]", claims.Audience, AudiencePlatform)
		}
		if claims.Issuer != string(IssuerPlatform) {
			t.Errorf("Issuer = %q, want %q", claims.Issuer, IssuerPlatform)
		}
	})

	t.Run("empty orgName returns error", func(t *testing.T) {
		_, err := signer.IssueAccessToken("user-1", "", false)
		if err == nil {
			t.Error("expected error for empty orgName")
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

	t.Run("valid platform token roundtrip", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", false)
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
		if len(claims.Audience) == 0 || claims.Audience[0] != AudiencePlatform {
			t.Errorf("Audience = %v, want [%q]", claims.Audience, AudiencePlatform)
		}
		if claims.Key != ApisixConsumerKey {
			t.Errorf("Key = %q, want %q", claims.Key, ApisixConsumerKey)
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		_, err := signer.ParsePlatformClaims("not-a-valid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("tampered token returns error", func(t *testing.T) {
		tokenStr, err := signer.IssueAccessToken("user-1", "acme", false)
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

func TestIssueAccessTokenIsAdmin(t *testing.T) {
	signer, err := GenerateDevSigner()
	if err != nil {
		t.Fatalf("GenerateDevSigner: %v", err)
	}

	t.Run("admin flag is true", func(t *testing.T) {
		token, err := signer.IssueAccessToken("user-1", "my-org", true)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		claims, err := signer.ParsePlatformClaims(token)
		if err != nil {
			t.Fatalf("ParsePlatformClaims: %v", err)
		}
		if !claims.IsAdmin {
			t.Error("expected IsAdmin=true")
		}
	})

	t.Run("admin flag is false", func(t *testing.T) {
		token, err := signer.IssueAccessToken("user-1", "my-org", false)
		if err != nil {
			t.Fatalf("IssueAccessToken: %v", err)
		}
		claims, err := signer.ParsePlatformClaims(token)
		if err != nil {
			t.Fatalf("ParsePlatformClaims: %v", err)
		}
		if claims.IsAdmin {
			t.Error("expected IsAdmin=false")
		}
	})
}
