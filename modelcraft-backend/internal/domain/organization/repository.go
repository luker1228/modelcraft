package organization

import (
	"context"
)

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	// Create 创建组织
	Create(ctx context.Context, org *Organization) error

	// GetByName 根据名称获取组织
	GetByName(ctx context.Context, name string) (*Organization, error)

	// GetByPhone 根据手机号获取组织（全局唯一）
	// 返回 nil, shared.NewNotFoundError 当不存在时
	GetByPhone(ctx context.Context, phone string) (*Organization, error)

	// ListByUser 获取用户所属的所有组织
	ListByUser(ctx context.Context, userID string) ([]*Organization, error)

	// Update 更新组织
	Update(ctx context.Context, org *Organization) error

	// ExistsByName 检查组织名称是否已存在
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ExistsByPhone 检查 org phone 是否已被注册
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
}
