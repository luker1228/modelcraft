# RLS 数据库 Schema 终态

> RLS（行级数据隔离）模块的数据库表结构设计
> - 主 PRD: `ai-metadata/prd/rls/prd.md`
> - 后端计划: `backend-plan.md`

---

## 1. 新增表

### 1.1 model_rls_policies

Model RLS 策略配置表，与 Model 1:1 绑定。

```sql
-- 模型 RLS 策略表
CREATE TABLE model_rls_policies (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_id            BIGINT UNSIGNED NOT NULL COMMENT '模型 ID',

    -- 五件套 JSON 表达式（存储为 JSON 字符串）
    select_predicate    TEXT NOT NULL COMMENT 'SELECT USING 谓词 JSON',
    insert_check        TEXT NOT NULL COMMENT 'INSERT WITH CHECK 谓词 JSON',
    update_predicate    TEXT NOT NULL COMMENT 'UPDATE USING 谓词 JSON',
    update_check        TEXT NOT NULL COMMENT 'UPDATE WITH CHECK 谓词 JSON',
    delete_predicate    TEXT NOT NULL COMMENT 'DELETE USING 谓词 JSON',

    -- 元数据
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 外键约束
    CONSTRAINT fk_mrp_model
        FOREIGN KEY (model_id) REFERENCES models(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    -- 唯一约束：每个模型只有一个 Policy
    UNIQUE KEY uk_model_id (model_id)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Model RLS 策略配置';

-- 索引
CREATE INDEX idx_mrp_model_id ON model_rls_policies(model_id);
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `model_id` | BIGINT UNSIGNED | 关联 models.id，级联删除 |
| `select_predicate` | TEXT | SELECT USING 谓词 JSON，如 `{"owner":{"_eq":{"_auth":"uid"}}}` |
| `insert_check` | TEXT | INSERT WITH CHECK 谓词 JSON |
| `update_predicate` | TEXT | UPDATE USING 谓词 JSON |
| `update_check` | TEXT | UPDATE WITH CHECK 谓词 JSON |
| `delete_predicate` | TEXT | DELETE USING 谓词 JSON |

**数据示例**:

```json
{
  "model_id": 123,
  "select_predicate": "{\"owner\":{\"_eq\":{\"_auth\":\"uid\"}}}",
  "insert_check": "{\"owner\":{\"_eq\":{\"_auth\":\"uid\"}}}",
  "update_predicate": "{\"owner\":{\"_eq\":{\"_auth\":\"uid\"}}}",
  "update_check": "{\"owner\":{\"_eq\":{\"_auth\":\"uid\"}}}",
  "delete_predicate": "{\"owner\":{\"_eq\":{\"_auth\":\"uid\"}}}"
}
```

---

### 1.2 project_auth_schemas

Project 认证变量配置表，与 Project 1:1 绑定。

```sql
-- Project 认证变量配置表
CREATE TABLE project_auth_schemas (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id      BIGINT UNSIGNED NOT NULL COMMENT '项目 ID',

    -- 扩展变量配置（JSON 数组）
    variables       JSON NOT NULL COMMENT '认证变量列表 [{name, source, type}]',

    -- 元数据
    created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 外键约束
    CONSTRAINT fk_pas_project
        FOREIGN KEY (project_id) REFERENCES projects(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    -- 唯一约束：每个项目只有一个 AuthSchema
    UNIQUE KEY uk_project_id (project_id)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Project 认证变量配置';

-- 索引
CREATE INDEX idx_pas_project_id ON project_auth_schemas(project_id);
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `project_id` | BIGINT UNSIGNED | 关联 projects.id，级联删除 |
| `variables` | JSON | 认证变量列表，`uid` 内置不存储 |

**数据示例**:

```json
{
  "project_id": 456,
  "variables": "[{\"name\":\"tenant_id\",\"source\":\"jwt.tenant_id\",\"type\":\"uuid\"},{\"name\":\"role\",\"source\":\"jwt.role\",\"type\":\"string\"}]"
}
```

---

## 2. 修改现有表

### 2.1 fields 表说明

`field_definitions` 表的 `format` 字段为 `VARCHAR(50)` 类型（非 ENUM），新增 `END_USER_REF` 值无需 DDL 变更。

**说明**:
- `END_USER_REF` 类型字段名固定为 `owner`
- 数据库层外键约束: `REFERENCES private_{projectSlug}.users(id)`
- 应用层约束：通过 Go 代码 `FieldFormat` 枚举类型保证值合法性

---

### 2.2 models 表扩展

新增 `created_via` 字段，区分新建 Model 和导入 Model。

```sql
-- models 表新增 created_via 字段
ALTER TABLE models
    ADD COLUMN created_via ENUM('NEW', 'IMPORTED') NOT NULL DEFAULT 'NEW'
    COMMENT '模型创建来源：NEW=新建，IMPORTED=导入';

-- 为已有数据设置默认值（根据实际数据情况选择）
-- UPDATE models SET created_via = 'NEW' WHERE created_via IS NULL;

-- 添加索引
CREATE INDEX idx_models_created_via ON models(created_via);
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `created_via` | ENUM('NEW', 'IMPORTED') | 模型创建来源 |

**行为规则**:
- `NEW`: 自动添加 `owner` 字段 + 默认 Policy
- `IMPORTED`: 不自动添加 `owner` 字段，无 Policy

---

## 3. 数据库 Migration 文件

### 3.1 完整 Migration SQL

文件: `modelcraft-backend/db/schema/mysql/11_rls.sql`

```sql
-- ============================================
-- Migration: RLS (Row Level Security)
-- Description: 行级数据隔离功能
-- ============================================

-- ----------------------------------------
-- 1. models 表新增 created_via 字段
-- ----------------------------------------
ALTER TABLE models
    ADD COLUMN created_via ENUM('NEW', 'IMPORTED') NOT NULL DEFAULT 'NEW'
    COMMENT '模型创建来源：NEW=新建，IMPORTED=导入';

CREATE INDEX idx_models_created_via ON models(created_via);

-- ----------------------------------------
-- 3. 创建 model_rls_policies 表
-- ----------------------------------------
CREATE TABLE model_rls_policies (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_id            BIGINT UNSIGNED NOT NULL COMMENT '模型 ID',

    select_predicate    TEXT NOT NULL COMMENT 'SELECT USING 谓词 JSON',
    insert_check        TEXT NOT NULL COMMENT 'INSERT WITH CHECK 谓词 JSON',
    update_predicate    TEXT NOT NULL COMMENT 'UPDATE USING 谓词 JSON',
    update_check        TEXT NOT NULL COMMENT 'UPDATE WITH CHECK 谓词 JSON',
    delete_predicate    TEXT NOT NULL COMMENT 'DELETE USING 谓词 JSON',

    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    CONSTRAINT fk_mrp_model
        FOREIGN KEY (model_id) REFERENCES models(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    UNIQUE KEY uk_model_id (model_id)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Model RLS 策略配置';

CREATE INDEX idx_mrp_model_id ON model_rls_policies(model_id);

-- ----------------------------------------
-- 4. 创建 project_auth_schemas 表
-- ----------------------------------------
CREATE TABLE project_auth_schemas (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id      BIGINT UNSIGNED NOT NULL COMMENT '项目 ID',

    variables       JSON NOT NULL COMMENT '认证变量列表',

    created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    CONSTRAINT fk_pas_project
        FOREIGN KEY (project_id) REFERENCES projects(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    UNIQUE KEY uk_project_id (project_id)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Project 认证变量配置';

CREATE INDEX idx_pas_project_id ON project_auth_schemas(project_id);
```

---

## 4. 实体关系图

```
┌─────────────────┐     1:1      ┌─────────────────────┐
│     models      │──────────────│ model_rls_policies  │
├─────────────────┤   CASCADE    ├─────────────────────┤
│ id (PK)         │              │ id (PK)             │
│ name            │              │ model_id (FK, UQ)   │
│ created_via     │              │ select_predicate    │
│ ...             │              │ insert_check        │
└─────────────────┘              │ update_predicate    │
         │                       │ update_check        │
         │ 1:N                   │ delete_predicate    │
         ▼                       └─────────────────────┘
┌─────────────────┐
│     fields      │
├─────────────────┤
│ id (PK)         │
│ model_id (FK)   │
│ name            │
│ format (VARCHAR)│  ← 新增 END_USER_REF
│ ...             │
└─────────────────┘

┌─────────────────┐     1:1      ┌─────────────────────┐
│    projects     │──────────────│ project_auth_schemas│
├─────────────────┤   CASCADE    ├─────────────────────┤
│ id (PK)         │              │ id (PK)             │
│ name            │              │ project_id (FK, UQ) │
│ ...             │              │ variables (JSON)    │
└─────────────────┘              └─────────────────────┘
```

---

## 5. 与 Runtime 数据表的关联

### 5.1 EndUserRef 外键约束

对于包含 `owner` 字段（`format = END_USER_REF`）的 Model，在 Runtime 数据库中：

```sql
-- 示例：orders 表（Runtime 数据表）
CREATE TABLE private_projectslug.orders (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title       VARCHAR(255),
    amount      DECIMAL(10, 2),
    owner       VARCHAR(36) NOT NULL,  -- EndUser ID

    -- 外键约束指向 private schema 的 users 表
    CONSTRAINT fk_orders_owner
        FOREIGN KEY (owner) REFERENCES private_projectslug.users(id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
```

**说明**:
- `owner` 字段存储 EndUser ID（UUID 字符串）
- 数据库层外键约束保证数据完整性
- `ON DELETE RESTRICT` 防止误删 EndUser 导致数据孤立

---

## 6. 索引策略

| 表 | 索引 | 用途 |
|----|------|------|
| `model_rls_policies` | `uk_model_id` (UNIQUE) | 模型级 Policy 快速查询 |
| `project_auth_schemas` | `uk_project_id` (UNIQUE) | 项目级 AuthSchema 快速查询 |
| `models` | `idx_models_created_via` | 按创建来源筛选（统计/批量操作） |
| `fields` | 已有 `idx_fields_model_id` | 模型字段列表查询 |

---

## 7. 数据一致性约束

### 7.1 应用层约束

| 约束 | 实现位置 | 说明 |
|------|----------|------|
| 一个 Model 只有一个 END_USER_REF 字段 | FieldDefinition 领域逻辑 | 添加第二个时报错 `END_USER_REF_ALREADY_EXISTS` |
| 有 owner 字段必有 Policy | Model 领域逻辑 | `isRLSEnabled() == true` 时 `getPolicy() != nil` |
| 删除 owner 字段级联删除 Policy | FieldDefinition 领域逻辑 | `removeField()` 触发 Policy 删除 |

### 7.2 数据库层约束

| 约束 | 实现方式 | 说明 |
|------|----------|------|
| Model-Policy 1:1 | `uk_model_id` UNIQUE | 数据库层保证唯一性 |
| Project-AuthSchema 1:1 | `uk_project_id` UNIQUE | 数据库层保证唯一性 |
| Policy 级联删除 | `ON DELETE CASCADE` | Model 删除时自动删除 Policy |
| AuthSchema 级联删除 | `ON DELETE CASCADE` | Project 删除时自动删除 AuthSchema |

---

## 附录：SQLc 查询定义

文件: `modelcraft-backend/db/queries/rls.sql`

```sql
-- name: GetModelRLSPolicy :one
SELECT * FROM model_rls_policies
WHERE model_id = ?;

-- name: UpsertModelRLSPolicy :exec
INSERT INTO model_rls_policies (
    model_id, select_predicate, insert_check,
    update_predicate, update_check, delete_predicate
) VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    select_predicate = VALUES(select_predicate),
    insert_check = VALUES(insert_check),
    update_predicate = VALUES(update_predicate),
    update_check = VALUES(update_check),
    delete_predicate = VALUES(delete_predicate);

-- name: DeleteModelRLSPolicy :exec
DELETE FROM model_rls_policies
WHERE model_id = ?;

-- name: GetProjectAuthSchema :one
SELECT * FROM project_auth_schemas
WHERE project_id = ?;

-- name: UpsertProjectAuthSchema :exec
INSERT INTO project_auth_schemas (project_id, variables)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
    variables = VALUES(variables);

-- name: DeleteProjectAuthSchema :exec
DELETE FROM project_auth_schemas
WHERE project_id = ?;
```
