package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUserNotFound       = errors.New("user not found in ModelCraft database")
	ErrInvalidExternalJWT = errors.New("invalid external JWT")
	ErrTokenGeneration    = errors.New("failed to generate ModelCraft JWT")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid or malformed token")
	ErrPermissionsInvalid = errors.New("invalid permissions format in token")
)

// MembershipClaimInfo represents a user's membership in an organization (for JWT claims)
type MembershipClaimInfo struct {
	OrgName     string `json:"orgName"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	JoinedAt    int64  `json:"joinedAt"` // Unix timestamp in milliseconds
}

// ModelCraftClaims represents the custom claims for ModelCraft JWT tokens.
// This structure separates ModelCraft authentication/authorization from external provider tokens.
type ModelCraftClaims struct {
	jwt.RegisteredClaims

	// User identity
	UserID     string `json:"user_id"`     // ModelCraft user UUID
	ExternalID string `json:"external_id"` // External provider user ID
	Name       string `json:"name"`
	Email      string `json:"email"`

	// Organization
	Organization string `json:"organization"`

	// Authorization
	Roles       []string `json:"roles"`       // Role names: ["owner", "editor"]
	Permissions []string `json:"permissions"` // Permission strings: ["model:read", "model:write"]

	// Memberships (limited to 10 for JWT size control)
	Memberships        []MembershipClaimInfo `json:"memberships,omitempty"`
	HasMoreMemberships bool                  `json:"hasMoreMemberships,omitempty"` // True if user has >10 memberships

	// Issuer (always "modelcraft")
	Issuer string `json:"iss"`
}

// Validate checks that the ModelCraftClaims contains all required fields and is not expired.
func (c *ModelCraftClaims) Validate() error {
	if c.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidToken)
	}

	if c.Issuer != "modelcraft" {
		return fmt.Errorf("%w: expected issuer 'modelcraft', got '%s'", ErrInvalidToken, c.Issuer)
	}

	if c.ExpiresAt == nil || c.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	// Validate permissions format
	for _, perm := range c.Permissions {
		if perm == "" {
			return fmt.Errorf("%w: empty permission string", ErrPermissionsInvalid)
		}
	}

	return nil
}

// IsExpired checks if the token has expired.
func (c *ModelCraftClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time) || time.Now().Equal(c.ExpiresAt.Time)
}

// HasPermission checks if the token has a specific permission.
func (c *ModelCraftClaims) HasPermission(perm string) bool {
	for _, p := range c.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the token has any of the specified permissions.
func (c *ModelCraftClaims) HasAnyPermission(perms ...string) bool {
	for _, neededPerm := range perms {
		if c.HasPermission(neededPerm) {
			return true
		}
	}
	return false
}

// HasRole checks if the token has a specific role.
func (c *ModelCraftClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}
