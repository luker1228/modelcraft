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

### BM-20260513-0002 · protected admin 角色未短路通配权限导致 runtime 误拒绝

**严重程度**: HIGH  
**分类**: rbac / runtime-authz  
**发现日期**: 2026-05-13

#### 问题描述

`FindPermissionsByEndUserAndModel` 仅按 bundle 链路解析权限（显式角色 bundle + 用户直连 bundle），没有在入口对 `is_protected=true && name=admin` 做短路处理。结果是 admin 用户在 bundle 为空时仍会被判定 `insert` 不允许，GraphQL create 返回 `Permission denied: insert`。

#### 错误模式

```go
// ❌ 错误：admin 也走 bundle 解析链路
bundleIDs := collectBundleIDsFromRoleAndUser(...)
permissions := resolvePermissions(bundleIDs, modelID)
if !permissions.Insert.Allowed {
    return permissionDenied("insert")
}
```

```go
// ✅ 正确：protected admin 先短路为 wildcard
if isProtectedAdmin(endUserID, orgName, projectSlug) {
    return wildcardPermissions(modelID) // select/insert/update/delete 全开
}

bundleIDs := collectBundleIDsFromRoleAndUser(...)
permissions := resolvePermissions(bundleIDs, modelID)
```

#### 根因

把“admin 是受保护内置全权限角色”的业务语义，错误实现成“admin 也必须依赖 bundle 才有数据权限”。实现语义与产品规则不一致。

#### 症状

- GraphQL runtime create 失败：`[OPERATION_FAILED.PERMISSION] Permission denied: insert`
- 用户已是受保护 `admin`，但在 DB 中无 role/user bundle 绑定时必现
- 部署后若仍用旧镜像（未 build）会误判为“修复无效”

#### 修复范围

1. 在 `FindPermissionsByEndUserAndModel` 增加 `hasProtectedAdminRole` 前置判定
2. 命中后直接返回 model 级 wildcard（`select/insert/update/delete` + `ScopeAll`）
3. 增加回归测试覆盖：admin 命中时不应触发 bundle 查询链路

#### ⚡ Checklist 规则（Review 时必查）

> **受保护 `admin` 角色必须在 runtime 权限入口先短路为全权限，不得依赖 bundle 绑定。**  
> 判定条件：`project_slug` 匹配 + `is_protected=true` + `name` 大小写不敏感等于 `admin`。

---
