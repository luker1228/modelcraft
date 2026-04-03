package membership

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMembership(t *testing.T) {
	t.Run("should create active membership", func(t *testing.T) {
		m, err := NewMembership("m-001", "user-001", "org-001")
		assert.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "m-001", m.ID)
		assert.Equal(t, "user-001", m.UserID)
		assert.Equal(t, "org-001", m.OrgName)
		assert.Equal(t, MembershipStatusActive, m.Status)
		assert.NotNil(t, m.JoinedAt)
		assert.Nil(t, m.InvitedAt)
		assert.Equal(t, "", m.InvitedBy)
		assert.True(t, m.IsActive())
	})

	t.Run("should return error when ID is empty", func(t *testing.T) {
		m, err := NewMembership("", "user-001", "org-001")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "membership ID is required")
	})

	t.Run("should return error when user ID is empty", func(t *testing.T) {
		m, err := NewMembership("m-001", "", "org-001")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("should return error when org name is empty", func(t *testing.T) {
		m, err := NewMembership("m-001", "user-001", "")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "organization name is required")
	})
}

func TestNewInvitation(t *testing.T) {
	t.Run("should create invited membership", func(t *testing.T) {
		m, err := NewInvitation("m-001", "user-002", "org-001", "user-001")
		assert.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, MembershipStatusInvited, m.Status)
		assert.Equal(t, "user-001", m.InvitedBy)
		assert.NotNil(t, m.InvitedAt)
		assert.Nil(t, m.JoinedAt)
		assert.False(t, m.IsActive())
	})

	t.Run("should return error when org name is empty", func(t *testing.T) {
		m, err := NewInvitation("m-001", "user-002", "", "user-001")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "organization name is required")
	})
}

func TestMembership_AcceptInvitation(t *testing.T) {
	t.Run("should accept invitation and become active", func(t *testing.T) {
		m, err := NewInvitation("m-001", "user-002", "org-001", "user-001")
		assert.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, MembershipStatusInvited, m.Status)

		err = m.AcceptInvitation()
		assert.NoError(t, err)
		assert.Equal(t, MembershipStatusActive, m.Status)
		assert.NotNil(t, m.JoinedAt)
		assert.True(t, m.IsActive())
	})

	t.Run("should return error when accepting non-invited membership", func(t *testing.T) {
		m, err := NewMembership("m-001", "user-001", "org-001")
		assert.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, MembershipStatusActive, m.Status)

		err = m.AcceptInvitation()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only accept invitation when status is 'invited'")
	})
}

func TestMembership_Suspend(t *testing.T) {
	t.Run("should suspend an active membership", func(t *testing.T) {
		m, err := NewMembership("m-001", "user-001", "org-001")
		assert.NoError(t, err)
		assert.NotNil(t, m)
		assert.True(t, m.IsActive())

		m.Suspend()
		assert.Equal(t, MembershipStatusSuspended, m.Status)
		assert.False(t, m.IsActive())
	})
}

func TestMembership_Validate(t *testing.T) {
	t.Run("should return error when status is invalid", func(t *testing.T) {
		// Given: Create a membership with valid data
		m := &Membership{
			ID:      "m-001",
			UserID:  "user-001",
			OrgName: "org-001",
			Status:  "invalid-status", // When: Set an invalid status
		}

		// Then: Validation should fail
		err := m.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "membership status must be one of: active, suspended, invited")
	})
}
