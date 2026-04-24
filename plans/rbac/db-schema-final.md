# RBAC 行列级权限系统 — DB Schema 终态文档

> **版本**: v2.1  
> **日期**: 2026-04-24  
> **适用范围**: ModelCraft Project 维度终端用户行列级数据权限  
> **迁移文件**: `db/schema/mysql/13_rbac_permissions.sql`

---

## 1. 概述

本方案在 `12_end_user_auth.sql` 已有表结构的基础上，**最小侵入地**扩展 RBAC 权限体系。

### 变更范围

| 操作 | 对象 | 说明 |
|------|------|------|
| **ALTER** | `end_user_roles` | 追加 `is_implicit` 列，标识内置隐式角色 |
| **CREATE** | `end_user_permissions` | 权限点（模型 × 动作 × 行列策略） |
| **CREATE** | `end_user_permission_bundles` | 权限包（权限点的命名集合） |
| **CREATE** | `end_user_bundle_permissions` | 权限包 ↔ 权限点 有序中间表 |
| **CREATE** | `end_user_role_bundles` | 角色 ↔ 权限包 关联表 |
| **CREATE** | `end_user_user_bundles` | 用户直接授权 ↔ 权限包 关联表 |

### 与现有 end_user 表的关系

```
12_end_user_auth.sql（已有，不可修改 DDL）
  ├── end_user_users          ← user_id 来源
  ├── end_user_accounts       ← 登录账号，不直接参与鉴权
  ├── end_user_roles          ← ALTER 追加 is_implicit 列
  └── end_user_role_users     ← 用户-角色关联（隐式角色不向此表插行）

13_rbac_permissions.sql（本文件新增）
  ├── end_user_permissions         ← 权限点
  ├── end_user_permission_bundles  ← 权限包
  ├── end_user_bundle_permissions  ← 权限包-权限点 中间表
  ├── end_user_role_bundles        ← 角色-权限包 关联
  └── end_user_user_bundles        ← 用户直接授权-权限包 关联
```

### 设计原则

1. **不对 `projects` 使用 FK 约束** — Project 永不删除只归档，改用 INDEX
2. **`end_user_*` 表复合 FK 模式** — 跨表引用 `end_user_users` 时使用 `(org_name, project_slug, id)` 三元复合键
3. **时间字段** — `end_user_*` 新表统一用 `DATETIME`（无精度），与 `12_end_user_auth.sql` 保持一致
4. **隐式角色** — `is_implicit=1` 的角色由运行时自动注入，不在 `end_user_role_users` 维护归属关系

---

## 2. 完整 SQL DDL

**文件**: `db/schema/mysql/13_rbac_permissions.sql`

```sql
-- =============================================================
-- 13_rbac_permissions.sql
-- RBAC 行列级权限系统
-- 依赖: 12_end_user_auth.sql（end_user_roles / end_user_users）
-- =============================================================

-- -------------------------------------------------------------
-- 0. ALTER 现有表：end_user_roles 追加 is_implicit 列
-- -------------------------------------------------------------
ALTER TABLE `end_user_roles`
  ADD COLUMN `is_implicit` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '内置隐式角色标志：0=显式角色（用户手动分配），1=隐式角色（系统自动注入）',
  ADD INDEX `idx_end_user_roles_implicit` (`is_implicit`);

-- -------------------------------------------------------------
-- 1. end_user_permissions — 权限点
--    一行 = 一个「模型 × 动作 × 行范围」的细粒度权限
-- -------------------------------------------------------------
CREATE TABLE `end_user_permissions` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限点 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，不做 FK）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，不做 FK）',
  `model_id`      VARCHAR(36)  NOT NULL                    COMMENT '关联模型 ID，FK → models.id',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限点名称，人类可读',
  `description`   TEXT         NULL                        COMMENT '权限点描述',
  `action`        ENUM(
                    'select',
                    'insert',
                    'update',
                    'delete',
                    'export'
                  )            NOT NULL                    COMMENT '操作动作',
  `column_policy` JSON         NULL                        COMMENT '列策略 JSON，结构见注释',
  -- column_policy 结构示例：
  -- {
  --   "defaultMode": "VISIBLE",          // VISIBLE | HIDDEN | MASKED
  --   "rules": [
  --     { "fieldName": "salary", "mode": "MASKED", "maskPattern": "***" },
  --     { "fieldName": "id_card", "mode": "HIDDEN" }
  --   ]
  -- }
  `row_scope`     ENUM(
                    'ALL',
                    'SELF',
                    'DEPT',
                    'DEPT_AND_CHILDREN'
                  )            NOT NULL DEFAULT 'ALL'      COMMENT '行范围：ALL全量/SELF本人/DEPT本部门/DEPT_AND_CHILDREN含子部门',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 业务唯一键：同一模型下，相同动作+行范围+名称不可重复
  UNIQUE KEY `uq_permissions_model_action_scope_name`
    (`model_id`, `action`, `row_scope`, `name`),
  -- 按组织/项目快速检索（不做 FK，Project 不删除只归档）
  INDEX `idx_permissions_org_project` (`org_name`, `project_slug`),
  INDEX `idx_permissions_model_id` (`model_id`),
  -- FK → models.id（模型可删除，级联清理权限点）
  CONSTRAINT `fk_permissions_model`
    FOREIGN KEY (`model_id`) REFERENCES `models` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限点：每行描述对某模型某动作的行列级权限配置';

-- -------------------------------------------------------------
-- 2. end_user_permission_bundles — 权限包
--    权限点的命名集合，可跨模型聚合
-- -------------------------------------------------------------
CREATE TABLE `end_user_permission_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限包 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限包名称',
  `description`   TEXT         NULL                        COMMENT '权限包描述',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一项目下权限包名称唯一
  UNIQUE KEY `uq_bundles_org_project_name`
    (`org_name`, `project_slug`, `name`),
  -- 快速检索（不做 FK → projects）
  INDEX `idx_bundles_org_project` (`org_name`, `project_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包：权限点的命名集合，用于角色授权或用户直接授权';

-- -------------------------------------------------------------
-- 3. end_user_bundle_permissions — 权限包-权限点 有序中间表
-- -------------------------------------------------------------
CREATE TABLE `end_user_bundle_permissions` (
  `id`             VARCHAR(36) NOT NULL                    COMMENT 'UUID',
  `bundle_id`      VARCHAR(36) NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `permission_id`  VARCHAR(36) NOT NULL                    COMMENT '权限点 ID，FK → end_user_permissions.id',
  `sort_order`     INT         NOT NULL DEFAULT 0          COMMENT '显示排序权重（ASC）',
  `created_at`     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一权限包内权限点不重复
  UNIQUE KEY `uq_bundle_permissions_bundle_perm`
    (`bundle_id`, `permission_id`),
  INDEX `idx_bundle_permissions_permission_id` (`permission_id`),
  CONSTRAINT `fk_bundle_permissions_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_bundle_permissions_permission`
    FOREIGN KEY (`permission_id`) REFERENCES `end_user_permissions` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包-权限点 有序中间表';

-- -------------------------------------------------------------
-- 4. end_user_role_bundles — 角色-权限包 关联
-- -------------------------------------------------------------
CREATE TABLE `end_user_role_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，快速查询）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，快速查询）',
  `role_id`       VARCHAR(36)  NOT NULL                    COMMENT '角色 ID，FK → end_user_roles.id',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一角色不可重复授予同一权限包
  UNIQUE KEY `uq_role_bundles_role_bundle`
    (`role_id`, `bundle_id`),
  INDEX `idx_role_bundles_org_project` (`org_name`, `project_slug`),
  INDEX `idx_role_bundles_bundle_id` (`bundle_id`),
  CONSTRAINT `fk_role_bundles_role`
    FOREIGN KEY (`role_id`) REFERENCES `end_user_roles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_role_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='角色-权限包 关联：角色持有哪些权限包';

-- -------------------------------------------------------------
-- 5. end_user_user_bundles — 用户直接授权-权限包 关联
--    用户可绕过角色直接获得权限包（最高优先级通道）
-- -------------------------------------------------------------
CREATE TABLE `end_user_user_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `user_id`       VARCHAR(36)  NOT NULL                    COMMENT '用户 ID（复合 FK 的 id 部分）',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一用户在同一项目内不可重复获得同一权限包
  UNIQUE KEY `uq_user_bundles_org_project_user_bundle`
    (`org_name`, `project_slug`, `user_id`, `bundle_id`),
  INDEX `idx_user_bundles_bundle_id` (`bundle_id`),
  -- 复合 FK → end_user_users(org_name, project_slug, id)
  CONSTRAINT `fk_user_bundles_user`
    FOREIGN KEY (`org_name`, `project_slug`, `user_id`)
    REFERENCES `end_user_users` (`org_name`, `project_slug`, `id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户直接授权-权限包：绕过角色直接给用户授予权限包';
```

---

## 3. ER 关系图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  已有表（12_end_user_auth.sql，不可修改 DDL）                                  │
│                                                                             │
│  end_user_users                    end_user_roles [ALTER: +is_implicit]     │
│  ─────────────────                 ──────────────────────────────────       │
│  PK id                             PK id                                    │
│  org_name                          org_name                                 │
│  project_slug                      project_slug                             │
│  UNIQUE(org_name,project_slug,id)  name                                     │
│          │                         is_implicit ← 新增                       │
│          │                              │                                   │
│          └──────────┐       ┌───────────┘                                   │
│                     │       │                                               │
│              end_user_role_users                                            │
│              ────────────────────                                           │
│              user_id ─────────────── FK(org_name,project_slug,id)→users    │
│              role_id ─────────────── FK → end_user_roles.id                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                     │                         │
                     │ (via role_id)            │ (via user_id 复合 FK)
                     ▼                         ▼
        ┌──────────────────────┐   ┌────────────────────────┐
        │ end_user_role_bundles│   │ end_user_user_bundles   │
        │ ─────────────────────│   │ ────────────────────────│
        │ PK id                │   │ PK id                   │
        │ org_name             │   │ org_name                │
        │ project_slug         │   │ project_slug            │
        │ role_id ──FK─────────┘   │ user_id ─FK(复合)──────┘
        │ bundle_id ──┐            │ bundle_id ──┐           │
        │ granted_at  │            │ granted_at  │           │
        └─────────────┼────────────┘             │           │
                      │                          │           │
                      └──────────┬───────────────┘           │
                                 ▼                           │
              ┌──────────────────────────────────────┐       │
              │    end_user_permission_bundles        │       │
              │    ──────────────────────────────     │       │
              │    PK id                              │       │
              │    org_name  (INDEX，不 FK projects)  │       │
              │    project_slug                       │       │
              │    name (UNIQUE per project)          │       │
              └──────────────────┬────────────────────┘       │
                                 │                            │
                                 │ 1:N                        │
                                 ▼                            │
              ┌──────────────────────────────────────┐       │
              │   end_user_bundle_permissions         │       │
              │   ────────────────────────────────    │       │
              │   PK id                               │       │
              │   bundle_id ─── FK → bundles.id       │       │
              │   permission_id ─── FK → perms.id     │       │
              │   sort_order                          │       │
              └──────────────────┬────────────────────┘       │
                                 │                            │
                                 ▼                            │
              ┌──────────────────────────────────────┐       │
              │      end_user_permissions             │       │
              │      ────────────────────────         │       │
              │      PK id                            │       │
              │      org_name (INDEX)                 │       │
              │      project_slug                     │       │
              │      model_id ─── FK → models.id      │       │
              │      name                             │       │
              │      action (ENUM)                    │       │
              │      column_policy (JSON)             │       │
              │      row_scope (ENUM)                 │       │
              └───────────────────────────────────────┘       │
                                                              │
                      models（外部表，已有）                    │
                      ─────────────────────                   │
                      PK id  ◄── FK from permissions          │
```

---

## 4. 鉴权数据流（3 通道）

```
  用户请求 GraphQL 查询
        │
        ▼
  ┌─────────────────────────────────────────────────────────────┐
  │                    鉴权中间件                                  │
  │  从 JWT / Session 取出: org_name, project_slug, user_id      │
  └──────────────────────┬──────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────────┐
         ▼               ▼                   ▼
  ┌────────────┐  ┌──────────────┐  ┌─────────────────┐
  │  通道 A    │  │   通道 B     │  │    通道 C       │
  │ 用户直接   │  │  显式角色    │  │   隐式角色      │
  │ 授权权限包 │  │  持有权限包  │  │   持有权限包    │
  └─────┬──────┘  └──────┬───────┘  └────────┬────────┘
        │                │                   │
        │  end_user_      │  end_user_role_   │  is_implicit=1
        │  user_bundles   │  users JOIN        │  角色由 runtime
        │  WHERE          │  end_user_roles    │  按规则自动注入
        │  user_id=?      │  JOIN role_bundles │  (OWNER/ALL等)
        │                │                   │
        └────────────────┼───────────────────┘
                         │  UNION 三通道 bundle_id 集合
                         ▼
  ┌──────────────────────────────────────────────────────────────┐
  │  SELECT DISTINCT permission_id                               │
  │  FROM end_user_bundle_permissions                            │
  │  WHERE bundle_id IN (通道A ∪ 通道B ∪ 通道C)                  │
  └──────────────────────┬───────────────────────────────────────┘
                         │
                         ▼
  ┌──────────────────────────────────────────────────────────────┐
  │  SELECT action, column_policy, row_scope                     │
  │  FROM end_user_permissions                                   │
  │  WHERE id IN (上步结果) AND model_id = ?                     │
  └──────────────────────┬───────────────────────────────────────┘
                         │
         ┌───────────────┴──────────────┐
         ▼                              ▼
  ┌─────────────┐               ┌───────────────┐
  │  行范围过滤  │               │  列策略过滤   │
  │ row_scope   │               │ column_policy │
  │ ALL/SELF/   │               │ VISIBLE/HIDDEN│
  │ DEPT/...    │               │ /MASKED       │
  └─────────────┘               └───────────────┘
         │                              │
         └──────────────┬───────────────┘
                        ▼
               最终数据集返回给用户
```

---

## 5. 迁移步骤

### 前置检查

```bash
# 确认 12_end_user_auth.sql 已应用（以下表必须存在）
cd modelcraft-backend
just db login
# 在 MySQL shell 中执行：
# SHOW TABLES LIKE 'end_user_%';
# 应看到 end_user_users / end_user_accounts / end_user_roles / end_user_role_users
```

### 创建迁移文件

```bash
# 新建 SQL 文件（按实际路径）
cp /dev/null db/schema/mysql/13_rbac_permissions.sql
# 将上方 DDL 粘贴进去
```

### 应用迁移

```bash
# Atlas schema apply（just db up 内部调用）
just db up

# 验证表已创建
just db login
# SHOW TABLES LIKE 'end_user_%';
# 应多出: end_user_permissions / end_user_permission_bundles /
#         end_user_bundle_permissions / end_user_role_bundles /
#         end_user_user_bundles

# 验证 ALTER 已生效
# DESCRIBE end_user_roles;
# 应看到 is_implicit 列
```

### 回滚方案（如需）

```sql
-- 回滚顺序：先删子表，后删父表，最后回滚 ALTER
DROP TABLE IF EXISTS `end_user_user_bundles`;
DROP TABLE IF EXISTS `end_user_role_bundles`;
DROP TABLE IF EXISTS `end_user_bundle_permissions`;
DROP TABLE IF EXISTS `end_user_permission_bundles`;
DROP TABLE IF EXISTS `end_user_permissions`;
ALTER TABLE `end_user_roles`
  DROP INDEX `idx_end_user_roles_implicit`,
  DROP COLUMN `is_implicit`;
```

---

## 6. sqlc 查询文件规划

所有查询文件放在 `db/queries/` 下，文件名与表名对齐（`end_user_*` 前缀）。

### 文件清单

| 文件 | 对应表 | 核心查询 |
|------|--------|--------|
| `end_user_permissions.sql` | `end_user_permissions` | CRUD + 按 model_id 检索 |
| `end_user_permission_bundles.sql` | `end_user_permission_bundles` | CRUD + 按 org/project 列举 |
| `end_user_bundle_permissions.sql` | `end_user_bundle_permissions` | 包内权限点增删 + 有序列举 |
| `end_user_role_bundles.sql` | `end_user_role_bundles` | 角色-包 关联管理 |
| `end_user_user_bundles.sql` | `end_user_user_bundles` | 用户直接授权管理 |
| `end_user_rbac_resolve.sql` | 跨表 JOIN | 鉴权核心查询（3 通道合并） |

### 关键查询示例（end_user_rbac_resolve.sql）

```sql
-- name: ResolveUserPermissions :many
-- 给定 user_id，返回其在指定 model 上的所有有效权限点（3 通道合并）
SELECT DISTINCT
  p.id,
  p.action,
  p.column_policy,
  p.row_scope
FROM end_user_permissions p
JOIN end_user_bundle_permissions bp ON bp.permission_id = p.id
WHERE p.model_id = sqlc.arg(model_id)
  AND bp.bundle_id IN (

  -- 通道 A：用户直接授权
  SELECT bundle_id
  FROM end_user_user_bundles
  WHERE org_name      = sqlc.arg(org_name)
    AND project_slug  = sqlc.arg(project_slug)
    AND user_id       = sqlc.arg(user_id)

  UNION

  -- 通道 B：显式角色
  SELECT rb.bundle_id
  FROM end_user_role_bundles rb
  JOIN end_user_role_users ru
    ON ru.role_id = rb.role_id
  JOIN end_user_roles r
    ON r.id = ru.role_id
   AND r.is_implicit = 0
  WHERE ru.org_name      = sqlc.arg(org_name)
    AND ru.project_slug  = sqlc.arg(project_slug)
    AND ru.user_id       = sqlc.arg(user_id)

  UNION

  -- 通道 C：隐式角色（由 runtime 注入，直接按 role_id 查）
  SELECT rb.bundle_id
  FROM end_user_role_bundles rb
  JOIN end_user_roles r ON r.id = rb.role_id
  WHERE r.is_implicit   = 1
    AND rb.org_name     = sqlc.arg(org_name)
    AND rb.project_slug = sqlc.arg(project_slug)
    AND rb.role_id IN (sqlc.slice(implicit_role_ids))
);

-- name: ListBundlePermissions :many
-- 列出权限包内所有权限点（按 sort_order 排序）
SELECT
  p.id,
  p.name,
  p.action,
  p.column_policy,
  p.row_scope,
  bp.sort_order
FROM end_user_bundle_permissions bp
JOIN end_user_permissions p ON p.id = bp.permission_id
WHERE bp.bundle_id = sqlc.arg(bundle_id)
ORDER BY bp.sort_order ASC, p.name ASC;

-- name: GrantBundleToRole :exec
INSERT INTO end_user_role_bundles
  (id, org_name, project_slug, role_id, bundle_id, granted_at)
VALUES
  (sqlc.arg(id), sqlc.arg(org_name), sqlc.arg(project_slug),
   sqlc.arg(role_id), sqlc.arg(bundle_id), NOW());

-- name: GrantBundleToUser :exec
INSERT INTO end_user_user_bundles
  (id, org_name, project_slug, user_id, bundle_id, granted_at)
VALUES
  (sqlc.arg(id), sqlc.arg(org_name), sqlc.arg(project_slug),
   sqlc.arg(user_id), sqlc.arg(bundle_id), NOW());

-- name: RevokeRoleBundle :exec
DELETE FROM end_user_role_bundles
WHERE role_id = sqlc.arg(role_id) AND bundle_id = sqlc.arg(bundle_id);

-- name: RevokeUserBundle :exec
DELETE FROM end_user_user_bundles
WHERE org_name     = sqlc.arg(org_name)
  AND project_slug = sqlc.arg(project_slug)
  AND user_id      = sqlc.arg(user_id)
  AND bundle_id    = sqlc.arg(bundle_id);
```

### sqlc.yaml 配置追加

在现有 `sqlc.yaml` 的 `queries` 列表中追加：

```yaml
queries:
  # ... 现有配置 ...
  - "db/queries/end_user_permissions.sql"
  - "db/queries/end_user_permission_bundles.sql"
  - "db/queries/end_user_bundle_permissions.sql"
  - "db/queries/end_user_role_bundles.sql"
  - "db/queries/end_user_user_bundles.sql"
  - "db/queries/end_user_rbac_resolve.sql"
```

运行生成：

```bash
just generate-sqlc
```

---

## 7. 关键设计决策备注

| 决策 | 选择 | 原因 |
|------|------|------|
| 角色表复用 | ALTER `end_user_roles` 加列 | 避免语义重复表，`is_implicit` 是角色自身属性 |
| 命名前缀 | `end_user_*` | 与 `12_end_user_auth.sql` 风格一致 |
| projects FK | 不做，改 INDEX | Project 永不删除只归档，FK 无法级联处理归档场景 |
| 用户 FK 模式 | 复合三元 `(org_name, project_slug, id)` | 匹配 `end_user_users` 的 UNIQUE 约束 |
| 时间精度 | `DATETIME`（无精度） | 与已有 end_user 表保持一致，避免跨表 JOIN 类型不一致 |
| 隐式角色注入 | 运行时注入，不写 `role_users` 表 | 隐式角色成员关系由业务规则决定（如 Project Owner），非静态数据 |
