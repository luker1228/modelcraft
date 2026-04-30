# 📋 权限包（Permission Bundle）快速参考卡片

## 🎯 截图中的字段解读

当看到这样的列表项：
```
┌─────────────────────────┐
│ d97f7a43...  │ 读写全部 │
└─────────────────────────┘
```

实际上对应的是 **EndUserBundleDataPermissionItem** 类型中的：

| 显示字段 | GraphQL字段 | 数据库列 | 含义 |
|---------|----------|--------|------|
| `d97f7a43...` | `id` | `id` | Item的唯一标识符(UUID) |
| `读写全部` | `preset` | `preset` | 预设权限类型 |

---

## 🏗️ 核心数据结构

### EndUserBundleDataPermissionItem (数据权限配置项)

这是权限包里**最重要**的数据结构：

```graphql
type EndUserBundleDataPermissionItem {
  id: ID!                         # Item UUID (对应 d97f7a43...)
  bundleId: ID!                   # 所属权限包ID
  modelId: ID!                    # 关联的数据模型ID
  
  grantType: DataPermissionGrantType!  # 关键：PRESET 或 CUSTOM
  
  # 当 grantType=PRESET 时有值（对应"读写全部"等预设名称）
  preset: EndUserPermissionPreset  
    # 可选值: READ_WRITE_ALL / READ_ALL / READ_WRITE_OWNER / READ_ALL_WRITE_OWNER
  
  # 当 grantType=CUSTOM 时有值（自定义权限）
  customPermissionId: ID
  customPermission: EndUserPermission
  
  sortOrder: Int!                 # 显示排序顺序
  createdAt: Time!
  updatedAt: Time!
}
```

### 两种权限模式

#### 1️⃣ PRESET 模式（预设）
最常见，使用4个预定义的权限模板：

| 预设值 | 中文名 | 权限范围 | 依赖字段 |
|-------|------|--------|--------|
| `READ_WRITE_ALL` | **读写全部** | 对所有行/列读写 | 无 |
| `READ_ALL` | **只读全部** | 对所有行/列只读 | 无 |
| `READ_WRITE_OWNER` | **读写自己** | 只能读写自己的行 | END_USER_REF |
| `READ_ALL_WRITE_OWNER` | **读所有写自己** | 读全部，写自己的行 | END_USER_REF |

**数据库存储**：
```sql
grantType = 'PRESET'
preset = 'READ_WRITE_ALL'  -- 对应"读写全部"
customPermissionId = NULL
```

#### 2️⃣ CUSTOM 模式（自定义）
由管理员手工创建，支持精细化的列级和行级权限控制：

**数据库存储**：
```sql
grantType = 'CUSTOM'
preset = NULL
customPermissionId = 'uuid-of-permission'  -- 指向end_user_data_permissions表
```

对应的自定义权限包含：
- 列策略（Column Policy）：可见/只读/脱敏/隐藏
- 行策略（Row Policy）：WHERE条件过滤
- 操作权限（SELECT/INSERT/UPDATE/DELETE/EXPORT）

---

## 📊 数据库表关系

```
┌─ end_user_permission_bundles (权限包)
│  └─ 1..∞ 关系
│
├─ end_user_bundle_data_permission_items (数据权限配置项) ⭐
│  ├─ grant_type = PRESET
│  │  └─ preset = READ_WRITE_ALL | READ_ALL | ...
│  │
│  └─ grant_type = CUSTOM
│     └─ custom_permission_id → end_user_data_permissions (自定义权限)
│
├─ end_user_permission_bundle_snapshots (版本快照)
│  └─ items: JSON格式存储每个Item的快照
│
├─ end_user_role_bundles (角色关联)
│  └─ role_id → end_user_roles
│
└─ end_user_user_bundles (用户直接授权)
   └─ user_id → end_user_users
```

---

## 🔍 核心表定义

### end_user_bundle_data_permission_items 表

```sql
CREATE TABLE `end_user_bundle_data_permission_items` (
  `id`                     VARCHAR(36) PRIMARY KEY,
  `bundle_id`              VARCHAR(36) NOT NULL FK,
  `model_id`               VARCHAR(36) NOT NULL FK,
  
  # 核心约束：同一bundle下同一model只能有一个item
  UNIQUE KEY (`bundle_id`, `model_id`),
  
  # 授权来源类型
  `grant_type`             ENUM('PRESET', 'CUSTOM') NOT NULL,
  
  # PRESET专用字段
  `preset`                 ENUM('READ_WRITE_ALL', 'READ_ALL', 
                                'READ_WRITE_OWNER', 'READ_ALL_WRITE_OWNER') NULL,
  
  # CUSTOM专用字段
  `custom_permission_id`   VARCHAR(36) NULL FK,
  
  # 内部约束：PRESET和CUSTOM互斥
  CHECK (grant_type != 'PRESET' OR 
         (preset IS NOT NULL AND custom_permission_id IS NULL)),
  CHECK (grant_type != 'CUSTOM' OR 
         (custom_permission_id IS NOT NULL AND preset IS NULL)),
  
  `sort_order`             INT NOT NULL DEFAULT 0,
  `created_at`             DATETIME DEFAULT NOW(),
  `updated_at`             DATETIME DEFAULT NOW()
);
```

---

## 📁 关键文件位置

### GraphQL Schema
- **后端**: `modelcraft-backend/api/graph/project/schema/rbac.graphql`
  - 第186-204行：EndUserBundleDataPermissionItem 定义
  - 第62-67行：EndUserPermissionPreset enum定义
  - 第72-75行：DataPermissionGrantType enum定义

### 数据库Schema
- **Bundle表**: `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` (第70-90行)
- **Item表** ⭐: `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` (第98-138行)
- **自定义权限表**: `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` (第19-64行)
- **快照表**: `modelcraft-backend/db/schema/mysql/14_rbac_bundle_snapshots.sql` (第12-32行)

### SQLC查询
- **Item操作**: `modelcraft-backend/db/queries/rbac/bundle.sql` (第48-89行)
  - `UpsertBundleDataPermissionItem` - Replace语义
  - `ListBundleDataPermissionItems` - 按sort_order排序
  - `GetBundleDataPermissionItemByBundleAndModel` - 单项查询

### Go代码
- **数据模型**: `modelcraft-backend/internal/infrastructure/dbgen/models.go`
  - `type EndUserBundleDataPermissionItem struct`
  - `type EndUserDataPermission struct`

---

## 🚀 常见操作

### 1. 为Bundle添加PRESET权限
```graphql
mutation {
  bindPresetItemToBundle(input: {
    bundleId: "bundle-123"
    modelId: "model-456"
    preset: READ_WRITE_ALL      # 读写全部
    sortOrder: 0
  }) {
    bundle {
      dataPermissionItems {
        id
        preset
        modelId
      }
    }
  }
}
```

对应数据库操作：
```sql
INSERT INTO end_user_bundle_data_permission_items
  (id, bundle_id, model_id, grant_type, preset, sort_order)
VALUES (?, ?, ?, 'PRESET', 'READ_WRITE_ALL', 0)
ON DUPLICATE KEY UPDATE
  grant_type='PRESET', preset='READ_WRITE_ALL';
```

### 2. 为Bundle添加CUSTOM权限
```graphql
mutation {
  bindCustomItemToBundle(input: {
    bundleId: "bundle-123"
    modelId: "model-456"
    customPermissionId: "custom-perm-789"  # 引用自定义权限
    sortOrder: 1
  }) {
    bundle {
      dataPermissionItems {
        id
        customPermission {
          displayName
          columnPolicy { ... }
          rowScope
        }
      }
    }
  }
}
```

### 3. 查询Bundle所有权限配置
```graphql
query {
  endUserPermissionBundle(id: "bundle-123") {
    dataPermissionItems {
      id
      modelId
      grantType
      preset        # PRESET时有值
      customPermission {  # CUSTOM时有值
        displayName
        columnPolicy { ... }
      }
    }
  }
}
```

---

## 💡 关键设计原则

1. **Item-centric 模型**
   - 同一bundle下同一model最多一个item
   - 通过upsert实现replace语义（更新已有，新增不存在的）

2. **预设 vs 自定义**
   - 预设快速易用，适合80%的场景
   - 自定义支持复杂需求，如列级脱敏、动态行过滤

3. **版本控制**
   - 每次Item列表变更自动保存快照
   - 最多保留5个历史版本，支持回滚

4. **授权关系**
   - 支持两种授权渠道：
     - 角色授权（Role → Bundle）
     - 用户直接授权（User → Bundle）

---

## 🎓 学习路径

1. **理解Item数据结构** ← 从这里开始！
   - 核心字段：`grantType`, `preset`, `customPermissionId`
   - 两种模式的互斥约束

2. **学习两种权限模式**
   - PRESET的4种预设
   - CUSTOM的列级/行级策略

3. **理解Bundle组织**
   - Bundle = Item的集合
   - Item = 模型上的权限配置

4. **研究版本控制**
   - Snapshot记录历史变更
   - 支持版本回滚

5. **学习授权关系**
   - 如何通过角色/用户关联Bundle
   - 权限的最终生效过程

