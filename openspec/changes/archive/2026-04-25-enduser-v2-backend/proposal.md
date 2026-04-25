## Why

当前 EndUser 账号与 Project 强绑定（`org + project + username`），导致同一组织下跨项目复用账号困难、权限管理分散、登录后也无法一次性拿到可访问项目全集。`enduser-v2` 需要把账号所有权提升到 Org 层，同时保持运行时 JWT 仍按项目语义兼容现有 RLS。

## What Changes

- 将 EndUser 账号作用域从 `org+project` 调整为 `org`：用户名唯一键改为 `UNIQUE(org_name, username)`。
- 新增 EndUser 与 Project 的访问绑定实体（含权限包），用于表达“一个 EndUser 可访问多个项目”。
- 新增/调整 GraphQL 能力：
  - Org 侧负责 EndUser 账号生命周期（创建、列表、禁用、删除）。
  - Project 侧负责 EndUser 项目访问授权（授予、撤销、查询、更新权限包）。
- 登录改为一次签发：校验账号后直接返回 JWT，并同时返回该用户有权限的全部项目供自由选择。
- 数据库迁移：移除多处 `project_slug` 强绑定字段，新增访问绑定表并补齐索引与约束。
- **BREAKING**：EndUser 相关后端接口与仓储模型从项目作用域切换为组织作用域。

## Capabilities

### New Capabilities
- `enduser-org-account`: EndUser 账号在 Org 维度管理（创建、查询、状态更新、删除）与数据约束。
- `enduser-project-access`: EndUser 到 Project 的访问授权关系管理（授予/撤销/列表/权限包绑定）。
- `enduser-two-phase-auth`: EndUser 登录一次签发令牌，并返回全部可访问项目用于自由选择。

### Modified Capabilities
- （无）

## Impact

- 数据库：`end_user_users`、`end_user_accounts`、`end_user_roles`、`end_user_role_users` 结构调整；新增 `end_user_project_access`。
- 后端领域层与仓储层：EndUser 聚合模型、Repository 接口、查询条件与租户作用域传递。
- GraphQL：Org/Project 两套 schema 新增或调整 EndUser 相关类型与 mutation/query。
- BFF 鉴权接口：登录与刷新路径和返回结构调整（支持项目选择）。
- 兼容性：依赖旧 `project_slug` 绑定语义的调用方需要同步改造。