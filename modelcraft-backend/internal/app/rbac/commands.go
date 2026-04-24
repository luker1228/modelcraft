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

// UpdateBundleCommand 更新权限包命令
type UpdateBundleCommand struct {
	OrgName     string
	ID          string
	Name        string
	Description *string
}

// RemovePermissionFromBundleCommand 从权限包移除权限点命令
type RemovePermissionFromBundleCommand struct {
	OrgName      string
	BundleID     string
	PermissionID string
}

// DeleteBundleCommand 删除权限包命令
type DeleteBundleCommand struct {
	OrgName string
	ID      string
}

// DeletePermissionCommand 删除权限点命令
type DeletePermissionCommand struct {
	OrgName string
	ID      string
}

// DeleteRoleCommand 删除角色命令
type DeleteRoleCommand struct {
	OrgName string
	ID      string
}

// AssignBundleToRoleCommand 将权限包关联到角色命令
type AssignBundleToRoleCommand struct {
	OrgName  string
	RoleID   string
	BundleID string
}

// RevokeBundleFromRoleCommand 从角色解除权限包关联命令
type RevokeBundleFromRoleCommand struct {
	OrgName  string
	RoleID   string
	BundleID string
}

// GrantBundleToUserCommand 直接将权限包授予用户命令
type GrantBundleToUserCommand struct {
	UserID      string
	OrgName     string
	ProjectSlug string
	BundleID    string
}

// RevokeBundleFromUserCommand 撤销用户权限包命令
type RevokeBundleFromUserCommand struct {
	UserID      string
	OrgName     string
	ProjectSlug string
	BundleID    string
}

// AssignRoleToUserCommand 将角色分配给用户命令
type AssignRoleToUserCommand struct {
	UserID      string
	OrgName     string
	ProjectSlug string
	RoleID      string
}

// RevokeRoleFromUserCommand 撤销用户角色命令
type RevokeRoleFromUserCommand struct {
	UserID      string
	OrgName     string
	ProjectSlug string
	RoleID      string
}

// GetEffectivePermissionsQuery 获取用户有效权限集查询
type GetEffectivePermissionsQuery struct {
	UserID      string
	OrgName     string
	ProjectSlug string
}
