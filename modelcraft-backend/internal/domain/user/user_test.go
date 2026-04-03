package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("should create user with valid input", func(t *testing.T) {
		user, err := NewUser("uuid-123", "casdoor-user-001", "Luke", "13800138000")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "uuid-123", user.ID)
		assert.Equal(t, "casdoor-user-001", user.ExternalID)
		assert.Equal(t, "Luke", user.Name)
		assert.Equal(t, "13800138000", user.Phone)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("should return error when ID is empty", func(t *testing.T) {
		user, err := NewUser("", "casdoor-user-001", "", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("should return error when external ID is empty", func(t *testing.T) {
		user, err := NewUser("uuid-123", "", "", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "external ID is required")
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("should pass validation with valid fields", func(t *testing.T) {
		user := &User{ID: "uuid-123", ExternalID: "ext-001"}
		assert.NoError(t, user.Validate())
	})

	t.Run("should fail validation with empty ID", func(t *testing.T) {
		user := &User{ID: "", ExternalID: "ext-001"}
		assert.Error(t, user.Validate())
	})

	t.Run("should fail validation with empty external ID", func(t *testing.T) {
		user := &User{ID: "uuid-123", ExternalID: ""}
		assert.Error(t, user.Validate())
	})
}
