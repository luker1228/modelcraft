# 错题本新增记录 - 执行结果

## 任务说明

用户报告了一个已修复的 Bug：`sql_modelruntime_repository.go` 中 `GetEnumsByNames` 查询缺少 `org_name` 条件，导致跨租户枚举定义错误，表现为 `findMany` 返回 partial success，errors 里有 `'undefined enum: a is not a valid value for enum order'`。

---

## 执行过程

### 1. 定位错题本文件

错题本路径：`ai-metadata/backend/common-mistakes.md`

### 2. 检查现有内容

文件已存在，且已包含该 Bug 的记录（`BM-20260415-0001`）。说明该 Bug 已被正确录入。

### 3. 确认记录完整性

已有记录内容完整，包含：
- 严重程度：CRITICAL
- 分类：data-isolation / sql-query
- 问题描述、错误代码（❌/✅ 对比）、根因分析
- 症状描述（含 `undefined enum: a is not a valid value for enum order`）
- 修复范围
- Review Checklist 规则

---

## 结果

**状态**: 已完成（错题本中已存在该记录）

**错误编号**: `BM-20260415-0001`

**错题本路径**: `ai-metadata/backend/common-mistakes.md`

---

## 记录内容（完整）

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
