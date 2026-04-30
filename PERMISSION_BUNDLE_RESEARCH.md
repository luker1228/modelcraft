# 🔍 ModelCraft 权限包（Permission Bundle）数据权限配置 - 研究报告

**研究时间**: 2026-05-01  
**项目**: ModelCraft  
**研究对象**: 权限包下的"数据权限配置"列表项字段结构

---

## 📌 核心发现

### 1. 截图中显示的字段对应

当在权限包管理界面看到这样的列表项：
```
┌─────────────────────────┐
│ d97f7a43...  │ 读写全部 │
└─────────────────────────┘
```

实际上对应的 GraphQL 类型和数据库字段如下：

| 视觉显示 | GraphQL字段 | 数据库列 | 数据类型 | 说明 |
|--------|----------|--------|--------|------|
| `d97f7a43...` | `id` | `id` | VARCHAR(36) | 唯一标识符（Item UUID） |
| `读写全部` | `preset` | `preset` | ENUM | 预设权限类型 |

### 2. 关键数据结构

**类型名**: `EndUserBundleDataPermissionItem`  
**位置**: GraphQL Schema 中的 `rbac.graphql` 第 189-204 行  
**数据库表**: `end_user_bundle_data_permission_items`  
**数据库文件**: `13_rbac_permissions.sql` 第 98-138 行  

这个类型代表"权限包内的单个数据权限配置项"，是理解整个权限系统的核心。

---

## 📊 完整的 Item 字段列表

```graphql
type EndUserBundleDataPermissionItem {
  id: ID!                         # d97f7a43... (Item UUID)
  bundleId: ID!                   # 所属权限包ID
  modelId: ID!                    # 关联模型ID
  
  grantType: DataPermissionGrantType!  # 关键字段：PRESET 或 CUSTOM
  
  # PRESET模式下有值
  preset: EndUserPermissionPreset  # 对应"读写全部"等预设名称
    # 可选值：
    # - READ_WRITE_ALL (读写全部)
    # - READ_ALL (只读全部)
    # - READ_WRITE_OWNER (读写自己)
    # - READ_ALL_WRITE_OWNER (读所有写自己)
  
  # CUSTOM模式下有值
  customPermissionId: ID
  customPermission: EndUserPermission
  
  sortOrder: Int!                 # 显示顺序
  createdAt: Time!
  updatedAt: Time!
}
```

---

## 🎯 两种权限授权模式详解

### PRESET 模式（预设） - 80%场景

使用预定义的4个权限模板之一：

| 英文值 | 中文名 | 权限说明 | 数据库值 |
|--------|------|--------|---------|
| `READ_WRITE_ALL` | **读写全部** | 对所有行/列读写 | `'READ_WRITE_ALL'` |
| `READ_ALL` | **只读全部** | 对所有行/列只读 | `'READ_ALL'` |
| `READ_WRITE_OWNER` | **读写自己** | 只能读写自己的行 | `'READ_WRITE_OWNER'` |
| `READ_ALL_WRITE_OWNER` | **读所有写自己** | 读全部，写自己的行 | `'READ_ALL_WRITE_OWNER'` |

**数据库存储示例**：
```sql
INSERT INTO end_user_bundle_data_permission_items
VALUES (
  'item-uuid-123',
  'bundle-uuid-456',
  'model-uuid-789',
  'PRESET',                    -- grant_type
  'READ_WRITE_ALL',            -- preset (对应"读写全部")
  NULL,                        -- custom_permission_id
  0                            -- sort_order
);
```

### CUSTOM 模式（自定义） - 20%场景

由管理员手工创建复杂权限规则：

**数据库存储示例**：
```sql
INSERT INTO end_user_bundle_data_permission_items
VALUES (
  'item-uuid-124',
  'bundle-uuid-456',
  'model-uuid-789',
  'CUSTOM',                    -- grant_type
  NULL,                        -- preset
  'custom-perm-uuid',          -- custom_permission_id
  1                            -- sort_order
);

-- 对应的自定义权限存储在end_user_data_permissions表
INSERT INTO end_user_data_permissions
VALUES (
  'custom-perm-uuid',
  'org-name',
  'project-slug',
  'db-name',
  'model-name',
  'model-uuid-789',
  '销售员可见客户信息',
  '销售员只能看到自己跟进的客户数据',
  '{
    "defaultMode": "VISIBLE",
    "rules": [
      {"fieldName": "salary", "mode": "MASKED", "maskPattern": "***"},
      {"fieldName": "id_card", "mode": "HIDDEN"}
    ]
  }',                          -- column_policy (列级脱敏)
  '{
    "select": {"allowed": true, "scope": "custom", "predicate": {...}},
    "insert": {"allowed": true, "scope": "self", "check": {...}},
    "update": {"allowed": true, "scope": "self"},
    "delete": {"allowed": false}
  }',                          -- row_policy (行级过滤)
  NOW(),
  NOW()
);
```

---

## 🏗️ 数据库 Schema 关系图

### 核心表：end_user_bundle_data_permission_items

```sql
CREATE TABLE `end_user_bundle_data_permission_items` (
  `id`                   VARCHAR(36) PRIMARY KEY,
  `bundle_id`            VARCHAR(36) NOT NULL,
  `model_id`             VARCHAR(36) NOT NULL,
  
  # 核心约束：同一 bundle 下同一 model 最多一个 item
  UNIQUE KEY (`bundle_id`, `model_id`),
  
  # 关键字段：两种模式选择
  `grant_type`           ENUM('PRESET', 'CUSTOM') NOT NULL,
  
  # PRESET 专用
  `preset`               ENUM(
                          'READ_WRITE_ALL',
                          'READ_ALL',
                          'READ_WRITE_OWNER',
                          'READ_ALL_WRITE_OWNER'
                        ) NULL,
  
  # CUSTOM 专用
  `custom_permission_id` VARCHAR(36) NULL,
  
  # 内部约束：PRESET 和 CUSTOM 互斥
  CHECK (grant_type != 'PRESET' OR 
         (preset IS NOT NULL AND custom_permission_id IS NULL)),
  CHECK (grant_type != 'CUSTOM' OR 
         (custom_permission_id IS NOT NULL AND preset IS NULL)),
  
  `sort_order`           INT NOT NULL DEFAULT 0,
  `created_at`           DATETIME DEFAULT NOW(),
  `updated_at`           DATETIME DEFAULT NOW(),
  
  # 外键关系
  CONSTRAINT `fk_bundle_items_bundle`
    FOREIGN KEY (`bundle_id`) 
    REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE,
  
  CONSTRAINT `fk_bundle_items_model`
    FOREIGN KEY (`model_id`) 
    REFERENCES `models` (`id`)
    ON DELETE CASCADE,
  
  CONSTRAINT `fk_bundle_items_custom_permission`
    FOREIGN KEY (`custom_permission_id`) 
    REFERENCES `end_user_data_permissions` (`id`)
);
```

### 关联的表

1. **end_user_permission_bundles** (权限包)
   - 包含：id, slug, name, description, created_at, updated_at

2. **end_user_data_permissions** (自定义权限实体)
   - 包含：id, model_id, name, description, column_policy (JSON), row_policy (JSON)

3. **end_user_permission_bundle_snapshots** (版本快照)
   - 包含：id, bundle_id, version, items (JSON array), created_by, restored_from

---

## 📁 相关文件导航

### GraphQL Schema Files

| 文件 | 关键内容 | 行号 |
|-----|--------|------|
| `api/graph/project/schema/rbac.graphql` | EndUserBundleDataPermissionItem 类型定义 | 189-204 ⭐ |
| | EndUserPermissionPreset enum (4个预设值) | 62-67 |
| | DataPermissionGrantType enum (PRESET\|CUSTOM) | 72-75 |
| | EndUserPermissionBundle 完整定义 | 157-183 |

### Database Schema Files

| 文件 | 关键内容 | 行号 |
|-----|--------|------|
| `db/schema/mysql/13_rbac_permissions.sql` | Bundle 表定义 | 70-90 |
| | **Item 表定义** (核心) | **98-138 ⭐** |
| | 自定义权限表定义 | 19-64 |
| `db/schema/mysql/14_rbac_bundle_snapshots.sql` | 版本快照表定义 | 12-32 |

### SQLC Query Files

| 文件 | 关键查询 | 行号 |
|-----|--------|------|
| `db/queries/rbac/bundle.sql` | UpsertBundleDataPermissionItem | 50-67 |
| | ListBundleDataPermissionItems | 74-78 |
| | GetBundleDataPermissionItemByBundleAndModel | 84-88 |
| | RemoveBundleDataPermissionItem | 69-72 |

### Go Code Files

| 文件 | 关键内容 |
|-----|--------|
| `internal/infrastructure/dbgen/models.go` | EndUserBundleDataPermissionItem struct |
| | EndUserDataPermission struct |
| `internal/infrastructure/dbgen/bundle.sql.go` | sqlc 生成的 Go 代码 |

---

## 🔄 数据流示例

### 创建 PRESET Item 的完整流程

```
1. 前端用户操作
   用户在权限包界面选择"读写全部"预设
   
2. GraphQL Mutation
   bindPresetItemToBundle(
     bundleId: "bundle-123",
     modelId: "model-456",
     preset: READ_WRITE_ALL,
     sortOrder: 0
   )
   
3. 后端处理
   - 验证 bundle 存在
   - 验证 model 存在
   - 生成新 Item UUID
   
4. 数据库操作 (UPSERT语义)
   INSERT INTO end_user_bundle_data_permission_items
   (id, bundle_id, model_id, grant_type, preset, sort_order)
   VALUES ('item-uuid', 'bundle-123', 'model-456', 'PRESET', 'READ_WRITE_ALL', 0)
   ON DUPLICATE KEY UPDATE
   (同 bundle+model 已存在时更新现有记录)
   
5. 版本控制
   - 自动保存 snapshot
   - version 递增
   - 最多保留 5 个历史版本
   
6. GraphQL Response
   {
     "bundle": {
       "dataPermissionItems": [
         {
           "id": "item-uuid",
           "bundleId": "bundle-123",
           "modelId": "model-456",
           "grantType": "PRESET",
           "preset": "READ_WRITE_ALL",
           "customPermissionId": null,
           "sortOrder": 0
         }
       ]
     }
   }
```

### 查询 Bundle 的权限配置

```graphql
query {
  endUserPermissionBundle(id: "bundle-123") {
    id
    name
    dataPermissionItems {
      id              # d97f7a43...
      modelId
      grantType       # PRESET 或 CUSTOM
      preset          # 读写全部 (仅PRESET时有值)
      customPermission {  # (仅CUSTOM时有值)
        displayName
        columnPolicy { defaultMode, rules { fieldName, mode } }
        rowScope
      }
      sortOrder
    }
  }
}
```

---

## 💡 设计原理

### 1. Item-Centric 模型

- 每个 Item 代表 Bundle 在某个 Model 上的权限配置
- 同一 Bundle 下同一 Model 最多一个 Item
- 通过 UPSERT 实现 replace 语义（不重复，就地更新）

### 2. 二元模式设计

使用 `grant_type` 字段在两种模式间切换：
- **PRESET**：快速易用，覆盖 80% 的需求
- **CUSTOM**：灵活强大，支持复杂的列级和行级控制

两种模式的字段完全互斥，通过 CHECK 约束保证数据完整性。

### 3. 版本控制与回滚

- 每次权限配置变更都保存快照
- 快照存储 Item 的完整信息（JSON 格式）
- 支持回滚到任意历史版本
- 最多保留 5 个版本（FIFO 清理）

### 4. 授权双通道

权限最终生效通过两种方式：
- **角色授权**：Role → Bundle (end_user_role_bundles)
- **用户直接授权**：User → Bundle (end_user_user_bundles)

---

## 🚀 常见操作代码示例

### 添加 PRESET Item

```sql
-- 数据库层面
INSERT INTO end_user_bundle_data_permission_items
(id, bundle_id, model_id, grant_type, preset, sort_order, created_at, updated_at)
VALUES (
  UUID(),
  'bundle-uuid',
  'model-uuid',
  'PRESET',
  'READ_WRITE_ALL',
  0,
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  grant_type = VALUES(grant_type),
  preset = VALUES(preset),
  sort_order = VALUES(sort_order),
  updated_at = NOW();
```

### 添加 CUSTOM Item

```sql
-- 先创建自定义权限
INSERT INTO end_user_data_permissions
(id, org_name, project_slug, model_id, name, description, column_policy, row_policy)
VALUES (...);

-- 再绑定到 Bundle
INSERT INTO end_user_bundle_data_permission_items
(id, bundle_id, model_id, grant_type, custom_permission_id, sort_order)
VALUES (
  UUID(),
  'bundle-uuid',
  'model-uuid',
  'CUSTOM',
  'custom-perm-uuid',
  1
);
```

### 查询 Bundle 所有 Item

```sql
SELECT *
FROM end_user_bundle_data_permission_items
WHERE bundle_id = 'bundle-uuid'
ORDER BY sort_order, created_at;
```

---

## 📚 相关知识体系

### 权限系统分层

```
组织级权限 (PermissionRole)
    ↓
项目级权限 (RBAC - Row/Column级)
    ├─ 预设权限 (Preset)
    ├─ 自定义权限 (Custom)
    └─ 权限包 (Bundle) ← 本研究的焦点
        └─ 数据权限 Item
            ├─ 列级策略 (ColumnPolicy)
            └─ 行级策略 (RowPolicy)
```

### 数据权限配置项的层级

```
Bundle (权限包)
  └─ Item (权限配置项) ⭐
      ├─ PRESET 模式
      │  └─ 4个预设之一
      └─ CUSTOM 模式
          └─ EndUserPermission (自定义权限)
              ├─ ColumnPolicy (列级)
              └─ RowPolicy (行级)
```

---

## 📖 快速查阅表

### 核心枚举值

**DataPermissionGrantType**:
- `PRESET` - 使用预设模板
- `CUSTOM` - 使用自定义权限

**EndUserPermissionPreset** (4个值):
- `READ_WRITE_ALL` - 读写全部
- `READ_ALL` - 只读全部  
- `READ_WRITE_OWNER` - 读写自己
- `READ_ALL_WRITE_OWNER` - 读全部写自己

**ColumnAccessMode**:
- `VISIBLE` - 可见可编辑
- `READONLY` - 可见只读
- `MASKED` - 脱敏显示
- `HIDDEN` - 隐藏

**RowScopeType**:
- `ALL` - 所有行
- `SELF` - 自己的行
- `DEPT` - 部门的行
- `DEPT_AND_CHILDREN` - 部门及下级

---

## 🎓 学习建议

1. **从 Item 开始**
   - 理解 Item 是权限配置的最小单位
   - 掌握 PRESET vs CUSTOM 两种模式

2. **理解约束**
   - `grant_type` 决定了哪些字段有值
   - CHECK 约束保证数据一致性

3. **研究版本控制**
   - 了解快照如何记录历史
   - 理解回滚机制

4. **探索授权关系**
   - 权限如何从 Bundle 流向用户/角色
   - 最终权限生效的路径

---

## 📝 总结

**"数据权限配置"列表里的每个条目是一个 `EndUserBundleDataPermissionItem`，包含以下核心字段：**

| 字段 | 类型 | 说明 |
|-----|-----|------|
| `id` | UUID | 条目唯一标识 (对应截图中的 d97f7a43...) |
| `bundleId` | UUID | 所属权限包 |
| `modelId` | UUID | 关联的数据模型 |
| `grantType` | ENUM | PRESET 或 CUSTOM |
| `preset` | ENUM | 预设类型 (对应截图中的"读写全部") |
| `customPermissionId` | UUID | 自定义权限ID (CUSTOM模式下) |
| `sortOrder` | INT | 显示顺序 |
| `createdAt` | DATETIME | 创建时间 |
| `updatedAt` | DATETIME | 更新时间 |

**核心表**：`end_user_bundle_data_permission_items`  
**核心 GraphQL 类型**：`EndUserBundleDataPermissionItem`  
**文件位置**：见上述导航表格

---

*本研究报告为 ModelCraft 权限系统的快速参考文档。*
