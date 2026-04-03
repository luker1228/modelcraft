package middleware

import (
	"testing"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name            string
		userPermissions []string
		required        string
		want            bool
	}{
		// Exact match
		{
			name:            "exact match grants access",
			userPermissions: []string{"project:create"},
			required:        "project:create",
			want:            true,
		},
		{
			name:            "no match denies access",
			userPermissions: []string{"project:read"},
			required:        "project:create",
			want:            false,
		},
		// Global wildcard
		{
			name:            "global wildcard grants any permission",
			userPermissions: []string{"*"},
			required:        "project:create",
			want:            true,
		},
		{
			name:            "global wildcard grants nested permission",
			userPermissions: []string{"*"},
			required:        "model:delete",
			want:            true,
		},
		// Resource wildcard
		{
			name:            "resource wildcard grants action on resource",
			userPermissions: []string{"project:*"},
			required:        "project:create",
			want:            true,
		},
		{
			name:            "resource wildcard grants different action",
			userPermissions: []string{"project:*"},
			required:        "project:delete",
			want:            true,
		},
		{
			name:            "resource wildcard does not grant other resource",
			userPermissions: []string{"project:*"},
			required:        "model:create",
			want:            false,
		},
		// Multiple permissions
		{
			name:            "multiple permissions match second",
			userPermissions: []string{"model:read", "project:create"},
			required:        "project:create",
			want:            true,
		},
		{
			name:            "multiple permissions with wildcard",
			userPermissions: []string{"model:read", "project:*"},
			required:        "project:delete",
			want:            true,
		},
		// Edge cases
		{
			name:            "empty permissions deny access",
			userPermissions: []string{},
			required:        "project:create",
			want:            false,
		},
		{
			name:            "nil permissions deny access",
			userPermissions: nil,
			required:        "project:create",
			want:            false,
		},
		{
			name:            "partial resource name does not match",
			userPermissions: []string{"proj:*"},
			required:        "project:create",
			want:            false,
		},
		{
			name:            "user invite permission",
			userPermissions: []string{"user:invite", "user:remove"},
			required:        "user:invite",
			want:            true,
		},
		{
			name:            "owner permissions grant everything",
			userPermissions: []string{"*"},
			required:        "user:invite",
			want:            true,
		},
		{
			name: "admin with resource wildcards",
			userPermissions: []string{
				"project:*",
				"model:*",
				"cluster:*",
				"enum:*",
				"user:invite",
				"user:remove",
				"user:list",
			},
			required: "project:create",
			want:     true,
		},
		{
			name: "admin lacks wildcard user permissions",
			userPermissions: []string{
				"project:*",
				"model:*",
				"cluster:*",
				"enum:*",
				"user:invite",
				"user:remove",
				"user:list",
			},
			required: "user:delete",
			want:     false,
		},
		{
			name:            "viewer cannot create",
			userPermissions: []string{"project:read", "model:read", "cluster:read", "enum:read"},
			required:        "project:create",
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPermission(tt.userPermissions, tt.required)
			if got != tt.want {
				t.Errorf("CheckPermission(%v, %q) = %v, want %v", tt.userPermissions, tt.required, got, tt.want)
			}
		})
	}
}
