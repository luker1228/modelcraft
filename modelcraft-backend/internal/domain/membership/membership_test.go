package membership_test

import (
	"modelcraft/internal/domain/membership"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMembership(t *testing.T) {
	m, err := membership.NewMembership("mid-001", "user-001", "my-org", false)
	require.NoError(t, err)
	assert.Equal(t, "mid-001", m.ID)
	assert.Equal(t, "user-001", m.UserID)
	assert.Equal(t, "my-org", m.OrgName)
	assert.False(t, m.IsAdmin)
	assert.Equal(t, membership.MembershipStatusActive, m.Status)
	assert.NotZero(t, m.CreatedAt)
}

func TestNewMembershipAdmin(t *testing.T) {
	m, err := membership.NewMembership("mid-002", "user-002", "my-org", true)
	require.NoError(t, err)
	assert.True(t, m.IsAdmin)
	assert.Equal(t, membership.MembershipStatusActive, m.Status)
}

func TestMembershipValidate_RequiredFields(t *testing.T) {
	_, err := membership.NewMembership("", "user-001", "my-org", false)
	assert.Error(t, err)

	_, err = membership.NewMembership("mid-001", "", "my-org", false)
	assert.Error(t, err)

	_, err = membership.NewMembership("mid-001", "user-001", "", false)
	assert.Error(t, err)
}

func TestMembershipSuspend(t *testing.T) {
	m, err := membership.NewMembership("mid-001", "user-001", "my-org", false)
	require.NoError(t, err)
	m.Suspend()
	assert.Equal(t, membership.MembershipStatusSuspended, m.Status)
	assert.False(t, m.IsActive())
}
