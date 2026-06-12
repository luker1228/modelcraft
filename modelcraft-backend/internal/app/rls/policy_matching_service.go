package rls

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/rls"
	"strings"
)

// PolicyRepository 策略查询接口
type PolicyRepository interface {
	ListByAction(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, roles []string) ([]*rls.Policy, error)
}

// PolicyMatchingService 策略匹配 + OR 合并引擎
type PolicyMatchingService struct {
	repo     PolicyRepository
	compiler rls.PolicyCompiler
}

// NewPolicyMatchingService 创建 PolicyMatchingService
func NewPolicyMatchingService(repo PolicyRepository, compiler rls.PolicyCompiler) *PolicyMatchingService {
	return &PolicyMatchingService{repo: repo, compiler: compiler}
}

// ResolveUsing 匹配策略并 OR 合并 using 表达式
func (s *PolicyMatchingService) ResolveUsing(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, []interface{}, error) {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return "", nil, err
	}

	if len(policies) == 0 {
		return "", nil, fmt.Errorf("RLS deny: no matching policy for action=%s, roles=%v", action, userCtx.Roles)
	}

	var orClauses []string
	var allParams []interface{}

	for _, p := range policies {
		expr := p.UsingExpr
		if action == rls.ActionCreate {
			expr = p.WithCheckExpr
		}

		if expr == "" {
			continue
		}

		compiled, err := s.compiler.Compile(ctx, expr, userCtx)
		if err != nil {
			return "", nil, fmt.Errorf("compile policy %q: %w", p.PolicyName, err)
		}
		orClauses = append(orClauses, "("+compiled.SQL+")")
		allParams = append(allParams, compiled.Params...)
	}

	if len(orClauses) == 0 {
		return "1=1", nil, nil
	}

	return strings.Join(orClauses, " OR "), allParams, nil
}

// ResolveCheck 匹配策略并 OR 合并 withCheck 表达式
func (s *PolicyMatchingService) ResolveCheck(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, []interface{}, error) {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return "", nil, err
	}

	if len(policies) == 0 {
		return "", nil, fmt.Errorf("RLS deny: no matching policy for action=%s", action)
	}

	var orClauses []string
	var allParams []interface{}

	for _, p := range policies {
		expr := p.WithCheckExpr
		if expr == "" {
			continue
		}

		compiled, err := s.compiler.Compile(ctx, expr, userCtx)
		if err != nil {
			return "", nil, fmt.Errorf("compile check expression for policy %q: %w", p.PolicyName, err)
		}
		orClauses = append(orClauses, "("+compiled.SQL+")")
		allParams = append(allParams, compiled.Params...)
	}

	if len(orClauses) == 0 {
		return "1=0", nil, nil // 无 CHECK 策略 → 拒绝
	}

	return strings.Join(orClauses, " OR "), allParams, nil
}

// Ensure PolicyMatchingService implements MatchingService
var _ MatchingService = (*PolicyMatchingService)(nil)

// MatchingService 匹配引擎接口（runtime 层依赖）
type MatchingService interface {
	ResolveUsing(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
	ResolveCheck(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
}
