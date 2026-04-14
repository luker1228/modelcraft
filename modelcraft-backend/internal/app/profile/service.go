package profile

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"strings"

	domainProfile "modelcraft/internal/domain/profile"

	domainUser "modelcraft/internal/domain/user"
)

const (
	defaultUserStatus = "REGISTERED"
	timeLayout        = "2006-01-02T15:04:05Z07:00"
)

// AppService handles profile related application use cases.
type AppService struct {
	userRepo    domainUser.UserRepository
	profileRepo domainProfile.Repository
}

// NewAppService creates a new profile app service.
func NewAppService(userRepo domainUser.UserRepository, profileRepo domainProfile.Repository) *AppService {
	return &AppService{
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

// GetMyUserProfile returns current user and profile in one aggregated payload.
func (s *AppService) GetMyUserProfile(ctx context.Context, cmd GetMyUserProfileCommand) (*UserProfileView, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "organization name is required")
	}
	if cmd.UserID == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "user id is required")
	}

	u, err := s.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.UserNotFound, cmd.UserID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if u == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.UserNotFound, cmd.UserID)
	}

	p, err := s.profileRepo.FindByUserID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProfileNotFound, cmd.UserID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if p == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProfileNotFound, cmd.UserID)
	}

	return &UserProfileView{
		User: UserView{
			ID:        u.ID,
			Phone:     u.Phone.String(),
			UserName:  u.Name,
			Status:    defaultUserStatus,
			CreatedAt: u.CreatedAt.Format(timeLayout),
			UpdatedAt: u.UpdatedAt.Format(timeLayout),
		},
		Profile: ProfileView{
			ID:        p.ID,
			UserID:    p.UserID,
			Nickname:  p.Nickname,
			AvatarURL: p.AvatarURL,
			Bio:       p.Bio,
			CreatedAt: p.CreatedAt.Format(timeLayout),
			UpdatedAt: p.UpdatedAt.Format(timeLayout),
		},
	}, nil
}

// UpdateMyProfile updates current user's profile with PATCH semantics.
func (s *AppService) UpdateMyProfile(ctx context.Context, cmd UpdateMyProfileCommand) (*ProfileView, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "organization name is required")
	}
	if cmd.UserID == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "user id is required")
	}

	patch := domainProfile.UpdatePatch{}
	if cmd.Nickname != nil {
		nickname := strings.TrimSpace(*cmd.Nickname)
		patch.Nickname = &nickname
	}
	if cmd.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*cmd.AvatarURL)
		patch.AvatarURL = &avatarURL
	}
	if cmd.Bio != nil {
		bio := strings.TrimSpace(*cmd.Bio)
		patch.Bio = &bio
	}

	if patch.IsEmpty() {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.InvalidProfileInput,
			"at least one field must be provided",
		)
	}

	if err := s.profileRepo.UpdateByUserID(ctx, cmd.OrgName, cmd.UserID, patch); err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProfileNotFound, cmd.UserID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	updated, err := s.profileRepo.FindByUserID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProfileNotFound, cmd.UserID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if updated == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProfileNotFound, cmd.UserID)
	}

	return &ProfileView{
		ID:        updated.ID,
		UserID:    updated.UserID,
		Nickname:  updated.Nickname,
		AvatarURL: updated.AvatarURL,
		Bio:       updated.Bio,
		CreatedAt: updated.CreatedAt.Format(timeLayout),
		UpdatedAt: updated.UpdatedAt.Format(timeLayout),
	}, nil
}
