# RBAC 行列级权限系统 — 后端实现计划

> **版本**: v1.2  
> **日期**: 2026-04-24  
> **依赖文档**:
> - PRD: `ai-metadata/prd/rbac/00-rbac-overview.md` ~ `04-department-scope.md`
> - DB Schema: `plans/rbac/db-schema-final.md`（5 张新表 + 1 条 ALTER，`13_rbac_permissions.sql`）
> - API Contract: `plans/rbac/api-contract.md`
> - 架构规范: `ai-metadata/backend/development/architecture.md`
> - Repository 规范: `ai-metadata/backend/development/repo-develop.md`
> - Domain 接口规范: `ai-metadata/backend/development/domain-development.md`

---

## 1. 实现范围概述

### 1.1 功能边界

本次新增 RBAC 系统控制**终端用户（End User）**在 Runtime 阶段对数据表行列的访问。  
与现有 Casbin 系统（`07_roles_permissions.sql`，控制开发者对 project/model/cluster 的 CRUD）**完全独立**，通过 `end_user_` 前缀区分。

**新增能力**：
- Project 维度的权限点（EndUserPermission）、权限包（EndUserPermissionBundle）、业务角色（EndUserRole）CRUD
- 用户授权：直接绑定权限包（UserBundle）、通过角色间接授权（UserRole → RoleBundle）
- 隐式角色：落库定义，鉴权时运行时自动注入，关系不落库
- 鉴权服务：`EndUserAuthzService.GetEffectivePermissions()` 返回用户对指定 Model 的有效权限集
- RLS 集成接口：将 `rowScope` 传递给 RLS 引擎，生成 `WHERE` 子句注入

### 1.2 涉及 DDD 层次与目录

```
modelcraft-backend/
├── db/
│   ├── schema/mysql/
│   │   └── 13_rbac_permissions.sql              ← Wave 1: 5 张新表 + ALTER end_user_roles
│   └── queries/
│       └── rbac/                                ← Wave 1: sqlc 查询文件
│           ├── permission.sql
│           ├── bundle.sql
│           ├── role.sql
│           ├── user_bundle.sql
│           ├── user_role.sql
│           └── authz.sql                        ← 鉴权 3 条链式查询
│
├── internal/
│   ├── domain/
│   │   └── rbac/                                ← Wave 2: 领域实体 + Repository 接口
│   │       ├── permission.go                    # EndUserPermission 实体
│   │       ├── bundle.go                        # EndUserPermissionBundle 实体
│   │       ├── role.go                          # EndUserRole 实体
│   │       ├── authz.go                         # 鉴权结果值对象 EffectivePermissionSet
│   │       ├── row_scope.go                     # RowScope 枚举值对象 + MergeRowScope()
│   │       ├── column_policy.go                 # ColumnPolicy 值对象（对齐 API 合约枚举模型）
│   │       ├── repository.go                    # EndUserPermissionRepository 接口（单一接口文件）
│   │       └── errors.go                        # 领域业务错误码常量
│   │
│   ├── infrastructure/
│   │   ├── dbgen/                               ← sqlc 生成（运行 just generate-sqlc 后自动产生）
│   │   └── repository/
│   │       └── sql_end_user_permission_repository.go  ← Wave 2: Repository 实现（sqlc wrapper）
│   │
│   ├── app/
│   │   └── rbac/                                ← Wave 2: Application Service
│   │       ├── commands.go                      # Command / Query 对象定义
│   │       ├── permission_app.go                # EndUserPermissionAppService（含 rowScope 前提校验）
│   │       ├── bundle_app.go                    # EndUserBundleAppService
│   │       ├── role_app.go                      # EndUserRoleAppService（含隐式角色保护）
│   │       └── authz_app.go                     # EndUserAuthzService（5 步鉴权编排，核心）
│   │
│   └── interfaces/
│       └── graphql/
│           └── project/
│               ├── rbac.resolvers.go            ← Wave 3: GraphQL Resolver
│               └── adapter/
│                   └── rbac_adapter.go          # 实体 → GraphQL DTO 映射 + 错误转换
│
└── api/graph/project/schema/
    └── rbac.graphql                             ← Wave 3: GraphQL Schema（extend type Query/Mutation）
```

---

## 2. 分层实现计划

### Wave 1 — DB & Repository（基础设施层）

**目标**：5 张新表 + 1 条 ALTER 落库 + sqlc 查询文件 + Repository 接口定义 + Infrastructure 实现

#### 2.1.1 迁移文件

文件路径：`db/schema/mysql/13_rbac_permissions.sql`

> Schema 终态已在 `plans/rbac/db-schema-final.md` 中完整定义，直接复制 SQL DDL 即可。

**5 张新表 + 1 条 ALTER**：

| 表名 | 说明 | 是否新建 | 鉴权步骤 |
|------|------|----------|----------|
| `end_user_permissions` | 权限点（model_id × action × column_policy × row_scope） | 新建 | Step 4 展开 |
| `end_user_permission_bundles` | 权限包（唯一正式授权单位） | 新建 | - |
| `end_user_bundle_permissions` | 权限包-权限点中间表（有序，M:N） | 新建 | Step 4 展开 |
| `end_user_role_bundles` | 角色-权限包关联（M:N） | 新建 | Step 2/3 |
| `end_user_user_bundles` | 用户直接授权-权限包 | 新建 | **Step 1** |
| `end_user_roles` | 业务角色（已有表，新增 `is_implicit` 列） | ALTER | Step 2/3 |
| `end_user_role_users` | 用户-业务角色关联（已有表，复用） | 已有 | **Step 2** |

**执行方式**：

```bash
# 确保 Atlas 已安装（见 ai-metadata/backend/tools/tools-installation.md）
just migrate-up
```

#### 2.1.2 sqlc 查询文件

目录：`db/queries/rbac/`

**`permission.sql`** — 权限点 CRUD：

```sql
-- name: CreateEndUserPermission :exec
INSERT INTO end_user_permissions (id, org_name, project_slug, model_id, name, description,
  action, column_policy, row_scope)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetEndUserPermissionByID :one
SELECT * FROM end_user_permissions WHERE id = ? AND org_name = ?;

-- name: ListEndUserPermissionsByProject :many
SELECT * FROM end_user_permissions
WHERE org_name = ? AND project_slug = ?
ORDER BY created_at;

-- name: ListEndUserPermissionsByModel :many
SELECT * FROM end_user_permissions
WHERE model_id = ? AND org_name = ?
ORDER BY action, row_scope;

-- name: UpdateEndUserPermission :execresult
UPDATE end_user_permissions
SET name = ?, description = ?, column_policy = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ?;

-- name: DeleteEndUserPermission :execresult
DELETE FROM end_user_permissions WHERE id = ? AND org_name = ?;
```

**`bundle.sql`** — 权限包 CRUD + 权限点组合：

```sql
-- name: CreateEndUserBundle :exec
INSERT INTO end_user_permission_bundles (id, org_name, project_slug, name, description)
VALUES (?, ?, ?, ?, ?);

-- name: GetEndUserBundleByID :one
SELECT * FROM end_user_permission_bundles WHERE id = ? AND org_name = ?;

-- name: ListEndUserBundlesByProject :many
SELECT * FROM end_user_permission_bundles
WHERE org_name = ? AND project_slug = ?
ORDER BY name;

-- name: UpdateEndUserBundle :execresult
UPDATE end_user_permission_bundles
SET name = ?, description = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ?;

-- name: DeleteEndUserBundle :execresult
DELETE FROM end_user_permission_bundles WHERE id = ? AND org_name = ?;

-- name: AddPermissionToBundle :exec
INSERT INTO end_user_bundle_permissions (id, bundle_id, permission_id, sort_order)
VALUES (?, ?, ?, ?);

-- name: RemovePermissionFromBundle :execresult
DELETE FROM end_user_bundle_permissions WHERE bundle_id = ? AND permission_id = ?;

-- name: ListPermissionsInBundle :many
SELECT p.* FROM end_user_permissions p
  JOIN end_user_bundle_permissions bp ON p.id = bp.permission_id
WHERE bp.bundle_id = ?
ORDER BY bp.sort_order, bp.created_at;
```

**`role.sql`** — 业务角色 CRUD：

```sql
-- name: CreateEndUserRole :exec
INSERT INTO end_user_roles (id, org_name, project_slug, name, description, is_implicit)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetEndUserRoleByID :one
SELECT * FROM end_user_roles WHERE id = ? AND org_name = ?;

-- name: ListEndUserRolesByProject :many
SELECT * FROM end_user_roles
WHERE org_name = ? AND project_slug = ?
ORDER BY is_implicit DESC, name;

-- name: UpdateEndUserRole :execresult
-- 注意：is_implicit=TRUE 的角色由业务层阻断，不走 SQL 层约束
UPDATE end_user_roles
SET name = ?, description = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND is_implicit = FALSE;

-- name: DeleteEndUserRole :execresult
-- 注意：is_implicit=TRUE 的角色由业务层阻断，不走 SQL 层约束
DELETE FROM end_user_roles
WHERE id = ? AND org_name = ? AND is_implicit = FALSE;

-- name: AssignBundleToRole :exec
INSERT INTO end_user_role_bundles (id, role_id, bundle_id) VALUES (?, ?, ?);

-- name: RevokeBundleFromRole :execresult
DELETE FROM end_user_role_bundles WHERE role_id = ? AND bundle_id = ?;

-- name: ListBundlesByRole :many
SELECT b.* FROM end_user_permission_bundles b
  JOIN end_user_role_bundles rb ON b.id = rb.bundle_id
WHERE rb.role_id = ?;
```

**`user_bundle.sql`** — 用户直接授权（鉴权 Step 1）：

```sql
-- name: GrantBundleToUser :exec
INSERT INTO end_user_user_bundles (id, user_id, org_name, project_slug, bundle_id)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeBundleFromUser :execresult
DELETE FROM end_user_user_bundles
WHERE user_id = ? AND bundle_id = ? AND org_name = ? AND project_slug = ?;

-- name: ListBundlesByUser :many
SELECT b.* FROM end_user_permission_bundles b
  JOIN end_user_user_bundles ub ON b.id = ub.bundle_id
WHERE ub.user_id = ? AND ub.org_name = ? AND ub.project_slug = ?;
```

**`user_role.sql`** — 显式角色关联：

```sql
-- name: AssignRoleToUser :exec
INSERT INTO end_user_role_users (id, user_id, role_id, org_name, project_slug)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeRoleFromUser :execresult
DELETE FROM end_user_role_users
WHERE user_id = ? AND role_id = ? AND org_name = ? AND project_slug = ?;

-- name: ListRolesByUser :many
SELECT role_id FROM end_user_role_users
WHERE user_id = ? AND org_name = ? AND project_slug = ?;
```

**`authz.sql`** — 鉴权核心链式查询（3 条，对应 Step 1~3）：

```sql
-- name: GetBundleIDsByUserDirect :many
-- ⚡ 鉴权链 Step 1: 用户直接关联的权限包 ID 列表
SELECT bundle_id FROM end_user_user_bundles
WHERE user_id = ? AND org_name = ? AND project_slug = ?;

-- name: GetBundleIDsByUserExplicitRoles :many
-- ⚡ 鉴权链 Step 2: 通过显式角色关联的权限包 ID 列表（单次 JOIN 查询，避免 N+1）
SELECT DISTINCT rb.bundle_id
FROM end_user_role_users ur
  JOIN end_user_role_bundles rb ON ur.role_id = rb.role_id
WHERE ur.user_id = ? AND ur.org_name = ? AND ur.project_slug = ?;

-- name: GetBundleIDsByImplicitRoles :many
-- ⚡ 鉴权链 Step 3: 隐式角色关联的权限包 ID 列表（对所有认证用户执行，无需 user_id）
SELECT DISTINCT rb.bundle_id
FROM end_user_roles r
  JOIN end_user_role_bundles rb ON r.id = rb.role_id
WHERE r.org_name = ? AND r.project_slug = ? AND r.is_implicit = TRUE;

-- name: GetPermissionsByBundleIDs :many
-- ⚡ 鉴权链 Step 4: 展开权限包 → 权限点（动态 IN，适用于 Step 1~3 合并后的 bundle_id 集合）
SELECT p.*
FROM end_user_permissions p
  JOIN end_user_bundle_permissions bp ON p.id = bp.permission_id
WHERE bp.bundle_id IN (sqlc.slice(bundleIDs))
  AND p.org_name = ?;
```

> **注意**：`GetPermissionsByBundleIDs` 使用 `sqlc.slice` 支持动态 IN 参数（需要 sqlc v1.20+）。

#### 2.1.3 Repository 接口

文件：`internal/domain/rbac/repository.go`

遵循 `ai-metadata/backend/development/domain-development.md` 规范：
- 所有方法第一个参数为 `ctx context.Context`
- 所有查询/删除方法包含显式 `orgName` 参数
- Project 域资源方法包含 `projectSlug` 参数
- `Create` / `Update` 方法通过实体对象携带 `orgName`

```go
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
    AssignBundleToRole(ctx context.Context, roleID, bundleID string) error

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
```

#### 2.1.4 Infrastructure 实现

文件：`internal/infrastructure/repository/sql_end_user_permission_repository.go`

结构体定义（遵循 Go Wrapper 架构，接收 `dbgen.Querier` 接口支持事务）：

```go
package repository

import (
    "context"
    "encoding/json"
    "modelcraft/internal/domain/rbac"
    "modelcraft/internal/infrastructure/dbgen"
    "modelcraft/internal/infrastructure/sqlerr"
)

type SqlEndUserPermissionRepository struct {
    q dbgen.Querier  // 接口，可以是 *sql.DB 或 *sql.Tx（TxManager 注入）
}

func NewSqlEndUserPermissionRepository(q dbgen.Querier) rbac.EndUserPermissionRepository {
    return &SqlEndUserPermissionRepository{q: q}
}

// 编译期接口满足检查（文件末尾必须有）
var _ rbac.EndUserPermissionRepository = (*SqlEndUserPermissionRepository)(nil)
```

**关键实现规则**（来自 `repo-develop.md`）：

| 规则 | 说明 |
|------|------|
| 写操作 | 用 `sqlerr.ExecWithErrorHandling`（INSERT/UPDATE/DELETE） |
| 读操作 | 用 `sqlerr.QueryWithSQLErrorHandling`（SELECT） |
| `column_policy` 字段 | JSON，需 `json.Marshal`/`json.Unmarshal`，错误需透传 |
| `GetPermissionByID` / `GetBundleByID` / `GetRoleByID` | **模式 A**（必须存在，返回 `(*T, error)`，不检查 IsNotFoundError） |
| `GetBundleIDs*` 3 条链式查询 | 返回 `([]string, error)`，空 slice 是合法状态（不是错误） |
| `UpdateRole` / `DeleteRole` | 检查 `RowsAffected`，为 0 时返回 `shared.NewRepositoryError(ErrTypeNoRowsAffected, ...)` |
| 隐式角色保护 | SQL 层通过 `AND is_implicit = FALSE` 条件隐式保护，业务层 App Service 主动前置校验 |

---

### Wave 2 — Domain & App（领域层 + 应用层）

**目标**：领域实体定义 + 业务规则校验 + Application Service 编排

#### 2.2.1 领域实体

**`internal/domain/rbac/column_policy.go`**

```go
package rbac

// ColumnAccessMode 列访问模式枚举（对齐 API 合约）
type ColumnAccessMode string

const (
    ColumnAccessModeVisible  ColumnAccessMode = "VISIBLE"   // 可见可编辑（完整访问）
    ColumnAccessModeReadonly ColumnAccessMode = "READONLY"  // 可见但不可编辑
    ColumnAccessModeMasked   ColumnAccessMode = "MASKED"    // 脱敏显示
    ColumnAccessModeHidden   ColumnAccessMode = "HIDDEN"    // 完全隐藏
)

// columnAccessModeOrder 列访问模式宽泛度排序（值越大权限越宽）
var columnAccessModeOrder = map[ColumnAccessMode]int{
    ColumnAccessModeHidden:   1,
    ColumnAccessModeMasked:   2,
    ColumnAccessModeReadonly: 3,
    ColumnAccessModeVisible:  4,
}

// IsValid 判断枚举值是否合法
func (m ColumnAccessMode) IsValid() bool {
    _, ok := columnAccessModeOrder[m]
    return ok
}

// ColumnRule 单个字段的列访问规则
type ColumnRule struct {
    FieldName   string           `json:"field_name"`
    Mode        ColumnAccessMode `json:"mode"`
    MaskPattern string           `json:"mask_pattern,omitempty"`
}

// ColumnPolicy 列策略（对齐 API 合约结构）
// nil 表示全列默认（按 DefaultMode 决定，默认 VISIBLE）
type ColumnPolicy struct {
    DefaultMode ColumnAccessMode `json:"default_mode"`
    Rules       []ColumnRule     `json:"rules"`
}

// mergeColumnPolicy 合并两个列策略，取更宽泛（更高权限）的结果
// 规则：VISIBLE > READONLY > MASKED > HIDDEN
// - DefaultMode 取两者中更宽泛的
// - Rules 中同 fieldName 的条目取更宽泛的 Mode；仅一方有的条目直接保留
func mergeColumnPolicy(a, b *ColumnPolicy) *ColumnPolicy {
    if a == nil {
        return b
    }
    if b == nil {
        return a
    }

    // DefaultMode 取更宽泛
    mergedDefault := a.DefaultMode
    if columnAccessModeOrder[b.DefaultMode] > columnAccessModeOrder[a.DefaultMode] {
        mergedDefault = b.DefaultMode
    }

    // 合并 Rules：按 fieldName 索引，取更宽泛的 Mode
    ruleMap := make(map[string]ColumnRule)
    for _, r := range a.Rules {
        ruleMap[r.FieldName] = r
    }
    for _, r := range b.Rules {
        if existing, ok := ruleMap[r.FieldName]; ok {
            if columnAccessModeOrder[r.Mode] > columnAccessModeOrder[existing.Mode] {
                ruleMap[r.FieldName] = r
            }
        } else {
            ruleMap[r.FieldName] = r
        }
    }

    mergedRules := make([]ColumnRule, 0, len(ruleMap))
    for _, r := range ruleMap {
        mergedRules = append(mergedRules, r)
    }

    return &ColumnPolicy{
        DefaultMode: mergedDefault,
        Rules:       mergedRules,
    }
}
```

**`internal/domain/rbac/permission.go`**

```go
package rbac

import "modelcraft/pkg/bizerrors"

// Action 操作动作枚举
type Action string

const (
    ActionSelect Action = "select"
    ActionInsert Action = "insert"
    ActionUpdate Action = "update"
    ActionDelete Action = "delete"
    ActionExport Action = "export"
)

// validActions 合法 Action 集合（用于 IsValid 校验）
var validActions = map[Action]struct{}{
    ActionSelect: {},
    ActionInsert: {},
    ActionUpdate: {},
    ActionDelete: {},
    ActionExport: {},
}

// IsValid 判断 Action 枚举值是否合法
func (a Action) IsValid() bool {
    _, ok := validActions[a]
    return ok
}

// EndUserPermission 权限点（最小权限定义单元）
// 包含四个维度：资源(ModelID) × 动作(Action) × 列策略(ColumnPolicy) × 行策略(RowScope)
type EndUserPermission struct {
    OrgName      string
    ProjectSlug  string
    ID           string
    ModelID      string
    Name         string
    Description  *string
    Action       Action
    ColumnPolicy *ColumnPolicy  // nil = 全列默认（DefaultMode=VISIBLE，无 Rules）
    RowScope     RowScope
}

// Validate 校验权限点合法性（纯域内校验，不查 DB）
func (p *EndUserPermission) Validate() error {
    if p.ModelID == "" {
        return bizerrors.NewValidationError("rbac.permission.modelID_required", "modelID is required")
    }
    if p.Name == "" {
        return bizerrors.NewValidationError("rbac.permission.name_required", "name is required")
    }
    if !p.Action.IsValid() {
        return bizerrors.NewValidationError("rbac.permission.invalid_action", "invalid action: "+string(p.Action))
    }
    if !p.RowScope.IsValid() {
        return bizerrors.NewValidationError("rbac.permission.invalid_row_scope", "invalid rowScope: "+string(p.RowScope))
    }
    if p.ColumnPolicy != nil && !p.ColumnPolicy.DefaultMode.IsValid() {
        return bizerrors.NewValidationError("rbac.permission.invalid_column_default_mode", "invalid default column access mode")
    }
    return nil
}
```

**`internal/domain/rbac/row_scope.go`**

```go
package rbac

// RowScope 行策略枚举
type RowScope string

const (
    RowScopeAll             RowScope = "ALL"
    RowScopeSelf            RowScope = "SELF"
    RowScopeDept            RowScope = "DEPT"
    RowScopeDeptAndChildren RowScope = "DEPT_AND_CHILDREN"
)

// rowScopeOrder 行策略范围宽泛度排序（值越大范围越宽）
var rowScopeOrder = map[RowScope]int{
    RowScopeSelf:            1,
    RowScopeDept:            2,
    RowScopeDeptAndChildren: 3,
    RowScopeAll:             4,
}

// IsValid 判断枚举值是否合法
func (r RowScope) IsValid() bool {
    _, ok := rowScopeOrder[r]
    return ok
}

// MergeRowScope 返回两个行策略中范围更宽泛的那个
// 规则：ALL > DEPT_AND_CHILDREN > DEPT > SELF
func MergeRowScope(a, b RowScope) RowScope {
    if rowScopeOrder[a] >= rowScopeOrder[b] {
        return a
    }
    return b
}
```

**`internal/domain/rbac/bundle.go`**

```go
package rbac

import "modelcraft/pkg/bizerrors"

// EndUserPermissionBundle 权限包（系统中唯一正式授权单位）
// 用户和角色只能关联权限包，不能直接关联权限点
type EndUserPermissionBundle struct {
    OrgName     string
    ProjectSlug string
    ID          string
    Name        string
    Description *string
    Permissions []*EndUserPermission  // 展开后按需填充（ListPermissionsInBundle）
}

func (b *EndUserPermissionBundle) Validate() error {
    if b.Name == "" {
        return bizerrors.NewValidationError("rbac.bundle.name_required", "bundle name is required")
    }
    return nil
}
```

**`internal/domain/rbac/role.go`**

```go
package rbac

import "modelcraft/pkg/bizerrors"

// EndUserRole 业务角色（Project 维度）
// 与 Casbin `roles` 表完全独立：roles = Org 维度系统角色，end_user_roles = Project 维度业务角色
type EndUserRole struct {
    OrgName     string
    ProjectSlug string
    ID          string
    Name        string
    Description *string
    // IsImplicit 内置隐式角色标志
    // true: 角色落库，但 end_user_role_users 表中不为每个用户插入关联行；
    //       鉴权时由系统自动注入（Step 3）；不可删除；name 不可更新
    IsImplicit  bool
    Bundles     []*EndUserPermissionBundle  // 按需填充
}

// GuardDelete 隐式角色删除保护（领域规则）
func (r *EndUserRole) GuardDelete() error {
    if r.IsImplicit {
        return bizerrors.NewValidationError(
            "rbac.role.implicit_protected",
            "implicit role cannot be deleted",
        )
    }
    return nil
}

// GuardUpdate 隐式角色更新保护：name 不可更新（领域规则）
func (r *EndUserRole) GuardUpdate() error {
    if r.IsImplicit {
        return bizerrors.NewValidationError(
            "rbac.role.implicit_name_immutable",
            "implicit role name cannot be updated",
        )
    }
    return nil
}
```

**`internal/domain/rbac/authz.go`**

```go
package rbac

// EffectivePermission 有效权限单条记录（Step 5 判定后的产物）
type EffectivePermission struct {
    ModelID      string
    Action       Action
    ColumnPolicy *ColumnPolicy  // nil = 全列默认
    RowScope     RowScope       // 已取最宽泛范围（多来源合并后）
}

// EffectivePermissionSet 用户在某 Project 下的有效权限集合（key = "modelID:action"）
type EffectivePermissionSet map[string]*EffectivePermission

// Merge 将一批权限点合并进有效权限集（Step 5 核心逻辑）
// rowScope 取并集最大范围，columnPolicy 取更宽泛的模式
func (eps EffectivePermissionSet) Merge(permissions []*EndUserPermission) EffectivePermissionSet {
    for _, p := range permissions {
        key := p.ModelID + ":" + string(p.Action)
        existing, ok := eps[key]
        if !ok {
            eps[key] = &EffectivePermission{
                ModelID:      p.ModelID,
                Action:       p.Action,
                ColumnPolicy: p.ColumnPolicy,
                RowScope:     p.RowScope,
            }
        } else {
            // 行策略取最宽泛（ALL > DEPT_AND_CHILDREN > DEPT > SELF）
            existing.RowScope = MergeRowScope(existing.RowScope, p.RowScope)
            // 列策略取并集（DefaultMode 取更宽泛，Rules 中同字段取更宽泛）
            existing.ColumnPolicy = mergeColumnPolicy(existing.ColumnPolicy, p.ColumnPolicy)
        }
    }
    return eps
}

// HasPermission 默认拒绝原则：只有命中 allow 才返回 true
func (eps EffectivePermissionSet) HasPermission(modelID string, action Action) bool {
    _, ok := eps[modelID+":"+string(action)]
    return ok
}

// GetPermission 获取有效权限（用于提取 rowScope 传给 RLS 引擎），不存在返回 nil
func (eps EffectivePermissionSet) GetPermission(modelID string, action Action) *EffectivePermission {
    return eps[modelID+":"+string(action)]
}
```

#### 2.2.2 业务规则：rowScope 字段前提校验

**位置**：`internal/app/rbac/permission_app.go` 的 `validateRowScopePrerequisite()`

**规则**（来自 `prd/rbac/01-permission-model.md`）：

| rowScope | Model 必须存在的字段 | 字段类型 |
|----------|---------------------|----------|
| `SELF` | `owner` | EndUserRef |
| `DEPT` / `DEPT_AND_CHILDREN` | `dept_id` | 任意类型 |
| `ALL` | 无要求 | - |

若字段不存在，返回 `bizerrors.NewValidationError("rbac.permission.row_scope_field_missing", "...")`

> **注意**：此校验需要访问 `ModelRepository.GetFieldsByModelID()`（已有接口），属于跨域协作，放在 App 层而不是 Domain 层。

#### 2.2.3 业务规则：隐式角色保护

**位置**：`internal/app/rbac/role_app.go`

- **删除前**：先 `GetRoleByID()`，调用 `role.GuardDelete()` — 隐式角色返回错误，阻断操作
- **更新前**：先 `GetRoleByID()`，调用 `role.GuardUpdate()` — 隐式角色 name 不可更新

#### 2.2.4 Application Service

**`internal/app/rbac/commands.go`**

```go
package rbac

import "modelcraft/internal/domain/rbac"

// CreatePermissionCommand 创建权限点命令
type CreatePermissionCommand struct {
    OrgName      string
    ProjectSlug  string
    ModelID      string
    Name         string
    Description  *string
    Action       rbac.Action
    ColumnPolicy *rbac.ColumnPolicy
    RowScope     rbac.RowScope
}

// UpdatePermissionCommand 更新权限点命令（只允许更新 name/description/columnPolicy）
type UpdatePermissionCommand struct {
    OrgName      string
    ID           string
    Name         string
    Description  *string
    ColumnPolicy *rbac.ColumnPolicy
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
```

**`internal/app/rbac/authz_app.go`** — 核心鉴权编排（5 步）：

```go
package rbac

import (
    "context"
    "fmt"
    "modelcraft/internal/domain/rbac"
)

// EndUserAuthzService 鉴权应用服务（编排 prd/rbac/03-auth-flow.md 中的 5 步鉴权流程）
type EndUserAuthzService struct {
    rbacRepo rbac.EndUserPermissionRepository
}

func NewEndUserAuthzService(rbacRepo rbac.EndUserPermissionRepository) *EndUserAuthzService {
    return &EndUserAuthzService{rbacRepo: rbacRepo}
}

// GetEffectivePermissions 获取用户在指定 Project 下的有效权限集
//
// 5 步鉴权（来自 prd/rbac/03-auth-flow.md）：
//   Step 1: 查 end_user_user_bundles（直接授权）
//   Step 2: 查 end_user_role_users → end_user_role_bundles（显式角色，单次 JOIN）
//   Step 3: 查 end_user_roles WHERE is_implicit=true → end_user_role_bundles（隐式角色，对所有认证用户执行）
//   Step 4: 展开所有权限点（GetPermissionsByBundleIDs，动态 IN）
//   Step 5: 合并取并集（rowScope 取最宽泛范围）
//
// 返回：
//   - 有效权限集，key = "modelID:action"；空集合（无任何授权）不报错
//   - error 仅在 DB 查询失败时返回
func (s *EndUserAuthzService) GetEffectivePermissions(
    ctx context.Context,
    q GetEffectivePermissionsQuery,
) (rbac.EffectivePermissionSet, error) {
    // Step 1
    directIDs, err := s.rbacRepo.GetBundleIDsByUserDirect(ctx, q.UserID, q.OrgName, q.ProjectSlug)
    if err != nil {
        return nil, fmt.Errorf("rbac authz step1 failed for user %s: %w", q.UserID, err)
    }

    // Step 2
    explicitIDs, err := s.rbacRepo.GetBundleIDsByUserExplicitRoles(ctx, q.UserID, q.OrgName, q.ProjectSlug)
    if err != nil {
        return nil, fmt.Errorf("rbac authz step2 failed for user %s: %w", q.UserID, err)
    }

    // Step 3（对所有认证用户执行，无需 userID）
    implicitIDs, err := s.rbacRepo.GetBundleIDsByImplicitRoles(ctx, q.OrgName, q.ProjectSlug)
    if err != nil {
        return nil, fmt.Errorf("rbac authz step3 failed for project %s/%s: %w", q.OrgName, q.ProjectSlug, err)
    }

    // 合并去重 bundle IDs
    allBundleIDs := deduplicateStrings(append(append(directIDs, explicitIDs...), implicitIDs...))
    if len(allBundleIDs) == 0 {
        return rbac.EffectivePermissionSet{}, nil  // 快速返回空集，无需查 Step 4
    }

    // Step 4: 展开权限点
    permissions, err := s.rbacRepo.GetPermissionsByBundleIDs(ctx, q.OrgName, allBundleIDs)
    if err != nil {
        return nil, fmt.Errorf("rbac authz step4 failed: %w", err)
    }

    // Step 5: 合并取并集（rowScope 取最宽泛范围）
    eps := rbac.EffectivePermissionSet{}
    return eps.Merge(permissions), nil
}
```

**`internal/app/rbac/permission_app.go`** — rowScope 前提校验摘要：

```go
// EndUserPermissionAppService 权限点应用服务
type EndUserPermissionAppService struct {
    rbacRepo  rbac.EndUserPermissionRepository
    modelRepo modeldesign.ModelRepository  // 跨域访问，用于 rowScope 字段前提校验
}

func NewEndUserPermissionAppService(
    rbacRepo rbac.EndUserPermissionRepository,
    modelRepo modeldesign.ModelRepository,
) *EndUserPermissionAppService {
    return &EndUserPermissionAppService{rbacRepo: rbacRepo, modelRepo: modelRepo}
}

// CreatePermission 创建权限点
// 业务规则：rowScope 字段前提校验（App 层，需跨域访问 ModelRepository）
func (s *EndUserPermissionAppService) CreatePermission(
    ctx context.Context,
    cmd CreatePermissionCommand,
) (*rbac.EndUserPermission, error) {
    // 1. rowScope 字段前提校验（SELF 需要 owner 字段，DEPT* 需要 dept_id 字段）
    if err := s.validateRowScopePrerequisite(ctx, cmd.ModelID, cmd.RowScope); err != nil {
        return nil, err
    }

    // 2. 构建实体
    perm := &rbac.EndUserPermission{
        OrgName:      cmd.OrgName,
        ProjectSlug:  cmd.ProjectSlug,
        ID:           bizutils.NewUUID(),
        ModelID:      cmd.ModelID,
        Name:         cmd.Name,
        Description:  cmd.Description,
        Action:       cmd.Action,
        ColumnPolicy: cmd.ColumnPolicy,
        RowScope:     cmd.RowScope,
    }

    // 3. 领域校验（纯域内规则）
    if err := perm.Validate(); err != nil {
        return nil, err
    }

    // 4. 持久化
    if err := s.rbacRepo.CreatePermission(ctx, perm); err != nil {
        return nil, s.convertRepoError(ctx, err, "CreatePermission")
    }
    return perm, nil
}
```

---

### Wave 3 — GraphQL Resolver（接口层）

**目标**：Schema 定义 → 代码生成 → Resolver 实现

#### 2.3.1 Schema 文件

文件：`api/graph/project/schema/rbac.graphql`

核心类型节选（完整 Schema 见 `plans/rbac/api-contract.md`）：

```graphql
# ─── 枚举 ─────────────────────────────────────────────────────────
enum RbacAction {
  SELECT
  INSERT
  UPDATE
  DELETE
  EXPORT
}

enum RbacRowScope {
  ALL
  SELF
  DEPT
  DEPT_AND_CHILDREN
}

enum ColumnAccessMode {
  VISIBLE
  READONLY
  MASKED
  HIDDEN
}

# ─── 列策略类型 ────────────────────────────────────────────────────
type ColumnRule {
  fieldName:   String!
  mode:        ColumnAccessMode!
  maskPattern: String
}

type ColumnPolicy {
  defaultMode: ColumnAccessMode!
  rules:       [ColumnRule!]!
}

input ColumnRuleInput {
  fieldName:   String!
  mode:        ColumnAccessMode!
  maskPattern: String
}

input ColumnPolicyInput {
  defaultMode: ColumnAccessMode!
  rules:       [ColumnRuleInput!]!
}

# ─── 核心类型 ──────────────────────────────────────────────────────
type EndUserPermission {
  id:           ID!
  name:         String!
  description:  String
  modelId:      String!
  action:       RbacAction!
  columnPolicy: ColumnPolicy     # null = 全列默认策略
  rowScope:     RbacRowScope!
}

type EndUserPermissionBundle {
  id:          ID!
  name:        String!
  description: String
  permissions: [EndUserPermission!]!
}

type EndUserRole {
  id:          ID!
  name:        String!
  description: String
  isImplicit:  Boolean!
  bundles:     [EndUserPermissionBundle!]!
}

type EffectivePermission {
  modelId:      String!
  action:       RbacAction!
  columnPolicy: ColumnPolicy
  rowScope:     RbacRowScope!
}

# ─── 错误类型 ──────────────────────────────────────────────────────
type EndUserPermissionNotFoundError        { message: String! }
type EndUserBundleNotFoundError            { message: String! }
type EndUserRoleNotFoundError              { message: String! }
type EndUserValidationError                { message: String! field: String }
type EndUserImplicitRoleProtectedError     { message: String! }
type EndUserRowScopeFieldMissingError      { message: String! requiredField: String! }

# ─── Mutation Payloads（模式：data + error 二选一，不混用） ─────────
type CreateEndUserPermissionPayload {
  permission: EndUserPermission
  error: CreateEndUserPermissionError
}
union CreateEndUserPermissionError =
    EndUserValidationError
  | EndUserRowScopeFieldMissingError

type DeleteEndUserRolePayload {
  error: DeleteEndUserRoleError
}
union DeleteEndUserRoleError =
    EndUserRoleNotFoundError
  | EndUserImplicitRoleProtectedError

# ─── Query & Mutation ─────────────────────────────────────────────
extend type Query {
  # 权限点
  endUserPermissions(projectSlug: String!): [EndUserPermission!]!
  endUserPermissionsByModel(projectSlug: String!, modelId: String!): [EndUserPermission!]!

  # 权限包
  endUserBundles(projectSlug: String!): [EndUserPermissionBundle!]!
  endUserBundle(projectSlug: String!, id: ID!): EndUserPermissionBundle

  # 角色
  endUserRoles(projectSlug: String!): [EndUserRole!]!

  # 鉴权（供 Runtime 层调用或管理界面展示）
  endUserEffectivePermissions(projectSlug: String!, userId: String!): [EffectivePermission!]!
}

extend type Mutation {
  createEndUserPermission(input: CreateEndUserPermissionInput!): CreateEndUserPermissionPayload!
  updateEndUserPermission(input: UpdateEndUserPermissionInput!): UpdateEndUserPermissionPayload!
  deleteEndUserPermission(projectSlug: String!, id: ID!): DeleteEndUserPermissionPayload!

  createEndUserBundle(input: CreateEndUserBundleInput!): CreateEndUserBundlePayload!
  updateEndUserBundle(input: UpdateEndUserBundleInput!): UpdateEndUserBundlePayload!
  addPermissionToBundle(input: AddPermissionToBundleInput!): AddPermissionToBundlePayload!
  removePermissionFromBundle(input: RemovePermissionFromBundleInput!): RemovePermissionFromBundlePayload!
  deleteEndUserBundle(projectSlug: String!, id: ID!): DeleteEndUserBundlePayload!

  createEndUserRole(input: CreateEndUserRoleInput!): CreateEndUserRolePayload!
  updateEndUserRole(input: UpdateEndUserRoleInput!): UpdateEndUserRolePayload!
  deleteEndUserRole(projectSlug: String!, id: ID!): DeleteEndUserRolePayload!
  assignBundleToRole(input: AssignBundleToRoleInput!): AssignBundleToRolePayload!
  revokeBundleFromRole(input: RevokeBundleFromRoleInput!): RevokeBundleFromRolePayload!

  grantBundleToUser(input: GrantBundleToUserInput!): GrantBundleToUserPayload!
  revokeBundleFromUser(input: RevokeBundleFromUserInput!): RevokeBundleFromUserPayload!
  assignRoleToUser(input: AssignRoleToUserInput!): AssignRoleToUserPayload!
  revokeRoleFromUser(input: RevokeRoleFromUserInput!): RevokeRoleFromUserPayload!
}
```

#### 2.3.2 代码生成

修改 `rbac.graphql` 后运行：

```bash
just generate-gql
```

> ⚠️ **严禁运行 `just clean-gql`**（会删除已实现的 resolver 代码）。  
> ⚠️ **严禁直接编辑 `internal/interfaces/graphql/generated/`**（自动生成，手动修改会被覆盖）。

#### 2.3.3 Resolver 实现

文件：`internal/interfaces/graphql/project/rbac.resolvers.go`

规范（来自 `ai-metadata/backend/development/architecture.md`）：
- 从 `ctx` 提取 `orgName`：`ctxutils.GetOrgName(ctx)`
- 错误转换前必须记录：`logfacade.Stack(ctx, err)`
- `*BusinessError` 通过 `adapter/rbac_adapter.go` 转为 GraphQL 联合错误类型

```go
// rbacEffectivePermissions Resolver 示例
func (r *queryResolver) RbacEffectivePermissions(
    ctx context.Context,
    projectSlug string,
    userID string,
) ([]*model.EffectivePermission, error) {
    orgName := ctxutils.GetOrgName(ctx)

    eps, err := r.rbacAuthzSvc.GetEffectivePermissions(ctx, rbacapp.GetEffectivePermissionsQuery{
        UserID:      userID,
        OrgName:     orgName,
        ProjectSlug: projectSlug,
    })
    if err != nil {
        logfacade.Stack(ctx, err)
        return nil, adapter.ToGraphQLError(ctx, err)
    }

    return adapter.ToEffectivePermissionsDTO(eps), nil
}

// createEndUserPermission Mutation Resolver 示例（错误联合类型模式）
func (r *mutationResolver) CreateEndUserPermission(
    ctx context.Context,
    input model.CreateEndUserPermissionInput,
) (*model.CreateEndUserPermissionPayload, error) {
    orgName := ctxutils.GetOrgName(ctx)

    perm, err := r.rbacPermSvc.CreatePermission(ctx, rbacapp.CreatePermissionCommand{
        OrgName:      orgName,
        ProjectSlug:  input.ProjectSlug,
        ModelID:      input.ModelID,
        Name:         input.Name,
        Description:  input.Description,
        Action:       rbac.Action(input.Action),
        RowScope:     rbac.RowScope(input.RowScope),
        ColumnPolicy: adapter.ToColumnPolicyDomain(input.ColumnPolicy),
    })
    if err != nil {
        logfacade.Stack(ctx, err)
        return &model.CreateEndUserPermissionPayload{
            Error: adapter.ToCreateEndUserPermissionError(ctx, err),
        }, nil
    }

    return &model.CreateEndUserPermissionPayload{
        Permission: adapter.ToEndUserPermissionDTO(perm),
    }, nil
}
```

---

### Wave 4 — 鉴权集成

**目标**：Runtime 查询时调用鉴权服务，将 `rowScope` 传递给 RLS 引擎

#### 2.4.1 Runtime 鉴权中间件接入

**位置**：`internal/interfaces/runtime/handler.go`（在现有 Runtime 请求处理链中插入）

```go
// 在 Runtime 请求处理链中注入 RBAC 鉴权
func (h *RuntimeHandler) handleWithRBAC(ctx context.Context, req RuntimeRequest) error {
    // 1. 从 JWT 提取终端用户 ID
    endUserID := ctxutils.GetEndUserID(ctx)

    // 2. 5 步鉴权，获取有效权限集
    eps, err := h.rbacAuthzSvc.GetEffectivePermissions(ctx, rbacapp.GetEffectivePermissionsQuery{
        UserID:      endUserID,
        OrgName:     req.OrgName,
        ProjectSlug: req.ProjectSlug,
    })
    if err != nil {
        return err
    }

    // 3. 将 HTTP Method → RBAC Action
    action := mapHTTPMethodToRbacAction(req.Method)  // GET→SELECT, POST→INSERT, etc.

    // 4. 默认拒绝：无权限则返回 403
    if !eps.HasPermission(req.ModelID, action) {
        return bizerrors.NewForbiddenError("rbac.forbidden", "permission denied for model: "+req.ModelID)
    }

    // 5. 提取 rowScope，注入 ctx 供 RLS 引擎读取
    perm := eps.GetPermission(req.ModelID, action)
    ctx = ctxutils.WithRbacContext(ctx, RbacContext{
        RowScope:     perm.RowScope,
        ColumnPolicy: perm.ColumnPolicy,
        EndUserID:    endUserID,
    })

    // 6. 继续执行查询（RLS 引擎从 ctx 读取 RbacContext）
    return h.queryExecutor.Execute(ctx, req)
}
```

#### 2.4.2 rowScope → RLS WHERE 子句转换接口

**位置**：`internal/infrastructure/database/dml/`（或现有 RLS 处理逻辑旁）

`rowScope` 到参数化 SQL WHERE 谓词的转换：

```go
// RowScopeToWhereClause 将 RBAC rowScope 转换为参数化 WHERE 子句
// 调用时机：Runtime DML 查询构建 SQL 前（与现有 RLS 策略编译器并行或复用）
func RowScopeToWhereClause(
    ctx context.Context,
    scope rbac.RowScope,
    endUserID string,
) (clause string, params []interface{}, err error) {
    switch scope {
    case rbac.RowScopeAll:
        return "", nil, nil  // 不注入 WHERE，查全表

    case rbac.RowScopeSelf:
        // 要求 Model 有 owner（EndUserRef）字段（创建权限点时已校验）
        return "owner = ?", []interface{}{endUserID}, nil

    case rbac.RowScopeDept:
        deptID := ctxutils.GetEndUserDeptID(ctx)  // 从会话提取当前用户部门 ID
        if deptID == "" {
            return "", nil, errors.New("dept_id not found in user session")
        }
        return "dept_id = ?", []interface{}{deptID}, nil

    case rbac.RowScopeDeptAndChildren:
        deptID := ctxutils.GetEndUserDeptID(ctx)
        if deptID == "" {
            return "", nil, errors.New("dept_id not found in user session")
        }
        // 递归查询子部门 ID（由 DeptService 提供，结果可缓存）
        deptIDs, err := h.deptSvc.GetDeptAndChildrenIDs(ctx, deptID)
        if err != nil {
            return "", nil, fmt.Errorf("failed to fetch dept tree: %w", err)
        }
        placeholders := strings.Repeat("?,", len(deptIDs))
        clause := "dept_id IN (" + strings.TrimRight(placeholders, ",") + ")"
        params := make([]interface{}, len(deptIDs))
        for i, id := range deptIDs {
            params[i] = id
        }
        return clause, params, nil
    }
    return "", nil, fmt.Errorf("unknown rowScope: %s", scope)
}
```

---

## 3. 关键接口定义（完整签名）

### 3.1 `EndUserPermissionRepository` 接口核心方法

```go
// EndUserPermissionRepository RBAC 权限系统仓储接口（Project 维度）
// 全量接口定义见 internal/domain/rbac/repository.go
type EndUserPermissionRepository interface {
    // 权限点
    CreatePermission(ctx context.Context, p *EndUserPermission) error
    GetPermissionByID(ctx context.Context, orgName, id string) (*EndUserPermission, error)
    ListPermissionsByModel(ctx context.Context, orgName, modelID string) ([]*EndUserPermission, error)
    UpdatePermission(ctx context.Context, p *EndUserPermission) error
    DeletePermission(ctx context.Context, orgName, id string) error

    // 权限包
    CreateBundle(ctx context.Context, b *EndUserPermissionBundle) error
    GetBundleByID(ctx context.Context, orgName, id string) (*EndUserPermissionBundle, error)
    ListBundlesByProject(ctx context.Context, orgName, projectSlug string) ([]*EndUserPermissionBundle, error)
    AddPermissionToBundle(ctx context.Context, bundleID, permissionID string, sortOrder int) error
    DeleteBundle(ctx context.Context, orgName, id string) error

    // 业务角色
    CreateRole(ctx context.Context, r *EndUserRole) error
    GetRoleByID(ctx context.Context, orgName, id string) (*EndUserRole, error)
    UpdateRole(ctx context.Context, r *EndUserRole) error
    DeleteRole(ctx context.Context, orgName, id string) error
    AssignBundleToRole(ctx context.Context, roleID, bundleID string) error
    RevokeBundleFromRole(ctx context.Context, roleID, bundleID string) error

    // 用户授权
    GrantBundleToUser(ctx context.Context, userID, orgName, projectSlug, bundleID string) error
    AssignRoleToUser(ctx context.Context, userID, orgName, projectSlug, roleID string) error
    RevokeRoleFromUser(ctx context.Context, userID, orgName, projectSlug, roleID string) error

    // ⚡ 鉴权核心：3 条链式查询（Step 1~3）
    GetBundleIDsByUserDirect(ctx context.Context, userID, orgName, projectSlug string) ([]string, error)
    GetBundleIDsByUserExplicitRoles(ctx context.Context, userID, orgName, projectSlug string) ([]string, error)
    GetBundleIDsByImplicitRoles(ctx context.Context, orgName, projectSlug string) ([]string, error)

    // ⚡ 鉴权核心：Step 4 展开
    GetPermissionsByBundleIDs(ctx context.Context, orgName string, bundleIDs []string) ([]*EndUserPermission, error)
}
```

### 3.2 `EndUserAuthzService.GetEffectivePermissions()` 完整签名

```go
// GetEffectivePermissionsQuery 鉴权查询参数
type GetEffectivePermissionsQuery struct {
    UserID      string  // 终端用户 ID（来自 JWT sub 字段）
    OrgName     string  // 组织名称（多租户隔离，来自 URL path）
    ProjectSlug string  // 项目标识符（Project 维度鉴权）
}

// GetEffectivePermissions 获取用户在指定 Project 下的有效权限集
//
// 执行 5 步鉴权流程（prd/rbac/03-auth-flow.md）：
//   Step 1: GetBundleIDsByUserDirect    → 直接授权包 ID
//   Step 2: GetBundleIDsByUserExplicitRoles → 显式角色包 ID（单次 JOIN）
//   Step 3: GetBundleIDsByImplicitRoles → 隐式角色包 ID（对所有认证用户执行）
//   Step 4: GetPermissionsByBundleIDs   → 展开权限点（动态 IN）
//   Step 5: EffectivePermissionSet.Merge → 合并取并集（rowScope 取最宽泛）
//
// 返回值约定：
//   - 空权限集（无任何授权）：返回 (EffectivePermissionSet{}, nil)，不报错
//   - 数据库查询失败：返回 (nil, error)
//   - 调用方通过 eps.HasPermission(modelID, action) 判定，默认拒绝
func (s *EndUserAuthzService) GetEffectivePermissions(
    ctx context.Context,
    q GetEffectivePermissionsQuery,
) (rbac.EffectivePermissionSet, error)
```

---

## 4. 验收标准

### 4.1 单元测试覆盖场景

**`internal/domain/rbac/`**

| 测试文件 | 核心场景 |
|----------|----------|
| `permission_test.go` | Validate() — ① 缺 modelID ② 缺 name ③ 非法 action ④ 非法 rowScope ⑤ columnPolicy 含非法 defaultMode ⑥ 全部合法 |
| `column_policy_test.go` | mergeColumnPolicy() — ① 两个 nil 返回 nil ② 一方 nil 返回另一方 ③ 同字段 MASKED∪VISIBLE=VISIBLE ④ DefaultMode HIDDEN∪READONLY=READONLY |
| `row_scope_test.go` | MergeRowScope() — ① SELF∪ALL=ALL ② DEPT∪DEPT_AND_CHILDREN=DEPT_AND_CHILDREN ③ 相同值=原值 |
| `role_test.go` | GuardDelete() — ① is_implicit=true 返回错误 ② is_implicit=false 返回 nil；GuardUpdate() 同上 |
| `authz_test.go` | Merge() — ① 单条权限 ② 同 model+action 多条 rowScope 取最大 ③ ColumnPolicy 合并取更宽泛 ④ 不同 model 不互相影响；HasPermission() ⑤ 命中 ⑥ 未命中（默认拒绝） |

**`internal/app/rbac/`**

| 测试文件 | 核心场景 |
|----------|----------|
| `permission_app_test.go` | ① rowScope=SELF，Model 有 owner → 成功 ② rowScope=SELF，Model 无 owner → ValidationError ③ rowScope=ALL → 不检查字段 ④ rowScope=DEPT，无 dept_id → ValidationError |
| `role_app_test.go` | ① 删除隐式角色 → 返回 ImplicitProtectedError ② 删除普通角色 → 成功 ③ 更新隐式角色 name → ImplicitNameImmutableError |
| `authz_app_test.go` | ① 三通道均无授权 → 返回空集 ② 仅直接授权 → 包含 Step 1 权限 ③ 直接+显式角色有重叠 bundle → 去重后 Step 4 不重复展开 ④ 隐式角色自动注入 → 无 user_roles 记录仍获得隐式权限 ⑤ 多通道 rowScope=SELF∪ALL → 有效结果=ALL |

**`internal/infrastructure/repository/`**

| 测试场景 | 验证内容 |
|----------|---------|
| CreatePermission → GetPermissionByID | 持久化 + 读取一致性，含 ColumnPolicy JSON 序列化 |
| DeleteRole(implicit=true) 被业务层阻断 | GuardDelete() 在 App 层调用，SQL 层 `AND is_implicit=FALSE` 兜底 |
| GetPermissionsByBundleIDs(空 bundleIDs) | 直接返回空 slice，不发 SQL 查询（短路优化） |
| 重复 GrantBundleToUser | 返回 `ErrTypeDuplicatedKey`（UK 约束） |

### 4.2 BDD 场景纲要（中文 Gherkin）

文件位置：`tests-bdd/features/rbac/`

> BDD 测试遵循 `ai-metadata/backend/testing/bdd-testing-guidelines.md`：  
> **认证优先级 TOKEN > 登录 > 注册**，固定 `TEST_ORG_NAME` + `TEST_PROJECT_SLUG`，不耦合注册流程。

**`permission.feature`**

```gherkin
功能: RBAC 权限点管理

  场景: 成功创建 SELECT ALL 权限点
    假如 项目 "demo" 下存在模型 "orders"
    当 设计者创建权限点 "订单查看-全部"，动作=SELECT，行策略=ALL
    那么 权限点创建成功，响应包含权限点 ID

  场景: 创建 SELF 权限点但 Model 无 owner 字段
    假如 模型 "orders" 没有 EndUserRef 类型字段
    当 设计者尝试创建权限点，动作=SELECT，行策略=SELF
    那么 返回错误码 "rbac.permission.missing_owner_field"

  场景: 创建 DEPT 权限点但 Model 无 dept_id 字段
    假如 模型 "orders" 没有 dept_id 字段
    当 设计者尝试创建权限点，动作=SELECT，行策略=DEPT
    那么 返回错误码 "rbac.permission.missing_dept_field"

  场景: 删除被权限包引用的权限点后自动解除关联
    假如 权限点 "订单查看-本人" 已加入权限包 "订单查看包"
    当 设计者删除权限点 "订单查看-本人"
    那么 删除成功
    并且 权限包 "订单查看包" 的权限点列表不再包含该权限点
```

**`bundle.feature`**

```gherkin
功能: RBAC 权限包管理

  场景: 成功创建权限包并有序添加权限点
    假如 权限点 "订单查看-本人"（排序=1）和 "订单查看-全部"（排序=2）已存在
    当 设计者创建权限包 "订单查看包"
    并且 按顺序将两个权限点添加到包
    那么 ListPermissionsInBundle 返回 2 个权限点，顺序正确

  场景: 删除权限包后关联用户授权自动清除
    假如 用户 "alice" 直接关联权限包 "订单查看包"
    当 设计者删除权限包 "订单查看包"
    那么 删除成功
    并且 用户 "alice" 对该包的直接授权自动清除（级联删除）
```

**`role.feature`**

```gherkin
功能: RBAC 业务角色管理

  场景: 成功创建普通角色并授权权限包
    当 设计者创建角色 "销售专员"，is_implicit=false
    并且 将权限包 "订单查看包" 授予角色 "销售专员"
    那么 ListBundlesByRole 返回 1 个权限包

  场景: 隐式角色不可删除
    假如 项目 "demo" 存在隐式角色 "SYSTEM_AUTHENTICATED_USER"
    当 设计者尝试删除该角色
    那么 返回错误码 "rbac.role.implicit_protected"

  场景: 隐式角色名称不可更新
    假如 项目 "demo" 存在隐式角色 "SYSTEM_AUTHENTICATED_USER"
    当 设计者尝试将角色名称修改为 "OTHER_NAME"
    那么 返回错误码 "rbac.role.implicit_name_immutable"

  场景: 隐式角色的权限包配置可以更新
    假如 项目 "demo" 存在隐式角色 "SYSTEM_AUTHENTICATED_USER"
    当 设计者将权限包 "基础自服务包" 授予该隐式角色
    那么 授权成功
```

**`authz.feature`**

```gherkin
功能: RBAC 五步鉴权流程

  场景: 用户通过直接授权获得权限（Step 1）
    假如 用户 "alice" 直接关联权限包 "订单查看包"
    并且 "订单查看包" 包含权限点 "orders:SELECT rowScope=ALL"
    当 查询 "alice" 对项目 "demo" 的有效权限集
    那么 有效权限集包含 modelId="orders"，action=SELECT，rowScope=ALL

  场景: 用户通过显式角色获得权限（Step 2）
    假如 用户 "bob" 关联角色 "销售专员"
    并且 角色 "销售专员" 关联权限包 "订单管理包"
    当 查询 "bob" 的有效权限集
    那么 有效权限集包含订单管理包内所有权限点

  场景: 隐式角色自动注入（Step 3）
    假如 隐式角色 "SYSTEM_AUTHENTICATED_USER" 关联权限包 "基础自服务包"
    并且 用户 "charlie" 没有任何直接授权或显式角色
    当 查询 "charlie" 的有效权限集
    那么 有效权限集包含基础自服务包内的所有权限点（无需任何显式关联）

  场景: 多通道 rowScope 取并集最宽泛（SELF ∪ ALL = ALL）
    假如 用户 "diana" 通过直接授权获得 "orders:SELECT rowScope=SELF"
    并且 通过隐式角色获得 "orders:SELECT rowScope=ALL"
    当 查询 "diana" 对 orders:SELECT 的有效行策略
    那么 有效 rowScope 为 "ALL"（取最宽泛范围）

  场景: 无任何授权时默认拒绝
    假如 用户 "eve" 没有任何直接授权或角色关联
    并且 项目 "demo" 的隐式角色没有关联任何权限包
    当 检查 "eve" 是否有 "orders:SELECT" 权限
    那么 HasPermission 返回 false（默认拒绝原则）

  场景: 多 bundle 展开去重不重复计入
    假如 用户 "frank" 直接关联权限包 "A"
    并且 "frank" 的显式角色也关联了权限包 "A"（重叠）
    当 获取有效权限集
    那么 权限包 "A" 只被展开一次（bundle ID 去重后再 Step 4）
```

---

## 5. 实现顺序与依赖关系

### 5.1 依赖关系图

```
Wave 1a ─ DB Schema (13_rbac_permissions.sql)
    │
    ├── Wave 1b ─ sqlc 查询文件 (db/queries/rbac/*.sql)
    │                 │
    │                 └── [just generate-sqlc]  ← 阻塞点①
    │                             │
    │           ┌─────────────────┘
    │           │
    ├── Wave 1c ─ Domain: 领域实体 + repository.go 接口
    │   (可与 1b 并行，接口不依赖 DB 生成)
    │           │
    │           ├── Wave 1d ─ Infrastructure: sql_end_user_permission_repository.go
    │           │             (依赖 1b 生成代码 + 1c 接口)
    │           │
    │           └── Wave 2a ─ App Service: commands.go + 4 个 *_app.go
    │                         (依赖 1c 接口，不依赖 1b/1d)
    │                             │
    │                             └── Wave 3a ─ GraphQL Schema (rbac.graphql)
    │                                               │
    │                                               └── [just generate-gql]  ← 阻塞点②
    │                                                           │
    │                                                           └── Wave 3b ─ rbac.resolvers.go
    │                                                                               │
    │                                                                               └── Wave 4 ─ Runtime 鉴权集成
```

### 5.2 并行开发机会

| 并行组 | 可同时进行的任务 | 说明 |
|--------|----------------|------|
| **组 A** | Wave 1a（DB Schema）+ Wave 1c 实体部分（permission.go / column_policy.go / bundle.go / role.go / authz.go） | 实体定义纯 Go 代码，不依赖 DB |
| **组 B（依赖 1a）** | Wave 1b（sqlc 文件）+ Wave 1c（repository.go 接口） | 接口不依赖生成代码 |
| **组 C（依赖 1b 生成完成）** | Wave 1d（Infrastructure 实现）+ Wave 2a（App Service） | App Service 依赖接口，不依赖 Infrastructure 实现 |
| **组 D（依赖 2a）** | Wave 3a（Schema 设计）+ Wave 4 接入点规划 | Schema 完成后运行 gql codegen，再实现 3b |

### 5.3 关键阻塞点

| 阻塞点 | 阻塞原因 | 解除条件 |
|--------|---------|---------|
| **① sqlc codegen** | `just generate-sqlc` 依赖所有 `db/queries/rbac/*.sql` 完成 | Wave 1b 全部 6 个 sql 文件完成后执行一次 |
| **② gql codegen** | `just generate-gql` 依赖 `rbac.graphql` 中的类型/接口完整定义 | Wave 3a 完成后执行，生成 Resolver 骨架 |
| **③ Infrastructure 编译** | `sql_end_user_permission_repository.go` 依赖 `dbgen.Querier` 包含 RBAC 相关方法 | 阻塞点①解除后 |
| **④ Runtime 鉴权集成** | 依赖 `EndUserAuthzService`（Wave 2a）+ Runtime 框架支持 ctx 注入 | Wave 2a 完成后可开始 Wave 4 |
| **⑤ BDD 测试** | 需要 Resolver 可调用 + 测试环境初始化（隐式角色种子数据） | Wave 3b 完成后编写，固定 TEST_PROJECT_SLUG 避免耦合注册 |

### 5.4 建议单人顺序交付（7 天）

| 天 | 完成内容 |
|----|---------|
| Day 1 | Wave 1a（13_rbac_permissions.sql）+ Wave 1c 实体（permission.go / column_policy.go / bundle.go / role.go / authz.go / row_scope.go）|
| Day 2 | Wave 1b（db/queries/rbac/ 6 个 sql 文件）+ Wave 1c（repository.go 接口）+ `just generate-sqlc` |
| Day 3 | Wave 1d（sql_end_user_permission_repository.go 实现）+ 单元测试（repository 层）|
| Day 4 | Wave 2a（commands.go + 4 个 App Service）+ 单元测试（App 层：rowScope 校验、隐式角色保护、5 步鉴权）|
| Day 5 | Wave 3a（rbac.graphql）+ `just generate-gql` + Wave 3b（rbac.resolvers.go + adapter）|
| Day 6 | Wave 4（Runtime 鉴权中间件 + rowScope → WHERE 转换接口）|
| Day 7 | BDD 场景编写 + 集成联调 + 隐式角色种子数据初始化脚本 |

---

## 附录：关键文件路径速查

| 文件/目录 | 说明 |
|-----------|------|
| `db/schema/mysql/13_rbac_permissions.sql` | DB 迁移（5 张新表 + ALTER end_user_roles，完整 DDL 见 `plans/rbac/db-schema-final.md`） |
| `db/queries/rbac/authz.sql` | 鉴权核心 3 条链式查询 + Step 4 展开查询（表名用 `end_user_` 前缀） |
| `internal/domain/rbac/repository.go` | `EndUserPermissionRepository` 接口（单一接口文件） |
| `internal/domain/rbac/authz.go` | `EffectivePermissionSet` 核心值对象（Merge / HasPermission）|
| `internal/domain/rbac/row_scope.go` | `RowScope` 枚举 + `MergeRowScope()`（并集最大范围）|
| `internal/domain/rbac/column_policy.go` | `ColumnAccessMode` 枚举 + `ColumnRule` + `ColumnPolicy` 结构体（对齐 API 合约）|
| `internal/infrastructure/repository/sql_end_user_permission_repository.go` | Infrastructure 实现（sqlc wrapper，接收 `dbgen.Querier`）|
| `internal/app/rbac/authz_app.go` | 5 步鉴权编排（`EndUserAuthzService`）|
| `internal/app/rbac/permission_app.go` | 权限点 CRUD + rowScope 字段前提校验（`EndUserPermissionAppService`）|
| `internal/app/rbac/bundle_app.go` | 权限包 CRUD（`EndUserBundleAppService`）|
| `internal/app/rbac/role_app.go` | 角色 CRUD + 隐式角色保护（`EndUserRoleAppService`，GuardDelete / GuardUpdate）|
| `api/graph/project/schema/rbac.graphql` | GraphQL Schema（含 `ColumnAccessMode` 枚举，extend type Query/Mutation）|
| `internal/interfaces/graphql/project/rbac.resolvers.go` | GraphQL Resolver（gql codegen 骨架 + 实现）|
| `internal/interfaces/graphql/project/adapter/rbac_adapter.go` | 实体 → DTO 映射 + `BusinessError` → 联合错误类型转换 |
| `plans/rbac/db-schema-final.md` | DB Schema 终态参考（含完整 DDL，Wave 1a 直接复制）|
| `plans/rbac/api-contract.md` | GraphQL API Contract 完整参考（Wave 3a 参考依据）|
| `ai-metadata/prd/rbac/03-auth-flow.md` | 鉴权流程 PRD（5 步，Wave 2a 设计依据）|
| `ai-metadata/backend/development/repo-develop.md` | Repository 层开发规范（错误处理模式、Go Wrapper 架构）|
| `ai-metadata/backend/development/domain-development.md` | Domain 层接口设计规范（ctx/orgName/projectSlug 规则）|
