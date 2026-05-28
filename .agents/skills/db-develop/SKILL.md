---
name: db-develop
description: >
  ModelCraft 数据库开发与管理指南。涵盖数据库连接方式、Schema 文件结构、
  Atlas 迁移原理以及 just db 命令完整用法。当用户涉及以下任何内容时，必须使用此 skill：
  (1) 数据库如何登录、连接、查看表结构，
  (2) 修改或新增 db/schema/mysql/*.sql 表结构，
  (3) 使用 just db up / reset / status / login 等命令同步数据库，
  (4) 理解 Atlas schema apply 的工作原理，
  (5) 询问哪个 SQL 文件负责哪个业务域，
  (6) 数据库相关报错、迁移失败、表不存在等问题排查。
  遇到任何 "数据库"、"建表"、"迁移"、"just db"、"atlas"、"schema" 相关问题时主动触发。
---

# ModelCraft 数据库开发指南

## 快速参考

```bash
just db up          # ⭐ 最常用：同步最新 schema 到数据库
just db status      # 查看当前数据库状态和表列表
just db login       # 进入 MySQL 交互 CLI
just db reset       # 危险：清空重建（开发调试用）
```

---

## 环境识别：先读 .env 文件

**在执行任何数据库操作前，先确认当前使用的是哪个 `.env` 文件**，因为不同环境指向不同数据库。

### 环境文件一览

| 文件 | 用途 | 数据库名 | 端口 |
|------|------|----------|------|
| `.env` | **本地开发**（默认） | `modelcraft` | `3307` |
| `.env.dev` | 本地开发（同 `.env`，备用） | `modelcraft` | `3307` |
| `.env.autotest` | 集成测试 | `modelcraft_test` | `3307` |
| `.env.docker.example` | Docker 部署参考模板 | `modelcraft` | `3306`（容器内） |

> `.env.*.example` 是模板文件，需复制并去掉 `.example` 后缀才能使用。

### 如何判断当前环境

`.env` 是一个**符号链接**，指向当前激活的环境文件。用 `just env-current` 查看：

```bash
just env-current
# 输出示例：
# ✅ Current environment: dev
#    Link: .env → .env.dev
```

其他环境管理命令：

```bash
just env-list              # 列出所有可用环境文件
just env-current           # 查看当前激活的环境
just env-switch dev        # 切换到 .env.dev（更新符号链接）
just env-switch autotest   # 切换到 .env.autotest（测试环境）
just env-diff autotest     # 对比当前 .env 与 .env.autotest 的差异
```

### just db 默认读取 `.env`，可以覆盖

```bash
just db up                   # 默认读取 .env（当前激活环境）
just db up .env.autotest     # 临时指定读取测试环境的配置
just db status .env.autotest # 查看测试库状态
just db login .env.autotest  # 登录测试库
```

---

## 数据库连接信息

### 本地开发环境（`.env` 默认）

| 参数 | 值 | 环境变量 |
|------|-----|----------|
| Host | `127.0.0.1` | `DB_HOST` |
| Port | `3307` | `DB_PORT` |
| User | `root` | `DB_USER` |
| Password | `modelcraft123` | `DB_PASSWORD` |
| Database | `modelcraft` | `DB_NAME` |

> **注意**：本地开发 MySQL 端口为 `3307`（非标准 3306），避免与系统 MySQL 冲突。

### 集成测试环境（`.env.autotest`）

`.env.autotest` 与 `.env` 连接参数相同，唯一区别是 `DB_NAME=modelcraft_test`，使用独立数据库隔离测试数据。

### 直接登录方式

```bash
# 方式 1：通过 just（推荐，自动读取 .env）
just db login

# 方式 2：登录测试数据库
just db login .env.autotest

# 方式 3：直接用 mysql 客户端
mysql -h 127.0.0.1 -P 3307 -u root -pmodelcraft123 modelcraft

# 方式 4：通过 Docker exec（当 mysql 客户端未安装时）
docker exec -it modelcraft-mysql-local mysql -h 127.0.0.1 -P 3306 -u root -pmodelcraft123 modelcraft
```

### Docker 环境

| 服务 | 容器名 | 宿主机端口 |
|------|--------|------------|
| MySQL 8.0 | `modelcraft-mysql` | `6033:3306` |
| Redis 7 | `modelcraft-redis` | `6379:6379` |

Docker 环境中 MySQL 通过 `./db/schema/mysql/` 目录在容器首次启动时自动初始化（`/docker-entrypoint-initdb.d`）。

---

## Schema 文件结构

Schema 定义在 `db/schema/mysql/`，**按编号顺序执行**：

| 文件 | 业务域 | 关键表 |
|------|--------|--------|
| `01_project.sql` | 项目域 | `projects`（复合主键：org_name + slug） |
| `02_database_cluster.sql` | 数据库集群 | `database_clusters` |
| `03_model_domain.sql` | 模型域 | `models`、`field_definitions`、`model_enums`、`logical_foreign_keys`、`model_groups` |
| `04_auth.sql` | 认证配置 | `project_auth_configs`（OIDC 配置） |
| `05_organizations.sql` | 组织域 | `organizations` |
| `06_users.sql` | 用户域 | `users`、`user_organizations` |
| `07_roles_permissions.sql` | 权限域（Casbin） | `roles`、`user_roles`、`role_permissions` |
| `08_refresh_tokens.sql` | Token 管理 | `refresh_tokens`（存 SHA256 hash，不存明文） |
| `09_api_keys.sql` | API Key | `api_keys`（hash + prefix） |
| `10_security_audit_logs.sql` | 安全审计 | `security_audit_logs` |

### 关键设计特征

- **多租户复合主键**：`projects` 使用 `(org_name, slug)` 作为复合主键，其他表通过 `(org_name, project_slug)` 引用
- **幂等建表**：所有表使用 `CREATE TABLE IF NOT EXISTS`，重复执行安全
- **所有表**：`ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

## 强制约束：唯一索引字段必须 `NOT NULL`

> **硬性规则**：凡是参与 `PRIMARY KEY` / `UNIQUE KEY`（含联合唯一索引）的字段，必须声明为 `NOT NULL`。

原因：MySQL 里 `NULL != NULL`，联合唯一索引中出现 `NULL` 时，可能出现“看似重复却不冲突”的脏数据。

```sql
-- ❌ 错误示例：c 允许 NULL，唯一约束会失效
UNIQUE KEY uk_abc (a, b, c)

-- 可能同时插入成功
(a, b, c) = (1, 2, NULL)
(a, b, c) = (1, 2, NULL)
```

```sql
-- ✅ 正确示例：唯一索引字段全部 NOT NULL
a BIGINT NOT NULL,
b BIGINT NOT NULL,
c BIGINT NOT NULL,
UNIQUE KEY uk_abc (a, b, c)
```

评审清单：
1. 新增/修改唯一索引时，逐个检查索引列是否 `NOT NULL`
2. 不要依赖 `NULL` 参与唯一约束表达业务语义
3. 软删除场景继续使用 `delete_token`（`NOT NULL DEFAULT 0`）做唯一避让

## 强制约束：字段默认值策略

> **硬性规则**：字段应尽量设置 `DEFAULT` 值；默认值必须是“系统保留值”，不得与业务有效值冲突。

目的：降低写入遗漏导致的 `NULL`/空值问题，同时避免默认值被误当成真实业务数据。

```sql
-- ❌ 错误示例：默认值与业务值冲突
status VARCHAR(32) NOT NULL DEFAULT 'active'
-- active 是真实业务状态，无法区分“未显式赋值”与“用户真实设置为 active”
```

```sql
-- ✅ 正确示例：使用保留默认值（不与业务值冲突）
status VARCHAR(32) NOT NULL DEFAULT '__UNSET__'
```

```sql
-- ✅ 正确示例：软删除/时间戳场景使用约定哨兵值
deleted_at BIGINT UNSIGNED NOT NULL DEFAULT 0
```

评审清单：
1. 新增字段时，优先提供 `NOT NULL + DEFAULT`
2. 先定义该字段的“业务值集合”，再选择不冲突的默认值
3. 若无法提供安全默认值，再允许 `NULL`，并在应用层显式处理
4. 禁止用会误导业务语义的默认值（如真实状态、真实枚举值）

## 强制约束：TEXT / BLOB 类型不支持 DEFAULT

> **硬性规则**：MySQL 的 `TEXT`、`BLOB`（及其变体 `TINYTEXT`、`MEDIUMTEXT`、`LONGTEXT` 等）**不支持 `DEFAULT` 值**。写入会报错或被静默忽略，取决于 MySQL 版本和 sql_mode。

```sql
-- ❌ 错误示例：TEXT 字段设置 DEFAULT ''（MySQL 报错或忽略）
`description` TEXT NOT NULL DEFAULT ''

-- ✅ 正确方案一：改用 VARCHAR，长度按业务需求选择
`description` VARCHAR(2000) NOT NULL DEFAULT ''

-- ✅ 正确方案二：允许 NULL（应用层保证非空）
`description` TEXT NULL
```

**选择依据**：
- 描述性短文本（≤ 2000 字符）→ 用 `VARCHAR(N) NOT NULL DEFAULT ''`，语义清晰，支持默认值
- 任意长文本（文章内容、日志等）→ 用 `TEXT NULL`，应用层判空处理

评审清单：
1. 新增 TEXT / BLOB 字段时，不要写 `DEFAULT ''` 或任何 DEFAULT 值
2. 若需要默认空字符串语义，改用 `VARCHAR`
3. 确保 GraphQL / Repository 层与 DDL 的 NULL 约定一致（`String!` 对应 `NOT NULL`，`String` 对应 `NULL`）

---

## 强制约束：软删除字段规则（`deleted_at` / `delete_token`）

### 核心字段定义

```sql
deleted_at   BIGINT UNSIGNED NOT NULL DEFAULT 0,
delete_token BIGINT UNSIGNED NOT NULL DEFAULT 0
```

语义约定：
- `deleted_at = 0`：活跃数据
- `deleted_at > 0`：已软删除（Unix 毫秒时间戳）
- `delete_token = 0`：活跃数据
- `delete_token > 0`：墓碑记录唯一逃逸值（仅用于唯一索引避让）

### 查询与写入规则

读路径（`SELECT` / `COUNT` / `EXISTS` / `JOIN` / 子查询）必须显式过滤：

```sql
deleted_at = 0
```

更新活跃数据必须带条件：

```sql
... WHERE ... AND deleted_at = 0
```

软删除禁止物理 `DELETE`，统一使用：

```sql
UPDATE <table>
SET deleted_at = <unix_millis_expr>,
    delete_token = <unique_token_expr>
WHERE ... AND deleted_at = 0
```

### 与唯一索引联动

若业务键需要“删后重建”，唯一键必须包含 `delete_token`：

```sql
UNIQUE KEY uk_xxx (<biz_columns...>, delete_token)
```

这样活跃数据（`delete_token=0`）保持唯一，墓碑数据可并存。

### 例外（黑名单机制）

黑名单表可保持物理删除，不强制要求 `deleted_at`；以 `modelcraft-backend/db/soft_delete.yaml` 配置为准。

---

## just db 命令详解

所有命令在 `modelcraft-backend/` 目录下执行。

### 常用命令

```bash
# 应用 schema（最常用，增量更新）
just db up

# 指定 env 文件（如使用测试环境）
just db up .env.autotest

# 查看数据库状态（连接信息 + 表列表 + schema 文件列表）
just db status

# 登录 MySQL CLI
just db login

# 创建数据库（首次或数据库不存在时）
just db create

# 重置数据库（清空所有表，重新应用 schema）
just db reset   # ⚠️ 危险操作！会丢失所有数据
```

### 完整命令列表

| 命令 | 说明 |
|------|------|
| `just db create` | 创建数据库（若不存在） |
| `just db drop` | 删除数据库 |
| `just db up` | 应用 schema（默认，使用 Atlas） |
| `just db down` | 回滚提示（需手动操作） |
| `just db status` | 查看状态（连接 + 表列表） |
| `just db reset` | 重置（drop + recreate + up） |
| `just db lint` | 检查迁移文件 |
| `just db diff` | 生成迁移 diff |
| `just db login` | 进入 MySQL 交互 CLI |

---

## Atlas 工作原理

`just db up` 底层调用：

```bash
atlas schema apply \
  -u "mysql://root:password@127.0.0.1:3307/modelcraft" \
  --to file://db/schema/mysql/ \
  --dev-url "mysql://root:password@127.0.0.1:3307/modelcraft_dev" \
  --auto-approve
```

### Atlas Schema Apply 原理

1. **读取期望状态**：从 `db/schema/mysql/*.sql` 文件中解析目标 schema（使用 `--dev-url` 临时数据库解析 SQL）
2. **读取当前状态**：连接 `-u` 指定的生产数据库，检查现有表结构
3. **计算 Diff**：对比期望状态与当前状态，生成最小变更集（ALTER TABLE、CREATE TABLE 等）
4. **执行变更**：`--auto-approve` 自动执行，不需手动确认

### Dev URL 的作用

`DB_DEV_URL`（`modelcraft_dev` 数据库）是 Atlas 用于**解析 SQL 文件**的临时沙盒：
- Atlas 先把 schema SQL 应用到 `_dev` 数据库，获取规范化的 schema 对象
- 再与目标数据库对比
- `_dev` 数据库会被 `just db create` 自动创建

### 与传统 Migration 的区别

| 特性 | Atlas Schema Apply（当前） | 传统 Migration（如 Flyway/Liquibase） |
|------|---------------------------|--------------------------------------|
| Schema 来源 | 声明式 SQL 文件（期望状态） | 顺序执行的迁移文件 |
| 执行方式 | 自动计算 diff，增量变更 | 按版本号顺序执行 |
| 幂等性 | 天然幂等 | 需要版本控制 |
| 回滚 | 手动写 alter | 有版本回滚机制 |

---

## 修改 Schema 的工作流

```bash
# 1. 编辑对应的 SQL 文件
vim db/schema/mysql/03_model_domain.sql

# 2. 同步到数据库
just db up

# 3. 验证变更
just db status
```

如果需要在应用层同步查询代码（sqlc），还需要：

```bash
# 4. 同步 sqlc 生成代码（如果有 SQL 查询变更）
just generate-sqlc
```

---

## 常见问题

### 数据库不存在 / 连接失败

```bash
just db create    # 创建数据库
just db up        # 应用 schema
```

### Atlas 未安装

`just db up` 会自动安装 Atlas：
```bash
curl -sSf https://atlasgo.sh | sh
```
也可手动安装：`just install-atlas`

### 需要完全重置（开发调试）

```bash
just db reset     # ⚠️ 会清空所有数据
```

### 查看当前表结构

```bash
just db login
# 进入 MySQL 后：
SHOW TABLES;
DESCRIBE models;
```
