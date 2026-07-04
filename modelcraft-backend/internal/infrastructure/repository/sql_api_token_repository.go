package repository

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"
)

// compile-time interface check
var _ enduser.APITokenRepository = (*SqlAPITokenRepository)(nil)

type SqlAPITokenRepository struct {
	q dbgen.Querier
}

func NewSqlAPITokenRepository(q dbgen.Querier) *SqlAPITokenRepository {
	return &SqlAPITokenRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

func (r *SqlAPITokenRepository) Save(ctx context.Context, token *enduser.APIToken) error {
	return r.q.InsertAPIToken(ctx, dbgen.InsertAPITokenParams{
		ID:        token.ID,
		OrgName:   token.OrgName,
		EndUserID: token.EndUserID,
		Name:      token.Name,
		TokenHash: token.TokenHash,
		ExpiresAt: ptrToNullTime(token.ExpiresAt),
		CreatedAt: token.CreatedAt,
	})
}

func (r *SqlAPITokenRepository) FindByHash(ctx context.Context, hash string) (*enduser.APIToken, error) {
	row, err := r.q.GetAPITokenByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find token by hash: %w", err)
	}
	return toDomainToken(&row), nil
}

func (r *SqlAPITokenRepository) ListByUser(
	ctx context.Context, orgName, endUserID string,
) ([]*enduser.APIToken, error) {
	rows, err := r.q.ListAPITokensByUser(ctx, dbgen.ListAPITokensByUserParams{
		OrgName:   orgName,
		EndUserID: endUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}

	tokens := make([]*enduser.APIToken, 0, len(rows))
	for i := range rows {
		tokens = append(tokens, toDomainToken(&rows[i]))
	}
	return tokens, nil
}

func (r *SqlAPITokenRepository) SoftDelete(ctx context.Context, id, orgName, endUserID string) error {
	now := uint64(time.Now().UnixMilli())
	affected, err := r.q.SoftDeleteAPIToken(ctx, dbgen.SoftDeleteAPITokenParams{
		DeletedAt:   now,
		DeleteToken: now,
		ID:          id,
		OrgName:     orgName,
		EndUserID:   endUserID,
	})
	if err != nil {
		return fmt.Errorf("soft delete api token: %w", err)
	}
	if affected == 0 {
		return shared.NewNotFoundError("api token not found or already deleted")
	}
	return nil
}

func (r *SqlAPITokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	return r.q.UpdateAPITokenLastUsed(ctx, dbgen.UpdateAPITokenLastUsedParams{
		ID:         id,
		LastUsedAt: sql.NullTime{Time: at, Valid: true},
	})
}

// toDomainToken converts a generated dbgen.UserApiToken to a domain enduser.APIToken.
func toDomainToken(row *dbgen.UserApiToken) *enduser.APIToken {
	return &enduser.APIToken{
		ID:          row.ID,
		OrgName:     row.OrgName,
		EndUserID:   row.EndUserID,
		Name:        row.Name,
		TokenHash:   row.TokenHash,
		ExpiresAt:   sqlerr.NullTimeToPtr(row.ExpiresAt),
		LastUsedAt:  sqlerr.NullTimeToPtr(row.LastUsedAt),
		CreatedAt:   row.CreatedAt,
		DeletedAt:   int64(row.DeletedAt),
		DeleteToken: int64(row.DeleteToken),
	}
}

func ptrToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
