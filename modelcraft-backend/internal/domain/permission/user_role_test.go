package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewUserRole tests the NewUserRole constructor
func TestNewUserRole(t *testing.T) {
	t.Run("should create user-role binding with valid input", func(t *testing.T) {
		// Arrange & Act
		userRole := NewUserRole("user123", 10, "org1")

		// Assert
		assert.NotNil(t, userRole)
		assert.Equal(t, "user123", userRole.UserID)
		assert.Equal(t, 10, userRole.RoleID)
		assert.Equal(t, "org1", userRole.OrgName)
		assert.NotZero(t, userRole.CreatedAt)
	})
}

// TestUserRole_Validate tests the Validate method
func TestUserRole_Validate(t *testing.T) {
	t.Run("valid user-role binding", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		err := userRole.Validate()
		assert.NoError(t, err)
	})

	t.Run("error: empty user_id", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "",
			RoleID:  10,
			OrgName: "org1",
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id cannot be empty")
	})

	t.Run("error: user_id too long", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  string(make([]byte, 37)),
			RoleID:  10,
			OrgName: "org1",
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id cannot exceed 36 characters")
	})

	t.Run("error: role_id is zero", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "user123",
			RoleID:  0,
			OrgName: "org1",
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role_id must be a positive integer")
	})

	t.Run("error: role_id is negative", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "user123",
			RoleID:  -1,
			OrgName: "org1",
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role_id must be a positive integer")
	})

	t.Run("error: empty org_name", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "",
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org_name cannot be empty")
	})

	t.Run("error: org_name too long", func(t *testing.T) {
		userRole := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: string(make([]byte, 256)),
		}
		err := userRole.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org_name cannot exceed 255 characters")
	})
}

// TestUserRole_IsSameBinding tests the IsSameBinding method
func TestUserRole_IsSameBinding(t *testing.T) {
	t.Run("same binding", func(t *testing.T) {
		ur1 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		ur2 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		result := ur1.IsSameBinding(ur2)
		assert.True(t, result)
	})

	t.Run("different user_id", func(t *testing.T) {
		ur1 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		ur2 := &UserRole{
			UserID:  "user456",
			RoleID:  10,
			OrgName: "org1",
		}
		result := ur1.IsSameBinding(ur2)
		assert.False(t, result)
	})

	t.Run("different role_id", func(t *testing.T) {
		ur1 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		ur2 := &UserRole{
			UserID:  "user123",
			RoleID:  20,
			OrgName: "org1",
		}
		result := ur1.IsSameBinding(ur2)
		assert.False(t, result)
	})

	t.Run("different org_name", func(t *testing.T) {
		ur1 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		ur2 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org2",
		}
		result := ur1.IsSameBinding(ur2)
		assert.False(t, result)
	})

	t.Run("nil other user-role", func(t *testing.T) {
		ur1 := &UserRole{
			UserID:  "user123",
			RoleID:  10,
			OrgName: "org1",
		}
		result := ur1.IsSameBinding(nil)
		assert.False(t, result)
	})
}
