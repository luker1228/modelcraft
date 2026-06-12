package rls

import (
	"context"
	"modelcraft/internal/domain/rls"
)

// PolicyCRUDService V2 策略 CRUD 应用服务
type PolicyCRUDService struct {
	repo rls.PolicyRepositoryV2
}

// NewPolicyCRUDService 创建 PolicyCRUDService
func NewPolicyCRUDService(repo rls.PolicyRepositoryV2) *PolicyCRUDService {
	return &PolicyCRUDService{repo: repo}
}

// ListByModel 查询模型的所有策略
func (s *PolicyCRUDService) ListByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) ([]*rls.Policy, error) {
	return s.repo.ListByModel(ctx, orgName, projectSlug, modelID)
}

// UpsertInput upsert 输入
type UpsertInput struct {
	ModelID       string
	PolicyName    string
	Action        rls.Action
	Role          string
	UsingExpr     rls.JsonExpr
	WithCheckExpr rls.JsonExpr
}

// Upsert 创建或更新策略
func (s *PolicyCRUDService) Upsert(
	ctx context.Context, orgName, projectSlug string, input UpsertInput,
) (*rls.Policy, error) {
	policy := &rls.Policy{
		ModelID:       input.ModelID,
		PolicyName:    input.PolicyName,
		Action:        input.Action,
		Role:          input.Role,
		UsingExpr:     input.UsingExpr,
		WithCheckExpr: input.WithCheckExpr,
	}

	if err := s.repo.Upsert(ctx, orgName, projectSlug, policy); err != nil {
		return nil, err
	}

	// Re-query to get the persisted record (with ID, timestamps)
	policies, err := s.repo.ListByModel(ctx, orgName, projectSlug, input.ModelID)
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		if p.PolicyName == input.PolicyName {
			return p, nil
		}
	}
	return policy, nil
}

// Delete 删除单条策略
func (s *PolicyCRUDService) Delete(
	ctx context.Context, orgName, projectSlug string, id int64,
) error {
	return s.repo.Delete(ctx, orgName, projectSlug, id)
}

// DeleteByModel 删除模型的所有策略
func (s *PolicyCRUDService) DeleteByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) error {
	return s.repo.DeleteByModel(ctx, orgName, projectSlug, modelID)
}
