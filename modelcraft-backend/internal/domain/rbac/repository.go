package rbac

import "context"

// EndUserPermissionRepository RBAC 权限系统仓储接口（Project 维度）
type EndUserPermissionRepository interface {
	// ─── 权限点 ────────────────────────────────────────────────

	// CreatePermission 创建权限点（org + project scoped，orgName 由实体携带）
	CreatePermission(ctx context.Context, p *EndUserPermission) error

	// GetPermissionByID 根据 ID 获取权限点（org scoped，防跨租户枚举）
	GetPermissionByID(ctx context.Context, orgName, id string) (*EndUserPermission, error)

	// ListPermissionsByProject 列出项目下所有权限点（org + project scoped）
	ListPermissionsByProject(ctx context.Context, orgName, projectSlug string) ([]*EndUserPermission, error)

	// ListPermissionsByModel 列出指定 Model 下的所有权限点（org scoped）
	ListPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*EndUserPermission, error)

	// UpdatePermission 更新权限点 name/description/columnPolicy
	// rowScope 和 action 不允许更新（需删除重建，由 App 层保证）
	UpdatePermission(ctx context.Context, p *EndUserPermission) error

	// DeletePermission 删除权限点（org scoped）
	// 级联删除 end_user_bundle_permissions 中的关联行（FK CASCADE）
	DeletePermission(ctx context.Context, orgName, id string) error

	// ─── 权限包 ────────────────────────────────────────────────

	// CreateBundle 创建权限包（org + project scoped，orgName 由实体携带）
	CreateBundle(ctx context.Context, b *EndUserPermissionBundle) error

	// GetBundleByID 根据 ID 获取权限包（org scoped）
	GetBundleByID(ctx context.Context, orgName, id string) (*EndUserPermissionBundle, error)

	// ListBundlesByProject 列出项目下所有权限包（org + project scoped）
	ListBundlesByProject(ctx context.Context, orgName, projectSlug string) ([]*EndUserPermissionBundle, error)

	// UpdateBundle 更新权限包 name/description
	UpdateBundle(ctx context.Context, b *EndUserPermissionBundle) error

	// DeleteBundle 删除权限包（org scoped）
	// 级联删除：end_user_bundle_permissions / end_user_role_bundles / end_user_user_bundles（FK CASCADE）
	DeleteBundle(ctx context.Context, orgName, id string) error

	// AddPermissionToBundle 向权限包添加权限点（有序，sort_order 由调用方提供）
	AddPermissionToBundle(ctx context.Context, bundleID, permissionID string, sortOrder int) error

	// RemovePermissionFromBundle 从权限包移除权限点
	RemovePermissionFromBundle(ctx context.Context, bundleID, permissionID string) error

	// ListPermissionsInBundle 列出权限包内所有权限点（按 sort_order 升序）
	ListPermissionsInBundle(ctx context.Context, bundleID string) ([]*EndUserPermission, error)

	// ─── 业务角色 ──────────────────────────────────────────────

	// CreateRole 创建 RBAC 业务角色（org + project scoped，orgName 由实体携带）
	CreateRole(ctx context.Context, r *EndUserRole) error

	// GetRoleByID 根据 ID 获取角色（org scoped）
	GetRoleByID(ctx context.Context, orgName, id string) (*EndUserRole, error)

	// ListRolesByProject 列出项目下所有角色（org + project scoped）
	// 隐式角色排在前面（is_implicit DESC）
	ListRolesByProject(ctx context.Context, orgName, projectSlug string) ([]*EndUserRole, error)

	// UpdateRole 更新角色 name/description（is_implicit=true 的角色由业务层阻断）
	UpdateRole(ctx context.Context, r *EndUserRole) error

	// DeleteRole 删除角色（org scoped，is_implicit=true 的角色由业务层阻断）
	DeleteRole(ctx context.Context, orgName, id string) error

	// AssignBundleToRole 将权限包授予角色（M:N）
	AssignBundleToRole(ctx context.Context, orgName, projectSlug, roleID, bundleID string) error

	// RevokeBundleFromRole 撤销角色对权限包的关联
	RevokeBundleFromRole(ctx context.Context, roleID, bundleID string) error

	// ListBundlesByRole 列出角色关联的所有权限包
	ListBundlesByRole(ctx context.Context, roleID string) ([]*EndUserPermissionBundle, error)

	// ─── 用户授权 ──────────────────────────────────────────────

	// GrantBundleToUser 直接将权限包授予用户（鉴权 Step 1 数据源）
	GrantBundleToUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error

	// RevokeBundleFromUser 撤销用户对权限包的直接授权
	RevokeBundleFromUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error

	// AssignRoleToUser 将显式角色授予用户（鉴权 Step 2 数据源）
	AssignRoleToUser(ctx context.Context, userID, orgName, projectSlug, roleID string) error

	// RevokeRoleFromUser 撤销用户的角色关联
	RevokeRoleFromUser(ctx context.Context, userID, orgName, projectSlug, roleID string) error

	// ─── 鉴权核心查询（3 条链式，对应 Step 1~3） ─────────────────

	// GetBundleIDsByUserDirect 获取用户直接关联的权限包 ID 列表（鉴权 Step 1）
	// 空列表（无授权）为合法状态，不返回错误
	GetBundleIDsByUserDirect(ctx context.Context, userID, orgName, projectSlug string) ([]string, error)

	// GetBundleIDsByUserExplicitRoles 获取用户显式角色关联的权限包 ID 列表（鉴权 Step 2）
	// 单次 JOIN 查询，避免 N+1；空列表为合法状态
	GetBundleIDsByUserExplicitRoles(ctx context.Context, userID, orgName, projectSlug string) ([]string, error)

	// GetBundleIDsByImplicitRoles 获取所有隐式角色关联的权限包 ID 列表（鉴权 Step 3）
	// 对所有认证用户执行，无需 userID；空列表为合法状态
	GetBundleIDsByImplicitRoles(ctx context.Context, orgName, projectSlug string) ([]string, error)

	// GetPermissionsByBundleIDs 展开权限包 → 权限点（鉴权 Step 4）
	// bundleIDs 为 Step 1~3 合并去重后的 ID 集合；bundleIDs 为空时直接返回空 slice
	GetPermissionsByBundleIDs(ctx context.Context, orgName string, bundleIDs []string) ([]*EndUserPermission, error)
}
