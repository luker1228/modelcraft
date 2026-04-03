package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPermission tests the NewPermission constructor
func TestNewPermission(t *testing.T) {
	t.Run("should create permission with valid input", func(t *testing.T) {
		// Arrange & Act
		perm := NewPermission("project", "create")

		// Assert
		assert.NotNil(t, perm)
		assert.Equal(t, "project", perm.Obj)
		assert.Equal(t, "create", perm.Act)
	})
}

// TestNewPermissionFromString tests parsing permission from string
func TestNewPermissionFromString(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		expectError bool
		expectedObj string
		expectedAct string
	}{
		{
			name:        "valid permission format",
			action:      "project:create",
			expectError: false,
			expectedObj: "project",
			expectedAct: "create",
		},
		{
			name:        "wildcard permission",
			action:      "*:*",
			expectError: false,
			expectedObj: "*",
			expectedAct: "*",
		},
		{
			name:        "permission with spaces",
			action:      " project : create ",
			expectError: false,
			expectedObj: "project",
			expectedAct: "create",
		},
		{
			name:        "error: invalid format without colon",
			action:      "projectcreate",
			expectError: true,
		},
		{
			name:        "error: invalid format with multiple colons",
			action:      "project:model:create",
			expectError: true,
		},
		{
			name:        "error: empty string",
			action:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := NewPermissionFromString(tt.action)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, perm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, perm)
				assert.Equal(t, tt.expectedObj, perm.Obj)
				assert.Equal(t, tt.expectedAct, perm.Act)
			}
		})
	}
}

// TestPermission_String tests the String method
func TestPermission_String(t *testing.T) {
	tests := []struct {
		name     string
		perm     *Permission
		expected string
	}{
		{
			name:     "basic permission",
			perm:     &Permission{Obj: "project", Act: "create"},
			expected: "project:create",
		},
		{
			name:     "wildcard permission",
			perm:     &Permission{Obj: "*", Act: "*"},
			expected: "*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.perm.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_IsWildcard tests the IsWildcard method
func TestPermission_IsWildcard(t *testing.T) {
	tests := []struct {
		name     string
		perm     *Permission
		expected bool
	}{
		{
			name:     "full wildcard",
			perm:     &Permission{Obj: "*", Act: "*"},
			expected: true,
		},
		{
			name:     "object wildcard only",
			perm:     &Permission{Obj: "*", Act: "create"},
			expected: false,
		},
		{
			name:     "action wildcard only",
			perm:     &Permission{Obj: "project", Act: "*"},
			expected: false,
		},
		{
			name:     "no wildcard",
			perm:     &Permission{Obj: "project", Act: "create"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.perm.IsWildcard()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_IsObjectWildcard tests the IsObjectWildcard method
func TestPermission_IsObjectWildcard(t *testing.T) {
	tests := []struct {
		name     string
		perm     *Permission
		expected bool
	}{
		{
			name:     "action wildcard",
			perm:     &Permission{Obj: "project", Act: "*"},
			expected: true,
		},
		{
			name:     "full wildcard",
			perm:     &Permission{Obj: "*", Act: "*"},
			expected: true,
		},
		{
			name:     "no wildcard",
			perm:     &Permission{Obj: "project", Act: "create"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.perm.IsObjectWildcard()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_Matches tests the Matches method
func TestPermission_Matches(t *testing.T) {
	tests := []struct {
		name     string
		perm     *Permission
		other    *Permission
		expected bool
	}{
		{
			name:     "exact match",
			perm:     &Permission{Obj: "project", Act: "create"},
			other:    &Permission{Obj: "project", Act: "create"},
			expected: true,
		},
		{
			name:     "wildcard matches anything",
			perm:     &Permission{Obj: "*", Act: "*"},
			other:    &Permission{Obj: "project", Act: "create"},
			expected: true,
		},
		{
			name:     "object wildcard matches any action on same resource",
			perm:     &Permission{Obj: "project", Act: "*"},
			other:    &Permission{Obj: "project", Act: "delete"},
			expected: true,
		},
		{
			name:     "resource wildcard matches any resource with same action",
			perm:     &Permission{Obj: "*", Act: "read"},
			other:    &Permission{Obj: "model", Act: "read"},
			expected: true,
		},
		{
			name:     "different object does not match",
			perm:     &Permission{Obj: "project", Act: "create"},
			other:    &Permission{Obj: "model", Act: "create"},
			expected: false,
		},
		{
			name:     "different action does not match",
			perm:     &Permission{Obj: "project", Act: "create"},
			other:    &Permission{Obj: "project", Act: "delete"},
			expected: false,
		},
		{
			name:     "nil other permission",
			perm:     &Permission{Obj: "project", Act: "create"},
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.perm.Matches(tt.other)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_Validate tests the Validate method
func TestPermission_Validate(t *testing.T) {
	t.Run("valid permission", func(t *testing.T) {
		perm := &Permission{Obj: "project", Act: "create"}
		err := perm.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid wildcard permission", func(t *testing.T) {
		perm := &Permission{Obj: "*", Act: "*"}
		err := perm.Validate()
		assert.NoError(t, err)
	})

	t.Run("error: empty object", func(t *testing.T) {
		perm := &Permission{Obj: "", Act: "create"}
		err := perm.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission object cannot be empty")
	})

	t.Run("error: object too long", func(t *testing.T) {
		perm := &Permission{Obj: string(make([]byte, 65)), Act: "create"}
		err := perm.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission object cannot exceed 64 characters")
	})

	t.Run("error: empty action", func(t *testing.T) {
		perm := &Permission{Obj: "project", Act: ""}
		err := perm.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission action cannot be empty")
	})

	t.Run("error: action too long", func(t *testing.T) {
		perm := &Permission{Obj: "project", Act: string(make([]byte, 65))}
		err := perm.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission action cannot exceed 64 characters")
	})
}
