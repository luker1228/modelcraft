# 后端错题本

> 记录真实发生过的后端 Bug，供代码审查时作为 Checklist 使用。
> 每条错误由 `backend-checklist` skill 写入，格式统一，便于自动化 review。

---

## 错误编号规则

`BM-YYYYMMDD-XXXX`（BM = Backend Mistake，XXXX 为当天顺序编号）

---

## ⚠️ 错误列表

---

### BM-20260415-0001 · SQL 查询缺少 org_name 过滤导致跨租户数据污染

**严重程度**: CRITICAL  
**分类**: data-isolation / sql-query  
**发现日期**: 2026-04-15

#### 问题描述

`GetEnumsByNames` SQL 查询只用 `project_slug` 过滤，缺少 `org_name` 条件。由于 `project_slug` 在不同 org 之间可能重名，导致 runtime 加载枚举时拿到了**其他 org 的同名枚举定义**，枚举 options 不一致，引发 `undefined enum: a is not a valid value for enum order` 错误。

#### 错误代码

```sql
-- ❌ 错误：缺少 org_name
SELECT * FROM model_enums
WHERE project_slug = ? AND name IN (sqlc.slice('names'));
```

```sql
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

#### ⚡ Checklist 规则（Review 时必查）

> **凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。**
> 适用表：`model_enums`、`field_definitions`、`models`、`logical_foreign_keys`、`model_field_enum_associations` 等所有 project 域资源表。

---
