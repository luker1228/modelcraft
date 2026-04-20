// Package rls provides RLS (Row Level Security) repository implementations.
package rls

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/rls"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlModelRLSPolicyRepository is the sqlc-based implementation of ModelRLSPolicyRepository.
type SqlModelRLSPolicyRepository struct {
	q dbgen.Querier
}

// NewSqlModelRLSPolicyRepository creates a new SqlModelRLSPolicyRepository.
func NewSqlModelRLSPolicyRepository(q dbgen.Querier) modeldesign.ModelRLSPolicyRepository {
	return &SqlModelRLSPolicyRepository{q: q}
}

// GetByModelID retrieves an RLS policy by model ID.
// Returns nil, nil if the policy does not exist.
func (r *SqlModelRLSPolicyRepository) GetByModelID(
	ctx context.Context,
	orgName, projectSlug, modelID string,
) (*modeldesign.ModelRLSPolicy, error) {
	// Note: Currently model_rls_policies table uses model_id as primary key lookup
	// Multi-tenant isolation is enforced via model_id uniqueness across the system
	// TODO: Add org_name and project_slug filtering when table schema is updated
	row, err := r.q.GetModelRLSPolicy(ctx, modelID)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // nil result with nil error indicates not found
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	return &modeldesign.ModelRLSPolicy{
		ModelID:         row.ModelID,
		SelectPredicate: rls.JsonExpr(row.SelectPredicate),
		InsertCheck:     rls.JsonExpr(row.InsertCheck),
		UpdatePredicate: rls.JsonExpr(row.UpdatePredicate),
		UpdateCheck:     rls.JsonExpr(row.UpdateCheck),
		DeletePredicate: rls.JsonExpr(row.DeletePredicate),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

// Save saves an RLS policy (upsert operation).
func (r *SqlModelRLSPolicyRepository) Save(
	ctx context.Context,
	orgName, projectSlug string,
	policy *modeldesign.ModelRLSPolicy,
) error {
	err := r.q.UpsertModelRLSPolicy(ctx, dbgen.UpsertModelRLSPolicyParams{
		ModelID:         policy.ModelID,
		SelectPredicate: string(policy.SelectPredicate),
		InsertCheck:     string(policy.InsertCheck),
		UpdatePredicate: string(policy.UpdatePredicate),
		UpdateCheck:     string(policy.UpdateCheck),
		DeletePredicate: string(policy.DeletePredicate),
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// DeleteByModelID deletes the RLS policy for a given model ID.
func (r *SqlModelRLSPolicyRepository) DeleteByModelID(ctx context.Context, orgName, projectSlug, modelID string) error {
	err := r.q.DeleteModelRLSPolicy(ctx, modelID)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// ExistsByModelID checks if an RLS policy exists for the given model ID.
func (r *SqlModelRLSPolicyRepository) ExistsByModelID(
	ctx context.Context,
	orgName, projectSlug, modelID string,
) (bool, error) {
	exists, err := r.q.ExistsModelRLSPolicy(ctx, modelID)
	if err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return exists, nil
}
