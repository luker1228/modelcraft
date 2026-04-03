package auth

import (
	"testing"
	"time"
)

func TestRefreshToken_IsValid_ValidToken(t *testing.T) {
	token := &RefreshToken{
		ID:        "rt_1",
		UserID:    "user_1",
		TokenHash: "hash_1",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if !token.IsValid() {
		t.Fatalf("expected token to be valid")
	}

	if token.IsRevoked() {
		t.Fatalf("expected token not to be revoked")
	}
}

func TestRefreshToken_IsValid_RevokedToken(t *testing.T) {
	revokedAt := time.Now()
	token := &RefreshToken{
		ID:        "rt_1",
		UserID:    "user_1",
		TokenHash: "hash_1",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
		RevokedAt: &revokedAt,
	}

	if token.IsValid() {
		t.Fatalf("expected revoked token to be invalid")
	}

	if !token.IsRevoked() {
		t.Fatalf("expected token to be revoked")
	}
}

func TestRefreshToken_IsValid_ExpiredToken(t *testing.T) {
	token := &RefreshToken{
		ID:        "rt_1",
		UserID:    "user_1",
		TokenHash: "hash_1",
		ExpiresAt: time.Now().Add(-10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if token.IsValid() {
		t.Fatalf("expected expired token to be invalid")
	}

	if token.IsRevoked() {
		t.Fatalf("expected expired token not to be revoked")
	}
}
