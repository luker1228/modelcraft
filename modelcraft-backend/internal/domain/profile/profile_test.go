package profile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	t.Run("should create profile with valid input", func(t *testing.T) {
		avatarURL := "mock://avatar/default-1.png"
		bio := "hello"

		p, err := NewProfile("profile-1", "user-1", "user_A1B2C3", &avatarURL, &bio)
		require.NoError(t, err)
		assert.Equal(t, "profile-1", p.ID)
		assert.Equal(t, "user-1", p.UserID)
		assert.Equal(t, "user_A1B2C3", p.Nickname)
		assert.Equal(t, &avatarURL, p.AvatarURL)
		assert.Equal(t, &bio, p.Bio)
		assert.False(t, p.CreatedAt.IsZero())
		assert.False(t, p.UpdatedAt.IsZero())
	})

	t.Run("should fail when nickname is empty", func(t *testing.T) {
		p, err := NewProfile("profile-1", "user-1", "", nil, nil)
		require.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "profile nickname is required")
	})

	t.Run("should fail when nickname is too long", func(t *testing.T) {
		p, err := NewProfile("profile-1", "user-1", strings.Repeat("a", maxNicknameLength+1), nil, nil)
		require.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "profile nickname must be <=")
	})

	t.Run("should fail when userID is empty", func(t *testing.T) {
		p, err := NewProfile("profile-1", "", "nick", nil, nil)
		require.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "profile userID is required")
	})
}

func TestNewInitialProfile(t *testing.T) {
	t.Run("should generate default nickname when empty", func(t *testing.T) {
		p, err := NewInitialProfile("profile-1", "user-1", "", nil, nil)
		require.NoError(t, err)
		assert.Regexp(t, `^user_[A-Z0-9]{6}$`, p.Nickname)
	})

	t.Run("should keep provided nickname", func(t *testing.T) {
		p, err := NewInitialProfile("profile-1", "user-1", "my_nick", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "my_nick", p.Nickname)
	})
}

func TestUpdatePatchIsEmpty(t *testing.T) {
	t.Run("should return true when all fields nil", func(t *testing.T) {
		patch := UpdatePatch{}
		assert.True(t, patch.IsEmpty())
	})

	t.Run("should return false when nickname is set", func(t *testing.T) {
		nickname := "new_nick"
		patch := UpdatePatch{Nickname: &nickname}
		assert.False(t, patch.IsEmpty())
	})
}

func TestProfileApplyPatch(t *testing.T) {
	avatarURL := "mock://avatar/default-1.png"
	bio := "hello"
	p, err := NewProfile("profile-1", "user-1", "nick", &avatarURL, &bio)
	require.NoError(t, err)

	t.Run("should apply partial update", func(t *testing.T) {
		newNick := "new_nick"
		newBio := "new bio"
		patch := UpdatePatch{
			Nickname: &newNick,
			Bio:      &newBio,
		}

		err := p.ApplyPatch(patch)
		require.NoError(t, err)
		assert.Equal(t, newNick, p.Nickname)
		assert.Equal(t, newBio, *p.Bio)
		assert.Equal(t, avatarURL, *p.AvatarURL)
	})

	t.Run("should fail on invalid patch value", func(t *testing.T) {
		invalidNickname := strings.Repeat("a", maxNicknameLength+1)
		patch := UpdatePatch{Nickname: &invalidNickname}

		err := p.ApplyPatch(patch)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "profile nickname must be <=")
	})
}
