package repository

import (
	"context"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlEndUserSessionRepository 将 end-user refresh token 存储在统一的 refresh_tokens 表。
// 统一用户体系后，end_user_accounts 表已废弃。
type SqlEndUserSessionRepository struct {
	q dbgen.Querier
}

// NewSqlEndUserSessionRepository creates a SqlEndUserSessionRepository.
// orgName and projectSlug are retained in the signature for call-site compatibility but are unused.
func NewSqlEndUserSessionRepository(db endUserDBTX, _, _ string) enduser.EndUserSessionRepository {
	return &SqlEndUserSessionRepository{q: dbgenwrap.NewSafeQuerier(dbgen.New(db))}
}

// Save inserts a new refresh token into the unified refresh_tokens table.
func (r *SqlEndUserSessionRepository) Save(ctx context.Context, session *enduser.EndUserSession) error {
	err := r.q.InsertRefreshToken(ctx, dbgen.InsertRefreshTokenParams{
		ID:        session.ID,
		UserID:    session.UserID,
		TokenHash: session.RefreshTokenHash,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// GetByTokenHash retrieves a session by its token hash.
// Returns (nil, nil) when not found.
func (r *SqlEndUserSessionRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*enduser.EndUserSession, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // per contract: not found is expected in repo layer
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	return &enduser.EndUserSession{
		ID:               row.ID,
		UserID:           row.UserID,
		RefreshTokenHash: row.TokenHash,
		ExpiresAt:        row.ExpiresAt,
		Revoked:          row.RevokedAt.Valid,
		CreatedAt:        row.CreatedAt,
	}, nil
}

// RevokeByID marks a session as revoked by setting revoked_at.
func (r *SqlEndUserSessionRepository) RevokeByID(ctx context.Context, id string) error {
	err := r.q.RevokeRefreshToken(ctx, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// RevokeAllByUserID revokes all active sessions for a user.
func (r *SqlEndUserSessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	err := r.q.RevokeAllRefreshTokensByUserID(ctx, userID)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// Compile-time interface check.
var _ enduser.EndUserSessionRepository = (*SqlEndUserSessionRepository)(nil)
