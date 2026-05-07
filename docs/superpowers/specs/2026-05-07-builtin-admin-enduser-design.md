# 设计文档：每个 Org 内置 admin EndUser

**日期：** 2026-05-07  
**状态：** 已批准，待实现  
**背景：** 当 tenant 管理员创建数据记录时，`owner` 字段（EndUser ref）目前在选择"不指定"时传 `undefined`，导致数据库中 owner 为空，违反 RLS 的数据所有权语义。

---

## 问题陈述

每条用户端数据记录必须有合法的 owner（EndUser UUID）才能参与 RLS。  
当前的"不指定"选项让 tenant 管理员可以创建无主数据，这不符合"数据不属于任何人时归平台侧超管"的业务语义。

---

## 目标

- 每个 Org 在创建时自动内置一个 `admin` EndUser，代表"平台公共数据"所有者
- 该账号可由平台侧配置密码并登录用户端
- Dropdown 去掉"不指定"，将内置 admin 置顶展示，新建记录默认预填为 admin
- 内置账号受保护，tenant 管理员不能删除或禁用它

---

## 方案：Org 内置 admin EndUser

### 1. 领域模型变更

在 `EndUser` 实体和 `end_user_users` 表中增加 `is_builtin` 字段：

```
EndUser
  + is_builtin: bool  ← 新增，标记为平台内置账号，不可删除/禁用
```

**业务规则：**
- 每个 Org 有且仅有一个 `is_builtin = true` 的 EndUser，用户名固定为 `admin`
- `is_builtin = true` 的账号不能被 tenant 管理员删除或禁用
- 密码由 Org 创建者（Developer）在创建 Org 时设置，后续可通过正常 `updateEndUserPassword` 修改
- 可以正常登录用户端（走现有 EndUser 登录流程）
- 用户名不可修改（现有约束，无需额外处理）

### 2. 数据库 Schema 变更

```sql
-- end_user_users 表新增字段
ALTER TABLE end_user_users
  ADD COLUMN is_builtin TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '是否为平台内置账号（每个 Org 唯一，不可删除/禁用）';
```

> **唯一性保证由应用层负责**：不加数据库唯一约束（因为普通用户 `is_builtin=0` 有多条，约束会误判）。应用层在 `CreateOrganizationService` 中保证每个 Org 只创建一条 `is_builtin=1` 记录；`EndUserRepository` 提供 `GetBuiltinByOrg(orgName)` 查询方法。

> Atlas 迁移文件放在 `db/schema/mysql/` 下，按现有序号命名。

### 3. Org 初始化流程变更

在 `CreateOrganizationService.createOrganizationInTransaction()` 中，事务内最后一步新增：

```
现有流程：
  1. 创建 Organization 实体
  2. 创建 Membership（创建者 → Org）
  3. 分配 Owner 角色

新增第 4 步：
  4. 创建内置 admin EndUser
       username    = "admin"
       password    = endUserAdminPassword（明文，bcrypt 哈希后存储）
       is_builtin  = true
       is_forbidden = false
       created_by  = 创建 Org 的 Developer user_id
```

**幂等保护：** 若 `(org_name, username="admin")` 已存在，跳过创建，不报错（支持重试）。

**密码来源：** `CreateOrganizationInput` 新增 `endUserAdminPassword: String!`，由前端在创建 Org 时让用户填写（与 Developer 登录密码一致，用户主动输入，服务层不做验证对比，直接作为内置 admin 的初始密码）。

### 4. GraphQL Schema 变更

**`api/graph/org/schema/end_user.graphql`：**

```graphql
type EndUser {
  # ...现有字段...
  isBuiltin: Boolean!  # 新增：是否为平台内置账号
}
```

**`api/graph/org/schema/project.graphql`（或 organization.graphql）：**

```graphql
input CreateOrganizationInput {
  # ...现有字段...
  endUserAdminPassword: String!  # 新增：内置 admin EndUser 初始密码
}
```

**错误类型（`api/graph/org/schema/end_user.graphql`）：**

```graphql
# 新增错误类型（用于 deleteEndUser / updateEndUserStatus 的 union）
type BuiltinUserCannotBeDeleted {
  message: String!
}

type BuiltinUserCannotBeDisabled {
  message: String!
}
```

> 具体 union 接入方式以现有 Error 处理规范为准（参见 `ai-metadata/backend/development/error-handling.md`）。

### 5. 应用层守卫

**`EndUserManagementAppService.DeleteEndUser`：**
```go
if user.IsBuiltin {
    return ErrBuiltinUserCannotBeDeleted
}
```

**`EndUserManagementAppService.UpdateEndUserStatus`（禁用）：**
```go
if user.IsBuiltin && isForbidden {
    return ErrBuiltinUserCannotBeDisabled
}
```

改密码（`UpdateEndUserPassword`，如需新增 mutation）：**允许**，不做额外限制。

### 6. 前端 Dropdown 变更

**`EndUserSelectorWidget.tsx`：**

- 移除"不指定（`__none__`）"选项
- 内置 admin（`isBuiltin = true`）固定置顶，带特殊标识（如「系统」chip）
- 其余 EndUser 保持原有排序
- 新建记录时，owner 字段默认预填为内置 admin 的 ID（由后端在 `listEndUsers` 返回时前端识别 `isBuiltin = true` 的那条）

**创建 Org 页面：**

- 在现有创建 Org 表单中新增「用户端管理员密码」输入框，`required`
- placeholder 提示：「将作为用户端 admin 账号的初始密码，可与您的登录密码相同」

---

## 影响范围

| 层 | 变更内容 |
|----|---------|
| DB Schema | `end_user_users` 新增 `is_builtin` 列，Atlas 迁移文件 |
| Domain | `EndUser` 实体新增 `IsBuiltin` 字段，守卫方法 |
| Repository | sqlc query 更新，支持按 `is_builtin` 查询 |
| App Service | `CreateOrganizationService` 事务内新增创建内置 admin；`EndUserManagementAppService` 守卫 |
| GraphQL Schema | `EndUser` 类型新增 `isBuiltin`；`CreateOrganizationInput` 新增 `endUserAdminPassword` |
| Code Gen | `just generate-gql` 重新生成 |
| Frontend | `EndUserSelectorWidget` 去掉"不指定"，置顶 admin；创建 Org 页新增密码输入框 |

---

## 不在本期范围内

- 内置 admin 的密码重置 mutation（可后续单独加）
- 内置 admin 在用户端的特殊权限（如绕过 RLS）
- 多个内置账号支持

---

## 参考文档

- `ai-metadata/backend/design/domain-model/7-sql-editor.md`（RLS 设计）
- `ai-metadata/backend/development/error-handling.md`（错误处理规范）
- `ai-metadata/backend/development/contract-sync.md`（GraphQL Schema 工作流）
- `ai-metadata/prd/enduser-v2/11-domain-model-changes.md`（EndUser v2 领域模型）
