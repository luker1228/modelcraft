package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlEndUserSessionRepository is the MySQL implementation of enduser.EndUserSessionRepository.
//
// Note:
// - It operates in private_{projectSlug} database.
// - not found -> (nil, nil)
// - update by id checks RowsAffected.
type endUserSessionDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SqlEndUserSessionRepository struct {
	db endUserSessionDBTX
}

// NewSqlEndUserSessionRepository creates a SqlEndUserSessionRepository.
func NewSqlEndUserSessionRepository(db endUserSessionDBTX) enduser.EndUserSessionRepository {
	return &SqlEndUserSessionRepository{db: db}
}

// Save creates a new session record.
func (r *SqlEndUserSessionRepository) Save(ctx context.Context, session *enduser.EndUserSession) error {
	const query = `
		INSERT INTO accounts (id, user_id, refresh_token_hash, expires_at, revoked, created_at)
		VALUES (?, ?, ?, ?, 0, NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
	)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// GetByTokenHash retrieves a session by token hash.
// Returns (nil, nil) when not found.
func (r *SqlEndUserSessionRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*enduser.EndUserSession, error) {
	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked, created_at
		FROM accounts
		WHERE refresh_token_hash = ?
	`

	row := r.db.QueryRowContext(ctx, query, tokenHash)

	var (
		sessionID        string
		userID           string
		refreshTokenHash string
		expiresAt        time.Time
		revoked          int
		createdAt        time.Time
	)

	err := row.Scan(
		&sessionID,
		&userID,
		&refreshTokenHash,
		&expiresAt,
		&revoked,
		&createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //nolint:nilnil // per contract: not found is expected in repo layer
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	return &enduser.EndUserSession{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
		Revoked:          revoked == 1,
		CreatedAt:        createdAt,
	}, nil
}

// RevokeByID marks a session as revoked.
// Returns NO_ROWS_AFFECTED when session does not exist.
func (r *SqlEndUserSessionRepository) RevokeByID(ctx context.Context, id string) error {
	const query = `UPDATE accounts SET revoked = 1 WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end-user session not found: %s", id))
	}

	return nil
}

// RevokeAllByUserID marks all sessions for the user as revoked.
func (r *SqlEndUserSessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	const query = `UPDATE accounts SET revoked = 1 WHERE user_id = ?`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// Compile-time interface satisfaction check.
var _ enduser.EndUserSessionRepository = (*SqlEndUserSessionRepository)(nil)
