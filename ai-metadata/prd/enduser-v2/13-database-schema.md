# EndUser v2 数据库 Schema 变更

> 本文档描述 EndUser v2 改造所需的数据库 Schema 变更，覆盖 DDL 设计、迁移策略和 Atlas 操作指南。

---

## 变更总览

| 表名 | 操作 | 说明 |
|------|------|------|
| `end_user_users` | **ALTER** | 去掉 `project_slug`，调整唯一约束，scope 变为 `(org_name)` |
| `end_user_accounts` | **ALTER** | 去掉 `project_slug`，FK 改为引用 `(org_name, user_id)` |
| `end_user_project_access` | **CREATE** | 新增多对多关联表：EndUser ↔ Project |
| `end_user_roles` | **ALTER** | 去掉 `project_slug`（迁移到 v2，角色与 Project 的关联通过 access 表管理） |
| `end_user_role_users` | **ALTER** | 去掉 `project_slug`，更新 FK 引用 |

---

## v1 vs v2 Schema 对比

### `end_user_users`

#### v1（当前）
```sql
CREATE TABLE `end_user_users` (
  `id`           VARCHAR(36)  NOT NULL,
  `org_name`     VARCHAR(36)  NOT NULL,
  `project_slug` VARCHAR(64)  NOT NULL,   -- ← 需要删除
  `username`     VARCHAR(64)  NOT NULL,
  `password`     VARCHAR(255) NOT NULL,
  `is_forbidden` TINYINT(1)   NOT NULL DEFAULT 0,
  `created_by`   VARCHAR(36)  NULL,
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_users_scope_username` (`org_name`, `project_slug`, `username`),
  UNIQUE KEY `uk_end_user_users_scope_id`       (`org_name`, `project_slug`, `id`),
  KEY           `idx_end_user_users_scope`       (`org_name`, `project_slug`),
  KEY           `idx_end_user_users_created_by`  (`created_by`)
);
```

#### v2（目标）
```sql
CREATE TABLE `end_user_users` (
  `id`           VARCHAR(36)  NOT NULL COMMENT '终端用户 ID (UUID)',
  `org_name`     VARCHAR(36)  NOT NULL COMMENT '所属 Org',
  `username`     VARCHAR(64)  NOT NULL COMMENT 'Org 内唯一用户名',
  `password`     VARCHAR(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  `is_forbidden` TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否禁用',
  `created_by`   VARCHAR(36)  NULL     COMMENT '创建者（平台用户 ID）',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_users_org_username` (`org_name`, `username`),
  UNIQUE KEY `uk_end_user_users_org_id`       (`org_name`, `id`),
  KEY           `idx_end_user_users_org`       (`org_name`),
  KEY           `idx_end_user_users_created_by`(`created_by`),

  CONSTRAINT `fk_end_user_users_created_by`
    FOREIGN KEY (`created_by`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='终端用户账号表（Org 级隔离）';
```

**变更说明**：
- 删除 `project_slug` 列
- 唯一约束 `(org_name, project_slug, username)` → `(org_name, username)`
- 辅助唯一键 `(org_name, project_slug, id)` → `(org_name, id)`
- 删除索引 `(org_name, project_slug)` → `(org_name)`

---

### `end_user_accounts`（会话表）

#### v1（当前）
```sql
CREATE TABLE `end_user_accounts` (
  `id`                 VARCHAR(36)  NOT NULL,
  `org_name`           VARCHAR(36)  NOT NULL,
  `project_slug`       VARCHAR(64)  NOT NULL,   -- ← 需要删除
  `user_id`            VARCHAR(36)  NOT NULL,
  `refresh_token_hash` VARCHAR(255) NOT NULL,
  `expires_at`         DATETIME     NOT NULL,
  `revoked`            TINYINT(1)   NOT NULL DEFAULT 0,
  `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_accounts_token_hash`  (`refresh_token_hash`),
  KEY           `idx_end_user_accounts_scope_user` (`org_name`, `project_slug`, `user_id`),
  KEY           `idx_end_user_accounts_scope`       (`org_name`, `project_slug`),

  CONSTRAINT `fk_end_user_accounts_user`
    FOREIGN KEY (`org_name`, `project_slug`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `project_slug`, `id`) ON DELETE CASCADE
);
```

#### v2（目标）
```sql
CREATE TABLE `end_user_accounts` (
  `id`                 VARCHAR(36)  NOT NULL COMMENT '会话 ID (UUID)',
  `org_name`           VARCHAR(36)  NOT NULL COMMENT '所属 Org',
  `user_id`            VARCHAR(36)  NOT NULL COMMENT '终端用户 ID',
  `refresh_token_hash` VARCHAR(255) NOT NULL COMMENT '刷新令牌哈希',
  `expires_at`         DATETIME     NOT NULL COMMENT '过期时间',
  `revoked`            TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否已撤销',
  `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_accounts_token_hash`  (`refresh_token_hash`),
  KEY           `idx_end_user_accounts_org_user` (`org_name`, `user_id`),
  KEY           `idx_end_user_accounts_org`      (`org_name`),

  CONSTRAINT `fk_end_user_accounts_user`
    FOREIGN KEY (`org_name`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='终端用户会话表（Org 级隔离）';
```

**变更说明**：
- 删除 `project_slug` 列（会话是 Org 级的，与 Project 无关）
- FK 从 `(org_name, project_slug, user_id) → end_user_users(org_name, project_slug, id)` 改为 `(org_name, user_id) → end_user_users(org_name, id)`

---

### `end_user_project_access`（新增）

```sql
CREATE TABLE `end_user_project_access` (
  `id`                   VARCHAR(36)  NOT NULL COMMENT '记录 ID (UUID)',
  `end_user_id`          VARCHAR(36)  NOT NULL COMMENT '终端用户 ID',
  `org_name`             VARCHAR(36)  NOT NULL COMMENT '所属 Org',
  `project_slug`         VARCHAR(64)  NOT NULL COMMENT '项目标识',
  `permission_bundle_id` VARCHAR(36)  NULL     COMMENT '分配的权限包 ID（NULL = 无特殊权限，仅访问）',
  `granted_by`           VARCHAR(36)  NULL     COMMENT '授权者（平台用户 ID）',
  `granted_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_eu_project_access_user_project` (`end_user_id`, `org_name`, `project_slug`),
  KEY           `idx_eu_project_access_project`   (`org_name`, `project_slug`),
  KEY           `idx_eu_project_access_user`      (`end_user_id`, `org_name`),

  CONSTRAINT `fk_eu_project_access_user`
    FOREIGN KEY (`org_name`, `end_user_id`)
    REFERENCES `end_user_users`(`org_name`, `id`) ON DELETE CASCADE,

  CONSTRAINT `fk_eu_project_access_granted_by`
    FOREIGN KEY (`granted_by`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='EndUser ↔ Project 授权关联表（多对多）';
```

**设计说明**：
- `permission_bundle_id` 允许 NULL：允许授权但不分配特殊权限包（使用 Project 默认权限）
- `granted_by` 记录授权人（Developer），便于审计
- 唯一约束 `(end_user_id, org_name, project_slug)` 确保一个用户在同一 Project 只有一条授权记录
- `project_slug` 不做外键约束（Project 属于不同 DB/schema，跨库 FK 不可行）

---

### `end_user_roles`

#### v2（目标）
```sql
CREATE TABLE `end_user_roles` (
  `id`          VARCHAR(36)  NOT NULL COMMENT '角色 ID (UUID)',
  `org_name`    VARCHAR(36)  NOT NULL COMMENT '所属 Org',
  `name`        VARCHAR(64)  NOT NULL COMMENT 'Org 内唯一角色名',
  `description` VARCHAR(255) NULL     COMMENT '角色描述',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_roles_org_name` (`org_name`, `name`),
  UNIQUE KEY `uk_end_user_roles_org_id`   (`org_name`, `id`),
  KEY           `idx_end_user_roles_org`  (`org_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='终端用户角色表（Org 级）';
```

---

### `end_user_role_users`

#### v2（目标）
```sql
CREATE TABLE `end_user_role_users` (
  `id`         VARCHAR(36) NOT NULL COMMENT '关联 ID (UUID)',
  `org_name`   VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `role_id`    VARCHAR(36) NOT NULL COMMENT '角色 ID',
  `user_id`    VARCHAR(36) NOT NULL COMMENT '终端用户 ID',
  `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_eu_role_users_org_role_user` (`org_name`, `role_id`, `user_id`),
  KEY           `idx_eu_role_users_org_role`  (`org_name`, `role_id`),
  KEY           `idx_eu_role_users_org_user`  (`org_name`, `user_id`),

  CONSTRAINT `fk_eu_role_users_role`
    FOREIGN KEY (`org_name`, `role_id`)
    REFERENCES `end_user_roles`(`org_name`, `id`) ON DELETE CASCADE,

  CONSTRAINT `fk_eu_role_users_user`
    FOREIGN KEY (`org_name`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='终端用户角色-用户关联表（Org 级）';
```

---

## 迁移策略

### 原则

- **破坏性变更，需要数据迁移**：`project_slug` 列的删除是不可逆操作，迁移前必须确认数据处理方式。
- **Atlas 管理**：所有 Schema 变更通过 `just db-diff` + `just db-apply` 执行，不直接修改 Atlas 托管的 schema 文件。
- **迁移窗口**：建议在 EndUser 功能上线前执行，避免双写冲突。

### 数据迁移方案

#### 方案A：清空重建（推荐，适用于测试/预发）

EndUser 数据为业务数据（非平台核心数据），如果当前环境为测试或早期预发，可直接：

```sql
-- 1. 清空关联数据
TRUNCATE TABLE `end_user_role_users`;
TRUNCATE TABLE `end_user_roles`;
TRUNCATE TABLE `end_user_accounts`;
TRUNCATE TABLE `end_user_users`;
-- 2. 执行 DDL 变更（由 Atlas 生成的迁移文件完成）
```

#### 方案B：保留数据迁移（适用于生产）

```sql
-- 1. 去重：同一 (org_name, username) 可能有多条来自不同 project 的记录
--    以 created_at 最早的记录为准，删除重复项
DELETE u1 FROM end_user_users u1
INNER JOIN end_user_users u2
  ON u1.org_name = u2.org_name
  AND u1.username = u2.username
  AND u1.created_at > u2.created_at;

-- 2. 为所有现有用户创建 project_access 记录（保留原有 project 关联）
INSERT INTO end_user_project_access (id, end_user_id, org_name, project_slug, granted_by, granted_at)
SELECT UUID(), id, org_name, project_slug, created_by, created_at
FROM end_user_users;

-- 3. 执行 DDL：删除 project_slug 列（由 Atlas 迁移文件完成）
```

> ⚠️ **方案B 前提假设**：同一用户在不同 Project 的账号是独立的（不同 password hash），迁移时需人工决策如何合并（取哪个 password）。实际操作建议与产品确认。

---

## Atlas 操作指南

### 1. 在 `db/schema/mysql/` 创建新迁移文件

```bash
# 新建迁移文件（命名规则：序号_描述.sql）
# 序号接上一个迁移文件，例如当前最大是 12_end_user_auth.sql
# 新文件命名：13_end_user_auth_v2.sql
```

### 2. 生成 Atlas 差异

```bash
cd modelcraft-backend
just db-diff
```

### 3. 检查生成的迁移计划

```bash
# 查看 Atlas 生成的迁移 SQL
cat db/migrations/xxx.sql
```

### 4. 应用迁移

```bash
just db-apply
```

### 注意事项

- **禁止直接修改 Atlas 已执行的迁移文件**
- 新增迁移文件必须保证幂等（或依赖 Atlas 的迁移状态管理）
- 生产环境先在预发执行并验证，再推生产

---

## Schema 影响范围

### SQLC 查询文件需更新

| 文件路径 | 变更说明 |
|---------|---------|
| `db/queries/end_user_users.sql` | 删除所有 `project_slug` 条件，修改插入语句 |
| `db/queries/end_user_accounts.sql` | 同上 |
| `db/queries/end_user_project_access.sql` | **新建文件**，增加 CRUD 查询 |

### Go 代码影响

| 层 | 文件 | 变更说明 |
|----|------|---------|
| Domain | `internal/domain/enduser/end_user.go` | 添加 `OrgName` 字段，删除通过 Context 隐式传递的 project scope |
| Domain | `internal/domain/enduser/end_user_repository.go` | `ListEndUsersQuery` 添加 `OrgName`；新增 `EndUserProjectAccessRepository` |
| Infrastructure | `internal/infrastructure/enduser/` | 重写所有 SQLC 查询调用，适配新 Schema |
| App | `internal/app/enduser/` | 业务逻辑调整，涉及登录流程新增 Project 选择 |
| Interfaces | `internal/interfaces/graphql/` | GraphQL resolver 适配新接口 |

---

## 参考

- [11-domain-model-changes.md](./11-domain-model-changes.md) — 领域模型变更详情
- [12-graphql-api-design.md](./12-graphql-api-design.md) — GraphQL API 变更
- `db/schema/mysql/12_end_user_auth.sql` — v1 原始 Schema
- Atlas 文档：`ai-metadata/backend/tools/justfile-guide.md`
