package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	validPhone, err := NewPhoneNumber("13800138000")
	require.NoError(t, err)

	t.Run("should create user with valid phone and password hash", func(t *testing.T) {
		user, err := NewUser("uuid-123", validPhone, "$2a$10$hashedpassword")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "uuid-123", user.ID)
		assert.Equal(t, "138****8000", user.Name) // auto-masked
		assert.Equal(t, "13800138000", user.Phone.String())
		assert.Equal(t, "$2a$10$hashedpassword", user.PasswordHash)
		assert.Empty(t, user.ExternalID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("should return error when ID is empty", func(t *testing.T) {
		user, err := NewUser("", validPhone, "$2a$10$hashedpassword")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("should return error when phone is zero", func(t *testing.T) {
		user, err := NewUser("uuid-123", PhoneNumber{}, "$2a$10$hashedpassword")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "phone number is required")
	})

	t.Run("should return error when password hash is empty", func(t *testing.T) {
		user, err := NewUser("uuid-123", validPhone, "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "password hash is required")
	})
}

func TestNewOAuthUser(t *testing.T) {
	t.Run("should create OAuth user with valid input", func(t *testing.T) {
		user, err := NewOAuthUser("uuid-123", "casdoor-user-001", "Luke", "13800138000")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "uuid-123", user.ID)
		assert.Equal(t, "casdoor-user-001", user.ExternalID)
		assert.Equal(t, "Luke", user.Name)
		assert.Equal(t, "13800138000", user.Phone.String())
		assert.Empty(t, user.PasswordHash)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("should create OAuth user without phone", func(t *testing.T) {
		user, err := NewOAuthUser("uuid-123", "ext-001", "Name", "")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.True(t, user.Phone.IsZero())
	})

	t.Run("should create OAuth user with invalid phone gracefully", func(t *testing.T) {
		// OAuth user may have non-standard phone format; should not fail
		user, err := NewOAuthUser("uuid-123", "ext-001", "Name", "not-a-phone")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.True(t, user.Phone.IsZero())
	})

	t.Run("should return error when ID is empty", func(t *testing.T) {
		user, err := NewOAuthUser("", "casdoor-user-001", "", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("should return error when external ID is empty", func(t *testing.T) {
		user, err := NewOAuthUser("uuid-123", "", "Name", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "external ID is required")
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("should pass validation with valid fields", func(t *testing.T) {
		user := &User{ID: "uuid-123"}
		assert.NoError(t, user.Validate())
	})

	t.Run("should fail validation with empty ID", func(t *testing.T) {
		user := &User{ID: ""}
		assert.Error(t, user.Validate())
	})
}
