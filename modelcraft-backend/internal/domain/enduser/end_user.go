package enduser

import (
	"fmt"
	"time"
)

// EndUser represents an end-user entity (aggregate root).
// EndUser is stored in mc_meta.end_user_users with org tenant scope (org_name).
type EndUser struct {
	ID          string         // UUID, primary key
	OrgName     string         // org scope key
	Username    string         // 3-64 chars, ^[a-zA-Z0-9_-]+$, unique within org
	Phone       string         // 11-digit mainland China mobile number
	Password    HashedPassword // bcrypt hashed
	IsForbidden bool           // whether the account is disabled
	IsAdmin     bool           // whether the user is an org admin (user_orgs.is_admin = 1)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewEndUser creates a new EndUser with validation.
// The password must already be hashed.
func NewEndUser(id, orgName, username, phone string, hashedPwd HashedPassword) (*EndUser, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgName == "" {
		return nil, fmt.Errorf("org name is required")
	}

	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	if phone == "" {
		return nil, fmt.Errorf("phone is required")
	}

	now := time.Now()
	return &EndUser{
		ID:          id,
		OrgName:     orgName,
		Username:    username,
		Phone:       phone,
		Password:    hashedPwd,
		IsForbidden: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Enable enables the account (DISABLED → ACTIVE).
func (u *EndUser) Enable() {
	u.IsForbidden = false
	u.UpdatedAt = time.Now()
}

// Disable disables the account (ACTIVE → DISABLED).
func (u *EndUser) Disable() {
	u.IsForbidden = true
	u.UpdatedAt = time.Now()
}

// IsActive returns whether the user can log in (not disabled).
func (u *EndUser) IsActive() bool {
	return !u.IsForbidden
}

// VerifyPassword verifies the plaintext password against the stored hash.
func (u *EndUser) VerifyPassword(plainPassword string) bool {
	return u.Password.Verify(plainPassword)
}
