package auth

import (
	"context"
	"time"
)

// APIKeyRepository defines persistence operations for API keys.
type APIKeyRepository interface {
	// Save persists a new API key.
	Save(ctx context.Context, key *APIKey) error
	// FindByHash returns (nil, nil) when not found.
	FindByHash(ctx context.Context, hash string) (*APIKey, error)
	// FindByID returns (nil, nil) when not found.
	FindByID(ctx context.Context, id string) (*APIKey, error)
	// ListByUserID lists all active API keys for a user.
	ListByUserID(ctx context.Context, userID string) ([]*APIKey, error)
	// CountActiveByUserID counts active API keys for quota checks.
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
	// Revoke revokes a key by id/user scope.
	Revoke(ctx context.Context, id, userID string) error
	// Update updates key display name and expiration.
	Update(ctx context.Context, id, userID, name string, roleIDs []int, expiresAt *time.Time) error
	// UpdateLastUsed updates the last-used timestamp.
	UpdateLastUsed(ctx context.Context, id string) error
	// DeleteRevoked cleans long-revoked API keys.
	DeleteRevoked(ctx context.Context) error
}
