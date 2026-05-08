package repository

import (
	"context"
	"modelcraft/internal/domain/profile"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"modelcraft/pkg/bizerrors"
)

// SqlProfileRepository is the sqlc-based implementation of profile.Repository.
type SqlProfileRepository struct {
	q dbgen.Querier
}

// NewSqlProfileRepository creates a new SqlProfileRepository backed by the given sqlc Querier.
func NewSqlProfileRepository(q dbgen.Querier) profile.Repository {
	return &SqlProfileRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

func profileToDomain(row dbgen.GetProfileByUserIDRow) *profile.Profile {
	return &profile.Profile{
		ID:        row.ID,
		UserID:    row.UserID,
		Nickname:  row.Nickname,
		AvatarURL: sqlerr.NullStrToPtr(row.AvatarUrl),
		Bio:       sqlerr.NullStrToPtr(row.Bio),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// Create is kept for backward compatibility and delegates to CreateInitialProfile.
func (r *SqlProfileRepository) Create(ctx context.Context, p *profile.Profile) error {
	return r.CreateInitialProfile(ctx, p)
}

// CreateInitialProfile creates an initial profile for a user.
func (r *SqlProfileRepository) CreateInitialProfile(ctx context.Context, p *profile.Profile) error {
	params := dbgen.CreateInitialProfileParams{
		ID:        p.ID,
		UserID:    p.UserID,
		Nickname:  p.Nickname,
		AvatarUrl: sqlerr.PtrToNullStr(p.AvatarURL),
		Bio:       sqlerr.PtrToNullStr(p.Bio),
	}

	if err := r.q.CreateInitialProfile(ctx, params); err != nil {
		return bizerrors.Wrapf(err, "failed to create initial profile for user: %s", p.UserID)
	}
	return nil
}

// FindByUserID retrieves profile by user ID and org name.
func (r *SqlProfileRepository) FindByUserID(ctx context.Context, orgName, userID string) (*profile.Profile, error) {
	row, err := r.q.GetProfileByUserID(ctx, dbgen.GetProfileByUserIDParams{
		UserID:  userID,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("profile not found by user id: " + userID)
		}
		return nil, bizerrors.Wrapf(err, "failed to get profile by user id: %s", userID)
	}

	return profileToDomain(row), nil
}

// UpdateByUserID updates profile by user ID and org name with patch semantics.
func (r *SqlProfileRepository) UpdateByUserID(
	ctx context.Context,
	orgName, userID string,
	patch profile.UpdatePatch,
) error {
	existing, err := r.FindByUserID(ctx, orgName, userID)
	if err != nil {
		return err
	}

	nickname := existing.Nickname
	avatarURL := existing.AvatarURL
	bio := existing.Bio

	if patch.Nickname != nil {
		nickname = *patch.Nickname
	}
	if patch.AvatarURL != nil {
		avatarURL = patch.AvatarURL
	}
	if patch.Bio != nil {
		bio = patch.Bio
	}

	result, err := r.q.UpdateProfileByUserID(ctx, dbgen.UpdateProfileByUserIDParams{
		Nickname:  nickname,
		AvatarUrl: sqlerr.PtrToNullStr(avatarURL),
		Bio:       sqlerr.PtrToNullStr(bio),
		UserID:    userID,
		OrgName:   orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return shared.NewNotFoundError("profile not found by user id: " + userID)
		}
		return bizerrors.Wrapf(err, "failed to update profile by user id: %s", userID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return bizerrors.Wrapf(err, "failed to get rows affected for profile update by user id: %s", userID)
	}
	if rowsAffected == 0 {
		return shared.NewNotFoundError("profile not found by user id: " + userID)
	}

	return nil
}

var _ profile.Repository = (*SqlProfileRepository)(nil)
