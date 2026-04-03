package permission

import (
	"modelcraft/pkg/bizerrors"
	"time"
)

// UserRole represents a user-role binding in a specific organization.
// A user can have different roles in different organizations.
type UserRole struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    int       `json:"role_id"`
	OrgName   string    `json:"org_name"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserRole creates a new user-role binding
func NewUserRole(userID string, roleID int, orgName string) *UserRole {
	return &UserRole{
		UserID:    userID,
		RoleID:    roleID,
		OrgName:   orgName,
		CreatedAt: time.Now(),
	}
}

// Validate validates the user-role binding
func (ur *UserRole) Validate() error {
	// Validate user_id
	if ur.UserID == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "user_id cannot be empty")
	}
	if len(ur.UserID) > 36 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "user_id cannot exceed 36 characters")
	}

	// Validate role_id
	if ur.RoleID <= 0 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "role_id must be a positive integer")
	}

	// Validate org_name
	if ur.OrgName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "org_name cannot be empty")
	}
	if len(ur.OrgName) > 255 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "org_name cannot exceed 255 characters")
	}

	return nil
}

// IsSameBinding checks if this user-role binding is the same as another
func (ur *UserRole) IsSameBinding(other *UserRole) bool {
	if other == nil {
		return false
	}
	return ur.UserID == other.UserID &&
		ur.RoleID == other.RoleID &&
		ur.OrgName == other.OrgName
}
