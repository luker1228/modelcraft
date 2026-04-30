# ModelCraft Permission Bundle 数据权限配置文档

## 目录
1. [GraphQL Schema 定义](#graphql-schema-定义)
2. [数据库 Schema](#数据库-schema)
3. [SQLC 查询](#sqlc-查询)
4. [关键字段映射](#关键字段映射)
5. [数据权限配置项结构](#数据权限配置项结构)

---

## GraphQL Schema 定义

### 文件位置
- **后端**: `/Users/luke/my_projects/modelcraft/modelcraft-backend/api/graph/project/schema/rbac.graphql`
- **前端**: `/Users/luke/my_projects/modelcraft/modelcraft-front/contract/graph/org/schema/permission.graphql`

### 核心类型定义

#### 1. **EndUserPermissionBundle** (权限包)
```graphql
type EndUserPermissionBundle implements Node {
  id: ID!
  
  # URL友好的标识符，同项目内唯一，创建时由用户指定或从名称自动派生，之后不可修改
  slug: String!
  
  name: String!
  description: String
  
  # Item-centric 数据权限列表：每个模型最多一个 item
  dataPermissionItems: [EndUserBundleDataPermissionItem!]!
  
  # 兼容旧字段（将逐步废弃），从 item 导出的 permission 视图
  permissions: [EndUserBundlePermissionEntry!]!
  
  # 当前版本号（每次权限列表变更后递增）。初始创建时为 0，首次修改后变为 1
  currentVersion: Int!
  
  # 最近历史快照列表（最多 5 个，按 version DESC 排列）
  snapshots: [EndUserPermissionBundleSnapshot!]!
  
  createdAt: Time!
  updatedAt: Time!
}
```

#### 2. **EndUserBundleDataPermissionItem** (数据权限配置项) ⭐ 核心
```graphql
type EndUserBundleDataPermissionItem {
  id: ID!
  bundleId: ID!
  modelId: ID!
  
  # 来源类型：PRESET（预设模板）或 CUSTOM（管理员自定义）
  grantType: DataPermissionGrantType!
  
  # 当 grantType=PRESET 时非空
  preset: EndUserPermissionPreset
  
  # 当 grantType=CUSTOM 时非空
  customPermissionId: ID
  
  # 当 grantType=CUSTOM 时，引用的自定义权限点摘要
  customPermission: EndUserPermission
  
  sortOrder: Int!
  createdAt: Time!
  updatedAt: Time!
}
```

#### 3. **数据权限预设类型** (Preset Enum)
```graphql
enum EndUserPermissionPreset {
  READ_WRITE_ALL          # 读写全部（不依赖 END_USER_REF 字段）
                          # SELECT ALL + INSERT ALL + UPDATE ALL + DELETE ALL
  
  READ_ALL                # 只读全部（不依赖 END_USER_REF 字段）
                          # SELECT ALL
  
  READ_WRITE_OWNER        # 读写自己（依赖 END_USER_REF 字段）
                          # SELECT SELF + INSERT SELF + UPDATE SELF + DELETE SELF
  
  READ_ALL_WRITE_OWNER    # 读所有写自己（依赖 END_USER_REF 字段）
                          # SELECT ALL + INSERT SELF + UPDATE SELF + DELETE SELF
}
```

#### 4. **数据权限来源类型** (GrantType Enum)
```graphql
enum DataPermissionGrantType {
  PRESET    # 预设模板
  CUSTOM    # 管理员自定义
}
```

#### 5. **自定义权限点** (EndUserPermission)
```graphql
type EndUserPermission implements Node {
  id: ID!
  modelId: ID!
  databaseName: String
  modelName: String
  action: RbacAction!                # SELECT, INSERT, UPDATE, DELETE, EXPORT
  columnPolicy: ColumnPolicy!        # 列级访问策略
  rowScope: RowScopeType!            # ALL, SELF, DEPT, DEPT_AND_CHILDREN
  
  # 来源预设，null 表示手动创建的自定义权限点
  preset: EndUserPermissionPreset
  
  displayName: String
  description: String
  createdAt: Time!
  updatedAt: Time!
}
```

#### 6. **权限包快照** (EndUserPermissionBundleSnapshot)
```graphql
type EndUserPermissionBundleSnapshot {
  version:      Int!
  createdAt:    Time!
  createdBy:    String
  
  # 若为回滚操作，指向来源版本号
  restoredFrom: Int
  
  # Item-centric 快照条目列表
  items: [EndUserPermissionSnapshotItemEntry!]!
  
  # 兼容旧字段
  permissions:  [EndUserPermissionSnapshotEntry!]!
}
```

---

## 数据库 Schema

### 核心表结构

#### 1. **end_user_permission_bundles** - 权限包表
```sql
CREATE TABLE `end_user_permission_bundles` (
  `id`            VARCHAR(36)   PRIMARY KEY      -- 权限包 UUID
  `slug`          VARCHAR(64)   NOT NULL UNIQUE  -- URL友好标识符（同项目内唯一）
  `org_name`      VARCHAR(64)   NOT NULL         -- 所属组织
  `project_slug`  VARCHAR(64)   NOT NULL         -- 所属项目
  `name`          VARCHAR(128)  NOT NULL UNIQUE  -- 权限包名称（同项目内唯一）
  `description`   TEXT          NULL             -- 权限包描述
  `created_at`    DATETIME      DEFAULT NOW()
  `updated_at`    DATETIME      DEFAULT NOW()
)
UNIQUE KEY `uq_bundles_org_project_slug` (`org_name`, `project_slug`, `slug`)
UNIQUE KEY `uq_bundles_org_project_name` (`org_name`, `project_slug`, `name`)
INDEX `idx_bundles_org_project` (`org_name`, `project_slug`)
```

#### 2. **end_user_bundle_data_permission_items** - Bundle数据权限配置项表 ⭐ 核心
```sql
CREATE TABLE `end_user_bundle_data_permission_items` (
  `id`                   VARCHAR(36)  PRIMARY KEY              -- Item UUID
  `bundle_id`            VARCHAR(36)  NOT NULL FK              -- 所属权限包
  `model_id`             VARCHAR(36)  NOT NULL FK              -- 目标模型
  
  # 核心字段：来源类型
  `grant_type`           ENUM('PRESET','CUSTOM') NOT NULL      -- 授权来源类型
  
  # 与 grant_type 互斥的字段
  `preset`               ENUM(
                           'READ_WRITE_ALL',
                           'READ_ALL',
                           'READ_WRITE_OWNER',
                           'READ_ALL_WRITE_OWNER'
                         ) NULL DEFAULT NULL                   -- PRESET值（仅grant_type=PRESET时有值）
  
  `custom_permission_id` VARCHAR(36)  NULL DEFAULT NULL        -- 自定义权限实体ID（仅grant_type=CUSTOM时有值）
  
  `sort_order`           INT NOT NULL DEFAULT 0                -- 显示排序权重
  `created_at`           DATETIME DEFAULT NOW()
  `updated_at`           DATETIME DEFAULT NOW()
)
UNIQUE KEY `uq_bundle_items_bundle_model` (`bundle_id`, `model_id`)  -- 同一bundle下同一model最多一个item
CHECK (grant_type != 'PRESET' OR (preset IS NOT NULL AND custom_permission_id IS NULL))
CHECK (grant_type != 'CUSTOM' OR (custom_permission_id IS NOT NULL AND preset IS NULL))
```

#### 3. **end_user_data_permissions** - 自定义权限实体表
```sql
CREATE TABLE `end_user_data_permissions` (
  `id`            VARCHAR(36)  PRIMARY KEY       -- 权限实体UUID
  `org_name`      VARCHAR(64)  NOT NULL
  `project_slug`  VARCHAR(64)  NOT NULL
  `database_name` VARCHAR(128) NULL              -- 数据源名称
  `model_name`    VARCHAR(128) NULL              -- 模型名称
  `model_id`      VARCHAR(36)  NOT NULL FK       -- 关联模型ID
  `name`          VARCHAR(128) NOT NULL          -- 权限名称
  `description`   TEXT         NULL              -- 权限描述
  
  # JSON策略字段
  `column_policy` JSON         NULL              -- 列策略JSON
  # 结构: {
  #   "defaultMode": "VISIBLE|READONLY|MASKED|HIDDEN",
  #   "rules": [
  #     { "fieldName": "salary", "mode": "MASKED", "maskPattern": "***" },
  #     { "fieldName": "id_card", "mode": "HIDDEN" }
  #   ]
  # }
  
  `row_policy`    JSON         NULL              -- 行策略JSON
  # 结构: {
  #   "select": { "allowed": true, "scope": "custom", "predicate": {...} },
  #   "insert": { "allowed": true, "scope": "custom", "check": {...} },
  #   "update": { "allowed": true, "scope": "custom", "predicate": {...}, "check_scope": "custom", "check": {...} },
  #   "delete": { "allowed": false }
  # }
  
  `created_at`    DATETIME DEFAULT NOW()
  `updated_at`    DATETIME DEFAULT NOW()
)
UNIQUE KEY `uq_permissions_model_name` (`model_id`, `name`)
```

#### 4. **end_user_permission_bundle_snapshots** - 权限包历史快照表
```sql
CREATE TABLE `end_user_permission_bundle_snapshots` (
  `id`            VARCHAR(36)  PRIMARY KEY       -- 快照UUID
  `bundle_id`     VARCHAR(36)  NOT NULL FK       -- 所属权限包ID
  `version`       INT          NOT NULL          -- 版本号（从1开始）
  
  # JSON结构：快照时刻的data permission item列表
  `items`         JSON         NOT NULL
  # 结构: [
  #   {
  #     "modelId": "uuid",
  #     "grantType": "PRESET|CUSTOM",
  #     "preset": "READ_ALL|...",  // PRESET值或null
  #     "customPermissionId": "uuid|null",
  #     "sortOrder": 0
  #   }
  # ]
  
  `created_at`    DATETIME DEFAULT NOW()
  `created_by`    VARCHAR(128) NULL              -- 操作人标识
  `restored_from` INT          NULL              -- 若为回滚操作，指向来源版本号
)
UNIQUE KEY `uq_bundle_snapshots_bundle_version` (`bundle_id`, `version`)
INDEX `idx_bundle_snapshots_bundle_version_desc` (`bundle_id`, `version` DESC)
```

#### 5. **end_user_role_bundles** - 角色-权限包关联表
```sql
CREATE TABLE `end_user_role_bundles` (
  `id`            VARCHAR(36)  PRIMARY KEY
  `org_name`      VARCHAR(64)  NOT NULL
  `project_slug`  VARCHAR(64)  NOT NULL
  `role_id`       VARCHAR(36)  NOT NULL FK       -- → end_user_roles.id
  `bundle_id`     VARCHAR(36)  NOT NULL FK       -- → end_user_permission_bundles.id
  `granted_at`    DATETIME DEFAULT NOW()
)
UNIQUE KEY `uq_role_bundles_role_bundle` (`role_id`, `bundle_id`)
```

#### 6. **end_user_user_bundles** - 用户直接授权-权限包关联表
```sql
CREATE TABLE `end_user_user_bundles` (
  `id`            VARCHAR(36)  PRIMARY KEY
  `org_name`      VARCHAR(64)  NOT NULL
  `project_slug`  VARCHAR(64)  NOT NULL
  `user_id`       VARCHAR(36)  NOT NULL          -- 用户ID（复合FK的id部分）
  `bundle_id`     VARCHAR(36)  NOT NULL FK       -- → end_user_permission_bundles.id
  `granted_at`    DATETIME DEFAULT NOW()
)
UNIQUE KEY `uq_user_bundles_org_project_user_bundle` 
  (`org_name`, `project_slug`, `user_id`, `bundle_id`)
```

---

## SQLC 查询

### 文件位置
`/Users/luke/my_projects/modelcraft/modelcraft-backend/db/queries/rbac/`

#### 1. bundle.sql - Bundle相关查询
```sql
-- 创建Bundle
CreateEndUserBundle :exec
  INSERT INTO end_user_permission_bundles 
    (id, slug, org_name, project_slug, name, description)

-- 获取单个Bundle
GetEndUserBundleByID :one
  WHERE id = ? AND org_name = ? AND project_slug = ?

GetEndUserBundleBySlug :one
  WHERE slug = ? AND org_name = ? AND project_slug = ?

-- 列表查询
ListEndUserBundlesByProject :many
  WHERE org_name = ? AND project_slug = ?
  ORDER BY name

-- 更新Bundle
UpdateEndUserBundle :execresult
  UPDATE name, description WHERE id = ?

-- 删除Bundle
DeleteEndUserBundle :execresult
  WHERE id = ?
```

#### 2. Bundle Data Permission Items查询 (核心)
```sql
-- Upsert Item（Replace语义：同一bundle+model最多一个item）
UpsertBundleDataPermissionItem :exec
  INSERT INTO end_user_bundle_data_permission_items
    (id, bundle_id, model_id, grant_type, preset, custom_permission_id, sort_order)
  ON DUPLICATE KEY UPDATE
    grant_type, preset, custom_permission_id, sort_order

-- 删除Item
RemoveBundleDataPermissionItem :execresult
  WHERE bundle_id = ? AND model_id = ?

-- 列表查询 (按sort_order和created_at排序)
ListBundleDataPermissionItems :many
  WHERE bundle_id = ?
  ORDER BY sort_order, created_at

-- 清空Bundle所有Item
ClearBundleDataPermissionItems :exec
  WHERE bundle_id = ?

-- 获取特定Item
GetBundleDataPermissionItemByBundleAndModel :one
  WHERE bundle_id = ? AND model_id = ?
```

#### 3. permission.sql - 自定义权限查询
```sql
-- 创建自定义权限
CreateEndUserPermission :exec
  INSERT INTO end_user_data_permissions
    (id, org_name, project_slug, database_name, model_name, model_id, 
     name, description, column_policy, row_policy)

-- 获取权限
GetEndUserPermissionByID :one
  WHERE id = ? AND org_name = ?

GetEndUserPermissionByModelAndName :one
  WHERE model_id = ? AND org_name = ? AND name = ?

-- 列表查询
ListEndUserPermissionsByProject :many
  WHERE org_name = ? AND project_slug = ?

ListEndUserPermissionsByModel :many
  WHERE model_id = ? AND org_name = ?

-- 检查权限是否被bundle引用
IsPermissionReferencedByBundleItem :one
  SELECT COUNT(*) > 0 FROM end_user_bundle_data_permission_items
  WHERE custom_permission_id = ?
```

#### 4. bundle_snapshot.sql - 快照相关查询
```sql
-- 插入快照
InsertBundleSnapshot :exec
  INSERT INTO end_user_permission_bundle_snapshots
    (id, bundle_id, version, items, created_by, restored_from)

-- 列表查询（最多5个，DESC排序）
ListBundleSnapshots :many
  WHERE bundle_id = ?
  ORDER BY version DESC
  LIMIT 5

-- 获取当前版本号
GetBundleCurrentVersion :one
  SELECT COALESCE(MAX(version), 0) FROM end_user_permission_bundle_snapshots
  WHERE bundle_id = ?

-- 按版本号获取快照
GetBundleSnapshotByVersion :one
  WHERE bundle_id = ? AND version = ?

-- 删除旧快照（只保留最近5个）
DeleteOldBundleSnapshots :exec
  DELETE FROM snapshots WHERE bundle_id = ? 
  AND NOT in (latest 5 by version DESC)
```

---

## 关键字段映射

### 截图中的字段对应

根据描述 `d97f7a43...` 和 `读写全部` 的数据权限配置列表项：

| 字段 | 数据库列 | GraphQL字段 | 说明 |
|------|--------|-----------|------|
| `d97f7a43...` | `id` | `id` | Item UUID |
| `读写全部` | `preset` (ENUM) | `preset` (Enum) | 预设类型：READ_WRITE_ALL |
| 模型关联 | `model_id` | `modelId` | 目标模型ID |
| 授权来源 | `grant_type` | `grantType` | PRESET 或 CUSTOM |
| 排序 | `sort_order` | `sortOrder` | 显示顺序 |
| 时间戳 | `created_at`, `updated_at` | `createdAt`, `updatedAt` | 创建/更新时间 |

### Item的两种模式

#### PRESET Mode（预设）
```
grantType = "PRESET"
preset = "READ_WRITE_ALL" | "READ_ALL" | "READ_WRITE_OWNER" | "READ_ALL_WRITE_OWNER"
customPermissionId = NULL
customPermission = NULL
```

#### CUSTOM Mode（自定义）
```
grantType = "CUSTOM"
preset = NULL
customPermissionId = "uuid-of-permission"
customPermission = {EndUserPermission object with columnPolicy, rowPolicy}
```

---

## 数据权限配置项结构

### 完整Item示例

```json
{
  "id": "d97f7a43-1234-5678-90ab-cdef12345678",
  "bundleId": "bundle-id-12345",
  "modelId": "model-id-67890",
  "grantType": "PRESET",
  "preset": "READ_WRITE_ALL",
  "customPermissionId": null,
  "customPermission": null,
  "sortOrder": 0,
  "createdAt": "2024-04-20T10:30:00Z",
  "updatedAt": "2024-04-20T10:30:00Z"
}
```

### 自定义权限示例

```json
{
  "id": "perm-id-98765",
  "bundleId": "bundle-id-12345",
  "modelId": "model-id-67890",
  "grantType": "CUSTOM",
  "preset": null,
  "customPermissionId": "custom-perm-id",
  "customPermission": {
    "id": "custom-perm-id",
    "modelId": "model-id-67890",
    "action": "SELECT",
    "rowScope": "SELF",
    "columnPolicy": {
      "defaultMode": "VISIBLE",
      "rules": [
        {
          "fieldName": "salary",
          "mode": "MASKED",
          "maskPattern": "***"
        }
      ]
    },
    "preset": null,
    "displayName": "销售员可见客户信息",
    "description": "销售员只能看到自己跟进的客户"
  },
  "sortOrder": 1,
  "createdAt": "2024-04-20T11:00:00Z",
  "updatedAt": "2024-04-20T11:00:00Z"
}
```

---

## 数据流关系图

```
Bundle (权限包)
│
├── dataPermissionItems: [Item]  ← 核心：每个模型最多一个item
│   │
│   ├── Item (PRESET Mode)
│   │   ├── grantType: "PRESET"
│   │   ├── preset: "READ_WRITE_ALL" | "READ_ALL" | ...
│   │   └── customPermissionId: null
│   │
│   └── Item (CUSTOM Mode)
│       ├── grantType: "CUSTOM"
│       ├── preset: null
│       ├── customPermissionId: UUID
│       └── customPermission: EndUserPermission {
│           ├── columnPolicy: {rules}
│           ├── rowScope: "ALL" | "SELF" | ...
│           └── ...
│       }
│
├── snapshots: [Snapshot]  ← 版本历史
│   └── items: [SnapshotItem]  ← JSON: modelId, grantType, preset, customPermissionId
│
├── Role-Bundle (关联)
└── User-Bundle (关联)
```

---

## 文件位置速查

| 内容 | 文件路径 |
|------|--------|
| GraphQL Schema | `/modelcraft-backend/api/graph/project/schema/rbac.graphql` |
| Bundle表 | `/modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` (表3) |
| Item表 | `/modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` (表3) |
| Snapshot表 | `/modelcraft-backend/db/schema/mysql/14_rbac_bundle_snapshots.sql` |
| Bundle查询 | `/modelcraft-backend/db/queries/rbac/bundle.sql` |
| Item查询 | `/modelcraft-backend/db/queries/rbac/bundle.sql` (第3部分) |
| 权限查询 | `/modelcraft-backend/db/queries/rbac/permission.sql` |
| 快照查询 | `/modelcraft-backend/db/queries/rbac/bundle_snapshot.sql` |
| Go生成代码 | `/modelcraft-backend/internal/infrastructure/dbgen/bundle.sql.go` |
| Go数据模型 | `/modelcraft-backend/internal/infrastructure/dbgen/models.go` |

