## MODIFIED Requirements

### Requirement: EndUser 可访问 Project 由 Role Assignment 决定
系统 SHALL 通过查询 `end_user_role_users` 表确定用户级主体可访问的 Project 列表，不再依赖 `end_user_project_access` 表。

- 若用户级主体在某 Project 下有至少一条 `EndUserRoleUser`，则该 Project 出现在其可访问列表中
- 若用户级主体在某 Project 下的所有 Role 均被撤销，则该 Project 从其可访问列表中移除
- 组织级与项目级管理主体在进入 data plane 时，MUST 能在满足访问模式约束的前提下访问同一 Project 集合中的用户级能力

#### Scenario: 用户级主体登录后查询可访问 Project
- **WHEN** 用户级主体登录并请求可访问 Project 列表
- **THEN** 系统 SHALL 查询 `end_user_role_users WHERE end_user_id = ?`
- **THEN** 提取不重复的 `project_slug` 列表
- **THEN** 返回这些 Project 的信息（含 `project_title`）

#### Scenario: 用户级主体所有 Role 被撤销后
- **WHEN** 管理员撤销某用户级主体在某 Project 下的所有 Role
- **THEN** 该 Project 不再出现在该用户级主体的可访问 Project 列表中

#### Scenario: 上层管理主体兼容访问用户级能力
- **WHEN** 组织级或项目级主体以扮演用户或全权限模式进入某个已授权 Project 的 data plane
- **THEN** 系统允许该主体访问该 Project 下与用户级主体一致的数据面能力集合

## ADDED Requirements

### Requirement: 用户级主体的数据访问 MUST 先具备 catalog 可见性
系统 SHALL 将用户级主体的 Project 可访问性解释为 data-plane 访问前提，而不仅是“可以进入项目”。一旦某用户级主体可访问某 Project，系统 MUST 允许其在授权范围内查看该 Project 下的 database catalog、model catalog 与必要的 model schema subset。

#### Scenario: 可访问 Project 时可查看 catalog
- **WHEN** 某用户级主体出现在 Project P 的可访问 Project 列表中
- **THEN** 该主体可以查看 Project P 授权范围内的 database catalog 与 model catalog

#### Scenario: catalog 是进入具体数据查询的前提
- **WHEN** 某用户级主体尝试查看 Project P 下某个 model 的运行态数据
- **THEN** 系统先按其 catalog 可见范围决定该 model 是否可见
- **THEN** 仅在 model 可见时继续执行数据查询授权判断
