package enduser

import "context"

// EndUserSessionRepository defines persistence operations for end-user sessions.
// Operations target mc_meta.end_user_accounts and are tenant-scoped by org/project.
type EndUserSessionRepository interface {
	// Save creates a new session record.
	Save(ctx context.Context, session *EndUserSession) error

	// GetByTokenHash retrieves a session by sha256(token) (returns (nil, nil) when not found).
	GetByTokenHash(ctx context.Context, tokenHash string) (*EndUserSession, error)

	// RevokeByID marks the specified session as revoked=1.
	RevokeByID(ctx context.Context, id string) error

	// RevokeAllByUserID revokes all active sessions for a user (called when deleting a user).
	RevokeAllByUserID(ctx context.Context, userID string) error
}
