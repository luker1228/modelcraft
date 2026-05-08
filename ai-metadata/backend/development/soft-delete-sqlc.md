# MySQL + sqlc 软删除规范（`deleted_at` / `delete_token`）

> 适用范围：`modelcraft-backend` 的 MySQL 表结构与 `db/queries/**/*.sql` 查询定义。

## 目标

- 统一软删除语义，避免物理删除带来的审计与恢复困难。
- 保证“源码 SQL 就是实际 SQL”，不依赖运行时隐式拼接条件。
- 兼容 `sqlc`（含 `sqlc.slice(...)`）生成流程。

## 核心字段约定

### `deleted_at`

- 类型：`BIGINT UNSIGNED NOT NULL DEFAULT 0`
- 语义：
- `0`：活跃数据
- `> 0`：已软删除（Unix 毫秒时间戳）

### `delete_token`

- 类型：`BIGINT UNSIGNED NOT NULL DEFAULT 0`
- 语义：
- `0`：活跃数据
- `> 0`：墓碑记录唯一逃逸值（仅用于唯一索引避让）

## 查询与写入规则

### 读路径

软删除表的默认查询必须显式带：

```sql
deleted_at = 0
```

适用：`SELECT` / `COUNT` / `EXISTS` / `JOIN` / 子查询。

### 更新路径

更新活跃数据时必须显式限制：

```sql
... WHERE ... AND deleted_at = 0
```

### 删除路径

软删除表禁止物理 `DELETE`，改为：

```sql
UPDATE <table>
SET deleted_at = <unix_millis_expr>,
    delete_token = <unique_token_expr>
WHERE ... AND deleted_at = 0
```

## 唯一索引规则

- 若业务键允许“删后重建”，唯一键需包含 `delete_token`。
- 示例：

```sql
UNIQUE KEY uk_xxx (<biz_columns...>, delete_token)
```

这样活跃数据（`delete_token=0`）仍保持唯一，墓碑数据可并存。

## 黑名单机制

默认采用“黑名单排除软删除”：

- 黑名单表保持物理删除，不要求 `deleted_at`。
- 黑名单配置文件：`modelcraft-backend/db/soft_delete.yaml`。

## `sqlc` 约束

- `sqlc.slice(...)` 等宏语法必须保持原样，不能被工具重写成其他形式。
- 任何自动改写工具都必须以“最小文本补丁”为原则，禁止整句格式化回写。

## 工具链约定

- 一次性迁移：`codemod`
- 持续守门：`lint`
- 生成链路：先 `soft-delete-lint`，再 `sqlc generate`

当前命令入口：

- `go run ./cmd/sqlsoftdelete codemod --config db/soft_delete.yaml`
- `go run ./cmd/sqlsoftdelete lint --config db/soft_delete.yaml`
- `just generate-sqlc`（含 soft-delete lint 前置）

## 常见误区

- 误区 1：`deleted_at` 用 `NULL` 表示活跃。  
  建议统一 `0`，避免 `NULL` 语义和唯一索引行为歧义。

- 误区 2：只改主查询，漏掉 `JOIN`/子查询。  
  结果会把墓碑数据带回业务层。

- 误区 3：工具整句重写 SQL。  
  会破坏 `sqlc` 宏或注释布局，导致 `sqlc generate` 失败。

