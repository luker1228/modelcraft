package projectgraphql

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/ctxutils"
	"time"
)

// stringPtr returns nil for empty string, otherwise a pointer to s.
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// getOrgAndProjectFromContext extracts orgName and projectSlug from GraphQL context.
func getOrgAndProjectFromContext(ctx context.Context) (orgName, projectSlug string, err error) {
	orgName, err = ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get organization name from context: %w", err)
	}

	projectSlug, err = ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project slug from context: %w", err)
	}

	return orgName, projectSlug, nil
}

func convertGraphQLExprTypeToDomain(exprType generated.RLSExprType) string {
	switch exprType {
	case generated.RLSExprTypeSelectPredicate:
		return "SELECT_PREDICATE"
	case generated.RLSExprTypeInsertCheck:
		return "INSERT_CHECK"
	case generated.RLSExprTypeUpdatePredicate:
		return "UPDATE_PREDICATE"
	case generated.RLSExprTypeUpdateCheck:
		return "UPDATE_CHECK"
	case generated.RLSExprTypeDeletePredicate:
		return "DELETE_PREDICATE"
	default:
		return "SELECT_PREDICATE"
	}
}

func toActionPolicy(p *rls.Policy) *generated.RlsActionPolicy {
	policyName := p.PolicyName
	return &generated.RlsActionPolicy{
		ID:            fmt.Sprintf("%d", p.ID),
		Action:        generated.RlsAction(p.Action),
		PolicyName:    &policyName,
		UsingExpr:     stringPtr(string(p.UsingExpr)),
		WithCheckExpr: stringPtr(string(p.WithCheckExpr)),
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}
