package rls

import (
	"context"
	"errors"
	"fmt"
	"modelcraft/internal/domain/rls"
)

// DataPolicyService V2 策略 CRUD 应用服务
type DataPolicyService struct {
	repo rls.PolicyRepositoryV2
}

// NewDataPolicyService 创建 DataPolicyService
func NewDataPolicyService(repo rls.PolicyRepositoryV2) *DataPolicyService {
	return &DataPolicyService{repo: repo}
}

// ListByModel 查询模型的所有策略
func (s *DataPolicyService) ListByModel(
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

// Upsert 创建或更新策略（按 model+role+action 唯一键）
func (s *DataPolicyService) Upsert(
	ctx context.Context, orgName, projectSlug string, input UpsertInput,
) (*rls.Policy, error) {
	if input.Role == "" {
		return nil, fmt.Errorf("role must not be empty: use \"*\" to match all end-users")
	}

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
	persisted, err := s.repo.GetByRoleAction(ctx, orgName, projectSlug, input.ModelID, input.Action, input.Role)
	if err != nil {
		if errors.Is(err, rls.ErrPolicyNotFound) {
			return policy, nil
		}
		return nil, err
	}
	return persisted, nil
}

// Delete 删除单条策略
func (s *DataPolicyService) Delete(
	ctx context.Context, orgName, projectSlug string, id int64,
) error {
	return s.repo.Delete(ctx, orgName, projectSlug, id)
}

// DeleteByModel 删除模型的所有策略
func (s *DataPolicyService) DeleteByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) error {
	return s.repo.DeleteByModel(ctx, orgName, projectSlug, modelID)
}

// DeleteByRole 删除模型下某角色的所有策略
func (s *DataPolicyService) DeleteByRole(
	ctx context.Context, orgName, projectSlug, modelID, role string,
) error {
	return s.repo.DeleteByRole(ctx, orgName, projectSlug, modelID, role)
}
