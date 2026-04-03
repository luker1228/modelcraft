package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {
	t.Run("should create role with valid input", func(t *testing.T) {
		role, err := NewRole(
			"role-001",
			"Editor",
			"Can edit resources",
			[]string{"project:read", "project:update"},
			false,
		)
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "role-001", role.ID)
		assert.Equal(t, "Editor", role.Name)
		assert.Equal(t, "Can edit resources", role.Description)
		assert.Equal(t, []string{"project:read", "project:update"}, role.Permissions)
		assert.False(t, role.IsSystem)
		assert.False(t, role.CreatedAt.IsZero())
	})

	t.Run("should create system role", func(t *testing.T) {
		role, err := NewRole("role-owner", "Owner", "Full access", []string{"*"}, true)
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.True(t, role.IsSystem)
	})

	t.Run("should return error when ID is empty", func(t *testing.T) {
		role, err := NewRole("", "Editor", "desc", []string{"project:read"}, false)
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "role ID is required")
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		role, err := NewRole("role-001", "", "desc", []string{"project:read"}, false)
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "role name is required")
	})

	t.Run("should return error when permissions is nil", func(t *testing.T) {
		role, err := NewRole("role-001", "Editor", "desc", nil, false)
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "role permissions are required")
	})

	t.Run("should return error for invalid permission format", func(t *testing.T) {
		role, err := NewRole("role-001", "Editor", "desc", []string{"invalid"}, false)
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "invalid permission format")
	})
}

func TestRole_HasPermission(t *testing.T) {
	t.Run("wildcard * matches everything", func(t *testing.T) {
		r, _ := NewRole("r1", "Owner", "", []string{"*"}, true)
		assert.True(t, r.HasPermission("project:create"))
		assert.True(t, r.HasPermission("model:delete"))
		assert.True(t, r.HasPermission("user:invite"))
	})

	t.Run("exact match", func(t *testing.T) {
		r, _ := NewRole("r1", "Viewer", "", []string{"project:read", "model:read"}, false)
		assert.True(t, r.HasPermission("project:read"))
		assert.True(t, r.HasPermission("model:read"))
		assert.False(t, r.HasPermission("project:create"))
		assert.False(t, r.HasPermission("model:delete"))
	})

	t.Run("resource wildcard matches all actions for that resource", func(t *testing.T) {
		r, _ := NewRole("r1", "ProjectAdmin", "", []string{"project:*", "model:read"}, false)
		assert.True(t, r.HasPermission("project:create"))
		assert.True(t, r.HasPermission("project:read"))
		assert.True(t, r.HasPermission("project:update"))
		assert.True(t, r.HasPermission("project:delete"))
		assert.True(t, r.HasPermission("model:read"))
		assert.False(t, r.HasPermission("model:create"))
		assert.False(t, r.HasPermission("cluster:read"))
	})

	t.Run("no permissions matches nothing", func(t *testing.T) {
		r, _ := NewRole("r1", "NoPerms", "", []string{"project:read"}, false)
		assert.False(t, r.HasPermission("model:read"))
	})
}

func TestRole_CanDelete(t *testing.T) {
	t.Run("system role cannot be deleted", func(t *testing.T) {
		r, _ := NewRole("r1", "Owner", "", []string{"*"}, true)
		assert.False(t, r.CanDelete())
	})

	t.Run("custom role can be deleted", func(t *testing.T) {
		r, _ := NewRole("r1", "Custom", "", []string{"project:read"}, false)
		assert.True(t, r.CanDelete())
	})
}

func TestIsValidPermission(t *testing.T) {
	validPerms := []string{
		"*",
		"project:create",
		"project:read",
		"project:update",
		"project:delete",
		"project:*",
		"model:*",
		"user:invite",
		"user:remove",
		"user:list",
	}
	for _, p := range validPerms {
		t.Run("valid: "+p, func(t *testing.T) {
			assert.True(t, isValidPermission(p))
		})
	}

	invalidPerms := []string{
		"",
		"project",  // 缺少 action
		":create",  // 缺少 resource
		"project:", // 缺少 action
	}
	for _, p := range invalidPerms {
		t.Run("invalid: "+p, func(t *testing.T) {
			assert.False(t, isValidPermission(p))
		})
	}
}
