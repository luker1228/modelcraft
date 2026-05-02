package rbac

import (
	"modelcraft/internal/domain/project"
	rbacdomain "modelcraft/internal/domain/rbac"
)

// CreatePermissionCommand 创建权限点命令
type CreatePermissionCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	ModelID              string
	Name                 string
	Description          *string
	Action               rbacdomain.Action
	ColumnPolicy         *rbacdomain.ColumnPolicy
	RowScope             rbacdomain.RowScope
}

// ApplyPresetPolicyCommand 应用预设策略命令。
// Preset 为空时：执行模型级 reconcile（同步该模型全部可适配内置预设）。
// Preset 非空时：兼容显式预设请求（用于特定调用路径的 owner 校验语义）。
type ApplyPresetPolicyCommand struct {
	project.ProjectScope
	ModelID string
	Preset  *rbacdomain.PermissionPreset
}

// UpdatePermissionCommand 更新权限点命令（只允许更新 name/description/columnPolicy；action 和 rowScope 不可变）
type UpdatePermissionCommand struct {
	OrgName      string
	ID           string
	Name         string
	Description  *string
	ColumnPolicy *rbacdomain.ColumnPolicy
}

// DeletePermissionCommand 删除权限点命令
type DeletePermissionCommand struct {
	OrgName string
	ID      string
}

// CreateBundleCommand 创建权限包命令
type CreateBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	Name                 string
	Description          *string
	// Slug 可选，不传时从 Name 自动派生。同项目内唯一，创建后不可修改。
	Slug *string
}

// UpdateBundleCommand 更新权限包命令
type UpdateBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	ID                   string
	Name                 string
	Description          *string
}

// DeleteBundleCommand 删除权限包命令
type DeleteBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	ID                   string
}

// AddPermissionToBundleCommand 向权限包添加权限点命令
type AddPermissionToBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	PermissionID         string
	SortOrder            int
}

// AddPresetToBundleCommand 向权限包添加模型预设权限命令（后端自动 ensure 预设权限点）
type AddPresetToBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	ModelID              string
	Preset               rbacdomain.PermissionPreset
	SortOrder            int
}

// RemovePermissionFromBundleCommand 从权限包移除权限点命令
type RemovePermissionFromBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	PermissionID         string
}

// CreateRoleCommand 创建 RBAC 角色命令
type CreateRoleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	Name                 string
	Description          *string
	IsImplicit           bool
}

// UpdateRoleCommand 更新角色命令（is_implicit=true 的角色会被 GuardUpdate() 阻断）
type UpdateRoleCommand struct {
	OrgName     string
	ID          string
	Name        string
	Description *string
}

// DeleteRoleCommand 删除角色命令
type DeleteRoleCommand struct {
	OrgName string
	ID      string
}

// AssignBundleToRoleCommand 将权限包关联到角色命令
type AssignBundleToRoleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	RoleID               string
	BundleID             string
}

// RevokeBundleFromRoleCommand 从角色解除权限包关联命令
type RevokeBundleFromRoleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	RoleID               string
	BundleID             string
}

// GrantBundleToUserCommand 直接将权限包授予用户命令（通道 1）
type GrantBundleToUserCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	UserID               string
	BundleID             string
}

// RevokeBundleFromUserCommand 撤销用户权限包直接授权命令
type RevokeBundleFromUserCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	UserID               string
	BundleID             string
}

// AssignRoleToUserCommand 将显式角色分配给用户命令（通道 2）
type AssignRoleToUserCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	UserID               string
	RoleID               string
}

// RevokeRoleFromUserCommand 撤销用户角色分配命令
type RevokeRoleFromUserCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	UserID               string
	RoleID               string
}

// ListProjectEndUserRoleUsersQuery 查询 Project 下有角色分配的用户列表命令
type ListProjectEndUserRoleUsersQuery struct {
	project.ProjectScope        // 嵌入: OrgName + ProjectSlug
	Search               string // username 模糊搜索
	RoleID               string // 按 Role 过滤（可选）
	First                int
	After                string
}

// RestoreBundleCommand 回滚权限包到历史快照命令
type RestoreBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	TargetVersion        int
}

// RestoreBundleResult 回滚权限包结果
type RestoreBundleResult struct {
	Bundle     *rbacdomain.EndUserPermissionBundle
	NewVersion int
}

// BindPresetItemToBundleCommand 绑定预设模板 item 到 bundle（replace 语义）
// 同一 bundle 下同一 model 最多一个 item；重复绑定会替换旧 item。
type BindPresetItemToBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	ModelID              string
	Preset               rbacdomain.PermissionPreset
	SortOrder            int
}

// BindCustomItemToBundleCommand 绑定自定义权限 item 到 bundle（replace 语义）
// 同一 bundle 下同一 model 最多一个 item；重复绑定会替换旧 item。
type BindCustomItemToBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	ModelID              string
	CustomPermissionID   string
	SortOrder            int
}

// RemoveDataPermissionItemFromBundleCommand 从 bundle 中移除指定模型的 item
type RemoveDataPermissionItemFromBundleCommand struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	BundleID             string
	ModelID              string
}

// GetEffectivePermissionsQuery 获取用户有效权限集查询
type GetEffectivePermissionsQuery struct {
	project.ProjectScope // 嵌入: OrgName + ProjectSlug
	UserID               string
}
