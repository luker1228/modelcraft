package repository

import (
	"context"
	domainauth "modelcraft/internal/domain/auth"
	"modelcraft/internal/infrastructure/dbgen"
)

// SqlRefreshTokenRepository is the sqlc-based implementation of auth.RefreshTokenRepository.
type SqlRefreshTokenRepository struct {
	q dbgen.Querier
}

// NewSqlRefreshTokenRepository creates a SqlRefreshTokenRepository.
func NewSqlRefreshTokenRepository(q dbgen.Querier) domainauth.RefreshTokenRepository {
	return &SqlRefreshTokenRepository{q: q}
}

// Save persists a new refresh token.
func (r *SqlRefreshTokenRepository) Save(ctx context.Context, token *domainauth.RefreshToken) error {
	return ExecWithErrorHandling(func() error {
		return r.q.InsertRefreshToken(ctx, dbgen.InsertRefreshTokenParams{
			ID:        token.ID,
			UserID:    token.UserID,
			TokenHash: token.TokenHash,
			ExpiresAt: token.ExpiresAt,
		})
	})
}

// FindByHash finds a refresh token by token hash.
// Returns (nil, nil) when not found.
func (r *SqlRefreshTokenRepository) FindByHash(
	ctx context.Context,
	hash string,
) (*domainauth.RefreshToken, error) {
	var row dbgen.RefreshToken
	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetRefreshTokenByHash(ctx, hash)
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // Pattern B: not found is a valid state, caller checks bool/nil
		}
		return nil, err
	}

	return toDomainRefreshToken(row), nil
}

// Revoke revokes a refresh token by ID.
func (r *SqlRefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	return ExecWithErrorHandling(func() error {
		return r.q.RevokeRefreshToken(ctx, id)
	})
}

// RevokeAllByUserID revokes all active refresh tokens for a user.
func (r *SqlRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	return ExecWithErrorHandling(func() error {
		return r.q.RevokeAllRefreshTokensByUserID(ctx, userID)
	})
}

// DeleteExpired removes expired/stale refresh token records.
func (r *SqlRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return ExecWithErrorHandling(func() error {
		return r.q.DeleteExpiredRefreshTokens(ctx)
	})
}

func toDomainRefreshToken(row dbgen.RefreshToken) *domainauth.RefreshToken {
	return &domainauth.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
		RevokedAt: NullTimeToPtr(row.RevokedAt),
	}
}

var _ domainauth.RefreshTokenRepository = (*SqlRefreshTokenRepository)(nil)
