package profile

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"testing"
	"time"

	domainProfile "modelcraft/internal/domain/profile"

	domainUser "modelcraft/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	user *domainUser.User
	err  error
}

func (m *mockUserRepository) Create(_ context.Context, _ *domainUser.User) error { return nil }
func (m *mockUserRepository) GetByID(_ context.Context, _ string) (*domainUser.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetByExternalID(_ context.Context, _ string) (*domainUser.User, error) {
	return nil, nil
}

func (m *mockUserRepository) ExistsByExternalID(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockUserRepository) FindIDByExternalID(_ context.Context, _ string) (string, bool, error) {
	return "", false, nil
}

func (m *mockUserRepository) GetByPhone(_ context.Context, _, _ string) (*domainUser.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByName(_ context.Context, _, _ string) (*domainUser.User, error) {
	return nil, nil
}

func (m *mockUserRepository) ExistsByPhone(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *mockUserRepository) ExistsByName(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *mockUserRepository) ListByOrg(_ context.Context, _ string) ([]*domainUser.User, error) {
	return nil, nil
}

type mockProfileRepository struct {
	profile          *domainProfile.Profile
	err              error
	updateErr        error
	lastUpdatePatch  domainProfile.UpdatePatch
	updatedByUserID  string
	updatedByOrgName string
}

func (m *mockProfileRepository) Create(_ context.Context, _ *domainProfile.Profile) error {
	return nil
}

func (m *mockProfileRepository) CreateInitialProfile(_ context.Context, _ *domainProfile.Profile) error {
	return nil
}

func (m *mockProfileRepository) FindByUserID(_ context.Context, _, _ string) (*domainProfile.Profile, error) {
	return m.profile, m.err
}

func (m *mockProfileRepository) UpdateByUserID(
	_ context.Context,
	orgName, userID string,
	patch domainProfile.UpdatePatch,
) error {
	m.lastUpdatePatch = patch
	m.updatedByUserID = userID
	m.updatedByOrgName = orgName
	return m.updateErr
}

func TestAppServiceGetMyUserProfile(t *testing.T) {
	ctx := context.Background()

	phone, err := domainUser.NewPhoneNumber("13812345678")
	require.NoError(t, err)

	userEntity, err := domainUser.NewUser("user-1", "test_user", phone, "hashed", "test-org")
	require.NoError(t, err)

	avatar := "mock://avatar/default-1.png"
	bio := "hello"
	profileEntity := &domainProfile.Profile{
		ID:        "profile-1",
		UserID:    "user-1",
		Nickname:  "user_A1B2C3",
		AvatarURL: &avatar,
		Bio:       &bio,
		CreatedAt: time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
	}

	t.Run("should return aggregated user and profile", func(t *testing.T) {
		svc := NewAppService(
			&mockUserRepository{user: userEntity},
			&mockProfileRepository{profile: profileEntity},
		)

		result, getErr := svc.GetMyUserProfile(ctx, GetMyUserProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
		})
		require.NoError(t, getErr)
		require.NotNil(t, result)
		assert.Equal(t, "user-1", result.User.ID)
		assert.Equal(t, "test_user", result.User.UserName)
		assert.Equal(t, "REGISTERED", result.User.Status)
		assert.Equal(t, "user_A1B2C3", result.Profile.Nickname)
		assert.Equal(t, "2026-01-01T08:00:00Z", result.Profile.CreatedAt)
	})

	t.Run("should return UserNotFound when user does not exist", func(t *testing.T) {
		svc := NewAppService(
			&mockUserRepository{err: shared.NewNotFoundError("user not found")},
			&mockProfileRepository{profile: profileEntity},
		)

		result, getErr := svc.GetMyUserProfile(ctx, GetMyUserProfileCommand{
			OrgName: "org-1",
			UserID:  "missing-user",
		})
		require.Nil(t, result)
		require.Error(t, getErr)
		bizErr, ok := getErr.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.UserNotFound.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("should return ProfileNotFound when profile does not exist", func(t *testing.T) {
		svc := NewAppService(
			&mockUserRepository{user: userEntity},
			&mockProfileRepository{err: shared.NewNotFoundError("profile not found")},
		)

		result, getErr := svc.GetMyUserProfile(ctx, GetMyUserProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
		})
		require.Nil(t, result)
		require.Error(t, getErr)
		bizErr, ok := getErr.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.ProfileNotFound.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("should convert repository error to system error", func(t *testing.T) {
		svc := NewAppService(
			&mockUserRepository{user: userEntity},
			&mockProfileRepository{err: shared.NewRepositoryError(shared.ErrTypeConnection, "db down")},
		)

		result, getErr := svc.GetMyUserProfile(ctx, GetMyUserProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
		})
		require.Nil(t, result)
		require.Error(t, getErr)
		bizErr, ok := getErr.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.SystemError.GetCode(), bizErr.Info().GetCode())
	})
}

func TestAppServiceUpdateMyProfile(t *testing.T) {
	ctx := context.Background()

	avatar := "mock://avatar/default-1.png"
	bio := "hello"
	profileEntity := &domainProfile.Profile{
		ID:        "profile-1",
		UserID:    "user-1",
		Nickname:  "user_A1B2C3",
		AvatarURL: &avatar,
		Bio:       &bio,
		CreatedAt: time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
	}

	t.Run("should update profile with patch semantics", func(t *testing.T) {
		repo := &mockProfileRepository{profile: profileEntity}
		svc := NewAppService(&mockUserRepository{}, repo)

		newNickname := "  new_nick  "
		result, err := svc.UpdateMyProfile(ctx, UpdateMyProfileCommand{
			OrgName:  "org-1",
			UserID:   "user-1",
			Nickname: &newNickname,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "new_nick", *repo.lastUpdatePatch.Nickname)
		assert.Nil(t, repo.lastUpdatePatch.AvatarURL)
		assert.Nil(t, repo.lastUpdatePatch.Bio)
		assert.Equal(t, "org-1", repo.updatedByOrgName)
		assert.Equal(t, "user-1", repo.updatedByUserID)
	})

	t.Run("should return InvalidProfileInput when no field provided", func(t *testing.T) {
		svc := NewAppService(&mockUserRepository{}, &mockProfileRepository{profile: profileEntity})

		result, err := svc.UpdateMyProfile(ctx, UpdateMyProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
		})
		require.Nil(t, result)
		require.Error(t, err)
		bizErr, ok := err.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.InvalidProfileInput.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("should return ProfileNotFound when update target missing", func(t *testing.T) {
		repo := &mockProfileRepository{updateErr: shared.NewNotFoundError("profile not found")}
		svc := NewAppService(&mockUserRepository{}, repo)
		newBio := "new bio"

		result, err := svc.UpdateMyProfile(ctx, UpdateMyProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
			Bio:     &newBio,
		})
		require.Nil(t, result)
		require.Error(t, err)
		bizErr, ok := err.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.ProfileNotFound.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("should convert update repository error to system error", func(t *testing.T) {
		repo := &mockProfileRepository{
			updateErr: shared.NewRepositoryError(shared.ErrTypeConnection, "db down"),
		}
		svc := NewAppService(&mockUserRepository{}, repo)
		newBio := "new bio"

		result, err := svc.UpdateMyProfile(ctx, UpdateMyProfileCommand{
			OrgName: "org-1",
			UserID:  "user-1",
			Bio:     &newBio,
		})
		require.Nil(t, result)
		require.Error(t, err)
		bizErr, ok := err.(*bizerrors.BusinessError)
		require.True(t, ok)
		assert.Equal(t, bizerrors.SystemError.GetCode(), bizErr.Info().GetCode())
	})
}
