# Spec：EndUser 访问授权模型

**能力领域**：`enduser-access-model`  
**变更**：`enduser-access-redesign`

---

## REMOVED Requirements

### Requirement: EndUserProjectAccess 实体与授权通道

**Reason**：`EndUserProjectAccess` 与 `EndUserRoleUser` 并行存在，形成两条语义重叠的授权通道。废弃后，Role Assignment 成为唯一授权通道，逻辑更清晰。

**Migration**：
- 删除 `end_user_project_access` 表（DB migration）
- 删除 `EndUserProjectAccess` domain entity 及其 Repository
- 删除以下 GraphQL 接口：`grantEndUserProjectAccess`、`updateEndUserProjectAccess`、`revokeEndUserProjectAccess`、`listProjectEndUserAccess`
- 登录流程中「可访问 Project 列表」改为查询 `end_user_role_users` 表

---

## MODIFIED Requirements

### Requirement: EndUser 存在性校验

在 Project Scope 下对 EndUser 执行操作时（如 `assignEndUserRole`、`revokeEndUserRole`），系统 SHALL 校验 EndUser 是否属于当前 Org，而非校验其是否在该 Project 下有 `EndUserProjectAccess` 记录。

- 若 EndUser 不属于当前 Org，返回错误 `EndUserNotFoundInProject`（语义重定义为「不属于该 Org」）
- 同 Org 下的任意 EndUser 均可被分配 Project Role，无需预注册到 Project

#### Scenario: 为同 Org 用户分配 Role

- **WHEN** 管理员在 Project 页面为某 EndUser 分配 Role
- **THEN** 系统 SHALL 检查 EndUser 是否属于该 Org（而非是否有 ProjectAccess 记录）
- **THEN** 若 EndUser 不属于该 Org，返回 `EndUserNotFoundInProject` 错误
- **THEN** 若 EndUser 属于同 Org，创建 `EndUserRoleUser` 记录并返回成功

#### Scenario: 尝试为外 Org 用户分配 Role

- **WHEN** 调用 `assignEndUserRole` 时传入不属于当前 Org 的 `endUserId`
- **THEN** 系统 SHALL 返回错误 `EndUserNotFoundInProject`（含说明：用户不属于该 Org）

---

## ADDED Requirements

### Requirement: Project 视角的 EndUser Role 分配列表查询

系统 SHALL 提供 `listProjectEndUserRoleUsers` GraphQL Query，在 Project Scope 下列出该 Project 内所有有 Role 分配的用户。

- 支持按 `username` 模糊搜索（`search` 参数）
- 支持按 `roleId` 过滤
- 返回 cursor-based 分页（`first` / `after`）
- 返回每条记录的 `endUser`、`role`、`assignedAt` 字段
- 需要 `rbac:read` 权限

#### Scenario: 列出 Project 下所有有 Role 分配的用户

- **WHEN** 管理员调用 `listProjectEndUserRoleUsers`（无过滤参数）
- **THEN** 系统 SHALL 返回该 Project 下所有 `end_user_role_users` 记录
- **THEN** 每条记录包含 `endUser { id, username, isForbidden }`、`role { id, name, description }`、`assignedAt`

#### Scenario: 按用户名搜索

- **WHEN** 调用时传入 `search: "alice"`
- **THEN** 系统 SHALL 仅返回 username 包含 "alice" 的用户的 Role 分配记录（模糊匹配）

#### Scenario: 按 Role 过滤

- **WHEN** 调用时传入 `roleId: "<id>"`
- **THEN** 系统 SHALL 仅返回该 Role 的分配记录

#### Scenario: Project 下无任何 Role 分配

- **WHEN** 该 Project 下无任何 `end_user_role_users` 记录
- **THEN** 系统 SHALL 返回空列表，`totalCount: 0`

---

### Requirement: EndUser 可在同一 Project 下持有多个 Role

系统 SHALL 允许同一 EndUser 在同一 Project 下持有多个 Role Assignment。

- 后端不限制同一用户在同 Project 下的 Role 数量
- `listProjectEndUserRoleUsers` 返回的每条记录代表一条独立的 Role 分配，同一用户可出现多行

#### Scenario: 为同一用户分配第二个 Role

- **WHEN** EndUser 在某 Project 下已有 Role A，管理员再为其分配 Role B
- **THEN** 系统 SHALL 创建第二条 `EndUserRoleUser` 记录
- **THEN** `listProjectEndUserRoleUsers` 返回该用户的两行记录（Role A 和 Role B）

---

### Requirement: EndUser 可访问 Project 由 Role Assignment 决定

系统 SHALL 通过查询 `end_user_role_users` 表确定 EndUser 可访问的 Project 列表，不再依赖 `end_user_project_access` 表。

- 若 EndUser 在某 Project 下有至少一条 `EndUserRoleUser`，则该 Project 出现在其可访问列表中
- 若 EndUser 在某 Project 下的所有 Role 均被撤销，则该 Project 从其可访问列表中移除

#### Scenario: EndUser 登录后查询可访问 Project

- **WHEN** EndUser 登录并请求可访问 Project 列表
- **THEN** 系统 SHALL 查询 `end_user_role_users WHERE end_user_id = ?`
- **THEN** 提取不重复的 `project_slug` 列表
- **THEN** 返回这些 Project 的信息（含 `project_title`）

#### Scenario: EndUser 所有 Role 被撤销后

- **WHEN** 管理员撤销 EndUser 在某 Project 下的所有 Role
- **THEN** 该 Project 不再出现在 EndUser 的可访问 Project 列表中
