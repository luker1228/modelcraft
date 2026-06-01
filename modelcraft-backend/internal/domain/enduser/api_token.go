package enduser

import "time"

// APIToken represents an EndUser Personal Access Token (PAT).
// Stored in mc_meta.end_user_api_tokens with org tenant scope.
type APIToken struct {
	ID          string     // UUID v7, primary key
	OrgName     string     // org scope key
	EndUserID   string     // FK → end_user_users.id
	Name        string     // human-readable label
	TokenHash   string     // SHA-256(plaintext token) hex
	ExpiresAt   *time.Time // nil = never expires
	LastUsedAt  *time.Time // nil = never used
	CreatedAt   time.Time
	DeletedAt   int64 // 0 = active, Unix ms = deleted
	DeleteToken int64 // 0 = active (soft-delete collision-avoidance token)
}

// IsValid returns true if the token is active and not expired.
func (t *APIToken) IsValid() bool {
	if t.DeletedAt != 0 {
		return false
	}
	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
		return false
	}
	return true
}
