## ADDED Requirements

### Requirement: Project Role Assignment 必须成为 ProjectPrincipal 展开基础
系统 SHALL 以 EndUser 在某 project 下的 role assignment 作为该主体进入 project API 时展开 `ProjectPrincipal` 的基础来源。

系统 SHALL 基于这些 assignment 展开：

- 是否可进入该 project
- project 级功能权限
- project 级数据权限来源

#### Scenario: 已授权用户进入 Project
- **WHEN** 某 EndUser 在某 project 下存在至少一条 role assignment
- **THEN** 系统 SHALL 允许其进入该 project
- **THEN** 系统 SHALL 基于 assignment 展开对应的功能权限与数据权限

## MODIFIED Requirements

### Requirement: EndUser 可访问 Project 由 Role Assignment 决定
系统 SHALL 通过查询 `end_user_role_users` 表确定 EndUser 可访问的 Project 列表，不再依赖 `end_user_project_access` 表。

- 若 EndUser 在某 Project 下有至少一条 `EndUserRoleUser`，则该 Project 出现在其可访问列表中
- 若 EndUser 在某 Project 下的所有 Role 均被撤销，则该 Project 从其可访问列表中移除
- 登录成功后返回的 project 列表与统一 workspace 中展示的可访问 project 列表 MUST 基于同一判定结果

#### Scenario: EndUser 登录后查询可访问 Project
- **WHEN** EndUser 登录并请求可访问 Project 列表
- **THEN** 系统 SHALL 查询 `end_user_role_users WHERE end_user_id = ?`
- **THEN** 提取不重复的 `project_slug` 列表
- **THEN** 返回这些 Project 的信息（含 `project_title`）

#### Scenario: EndUser 所有 Role 被撤销后
- **WHEN** 管理员撤销 EndUser 在某 Project 下的所有 Role
- **THEN** 该 Project 不再出现在 EndUser 的可访问 Project 列表中
