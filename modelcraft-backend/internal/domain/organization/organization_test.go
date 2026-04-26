package organization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrganization(t *testing.T) {
	t.Run("should create organization with valid input", func(t *testing.T) {
		org, err := NewOrganization("my-company", "My Company", "user-uuid-001")
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, "my-company", org.Name)
		assert.Equal(t, "My Company", org.DisplayName)
		assert.Equal(t, "user-uuid-001", org.OwnerID)
		assert.Equal(t, OrgStatusActive, org.Status)
		assert.False(t, org.CreatedAt.IsZero())
	})

	t.Run("should create organization with empty display name", func(t *testing.T) {
		org, err := NewOrganization("my-company", "", "user-uuid-001")
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, "", org.DisplayName)
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		org, err := NewOrganization("", "My Company", "user-uuid-001")
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "organization name is required")
	})

	t.Run("should return error when owner ID is empty", func(t *testing.T) {
		org, err := NewOrganization("my-company", "My Company", "")
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "organization owner ID is required")
	})

	t.Run("should return error when name format is invalid", func(t *testing.T) {
		org, err := NewOrganization("INVALID-NAME", "My Company", "user-uuid-001")
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "organization name must be 2-64 characters")
		assert.Contains(t, err.Error(), "lowercase letters/digits/underscores/hyphens only")
		assert.Contains(t, err.Error(), "start with a letter")
	})
}

func TestIsValidOrgName(t *testing.T) {
	validNames := []string{
		"ab",                   // 最短有效名称
		"my-company",           // 含连字符
		"my_company",           // 含下划线
		"acme123",              // 含数字
		"company-x9k2j7",       // AuthProvider 风格后缀
		"abcdefghijklmnopqrst", // 长名称
	}
	for _, name := range validNames {
		t.Run("valid: "+name, func(t *testing.T) {
			assert.True(t, isValidOrgName(name), "expected %q to be valid", name)
		})
	}

	invalidNames := []string{
		"a",          // 太短
		"A",          // 大写
		"1abc",       // 数字开头
		"-abc",       // 连字符开头
		"_abc",       // 下划线开头
		"my company", // 含空格
		"My-Company", // 含大写
		"abc.def",    // 含点号
	}
	for _, name := range invalidNames {
		t.Run("invalid: "+name, func(t *testing.T) {
			assert.False(t, isValidOrgName(name), "expected %q to be invalid", name)
		})
	}
}

func TestOrganization_StatusTransitions(t *testing.T) {
	t.Run("should suspend an active organization", func(t *testing.T) {
		org, _ := NewOrganization("my-org", "", "user-001")
		assert.True(t, org.IsActive())

		org.Suspend()
		assert.Equal(t, OrgStatusSuspended, org.Status)
		assert.False(t, org.IsActive())
	})

	t.Run("should activate a suspended organization", func(t *testing.T) {
		org, _ := NewOrganization("my-org", "", "user-001")
		org.Suspend()

		org.Activate()
		assert.Equal(t, OrgStatusActive, org.Status)
		assert.True(t, org.IsActive())
	})

	t.Run("should mark organization as deleted", func(t *testing.T) {
		org, _ := NewOrganization("my-org", "", "user-001")

		org.MarkDeleted()
		assert.Equal(t, OrgStatusDeleted, org.Status)
		assert.False(t, org.IsActive())
	})
}

func TestOrganization_UpdateDisplayName(t *testing.T) {
	t.Run("should update display name", func(t *testing.T) {
		org, _ := NewOrganization("my-org", "Old Name", "user-001")
		originalUpdatedAt := org.UpdatedAt

		org.UpdateDisplayName("New Name")
		assert.Equal(t, "New Name", org.DisplayName)
		assert.True(t, org.UpdatedAt.After(originalUpdatedAt) || org.UpdatedAt.Equal(originalUpdatedAt))
	})
}

// TestOrganization_Validate tests the Validate method directly
func TestOrganization_Validate(t *testing.T) {
	tests := []struct {
		name        string
		given       *Organization
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid organization with all fields",
			given: &Organization{
				Name:        "my-company",
				DisplayName: "My Company",
				OwnerID:     "user-uuid-001",
				Status:      OrgStatusActive,
			},
			expectError: false,
		},
		{
			name: "valid organization with suspended status",
			given: &Organization{
				Name:        "suspended-org",
				DisplayName: "Suspended Org",
				OwnerID:     "user-uuid-002",
				Status:      OrgStatusSuspended,
			},
			expectError: false,
		},
		{
			name: "valid organization with deleted status",
			given: &Organization{
				Name:        "deleted-org",
				DisplayName: "Deleted Org",
				OwnerID:     "user-uuid-003",
				Status:      OrgStatusDeleted,
			},
			expectError: false,
		},
		{
			name: "invalid organization with unrecognized status",
			given: &Organization{
				Name:        "test-org",
				DisplayName: "Test Org",
				OwnerID:     "user-uuid-004",
				Status:      "invalid_status",
			},
			expectError: true,
			errorMsg:    "organization status must be one of: active, suspended, deleted",
		},
		{
			name: "invalid organization with empty status",
			given: &Organization{
				Name:        "test-org",
				DisplayName: "Test Org",
				OwnerID:     "user-uuid-005",
				Status:      "",
			},
			expectError: true,
			errorMsg:    "organization status must be one of: active, suspended, deleted",
		},
		{
			name: "invalid organization name format - uppercase",
			given: &Organization{
				Name:        "UPPERCASE-NAME",
				DisplayName: "Test Org",
				OwnerID:     "user-uuid-006",
				Status:      OrgStatusActive,
			},
			expectError: true,
			errorMsg:    "organization name must be 2-64 characters",
		},
		{
			name: "invalid organization name format - starts with digit",
			given: &Organization{
				Name:        "1invalid",
				DisplayName: "Test Org",
				OwnerID:     "user-uuid-007",
				Status:      OrgStatusActive,
			},
			expectError: true,
			errorMsg:    "organization name must be 2-64 characters",
		},
		{
			name: "invalid organization name format - too short",
			given: &Organization{
				Name:        "x",
				DisplayName: "Test Org",
				OwnerID:     "user-uuid-008",
				Status:      OrgStatusActive,
			},
			expectError: true,
			errorMsg:    "organization name must be 2-64 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: calling Validate
			err := tt.given.Validate()

			// Then: verify the result
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
