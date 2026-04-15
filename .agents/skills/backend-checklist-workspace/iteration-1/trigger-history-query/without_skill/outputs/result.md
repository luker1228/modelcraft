# 跨租户数据污染历史案例与 SQL 层面踩坑记录

> 数据来源：`ai-metadata/backend/common-mistakes.md` + `ai-metadata/backend/development/domain-development.md`

---

## 历史案例

### BM-20260415-0001 · SQL 查询缺少 org_name 过滤导致跨租户数据污染

**严重程度**: CRITICAL  
**分类**: data-isolation / sql-query  
**发现日期**: 2026-04-15

#### 问题描述

`GetEnumsByNames` SQL 查询只用 `project_slug` 过滤，缺少 `org_name` 条件。由于 `project_slug` 在不同 org 之间可能重名，导致 runtime 加载枚举时拿到了**其他 org 的同名枚举定义**，枚举 options 不一致，引发 `undefined enum: a is not a valid value for enum order` 错误。

#### 错误代码 vs 修复代码

```sql
-- ❌ 错误：缺少 org_name
SELECT * FROM model_enums
WHERE project_slug = ? AND name IN (sqlc.slice('names'));

-- ✅ 正确：必须同时带 org_name
SELECT * FROM model_enums
WHERE org_name = ? AND project_slug = ? AND name IN (sqlc.slice('names'));
```

#### 根因

`project_slug` 不是全局唯一键，只在同一 org 内唯一。跨 org 时同一 `project_slug` + 同一 `enum name` 会命中不属于该 org 的记录。

#### 症状

- GraphQL `findMany` 返回 partial success，`errors` 数组含 `"undefined enum: X is not a valid value for enum Y"`
- `modelJsonSchema` 查询返回的枚举 options 与 runtime 实际使用的不一致
- 日志中 `enum.options` 里的 code 与数据库里该 org 的记录不符

#### 修复范围

1. SQL 查询加 `org_name = ?`
2. 对应 dbgen params struct 加 `OrgName` 字段
3. 调用处传入 `OrgName`

---

## SQL 层面的核心坑总结

### 坑 1：只按 project_slug 查询，忘记加 org_name

这是上述案例的根因。`project_slug` 在跨 org 时不唯一，凡是涉及以下表的 SQL，必须同时带 `org_name`：

- `model_enums`
- `field_definitions`
- `models`
- `logical_foreign_keys`
- `model_field_enum_associations`
- 所有 project 域资源表

**Review Checklist 规则：凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。**

---

### 坑 2：FindByID 不加 org_name，依赖 ID 全局唯一

即使是按 ID 查询，也不能省略 `org_name` 过滤。ID 可能被枚举、猜测或泄露，如果 `FindByID` 不过滤 `orgName`，攻击者知道另一个租户资源的 ID 后可以直接读取该资源。

```sql
-- ❌ 不安全：只靠 ID 查询
SELECT * FROM enum_definitions WHERE id = ?

-- ✅ 安全：即使 ID 泄露也无法跨租户读取
SELECT * FROM enum_definitions WHERE id = ? AND org_name = ?
```

对应 Repository 接口签名规范：

```go
// ❌ 错误：只有 id，无法保证租户隔离
FindByID(id string) (*EnumDefinition, error)

// ✅ 正确：FindByID 也必须有 orgName
FindByID(ctx context.Context, orgName, id string) (*EnumDefinition, error)
```

---

### 坑 3：Create/Update 方法未从实体中提取 org_name 写入 SQL

`Create` 和 `Update` 方法通过实体对象携带 `orgName`（嵌入 `ProjectScope`），但 Infrastructure 实现时必须**显式从实体中提取并写入 SQL 参数**，否则 `org_name` 字段为空，导致数据写入缺少租户标识。

```go
// Infrastructure 实现中必须：
func (r *SqlEnumRepo) Create(ctx context.Context, enum *EnumDefinition) error {
    params := dbgen.CreateEnumParams{
        OrgName:     enum.OrgName,     // ← 必须从实体取出，写入 DB
        ProjectSlug: enum.ProjectSlug,
        // ...
    }
}
```

---

### 坑 4：Repository 接口方法缺少 ctx

没有 `ctx` 的方法无法被正确追踪和取消（请求追踪 ID、超时控制、取消信号、日志字段注入全部丢失）。

```go
// ❌ 错误：缺少 ctx，无法追踪
Create(enum *EnumDefinition) error

// ✅ 正确
Create(ctx context.Context, enum *EnumDefinition) error
```

---

## 防御规则速查（Review Checklist）

| 场景 | 规则 |
|------|------|
| 任何带 `project_slug` 的 SQL | 必须同时带 `org_name` |
| `FindByID` / `GetByID` | 必须有 `orgName` 参数，防 ID 枚举攻击 |
| `Create` / `Update` 实体写入 | Infrastructure 实现必须从实体中提取 `OrgName` 写入 SQL |
| 所有 Repository 方法 | 第一个参数必须是 `ctx context.Context` |
| project 域资源方法 | 必须有 `projectSlug` 参数（`Field`、`Model`、`Enum` 等） |

---

## 来源文件

- `ai-metadata/backend/common-mistakes.md` — 后端错题本（真实 Bug 案例）
- `ai-metadata/backend/development/domain-development.md` — Domain 层 Repository 接口设计规范
