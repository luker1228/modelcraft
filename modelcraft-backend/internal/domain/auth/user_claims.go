package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUserClaimsInvalid = errors.New("invalid user claims")

// UserClaims represents a simplified JWT token containing ONLY user identity.
type UserClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
}

// RefreshClaims represents the JWT claims for refresh tokens.
type RefreshClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
}

// Validate checks that the UserClaims contains all required fields and is not expired.
func (c *UserClaims) Validate() error {
	if c.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrUserClaimsInvalid)
	}

	if c.Issuer != string(IssuerPlatform) {
		return fmt.Errorf("%w: expected issuer '%s', got '%s'", ErrUserClaimsInvalid, IssuerPlatform, c.Issuer)
	}

	if c.ExpiresAt == nil || c.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	return nil
}

// IsExpired checks if the token has expired.
func (c *UserClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time) || time.Now().Equal(c.ExpiresAt.Time)
}
