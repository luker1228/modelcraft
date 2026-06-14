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

type usingCompiler interface {
	CompileUsing(ctx context.Context, expr string, userCtx *rls.UserContext) (*rls.CompiledPolicy, error)
}

type inputCheckEvaluator interface {
	ValidateInput(ctx context.Context, expr string, input map[string]any, userCtx *rls.UserContext) error
}

// PolicyMatchingService 策略匹配 + OR 合并引擎
type PolicyMatchingService struct {
	repo           PolicyRepository
	usingCompiler  usingCompiler
	checkEvaluator inputCheckEvaluator
}

// NewPolicyMatchingService 创建 PolicyMatchingService
func NewPolicyMatchingService(repo PolicyRepository, usingCompiler usingCompiler, checkEvaluator inputCheckEvaluator) *PolicyMatchingService {
	return &PolicyMatchingService{repo: repo, usingCompiler: usingCompiler, checkEvaluator: checkEvaluator}
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
		return "", nil, fmt.Errorf("%w: action=%s, roles=%v", rls.ErrNoMatchingPolicy, action, userCtx.Roles)
	}

	var orClauses []string
	var allParams []interface{}

	for _, p := range policies {
		expr := p.UsingExpr
		if expr == "" {
			continue
		}

		compiled, err := s.usingCompiler.CompileUsing(ctx, string(expr), userCtx)
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

func (s *PolicyMatchingService) ValidateCheck(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	action rls.Action,
	input map[string]any,
	userCtx *rls.UserContext,
) error {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return fmt.Errorf("RLS deny: no matching policy for action=%s", action)
	}

	for _, p := range policies {
		if p.WithCheckExpr == "" {
			continue
		}
		if err := s.checkEvaluator.ValidateInput(ctx, string(p.WithCheckExpr), input, userCtx); err == nil {
			return nil
		}
	}

	return fmt.Errorf("RLS CHECK violation: no matching input check policy passed")
}

// GetCheckExpr returns the first non-empty raw CHECK expression string for the given action.
// Returns ("", nil) if no matching policy or no CHECK expression exists.
func (s *PolicyMatchingService) GetCheckExpr(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, error) {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return "", err
	}
	if len(policies) == 0 {
		return "", nil
	}
	for _, p := range policies {
		if p.WithCheckExpr != "" {
			return string(p.WithCheckExpr), nil
		}
	}
	return "", nil
}

// ResolveCheck is deprecated during the CEL transition. New write paths should use ValidateCheck.
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

		compiled, err := s.usingCompiler.CompileUsing(ctx, string(expr), userCtx)
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
	ValidateCheck(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, input map[string]any, userCtx *rls.UserContext) error
}
