package auth

import "time"

const APIKeyMaxPerUser = 20

// APIKey represents a persisted API key record.
type APIKey struct {
	ID         string
	UserID     string
	Name       string
	KeyHash    string
	KeyPrefix  string
	RoleIDs    []int
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
}

// IsValid checks whether the API key can still be used.
func (k *APIKey) IsValid() bool {
	if k == nil {
		return false
	}
	if k.ID == "" || k.UserID == "" || k.Name == "" || k.KeyHash == "" || k.KeyPrefix == "" {
		return false
	}
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && !k.ExpiresAt.After(time.Now()) {
		return false
	}
	return true
}
