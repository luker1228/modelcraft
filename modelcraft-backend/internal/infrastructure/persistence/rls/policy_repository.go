package rls

import (
	"context"
	"database/sql"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"

	domainRLS "modelcraft/internal/domain/rls"
)

// SqlPolicyRepository implements domain/rls.PolicyRepositoryV2 and app/rls.PolicyRepository.
type SqlPolicyRepository struct {
	q dbgen.Querier
}

// NewSqlPolicyRepository creates a new SqlPolicyRepository.
func NewSqlPolicyRepository(q dbgen.Querier) *SqlPolicyRepository {
	return &SqlPolicyRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// ListByModel returns all policies for a model.
func (r *SqlPolicyRepository) ListByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) ([]*domainRLS.Policy, error) {
	rows, err := r.q.ListPoliciesByModel(ctx, dbgen.ListPoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	return todomainPolicies(rows), nil
}

// ListByAction returns policies matching a specific action and roles.
func (r *SqlPolicyRepository) ListByAction(
	ctx context.Context, orgName, projectSlug, modelID string,
	action domainRLS.Action, roles []string,
) ([]*domainRLS.Policy, error) {
	rows, err := r.q.ListPoliciesByAction(ctx, dbgen.ListPoliciesByActionParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
		Action:      dbgen.ModelRlsPoliciesAction(action),
		Roles:       roles,
	})
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	return todomainPolicies(rows), nil
}

// Upsert creates or updates a policy (keyed by policy_name within a model).
func (r *SqlPolicyRepository) Upsert(
	ctx context.Context, orgName, projectSlug string, policy *domainRLS.Policy,
) error {
	err := r.q.UpsertPolicy(ctx, dbgen.UpsertPolicyParams{
		OrgName:       orgName,
		ProjectSlug:   projectSlug,
		ModelID:       policy.ModelID,
		PolicyName:    policy.PolicyName,
		Action:        dbgen.ModelRlsPoliciesAction(policy.Action),
		Role:          policy.Role,
		UsingExpr:     sql.NullString{String: string(policy.UsingExpr), Valid: policy.UsingExpr != ""},
		WithCheckExpr: sql.NullString{String: string(policy.WithCheckExpr), Valid: policy.WithCheckExpr != ""},
	})
	return sqlerr.WrapSQLError(err)
}

// Delete deletes a single policy by ID.
func (r *SqlPolicyRepository) Delete(
	ctx context.Context, orgName, projectSlug string, id int64,
) error {
	err := r.q.DeletePolicy(ctx, dbgen.DeletePolicyParams{
		ID:          uint64(id),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	return sqlerr.WrapSQLError(err)
}

// DeleteByModel deletes all policies for a model.
func (r *SqlPolicyRepository) DeleteByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) error {
	err := r.q.DeletePoliciesByModel(ctx, dbgen.DeletePoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	return sqlerr.WrapSQLError(err)
}

func todomainPolicies(rows []dbgen.ModelRlsPolicy) []*domainRLS.Policy {
	result := make([]*domainRLS.Policy, 0, len(rows))
	for _, r := range rows {
		result = append(result, &domainRLS.Policy{
			ID:            int64(r.ID),
			OrgName:       r.OrgName,
			ProjectSlug:   r.ProjectSlug,
			ModelID:       r.ModelID,
			PolicyName:    r.PolicyName,
			Action:        domainRLS.Action(r.Action),
			Role:          r.Role,
			UsingExpr:     domainRLS.JsonExpr(r.UsingExpr.String),
			WithCheckExpr: domainRLS.JsonExpr(r.WithCheckExpr.String),
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		})
	}
	return result
}

// compile-time interface checks
var _ domainRLS.PolicyRepositoryV2 = (*SqlPolicyRepository)(nil)
