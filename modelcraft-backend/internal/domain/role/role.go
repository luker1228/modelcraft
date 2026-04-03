package role

import (
	"fmt"
	"strings"
	"time"
)

// 系统预置角色 ID
const (
	OwnerRoleID  = "role-owner"
	AdminRoleID  = "role-admin"
	MemberRoleID = "role-member"
	ViewerRoleID = "role-viewer"
)

// Role 角色实体
// 每个角色定义一组权限，用于 RBAC 授权
type Role struct {
	ID          string   // UUID
	Name        string   // 角色名称：Owner、Admin、Member、Viewer 或自定义
	Description string   // 角色描述
	Permissions []string // 权限数组：["project:create", "model:*", "*"]
	IsSystem    bool     // 系统角色不可删除
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate 验证角色实体
func (r *Role) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("role ID is required")
	}
	if r.Name == "" {
		return fmt.Errorf("role name is required")
	}
	if len(r.Name) > 64 {
		return fmt.Errorf("role name must be at most 64 characters")
	}
	if r.Permissions == nil {
		return fmt.Errorf("role permissions are required")
	}
	for _, perm := range r.Permissions {
		if !isValidPermission(perm) {
			return fmt.Errorf("invalid permission format: %s (expected resource:action, resource:*, or *)", perm)
		}
	}
	return nil
}

// isValidPermission 验证权限格式
// 有效格式：
//   - "*" 全局通配
//   - "resource:action" 如 "project:create"
//   - "resource:*" 如 "project:*"
func isValidPermission(perm string) bool {
	if perm == "*" {
		return true
	}
	parts := strings.SplitN(perm, ":", 2)
	if len(parts) != 2 {
		return false
	}
	resource, action := parts[0], parts[1]
	if resource == "" || action == "" {
		return false
	}
	return true
}

// NewRole 创建角色实体
func NewRole(id, name, description string, permissions []string, isSystem bool) (*Role, error) {
	now := time.Now()
	role := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		Permissions: permissions,
		IsSystem:    isSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := role.Validate(); err != nil {
		return nil, err
	}
	return role, nil
}

// HasPermission 检查角色是否拥有指定权限
// 支持三种匹配：
//  1. 全局通配 "*" 匹配一切
//  2. 资源通配 "resource:*" 匹配该资源下所有操作
//  3. 精确匹配 "resource:action"
func (r *Role) HasPermission(required string) bool {
	for _, perm := range r.Permissions {
		// 全局通配
		if perm == "*" {
			return true
		}
		// 精确匹配
		if perm == required {
			return true
		}
		// 资源通配：project:* 匹配 project:create
		if strings.HasSuffix(perm, ":*") {
			permResource := strings.TrimSuffix(perm, ":*")
			requiredParts := strings.SplitN(required, ":", 2)
			if len(requiredParts) == 2 && requiredParts[0] == permResource {
				return true
			}
		}
	}
	return false
}

// CanDelete 判断角色是否可以删除（系统角色不可删除）
func (r *Role) CanDelete() bool {
	return !r.IsSystem
}
