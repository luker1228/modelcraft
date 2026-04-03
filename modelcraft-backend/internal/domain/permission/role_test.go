package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRole tests the NewRole constructor
func TestNewRole(t *testing.T) {
	t.Run("should create custom role with valid input", func(t *testing.T) {
		// Arrange & Act
		role := NewRole("data-analyst", "Read-only data access", "org1")

		// Assert
		assert.NotNil(t, role)
		assert.Equal(t, "data-analyst", role.Name)
		assert.Equal(t, "Read-only data access", role.Description)
		assert.Equal(t, "org1", role.OrgName)
		assert.False(t, role.IsSystem)
		assert.False(t, role.IsSystemRole())
	})
}

// TestNewSystemRole tests the NewSystemRole constructor
func TestNewSystemRole(t *testing.T) {
	t.Run("should create system role with correct properties", func(t *testing.T) {
		// Arrange & Act
		role := NewSystemRole("owner", "Full control")

		// Assert
		assert.NotNil(t, role)
		assert.Equal(t, "owner", role.Name)
		assert.Equal(t, "Full control", role.Description)
		assert.Equal(t, SystemOrgName, role.OrgName)
		assert.True(t, role.IsSystem)
		assert.True(t, role.IsSystemRole())
	})
}

// TestRole_IsSystemRole tests the IsSystemRole method
func TestRole_IsSystemRole(t *testing.T) {
	tests := []struct {
		name     string
		role     *Role
		expected bool
	}{
		{
			name: "system role with is_system=true",
			role: &Role{
				Name:     "owner",
				IsSystem: true,
				OrgName:  SystemOrgName,
			},
			expected: true,
		},
		{
			name: "system role with org_name=__SYSTEM__",
			role: &Role{
				Name:     "admin",
				IsSystem: false,
				OrgName:  SystemOrgName,
			},
			expected: true,
		},
		{
			name: "custom role",
			role: &Role{
				Name:     "data-analyst",
				IsSystem: false,
				OrgName:  "org1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.IsSystemRole()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRole_CanModify tests the CanModify method
func TestRole_CanModify(t *testing.T) {
	tests := []struct {
		name     string
		role     *Role
		expected bool
	}{
		{
			name: "system role cannot be modified",
			role: &Role{
				Name:     "owner",
				IsSystem: true,
				OrgName:  SystemOrgName,
			},
			expected: false,
		},
		{
			name: "custom role can be modified",
			role: &Role{
				Name:     "data-analyst",
				IsSystem: false,
				OrgName:  "org1",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.CanModify()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRole_CanDelete tests the CanDelete method
func TestRole_CanDelete(t *testing.T) {
	tests := []struct {
		name     string
		role     *Role
		expected bool
	}{
		{
			name: "system role cannot be deleted",
			role: &Role{
				Name:     "admin",
				IsSystem: true,
				OrgName:  SystemOrgName,
			},
			expected: false,
		},
		{
			name: "custom role can be deleted",
			role: &Role{
				Name:     "data-analyst",
				IsSystem: false,
				OrgName:  "org1",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.CanDelete()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSystemRoleName tests the IsSystemRoleName function
func TestIsSystemRoleName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		expected bool
	}{
		{name: "owner is system role", roleName: "owner", expected: true},
		{name: "admin is system role", roleName: "admin", expected: true},
		{name: "editor is system role", roleName: "editor", expected: true},
		{name: "viewer is system role", roleName: "viewer", expected: true},
		{name: "data-analyst is not system role", roleName: "data-analyst", expected: false},
		{name: "custom is not system role", roleName: "custom", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSystemRoleName(tt.roleName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRole_Validate tests the Validate method
func TestRole_Validate(t *testing.T) {
	t.Run("valid custom role", func(t *testing.T) {
		role := &Role{
			Name:     "data-analyst",
			OrgName:  "org1",
			IsSystem: false,
		}
		err := role.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid system role", func(t *testing.T) {
		role := &Role{
			Name:     "owner",
			OrgName:  SystemOrgName,
			IsSystem: true,
		}
		err := role.Validate()
		assert.NoError(t, err)
	})

	t.Run("error: empty name", func(t *testing.T) {
		role := &Role{
			Name:     "",
			OrgName:  "org1",
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role name cannot be empty")
	})

	t.Run("error: name too long", func(t *testing.T) {
		role := &Role{
			Name:     string(make([]byte, 65)),
			OrgName:  "org1",
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role name cannot exceed 64 characters")
	})

	t.Run("error: empty org_name", func(t *testing.T) {
		role := &Role{
			Name:     "data-analyst",
			OrgName:  "",
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org_name cannot be empty")
	})

	t.Run("error: org_name too long", func(t *testing.T) {
		role := &Role{
			Name:     "data-analyst",
			OrgName:  string(make([]byte, 256)),
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org_name cannot exceed 255 characters")
	})

	t.Run("error: system role with wrong org_name", func(t *testing.T) {
		role := &Role{
			Name:     "owner",
			OrgName:  "org1",
			IsSystem: true,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system role must have org_name='__SYSTEM__'")
	})

	t.Run("error: system role with invalid name", func(t *testing.T) {
		role := &Role{
			Name:     "custom",
			OrgName:  SystemOrgName,
			IsSystem: true,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system role name must be one of: owner, admin, editor, viewer")
	})

	t.Run("error: custom role using system role name", func(t *testing.T) {
		role := &Role{
			Name:     "owner",
			OrgName:  "org1",
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot create custom role with system role name")
	})

	t.Run("error: custom role using __SYSTEM__ org_name", func(t *testing.T) {
		role := &Role{
			Name:     "data-analyst",
			OrgName:  SystemOrgName,
			IsSystem: false,
		}
		err := role.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom role cannot use '__SYSTEM__' as org_name")
	})
}
