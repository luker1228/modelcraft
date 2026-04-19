package modeldesign

import "context"

// ModelRLSPolicyRepository RLS 策略 Repository 接口
type ModelRLSPolicyRepository interface {
	// GetByModelID 根据 Model ID 获取 Policy
	GetByModelID(ctx context.Context, orgName, projectSlug, modelID string) (*ModelRLSPolicy, error)

	// Save 保存 Policy（upsert）
	Save(ctx context.Context, orgName, projectSlug string, policy *ModelRLSPolicy) error

	// DeleteByModelID 删除指定 Model 的 Policy
	DeleteByModelID(ctx context.Context, orgName, projectSlug, modelID string) error

	// ExistsByModelID 判断指定 Model 是否有 Policy
	ExistsByModelID(ctx context.Context, orgName, projectSlug, modelID string) (bool, error)
}
