# Unified User System — Plan 1: DB Schema

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 破坏性更新数据库 Schema，合并 `end_user_users` 到 `users`，新增 `user_orgs`（含 `is_admin`），将 `end_user_role_users` 重命名为 `project_role_users`，删除 `end_user_accounts`。

**Architecture:** 使用 Atlas 声明式 schema（修改 `db/schema/mysql/` 下的 SQL 文件），然后运行 `just db reset` 破坏性重建数据库。同步修改 SQLC query 文件并重新生成代码。

**Tech Stack:** MySQL 8, Atlas CLI, SQLC, Go

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **修改** | `db/schema/mysql/06_users.sql` — `users` 表加字段，`user_organizations` 改为 `user_orgs` 加 `is_admin` |
| **修改** | `db/schema/mysql/12_end_user_auth.sql` — 删除 `end_user_users`、`end_user_accounts`；`end_user_role_users` 改为 `project_role_users`（FK 改指向 `users`）；`end_user_roles` 改名为 `project_roles` |
| **修改** | `db/schema/mysql/08_refresh_tokens.sql` — 确认 refresh tokens 表（tenant 用）不需要变更 |
| **修改** | `db/queries/org.sql` — 更新 user_organizations → user_orgs 相关查询，加 is_admin 字段 |
| **修改** | `db/queries/rbac/user_role.sql` — 更新 end_user_role_users → project_role_users，end_user_users → users |
| **修改** | `db/queries/rbac/authz.sql` — 同上 |
| **删除内容** | `db/queries/` 中所有引用 `end_user_users`、`end_user_accounts` 的查询 |
| **重新生成** | `internal/infrastructure/dbgen/` — 运行 `just generate-sqlc` |

---

## Task 1: 修改 `06_users.sql` — 将 `user_organizations` 替换为 `user_orgs`

**Files:**
- Modify: `modelcraft-backend/db/schema/mysql/06_users.sql`

`users` 表当前有 `name` 字段作为用户名。合并后 `name` 继续作为全局唯一 username（原 end_user_users.username 语义对齐）。无需新增字段，`name` + `password_hash` 已满足需求。

- [ ] **Step 1: 读取当前 06_users.sql，确认 users 表字段**

```bash
cat modelcraft-backend/db/schema/mysql/06_users.sql
```

预期：看到 `name`、`password_hash`、`deleted_at`、`delete_token` 字段。

- [ ] **Step 2: 修改 `user_organizations` 为 `user_orgs`，加 `is_admin` 字段，加单 Org 唯一约束**

将 `06_users.sql` 中的 `user_organizations` 表定义替换为：

```sql
CREATE TABLE IF NOT EXISTS `user_orgs` (
  `id`           VARCHAR(36)     NOT NULL PRIMARY KEY COMMENT 'UUID',
  `user_id`      VARCHAR(36)     NOT NULL COMMENT '用户 ID（引用 users.id）',
  `org_name`     VARCHAR(36)     NOT NULL COMMENT '组织名称（引用 organizations.name）',
  `is_admin`     TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否为管理员',
  `status`       VARCHAR(20)     NOT NULL DEFAULT 'active' COMMENT '状态：active | suspended',
  `created_at`   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位',

  CONSTRAINT `fk_user_orgs_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_orgs_org`  FOREIGN KEY (`org_name`) REFERENCES `organizations`(`name`) ON DELETE CASCADE,
  UNIQUE KEY `uk_user_orgs_user` (`user_id`, `delete_token`) COMMENT '每个用户只能属于一个 Org',
  UNIQUE KEY `uk_user_orgs_user_org` (`user_id`, `org_name`, `delete_token`),
  INDEX `idx_user_orgs_org` (`org_name`),
  INDEX `idx_user_orgs_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-组织绑定表（每人只属于一个 Org）';
```

注意：移除 `invited_by`、`invited_at`、`joined_at`（邀请流程不再需要）。

- [ ] **Step 3: 验证文件修改正确**

```bash
grep -n "user_orgs\|user_organizations\|is_admin" modelcraft-backend/db/schema/mysql/06_users.sql
```

预期：只出现 `user_orgs`，不出现 `user_organizations`；出现 `is_admin`。

- [ ] **Step 4: Commit**

```bash
cd modelcraft-backend && git add db/schema/mysql/06_users.sql
git commit -m "schema: rename user_organizations to user_orgs, add is_admin field"
```

---

## Task 2: 重写 `12_end_user_auth.sql`

**Files:**
- Modify: `modelcraft-backend/db/schema/mysql/12_end_user_auth.sql`

删除 `end_user_users`、`end_user_accounts`；重命名 `end_user_roles` → `project_roles`；重命名 `end_user_role_users` → `project_role_users`，FK 改为指向 `users` 表。

- [ ] **Step 1: 读取当前文件内容**

```bash
cat modelcraft-backend/db/schema/mysql/12_end_user_auth.sql
```

- [ ] **Step 2: 将整个文件内容替换为以下内容**

```sql
-- Project 级角色表（原 end_user_roles）
CREATE TABLE IF NOT EXISTS `project_roles` (
  `id`           VARCHAR(36)     NOT NULL COMMENT '角色 ID (UUID)',
  `org_name`     VARCHAR(36)     NOT NULL COMMENT '所属 Org',
  `project_slug` VARCHAR(64)     NOT NULL COMMENT '所属项目',
  `name`         VARCHAR(64)     NOT NULL COMMENT 'Project 内唯一角色名',
  `description`  VARCHAR(255)    NULL     COMMENT '角色描述',
  `is_implicit`  TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '内置隐式角色标志',
  `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`   BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_roles_name`   (`org_name`, `project_slug`, `name`, `delete_token`),
  UNIQUE KEY `uk_project_roles_org_id` (`org_name`, `id`, `delete_token`),
  KEY `idx_project_roles_org_id_fk`    (`org_name`, `id`),
  KEY `idx_project_roles_project`      (`org_name`, `project_slug`),
  KEY `idx_project_roles_implicit`     (`is_implicit`),
  KEY `idx_project_roles_live`         (`org_name`, `project_slug`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project 级数据角色表';

-- Project 级角色-用户关联表（原 end_user_role_users，FK 改指向 users）
CREATE TABLE IF NOT EXISTS `project_role_users` (
  `id`           VARCHAR(36) NOT NULL COMMENT '关联 ID (UUID)',
  `org_name`     VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `role_id`      VARCHAR(36) NOT NULL COMMENT '角色 ID（引用 project_roles.id）',
  `user_id`      VARCHAR(36) NOT NULL COMMENT '用户 ID（引用 users.id）',
  `created_at`   DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_role_users` (`org_name`, `role_id`, `user_id`),
  KEY `idx_project_role_users_role`  (`org_name`, `role_id`),
  KEY `idx_project_role_users_user`  (`org_name`, `user_id`),
  CONSTRAINT `fk_project_role_users_role` FOREIGN KEY (`org_name`, `role_id`) REFERENCES `project_roles`(`org_name`, `id`) ON DELETE CASCADE,
  CONSTRAINT `fk_project_role_users_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project 级角色-用户关联表';
```

- [ ] **Step 3: 验证文件内容**

```bash
grep -n "end_user" modelcraft-backend/db/schema/mysql/12_end_user_auth.sql
```

预期：**无任何输出**（所有 end_user 引用已消除）。

```bash
grep -n "project_roles\|project_role_users" modelcraft-backend/db/schema/mysql/12_end_user_auth.sql
```

预期：出现两个表定义。

- [ ] **Step 4: Commit**

```bash
cd modelcraft-backend && git add db/schema/mysql/12_end_user_auth.sql
git commit -m "schema: replace end_user tables with project_roles and project_role_users"
```

---

## Task 3: 应用 Schema 变更（破坏性重建）

**Files:** 无文件修改，运行命令

- [ ] **Step 1: 确认当前在 modelcraft-backend 目录**

```bash
cd modelcraft-backend && pwd
```

预期：输出以 `modelcraft-backend` 结尾。

- [ ] **Step 2: 破坏性重置数据库**

```bash
just db reset
```

预期：数据库删除并重新创建，所有表按新 schema 建立，无错误输出。

- [ ] **Step 3: 验证新表存在**

```bash
just db login
```

在 MySQL shell 中执行：

```sql
SHOW TABLES LIKE 'user_orgs';
SHOW TABLES LIKE 'project_roles';
SHOW TABLES LIKE 'project_role_users';
SHOW TABLES LIKE 'end_user_users';
SHOW TABLES LIKE 'end_user_accounts';
```

预期：`user_orgs`、`project_roles`、`project_role_users` 存在；`end_user_users`、`end_user_accounts` **不存在**。

```sql
DESCRIBE user_orgs;
```

预期：包含 `is_admin` 字段；`uk_user_orgs_user` 唯一约束（仅 `user_id`）存在。

退出 MySQL shell：`exit`

---

## Task 4: 更新 SQLC Query 文件 — org.sql

**Files:**
- Modify: `modelcraft-backend/db/queries/org.sql`

将所有 `user_organizations` 引用改为 `user_orgs`，加入 `is_admin` 字段的 query。

- [ ] **Step 1: 读取当前 org.sql**

```bash
cat modelcraft-backend/db/queries/org.sql
```

- [ ] **Step 2: 全局替换表名**

```bash
sed -i 's/user_organizations/user_orgs/g' modelcraft-backend/db/queries/org.sql
```

- [ ] **Step 3: 确认替换结果**

```bash
grep -n "user_organizations\|user_orgs" modelcraft-backend/db/queries/org.sql
```

预期：只出现 `user_orgs`，无 `user_organizations`。

- [ ] **Step 4: 在 org.sql 末尾追加 is_admin 相关 query**

在文件末尾添加以下内容：

```sql
-- name: GetUserOrgByUserID :one
SELECT id, user_id, org_name, is_admin, status, created_at, updated_at
FROM user_orgs
WHERE user_id = ? AND deleted_at = 0
LIMIT 1;

-- name: CreateUserOrg :exec
INSERT INTO user_orgs (id, user_id, org_name, is_admin, status)
VALUES (?, ?, ?, ?, 'active');

-- name: UpdateUserOrgAdmin :exec
UPDATE user_orgs
SET is_admin = ?, updated_at = CURRENT_TIMESTAMP(3)
WHERE user_id = ? AND org_name = ? AND deleted_at = 0;
```

- [ ] **Step 5: Commit**

```bash
cd modelcraft-backend && git add db/queries/org.sql
git commit -m "queries: update user_organizations -> user_orgs, add is_admin queries"
```

---

## Task 5: 更新 SQLC Query 文件 — rbac/user_role.sql 和 rbac/authz.sql

**Files:**
- Modify: `modelcraft-backend/db/queries/rbac/user_role.sql`
- Modify: `modelcraft-backend/db/queries/rbac/authz.sql`

- [ ] **Step 1: 读取两个文件**

```bash
cat modelcraft-backend/db/queries/rbac/user_role.sql
cat modelcraft-backend/db/queries/rbac/authz.sql
```

- [ ] **Step 2: 替换 user_role.sql 中的表名**

```bash
sed -i \
  -e 's/end_user_role_users/project_role_users/g' \
  -e 's/end_user_roles/project_roles/g' \
  -e 's/end_user_users/users/g' \
  modelcraft-backend/db/queries/rbac/user_role.sql
```

- [ ] **Step 3: 替换 authz.sql 中的表名**

```bash
sed -i \
  -e 's/end_user_role_users/project_role_users/g' \
  -e 's/end_user_roles/project_roles/g' \
  -e 's/end_user_users/users/g' \
  modelcraft-backend/db/queries/rbac/authz.sql
```

- [ ] **Step 4: 验证无残留 end_user 引用**

```bash
grep -rn "end_user_users\|end_user_role_users\|end_user_accounts" \
  modelcraft-backend/db/queries/
```

预期：**无任何输出**。

- [ ] **Step 5: 检查 end_user_roles 是否完全替换**

```bash
grep -rn "end_user_roles" modelcraft-backend/db/queries/
```

预期：**无任何输出**。

- [ ] **Step 6: Commit**

```bash
cd modelcraft-backend && git add db/queries/rbac/user_role.sql db/queries/rbac/authz.sql
git commit -m "queries: rename end_user tables to project_roles/project_role_users"
```

---

## Task 6: 重新生成 SQLC 代码

**Files:**
- Regenerate: `modelcraft-backend/internal/infrastructure/dbgen/`

- [ ] **Step 1: 运行 SQLC 代码生成**

```bash
cd modelcraft-backend && just generate-sqlc
```

预期：无错误输出，`internal/infrastructure/dbgen/` 下文件更新。

- [ ] **Step 2: 确认生成代码中无 end_user_users / end_user_accounts 引用**

```bash
grep -rn "end_user_users\|end_user_accounts\|EndUserUser\|EndUserAccount" \
  modelcraft-backend/internal/infrastructure/dbgen/
```

预期：**无任何输出**（或仅剩 end_user_roles 改名前的 comment，不影响编译）。

- [ ] **Step 3: 确认新表的 struct 已生成**

```bash
grep -rn "UserOrg\|ProjectRole\b\|ProjectRoleUser" \
  modelcraft-backend/internal/infrastructure/dbgen/
```

预期：出现 `UserOrg`、`ProjectRole`、`ProjectRoleUser` struct 定义。

- [ ] **Step 4: 尝试编译，确认 dbgen 层无报错**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/dbgen/...
```

预期：编译通过，无错误（上层 repository 层此时会有编译错误，属正常，将在 Plan 2 修复）。

- [ ] **Step 5: Commit**

```bash
cd modelcraft-backend && git add internal/infrastructure/dbgen/
git commit -m "chore: regenerate sqlc after schema rename"
```

---

## Task 7: 验证 Schema 完整性

- [ ] **Step 1: 运行 Atlas lint**

```bash
cd modelcraft-backend && just db lint
```

预期：无 lint 错误。

- [ ] **Step 2: 全局扫描残留 end_user 引用（schema + queries）**

```bash
grep -rn "end_user_users\|end_user_accounts\|end_user_role_users\|end_user_roles" \
  modelcraft-backend/db/
```

预期：**无任何输出**。（`end_user_roles` 已全部替换为 `project_roles`）

- [ ] **Step 3: 验证 user_orgs 唯一约束语义正确**

```bash
just db login
```

在 MySQL shell 中执行：

```sql
SHOW CREATE TABLE user_orgs\G
```

确认：
- `uk_user_orgs_user` 约束只包含 `(user_id, delete_token)`，不包含 `org_name`（强制单 Org 绑定）
- `is_admin` 字段存在，默认值为 `0`

退出：`exit`

- [ ] **Step 4: 最终 commit（如有未提交文件）**

```bash
cd modelcraft-backend && git status
# 如有未提交文件：
git add -A && git commit -m "schema: complete Plan 1 DB schema for unified user system"
```
