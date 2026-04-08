package repository

import (
	"context"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"

	domainauth "modelcraft/internal/domain/auth"
)

// SqlAPIKeyRepository is the sqlc-based implementation of auth.APIKeyRepository.
type SqlAPIKeyRepository struct {
	q dbgen.Querier
}

// NewSqlAPIKeyRepository creates a SqlAPIKeyRepository.
func NewSqlAPIKeyRepository(q dbgen.Querier) domainauth.APIKeyRepository {
	return &SqlAPIKeyRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Save persists a new API key.
func (r *SqlAPIKeyRepository) Save(ctx context.Context, key *domainauth.APIKey) error {
	return r.q.InsertAPIKey(ctx, dbgen.InsertAPIKeyParams{
		ID:        key.ID,
		UserID:    key.UserID,
		Name:      key.Name,
		KeyHash:   key.KeyHash,
		KeyPrefix: key.KeyPrefix,
		ExpiresAt: sqlerr.PtrToNullTime(key.ExpiresAt),
	})
}

// FindByHash finds an API key by hash.
// Returns (nil, nil) when not found.
func (r *SqlAPIKeyRepository) FindByHash(ctx context.Context, hash string) (*domainauth.APIKey, error) {
	row, err := r.q.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // Pattern B: not found is a valid state, caller checks bool/nil
		}
		return nil, err
	}

	return toDomainAPIKey(row), nil
}

// FindByID finds an API key by ID.
// Returns (nil, nil) when not found.
func (r *SqlAPIKeyRepository) FindByID(ctx context.Context, id string) (*domainauth.APIKey, error) {
	row, err := r.q.GetAPIKeyByID(ctx, id)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // Pattern B: not found is a valid state, caller checks bool/nil
		}
		return nil, err
	}

	return toDomainAPIKey(row), nil
}

// ListByUserID lists active API keys for a user.
func (r *SqlAPIKeyRepository) ListByUserID(ctx context.Context, userID string) ([]*domainauth.APIKey, error) {
	rows, err := r.q.ListAPIKeysByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*domainauth.APIKey, len(rows))
	for i, row := range rows {
		result[i] = toDomainAPIKey(row)
	}
	return result, nil
}

// CountActiveByUserID counts active API keys for a user.
func (r *SqlAPIKeyRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	count, err := r.q.CountActiveAPIKeysByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// Revoke revokes an API key scoped by id and user.
func (r *SqlAPIKeyRepository) Revoke(ctx context.Context, id, userID string) error {
	return r.q.RevokeAPIKey(ctx, dbgen.RevokeAPIKeyParams{ID: id, UserID: userID})
}

// Update updates key name and expiration.
func (r *SqlAPIKeyRepository) Update(
	ctx context.Context,
	id string,
	userID string,
	name string,
	expiresAt *time.Time,
) error {
	return r.q.UpdateAPIKey(ctx, dbgen.UpdateAPIKeyParams{
		Name:      name,
		ExpiresAt: sqlerr.PtrToNullTime(expiresAt),
		ID:        id,
		UserID:    userID,
	})
}

// UpdateLastUsed updates last_used_at timestamp.
func (r *SqlAPIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	return r.q.UpdateAPIKeyLastUsed(ctx, id)
}

// DeleteRevoked deletes stale revoked API keys.
func (r *SqlAPIKeyRepository) DeleteRevoked(ctx context.Context) error {
	return r.q.DeleteRevokedAPIKeys(ctx)
}

func toDomainAPIKey(row dbgen.ApiKey) *domainauth.APIKey {
	return &domainauth.APIKey{
		ID:         row.ID,
		UserID:     row.UserID,
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		KeyPrefix:  row.KeyPrefix,
		LastUsedAt: sqlerr.NullTimeToPtr(row.LastUsedAt),
		ExpiresAt:  sqlerr.NullTimeToPtr(row.ExpiresAt),
		CreatedAt:  row.CreatedAt,
		RevokedAt:  sqlerr.NullTimeToPtr(row.RevokedAt),
	}
}

var _ domainauth.APIKeyRepository = (*SqlAPIKeyRepository)(nil)
