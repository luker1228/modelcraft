package user

import (
	"context"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error

	// GetByID 根据内部 UUID 获取用户
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByExternalID 根据外部认证提供者 ID 获取用户
	GetByExternalID(ctx context.Context, externalID string) (*User, error)

	// ExistsByExternalID 检查外部 ID 是否已存在
	ExistsByExternalID(ctx context.Context, externalID string) (bool, error)

	// FindIDByExternalID retrieves the internal user ID by external authentication provider ID.
	// Returns ("", false, nil) if no user matches the given externalID.
	// Returns ("", false, err) on system failure.
	FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error)

	// GetByPhone 根据 org 和手机号获取用户 (Pattern A: 不存在返回 NotFoundError)
	GetByPhone(ctx context.Context, orgName, phone string) (*User, error)

	// GetByName 根据 org 和用户名获取用户 (Pattern A: 不存在返回 NotFoundError)
	GetByName(ctx context.Context, orgName, name string) (*User, error)

	// GetByNameGlobal 全局按 userName 查找用户（跨 org，用于管理员登录）
	GetByNameGlobal(ctx context.Context, name string) (*User, error)

	// ExistsByPhone 检查 org 内手机号是否已被注册 (Pattern B: 不存在返回 false)
	ExistsByPhone(ctx context.Context, orgName, phone string) (bool, error)

	// ExistsByName 检查 org 内用户名是否已被占用 (Pattern B: 不存在返回 false)
	ExistsByName(ctx context.Context, orgName, name string) (bool, error)

	// ListByOrg 返回 org 下所有用户列表
	ListByOrg(ctx context.Context, orgName string) ([]*User, error)
}
