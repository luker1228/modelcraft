## ADDED Requirements

### Requirement: EndUser 账号必须以 Org 作为唯一作用域
系统 MUST 将 EndUser 用户名唯一性约束定义为 `org_name + username`，并且 MUST 不再依赖 `project_slug` 参与账号唯一性判定。

#### Scenario: 在同一组织内创建重复用户名
- **WHEN** 管理员在同一 `org_name` 下创建已存在的 `username`
- **THEN** 系统 MUST 拒绝创建并返回“用户名已存在”错误

#### Scenario: 在不同组织创建同名账号
- **WHEN** 管理员在不同 `org_name` 下分别创建相同 `username`
- **THEN** 系统 SHALL 允许创建且两条账号记录彼此独立

### Requirement: Org GraphQL 必须提供 EndUser 账号生命周期管理
系统 SHALL 在 Org GraphQL 提供 EndUser 的创建、列表、禁用/启用、删除能力，且所有操作 MUST 基于组织作用域执行。

#### Scenario: 列表查询不传 Project 上下文
- **WHEN** 调用 `listEndUsers` 仅携带 Org 上下文
- **THEN** 系统 MUST 返回该组织内 EndUser 列表并支持分页

#### Scenario: 禁用账号后状态可见
- **WHEN** 调用 `updateEndUserStatus` 将 `isForbidden=true`
- **THEN** 后续 `listEndUsers` 返回该账号状态为禁用

### Requirement: 被禁用 EndUser 不得通过认证
系统 MUST 在认证阶段校验 `isForbidden`，对禁用账号拒绝登录。

#### Scenario: 禁用账号尝试登录
- **WHEN** `isForbidden=true` 的 EndUser 提交正确凭据
- **THEN** 系统 MUST 返回账号禁用错误且不得返回可访问项目
