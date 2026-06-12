package rls

import "context"

// PolicyRepositoryV2 V2 多策略 CRUD 接口
type PolicyRepositoryV2 interface {
	// ListByModel 查询模型的所有策略
	ListByModel(ctx context.Context, orgName, projectSlug, modelID string) ([]*Policy, error)

	// Upsert 创建或更新单条策略（按 policy_name 唯一键 upsert）
	Upsert(ctx context.Context, orgName, projectSlug string, policy *Policy) error

	// Delete 按 ID 删除单条策略
	Delete(ctx context.Context, orgName, projectSlug string, id int64) error

	// DeleteByModel 删除模型的所有策略
	DeleteByModel(ctx context.Context, orgName, projectSlug, modelID string) error
}
