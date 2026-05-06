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

	// ListPresetPermissionsByModel 列出指定 Model 下的 PRESET 权限点（org scoped）
	ListPresetPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*EndUserPermission, error)

	// GetPermissionByModelTypeName 通过 (model_id,type,name) 定位权限点（org scoped）
	GetPermissionByModelTypeName(
		ctx context.Context,
		orgName, modelID string,
		permissionType PermissionType,
		name string,
	) (*EndUserPermission, error)

	// UpdatePermission 更新权限点 name/description/columnPolicy
	// rowScope 和 action 不允许更新（需删除重建，由 App 层保证）
	UpdatePermission(ctx context.Context, p *EndUserPermission) error

	// DeletePermission 删除权限点（org scoped）
	// 级联删除 end_user_bundle_permissions 中的关联行（FK CASCADE）
	DeletePermission(ctx context.Context, orgName, id string) error

	// DeletePresetPermissionsByModel 删除指定 model 下的全部预设权限点（type=PRESET）
	DeletePresetPermissionsByModel(ctx context.Context, orgName, modelID string) error

	// UpdatePresetPermission 原地更新预设权限点（name/description/rowPolicy/preset）
	// 用于 reconcile 流程中的 toUpdate，保持 permission_id 稳定
	UpdatePresetPermission(ctx context.Context, p *EndUserPermission) error

	// IsPermissionReferencedByBundle 检查权限点是否被任何权限包引用
	IsPermissionReferencedByBundle(ctx context.Context, permissionID string) (bool, error)

	// ─── 权限包 ────────────────────────────────────────────────

	// CreateBundle 创建权限包（org + project scoped，orgName 由实体携带）
	CreateBundle(ctx context.Context, b *EndUserPermissionBundle) error

	// GetBundleByID 根据 ID 获取权限包（org + project scoped）
	GetBundleByID(ctx context.Context, orgName, projectSlug, id string) (*EndUserPermissionBundle, error)

	// GetBundleBySlug 根据 slug 获取权限包（org + project scoped）
	GetBundleBySlug(ctx context.Context, orgName, projectSlug, slug string) (*EndUserPermissionBundle, error)

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

	// UpsertBundleDataPermissionItem 写入或替换 bundle-model 唯一 item。
	UpsertBundleDataPermissionItem(ctx context.Context, item *EndUserBundleDataPermissionItem) error

	// RemoveBundleDataPermissionItem 按 bundle+model 删除 item。
	RemoveBundleDataPermissionItem(ctx context.Context, bundleID, modelID string) error

	// ListBundleDataPermissionItems 列出 bundle 中的 item。
	ListBundleDataPermissionItems(ctx context.Context, bundleID string) ([]*EndUserBundleDataPermissionItem, error)

	// GetBundleDataPermissionItemByBundleAndModel 获取 bundle-model 的唯一 item。
	GetBundleDataPermissionItemByBundleAndModel(
		ctx context.Context,
		bundleID, modelID string,
	) (*EndUserBundleDataPermissionItem, error)

	// ─── 权限包快照 ─────────────────────────────────────────────

	// SaveBundleSnapshot 写入权限包快照（含 restored_from 字段，回滚时非 nil）
	SaveBundleSnapshot(ctx context.Context, snapshot *BundleSnapshot) error

	// ListBundleSnapshots 列出权限包最近 5 个历史快照（按 version DESC）
	ListBundleSnapshots(ctx context.Context, bundleID string) ([]BundleSnapshot, error)

	// DeleteOldBundleSnapshots 删除超出保留上限的旧快照，只保留最近 5 个
	DeleteOldBundleSnapshots(ctx context.Context, bundleID string) error

	// GetBundleCurrentVersion 获取权限包当前最大版本号（无快照时返回 0）
	GetBundleCurrentVersion(ctx context.Context, bundleID string) (int, error)

	// GetBundleSnapshotByVersion 根据版本号获取快照（不存在时返回 NotFoundError）
	GetBundleSnapshotByVersion(ctx context.Context, bundleID string, version int) (*BundleSnapshot, error)

	// ClearBundlePermissions 清空权限包内所有权限点关联（用于回滚的第一步）
	ClearBundlePermissions(ctx context.Context, bundleID string) error

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

	// ListProjectEndUserRoleUsers 列出 Project 下所有有角色分配的用户（支持搜索和分页）
	ListProjectEndUserRoleUsers(
		ctx context.Context,
		query ListProjectEndUserRoleUsersQuery,
	) ([]*ProjectEndUserRoleUser, int64, error)

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

	// FindPermissionsByEndUserAndModel 查询指定 end-user 在某 model 上的
	// 所有有效权限点（跨 role → bundle → permission 链路）。
	// 仅查该 model，不全量拉取，用于 per-request 权限解析。
	// endUserID 或 modelID 为空时返回 nil（不报错）。
	FindPermissionsByEndUserAndModel(
		ctx context.Context,
		orgName, projectSlug, endUserID, modelID string,
	) ([]*EndUserPermission, error)
}
