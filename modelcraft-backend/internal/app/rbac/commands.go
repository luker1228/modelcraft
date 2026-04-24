package rbac

import rbacdomain "modelcraft/internal/domain/rbac"

// CreatePermissionCommand 创建权限点命令
type CreatePermissionCommand struct {
	OrgName      string
	ProjectSlug  string
	ModelID      string
	Name         string
	Description  *string
	Action       rbacdomain.Action
	ColumnPolicy *rbacdomain.ColumnPolicy
	RowScope     rbacdomain.RowScope
}

// UpdatePermissionCommand 更新权限点命令（只允许更新 name/description/columnPolicy）
type UpdatePermissionCommand struct {
	OrgName      string
	ID           string
	Name         string
	Description  *string
	ColumnPolicy *rbacdomain.ColumnPolicy
}

// CreateBundleCommand 创建权限包命令
type CreateBundleCommand struct {
	OrgName     string
	ProjectSlug string
	Name        string
	Description *string
}

// AddPermissionToBundleCommand 向权限包添加权限点命令
type AddPermissionToBundleCommand struct {
	OrgName      string
	BundleID     string
	PermissionID string
	SortOrder    int
}

// CreateRoleCommand 创建 RBAC 角色命令
type CreateRoleCommand struct {
	OrgName     string
	ProjectSlug string
	Name        string
	Description *string
	IsImplicit  bool
}

// UpdateRoleCommand 更新角色命令（is_implicit=true 的角色会被 GuardUpdate() 阻断）
type UpdateRoleCommand struct {
	OrgName     string
	ID          string
	Name        string
	Description *string
}

// GetEffectivePermissionsQuery 获取用户有效权限集查询
type GetEffectivePermissionsQuery struct {
	UserID      string
	OrgName     string
	ProjectSlug string
}
