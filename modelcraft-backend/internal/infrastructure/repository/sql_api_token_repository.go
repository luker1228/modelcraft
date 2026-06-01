package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modelcraft/internal/domain/enduser"
	"time"
)

// compile-time interface check
var _ enduser.APITokenRepository = (*SqlAPITokenRepository)(nil)

type SqlAPITokenRepository struct {
	db *sql.DB
}

func NewSqlAPITokenRepository(db *sql.DB) *SqlAPITokenRepository {
	return &SqlAPITokenRepository{db: db}
}

func (r *SqlAPITokenRepository) Save(ctx context.Context, token *enduser.APIToken) error {
	query := `
        INSERT INTO end_user_api_tokens
          (id, org_name, end_user_id, name, token_hash, expires_at, created_at, deleted_at, delete_token)
        VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0)`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.OrgName, token.EndUserID, token.Name,
		token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save api token: %w", err)
	}
	return nil
}

func (r *SqlAPITokenRepository) FindByHash(ctx context.Context, hash string) (*enduser.APIToken, error) {
	query := `
        SELECT id, org_name, end_user_id, name, token_hash,
               expires_at, last_used_at, created_at, deleted_at, delete_token
        FROM end_user_api_tokens
        WHERE token_hash = ? AND deleted_at = 0
        LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, hash)
	return scanAPIToken(row)
}

func (r *SqlAPITokenRepository) ListByUser(
	ctx context.Context, orgName, endUserID string,
) ([]*enduser.APIToken, error) {
	query := `
        SELECT id, org_name, end_user_id, name, token_hash,
               expires_at, last_used_at, created_at, deleted_at, delete_token
        FROM end_user_api_tokens
        WHERE org_name = ? AND end_user_id = ? AND deleted_at = 0
        ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgName, endUserID)
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*enduser.APIToken
	for rows.Next() {
		token, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *SqlAPITokenRepository) SoftDelete(ctx context.Context, id, orgName, endUserID string) error {
	now := time.Now().UnixMilli()
	query := `
        UPDATE end_user_api_tokens
        SET deleted_at = ?, delete_token = ?
        WHERE id = ? AND org_name = ? AND end_user_id = ? AND deleted_at = 0`
	result, err := r.db.ExecContext(ctx, query, now, now, id, orgName, endUserID)
	if err != nil {
		return fmt.Errorf("soft delete api token: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("api token not found or already deleted")
	}
	return nil
}

func (r *SqlAPITokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE end_user_api_tokens SET last_used_at = ? WHERE id = ?`, at, id)
	return err
}

type apiTokenScanner interface {
	Scan(dest ...any) error
}

func scanAPIToken(s apiTokenScanner) (*enduser.APIToken, error) {
	var t enduser.APIToken
	var expiresAt, lastUsedAt sql.NullTime
	err := s.Scan(
		&t.ID, &t.OrgName, &t.EndUserID, &t.Name, &t.TokenHash,
		&expiresAt, &lastUsedAt, &t.CreatedAt, &t.DeletedAt, &t.DeleteToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //nolint:nilnil // not found is a valid state
		}
		return nil, fmt.Errorf("scan api token: %w", err)
	}
	if expiresAt.Valid {
		t.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		t.LastUsedAt = &lastUsedAt.Time
	}
	return &t, nil
}
