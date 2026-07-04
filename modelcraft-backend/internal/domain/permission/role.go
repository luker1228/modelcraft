package permission

import (
	"modelcraft/pkg/bizerrors"
	"slices"
	"time"
)

// Role represents a role entity in the permission system.
// System roles (owner, admin, editor, viewer) are globally immutable with org_name='__SYSTEM__'.
// Custom roles are tenant-scoped and mutable.
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	OrgName     string    `json:"org_name"` // '__SYSTEM__' for system roles, org_name for custom roles
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	// SystemOrgName is the marker for system roles
	SystemOrgName = "__SYSTEM__"

	// System role names
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
	RoleGuest  = "guest" // anonymous demo access: read-only + create temporary API keys
)

// SystemRoles contains all system role names
var SystemRoles = []string{RoleOwner, RoleAdmin, RoleEditor, RoleViewer, RoleGuest}

// NewRole creates a new custom role
func NewRole(name, description, orgName string) *Role {
	return &Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
		OrgName:     orgName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewSystemRole creates a new system role (should only be used in migrations)
func NewSystemRole(name, description string) *Role {
	return &Role{
		Name:        name,
		Description: description,
		IsSystem:    true,
		OrgName:     SystemOrgName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// IsSystemRole checks if this role is a system role
func (r *Role) IsSystemRole() bool {
	return r.IsSystem || r.OrgName == SystemOrgName
}

// CanModify checks if this role can be modified
func (r *Role) CanModify() bool {
	return !r.IsSystemRole()
}

// CanDelete checks if this role can be deleted
func (r *Role) CanDelete() bool {
	return !r.IsSystemRole()
}

// IsSystemRoleName checks if a given name is a system role name
func IsSystemRoleName(name string) bool {
	return slices.Contains(SystemRoles, name)
}

// Validate validates the role entity
func (r *Role) Validate() error {
	// Validate name
	if r.Name == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "role name cannot be empty")
	}
	if len(r.Name) > 64 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "role name cannot exceed 64 characters")
	}

	// Validate org_name
	if r.OrgName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "org_name cannot be empty")
	}
	if len(r.OrgName) > 255 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "org_name cannot exceed 255 characters")
	}

	// System role validation
	if r.OrgName == SystemOrgName {
		return r.validateSystemRole()
	}

	return r.validateCustomRole()
}

func (r *Role) validateSystemRole() error {
	// If org_name is __SYSTEM__, it must be a valid system role
	if !r.IsSystem {
		return bizerrors.NewError(bizerrors.ParamInvalid, "custom role cannot use '__SYSTEM__' as org_name")
	}
	if !IsSystemRoleName(r.Name) {
		return bizerrors.NewError(
			bizerrors.ParamInvalid,
			"system role name must be one of: owner, admin, editor, viewer, guest",
		)
	}
	return nil
}

func (r *Role) validateCustomRole() error {
	// Custom role validation
	if r.IsSystem {
		return bizerrors.NewError(bizerrors.ParamInvalid, "system role must have org_name='__SYSTEM__'")
	}
	// Custom role cannot use system role names
	if IsSystemRoleName(r.Name) {
		return bizerrors.NewError(
			bizerrors.Conflict,
			"cannot create custom role with system role name: %s",
			r.Name,
		)
	}
	return nil
}
