package profile

import "context"

// Repository 定义 profile 聚合的持久化契约。
type Repository interface {
	// Create 兼容现有注册流程，创建用户初始 profile。
	Create(ctx context.Context, profile *Profile) error

	// CreateInitialProfile 创建用户初始 profile。
	CreateInitialProfile(ctx context.Context, profile *Profile) error

	// FindByUserID 按 userID 查询 profile（显式 orgName 做租户约束）。
	FindByUserID(ctx context.Context, orgName, userID string) (*Profile, error)

	// UpdateByUserID 按 userID 执行部分字段更新（PATCH 语义）。
	UpdateByUserID(ctx context.Context, orgName, userID string, patch UpdatePatch) error
}
