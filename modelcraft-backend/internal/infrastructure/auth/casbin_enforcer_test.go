package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetEnforcer tests the singleton enforcer initialization
func TestGetEnforcer(t *testing.T) {
	t.Run("should initialize enforcer successfully", func(t *testing.T) {
		// Act
		e, err := GetEnforcer()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("should return same instance on multiple calls", func(t *testing.T) {
		// Act
		e1, err1 := GetEnforcer()
		e2, err2 := GetEnforcer()

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Same(t, e1, e2, "Should return same enforcer instance")
	})
}

// TestLoadSystemRolePermissions tests loading hardcoded system role permissions
func TestLoadSystemRolePermissions(t *testing.T) {
	t.Run("should load all system role permissions", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)

		// Act
		err = LoadSystemRolePermissions(e)

		// Assert
		assert.NoError(t, err)

		// Verify owner has wildcard permission
		policies, _ := e.GetFilteredPolicy(0, "owner")
		assert.NotEmpty(t, policies, "Owner should have permissions")
		found := false
		for _, p := range policies {
			if len(p) >= 3 && p[1] == "*" && p[2] == "*" {
				found = true
				break
			}
		}
		assert.True(t, found, "Owner should have wildcard permission")

		// Verify admin has wildcard permission
		policies, _ = e.GetFilteredPolicy(0, "admin")
		assert.NotEmpty(t, policies, "Admin should have permissions")
		found = false
		for _, p := range policies {
			if len(p) >= 3 && p[1] == "*" && p[2] == "*" {
				found = true
				break
			}
		}
		assert.True(t, found, "Admin should have wildcard permission")

		// Verify viewer has read-only permissions
		policies, _ = e.GetFilteredPolicy(0, "viewer")
		assert.NotEmpty(t, policies, "Viewer should have permissions")
		for _, p := range policies {
			if len(p) >= 3 {
				assert.Equal(t, "read", p[2], "Viewer should only have 'read' actions")
			}
		}
	})
}

// TestAddUserRole tests adding user-role mappings
func TestAddUserRole(t *testing.T) {
	t.Run("should add user-role mapping successfully", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)
		err = LoadSystemRolePermissions(e)
		require.NoError(t, err)

		// Act
		err = AddUserRole(e, "user123", "admin")

		// Assert
		assert.NoError(t, err)

		// Verify mapping exists
		roles, _ := e.GetRolesForUser("user123")
		assert.Contains(t, roles, "admin")
	})

	t.Run("should allow multiple roles for same user", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)

		// Act
		err1 := AddUserRole(e, "user456", "admin")
		err2 := AddUserRole(e, "user456", "viewer")

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		roles, _ := e.GetRolesForUser("user456")
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "viewer")
	})
}

// TestRemoveUserRole tests removing user-role mappings
func TestRemoveUserRole(t *testing.T) {
	t.Run("should remove user-role mapping successfully", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)
		err = AddUserRole(e, "user789", "admin")
		require.NoError(t, err)

		// Act
		err = RemoveUserRole(e, "user789", "admin")

		// Assert
		assert.NoError(t, err)

		roles, _ := e.GetRolesForUser("user789")
		assert.NotContains(t, roles, "admin")
	})
}

// TestAddCustomRolePermission tests adding custom role permissions
func TestAddCustomRolePermission(t *testing.T) {
	t.Run("should add custom role permission successfully", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)

		// Act
		err = AddCustomRolePermission(e, "data-analyst", "model", "read")

		// Assert
		assert.NoError(t, err)

		// Verify permission exists
		policies, _ := e.GetFilteredPolicy(0, "data-analyst")
		assert.NotEmpty(t, policies)
		found := false
		for _, p := range policies {
			if len(p) >= 3 && p[1] == "model" && p[2] == "read" {
				found = true
				break
			}
		}
		assert.True(t, found, "Custom role should have model:read permission")
	})
}

// TestRemoveCustomRolePermission tests removing custom role permissions
func TestRemoveCustomRolePermission(t *testing.T) {
	t.Run("should remove custom role permission successfully", func(t *testing.T) {
		// Arrange
		e, err := GetEnforcer()
		require.NoError(t, err)
		err = AddCustomRolePermission(e, "researcher", "project", "read")
		require.NoError(t, err)

		// Act
		err = RemoveCustomRolePermission(e, "researcher", "project", "read")

		// Assert
		assert.NoError(t, err)

		// Verify permission removed
		policies, _ := e.GetFilteredPolicy(0, "researcher", "project", "read")
		assert.Empty(t, policies)
	})
}

// TestCheckPermission tests permission checking
func TestCheckPermission(t *testing.T) {
	// Setup: Load system roles and add user-role mappings
	e, err := GetEnforcer()
	require.NoError(t, err)
	err = LoadSystemRolePermissions(e)
	require.NoError(t, err)

	t.Run("owner has wildcard permission", func(t *testing.T) {
		// Arrange
		err := AddUserRole(e, "alice", "owner")
		require.NoError(t, err)

		// Act
		allowed, err := CheckPermission(e, "alice", "project", "create")

		// Assert
		assert.NoError(t, err)
		assert.True(t, allowed, "Owner should have project:create permission")

		// Test another resource
		allowed, err = CheckPermission(e, "alice", "model", "delete")
		assert.NoError(t, err)
		assert.True(t, allowed, "Owner should have model:delete permission")
	})

	t.Run("admin has specific permissions", func(t *testing.T) {
		// Arrange
		err := AddUserRole(e, "bob", "admin")
		require.NoError(t, err)

		// Act
		allowed, err := CheckPermission(e, "bob", "project", "create")

		// Assert
		assert.NoError(t, err)
		assert.True(t, allowed, "Admin should have project:create permission")
	})

	t.Run("viewer has read-only permissions", func(t *testing.T) {
		// Arrange
		err := AddUserRole(e, "charlie", "viewer")
		require.NoError(t, err)

		// Act - allowed action
		allowed, err := CheckPermission(e, "charlie", "project", "read")
		assert.NoError(t, err)
		assert.True(t, allowed, "Viewer should have project:read permission")

		// Act - denied action
		allowed, err = CheckPermission(e, "charlie", "project", "delete")
		assert.NoError(t, err)
		assert.False(t, allowed, "Viewer should NOT have project:delete permission")
	})

	t.Run("custom role with specific permission", func(t *testing.T) {
		// Arrange
		err := AddCustomRolePermission(e, "data-scientist", "model", "read")
		require.NoError(t, err)
		err = AddUserRole(e, "david", "data-scientist")
		require.NoError(t, err)

		// Act - allowed action
		allowed, err := CheckPermission(e, "david", "model", "read")
		assert.NoError(t, err)
		assert.True(t, allowed, "Custom role should have model:read permission")

		// Act - denied action
		allowed, err = CheckPermission(e, "david", "model", "delete")
		assert.NoError(t, err)
		assert.False(t, allowed, "Custom role should NOT have model:delete permission")
	})

	t.Run("user with no role has no permissions", func(t *testing.T) {
		// Act
		allowed, err := CheckPermission(e, "eve", "project", "read")

		// Assert
		assert.NoError(t, err)
		assert.False(t, allowed, "User with no role should have no permissions")
	})

	t.Run("user with multiple roles gets combined permissions", func(t *testing.T) {
		// Arrange
		err := AddUserRole(e, "frank", "viewer")
		require.NoError(t, err)
		err = AddCustomRolePermission(e, "editor-limited", "model", "update")
		require.NoError(t, err)
		err = AddUserRole(e, "frank", "editor-limited")
		require.NoError(t, err)

		// Act - permission from viewer role
		allowed, err := CheckPermission(e, "frank", "project", "read")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should have permission from viewer role")

		// Act - permission from custom role
		allowed, err = CheckPermission(e, "frank", "model", "update")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should have permission from custom role")

		// Act - no permission from any role
		allowed, err = CheckPermission(e, "frank", "project", "delete")
		assert.NoError(t, err)
		assert.False(t, allowed, "Should NOT have permission from any role")
	})
}

// TestGetSystemRolePermissions tests retrieving system role permissions
func TestGetSystemRolePermissions(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		expected bool
	}{
		{name: "owner has permissions", roleName: "owner", expected: true},
		{name: "admin has permissions", roleName: "admin", expected: true},
		{name: "editor has permissions", roleName: "editor", expected: true},
		{name: "viewer has permissions", roleName: "viewer", expected: true},
		{name: "custom role has no hardcoded permissions", roleName: "data-analyst", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := GetSystemRolePermissions(tt.roleName)
			if tt.expected {
				assert.NotNil(t, perms)
				assert.NotEmpty(t, perms)
			} else {
				assert.Nil(t, perms)
			}
		})
	}
}

// TestIsSystemRole tests system role checking
func TestIsSystemRole(t *testing.T) {
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
		{name: "unknown role is not system role", roleName: "unknown", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSystemRole(tt.roleName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
