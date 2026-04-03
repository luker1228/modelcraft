package auth

import "context"

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	// Save persists a new refresh token (hash should be precomputed by caller).
	Save(ctx context.Context, token *RefreshToken) error
	// FindByHash returns (nil, nil) when not found.
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	// Revoke revokes a refresh token by ID.
	Revoke(ctx context.Context, id string) error
	// RevokeAllByUserID revokes all active refresh tokens for a user.
	RevokeAllByUserID(ctx context.Context, userID string) error
	// DeleteExpired cleans expired and stale revoked refresh tokens.
	DeleteExpired(ctx context.Context) error
}
