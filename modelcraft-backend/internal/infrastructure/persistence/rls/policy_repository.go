package rls

import (
	"context"
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
			ID:           int64(row.ID),
			OrgName:      row.OrgName,
			ProjectSlug:  row.ProjectSlug,
			ModelID:      row.ModelID,
			PolicyName:   row.PolicyName,
			Action:       rls.Action(row.Action),
			Role:         row.Role,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
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
