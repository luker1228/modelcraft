package role

import (
	"context"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// GetByID 根据 UUID 获取角色
	GetByID(ctx context.Context, id string) (*Role, error)

	// GetByName 根据名称获取角色
	GetByName(ctx context.Context, name string) (*Role, error)

	// GetSystemRoleByName 根据名称获取系统角色（is_system=true）
	GetSystemRoleByName(ctx context.Context, name string) (*Role, error)

	// List 获取所有角色
	List(ctx context.Context) ([]*Role, error)

	// Create 创建自定义角色
	Create(ctx context.Context, role *Role) error

	// Update 更新角色
	Update(ctx context.Context, role *Role) error

	// Delete 删除角色（仅限非系统角色）
	Delete(ctx context.Context, id string) error
}
