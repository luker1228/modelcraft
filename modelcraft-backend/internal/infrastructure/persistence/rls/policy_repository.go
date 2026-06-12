package rls

import (
	"context"
	"encoding/json"

	"modelcraft/internal/domain/rls"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlPolicyRepository 策略多查询 Repository (sqlc-based)
type SqlPolicyRepository struct {
	q dbgen.Querier
}

// NewSqlPolicyRepository 创建 SqlPolicyRepository
func NewSqlPolicyRepository(q dbgen.Querier) *SqlPolicyRepository {
	return &SqlPolicyRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// ListByAction 查询匹配 action + role 的策略列表
func (r *SqlPolicyRepository) ListByAction(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, roles []string,
) ([]*rls.Policy, error) {
	// Ensure role="" is always included (default policy)
	rolesWithDefault := make([]string, len(roles)+1)
	copy(rolesWithDefault, roles)
	rolesWithDefault[len(roles)] = ""

	rows, err := r.q.ListPoliciesByAction(ctx, dbgen.ListPoliciesByActionParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
		Action:      dbgen.ModelRlsPoliciesAction(action),
		Roles:       rolesWithDefault,
	})
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	policies := make([]*rls.Policy, 0, len(rows))
	for _, row := range rows {
		p := &rls.Policy{
			ID:          int64(row.ID),
			OrgName:     row.OrgName,
			ProjectSlug: row.ProjectSlug,
			ModelID:     row.ModelID,
			PolicyName:  row.PolicyName,
			Action:      rls.Action(row.Action),
			Role:        row.Role,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}

		// Convert *json.RawMessage → JsonExpr
		if row.UsingExpr != nil {
			p.UsingExpr = rls.JsonExpr(*row.UsingExpr)
		}
		if row.WithCheckExpr != nil {
			p.WithCheckExpr = rls.JsonExpr(*row.WithCheckExpr)
		}

		policies = append(policies, p)
	}

	return policies, nil
}

// ListByModel 查询模型的所有策略
func (r *SqlPolicyRepository) ListByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) ([]*rls.Policy, error) {
	rows, err := r.q.ListPoliciesByModel(ctx, dbgen.ListPoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	policies := make([]*rls.Policy, 0, len(rows))
	for _, row := range rows {
		p := &rls.Policy{
			ID:          int64(row.ID),
			OrgName:     row.OrgName,
			ProjectSlug: row.ProjectSlug,
			ModelID:     row.ModelID,
			PolicyName:  row.PolicyName,
			Action:      rls.Action(row.Action),
			Role:        row.Role,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}

		// Convert *json.RawMessage -> JsonExpr
		if row.UsingExpr != nil {
			p.UsingExpr = rls.JsonExpr(*row.UsingExpr)
		}
		if row.WithCheckExpr != nil {
			p.WithCheckExpr = rls.JsonExpr(*row.WithCheckExpr)
		}

		policies = append(policies, p)
	}

	return policies, nil
}

// Upsert 创建或更新策略
func (r *SqlPolicyRepository) Upsert(ctx context.Context, orgName, projectSlug string, policy *rls.Policy) error {
	// Convert JsonExpr -> *json.RawMessage
	var usingExpr *json.RawMessage
	if policy.UsingExpr != "" {
		raw := json.RawMessage(policy.UsingExpr)
		usingExpr = &raw
	}
	var withCheckExpr *json.RawMessage
	if policy.WithCheckExpr != "" {
		raw := json.RawMessage(policy.WithCheckExpr)
		withCheckExpr = &raw
	}

	err := r.q.UpsertPolicy(ctx, dbgen.UpsertPolicyParams{
		OrgName:       orgName,
		ProjectSlug:   projectSlug,
		ModelID:       policy.ModelID,
		PolicyName:    policy.PolicyName,
		Action:        dbgen.ModelRlsPoliciesAction(policy.Action),
		Role:          policy.Role,
		UsingExpr:     usingExpr,
		WithCheckExpr: withCheckExpr,
	})
	return sqlerr.WrapSQLError(err)
}

// Delete 删除指定策略
func (r *SqlPolicyRepository) Delete(ctx context.Context, id int64, orgName, projectSlug string) error {
	err := r.q.DeletePolicy(ctx, dbgen.DeletePolicyParams{
		ID:          uint64(id),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	return sqlerr.WrapSQLError(err)
}

// DeleteByModel 删除模型的所有策略
func (r *SqlPolicyRepository) DeleteByModel(ctx context.Context, orgName, projectSlug, modelID string) error {
	err := r.q.DeletePoliciesByModel(ctx, dbgen.DeletePoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	return sqlerr.WrapSQLError(err)
}
