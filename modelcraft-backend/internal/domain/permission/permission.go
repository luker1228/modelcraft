package permission

import (
	"fmt"
	"modelcraft/pkg/bizerrors"
	"strings"
)

// Permission represents a permission value object.
// Permissions define what actions can be performed on resources.
// Format: obj:act (e.g., "project:create", "model:read", "*:*")
type Permission struct {
	Obj string `json:"obj"` // Resource object (e.g., "project", "model", "*")
	Act string `json:"act"` // Action (e.g., "create", "read", "update", "delete", "*")
}

// NewPermission creates a new permission value object
func NewPermission(obj, act string) *Permission {
	return &Permission{
		Obj: obj,
		Act: act,
	}
}

// NewPermissionFromString parses a permission string in format "obj:act"
func NewPermissionFromString(action string) (*Permission, error) {
	parts := strings.Split(action, ":")
	if len(parts) != 2 {
		return nil, bizerrors.NewError(
			bizerrors.ParamInvalid,
			"invalid permission format: must be 'resource:operation', "+
				"got: %s",
			action,
		)
	}

	return &Permission{
		Obj: strings.TrimSpace(parts[0]),
		Act: strings.TrimSpace(parts[1]),
	}, nil
}

// String returns the string representation of the permission
func (p *Permission) String() string {
	return fmt.Sprintf("%s:%s", p.Obj, p.Act)
}

// IsWildcard checks if this is a wildcard permission
func (p *Permission) IsWildcard() bool {
	return p.Obj == "*" && p.Act == "*"
}

// IsObjectWildcard checks if this permission has wildcard for actions
func (p *Permission) IsObjectWildcard() bool {
	return p.Act == "*"
}

// Matches checks if this permission matches another permission
// Supports wildcard matching (e.g., "*:*" matches anything)
func (p *Permission) Matches(other *Permission) bool {
	if other == nil {
		return false
	}

	// Wildcard permission matches everything
	if p.IsWildcard() {
		return true
	}

	// Object must match
	if p.Obj != "*" && p.Obj != other.Obj {
		return false
	}

	// Action must match or be wildcard
	if p.Act != "*" && p.Act != other.Act {
		return false
	}

	return true
}

// Validate validates the permission value object
func (p *Permission) Validate() error {
	// Validate obj
	if p.Obj == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "permission object cannot be empty")
	}
	if len(p.Obj) > 64 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "permission object cannot exceed 64 characters")
	}

	// Validate act
	if p.Act == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "permission action cannot be empty")
	}
	if len(p.Act) > 64 {
		return bizerrors.NewError(bizerrors.ParamInvalid, "permission action cannot exceed 64 characters")
	}

	return nil
}
