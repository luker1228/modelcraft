package auth

import (
	"context"
)

// AuthProvider defines the interface for authentication providers.
// Implementations include KeycloakProvider (future) and OIDCProvider (future).
type AuthProvider interface {
	// GetPublicKey returns the public key for JWT signature verification.
	// The returned key must be compatible with jwt.Parse (e.g., *rsa.PublicKey, *ecdsa.PublicKey).
	// Implementations may cache the key to avoid repeated parsing.
	GetPublicKey(ctx context.Context) (interface{}, error)

	// GetSigningMethod returns the JWT signing algorithm identifier (e.g., "RS256", "HS256", "ES256").
	GetSigningMethod() string

	// Type returns a unique identifier for this provider (e.g., "keycloak", "oidc").
	Type() string
}
