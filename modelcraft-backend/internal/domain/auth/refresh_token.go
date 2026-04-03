package auth

import "time"

// RefreshToken represents a persisted refresh token record.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// IsRevoked checks whether the refresh token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsValid checks whether the refresh token can still be used.
func (t *RefreshToken) IsValid() bool {
	if t == nil {
		return false
	}
	if t.ID == "" || t.UserID == "" || t.TokenHash == "" {
		return false
	}
	if t.IsRevoked() {
		return false
	}
	return t.ExpiresAt.After(time.Now())
}
