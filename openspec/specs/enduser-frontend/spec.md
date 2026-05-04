## Requirements

### Requirement: Org 侧用户详情页的「项目访问」区域

Org 用户详情页 (`/org/{orgName}/end-users/{userId}`) 的「项目访问」区域 SHALL 展示该用户在各 Project 下的 Role 分配情况（只读），并提供跳转链接。

- 数据源：`endUserRoleAssignments(endUserId)` 按 `projectSlug` 分组
- 展示列：项目名、角色名、授权时间
- 每行提供「前往管理 →」链接，跳转到 `/org/{orgName}/project/{slug}/end-user-access`
- 不提供任何增删改操作入口

#### Scenario: 展示用户已授权的 Project 列表

- **WHEN** 管理员访问某 EndUser 的详情页
- **THEN** 系统 SHALL 展示该用户所有 Role 分配记录，按 Project 分组
- **THEN** 每条记录展示：项目名、角色名、授权时间
- **THEN** 每行显示「前往管理 →」链接，跳转到对应 Project 的访问控制页

#### Scenario: 用户尚无任何 Role 分配

- **WHEN** EndUser 在任何 Project 下均无 Role 分配
- **THEN** 系统 SHALL 展示空状态提示，引导管理员到 Project 页面分配 Role

---

### Requirement: Project 访问控制页的用户 Role 分配管理

Project 访问控制页 (`/org/{orgName}/project/{slug}/end-user-access`) SHALL 提供完整的 EndUser Role 管理功能：列出、添加、修改、撤销。

- 列表通过 `listProjectEndUserRoleUsers` 查询，展示 username、角色名、授权时间及操作按钮
- 「添加用户」弹窗：下拉选 Org 内用户 + 下拉选 Role（默认预选第一个 `isImplicit=false` 的 Role）
- 「修改角色」：先 `revokeEndUserRole` 再 `assignEndUserRole`
- 「撤销」：调用 `revokeEndUserRole`
- 若 Project 下无任何 Role 可选，「添加用户」弹窗 SHALL 提示管理员先在 RBAC 页创建 Role

#### Scenario: 列出 Project 下所有有 Role 分配的用户

- **WHEN** 管理员访问 Project 访问控制页
- **THEN** 系统 SHALL 调用 `listProjectEndUserRoleUsers` 并展示结果列表
- **THEN** 每行显示：username、角色名、授权时间、「修改角色」和「撤销」按钮

#### Scenario: 添加用户

- **WHEN** 管理员点击「+ 添加用户」并填写用户名和角色后确认
- **THEN** 系统 SHALL 调用 `assignEndUserRole(endUserId, roleId)`
- **THEN** 成功后刷新列表

#### Scenario: Role 下拉默认选取

- **WHEN** 「添加用户」弹窗打开
- **THEN** Role 下拉 SHALL 默认预选第一个 `isImplicit = false` 的 Role
- **THEN** 若 Project 下无可用 Role，弹窗 SHALL 显示提示并禁用确认按钮

#### Scenario: 修改用户角色

- **WHEN** 管理员点击某行「修改角色」并选择新 Role 后确认
- **THEN** 系统 SHALL 先调用 `revokeEndUserRole` 撤销旧 Role
- **THEN** 再调用 `assignEndUserRole` 分配新 Role
- **THEN** 成功后刷新列表

#### Scenario: 撤销用户角色

- **WHEN** 管理员点击某行「撤销」并确认
- **THEN** 系统 SHALL 调用 `revokeEndUserRole`
- **THEN** 成功后从列表中移除该行

---

### Requirement: 前端 API Client 新增 LIST_PROJECT_END_USER_ROLE_USERS

`src/api-client/rbac/graphql-docs.ts` SHALL 新增 `LIST_PROJECT_END_USER_ROLE_USERS` GraphQL 文档常量，用于查询 Project 下的用户 Role 分配列表。

- 字段：`endUser { id, username, isForbidden }`、`role { id, name, description }`、`assignedAt`
- 支持 `pageInfo { hasNextPage, endCursor }` 和 `totalCount`
- 错误联合类型：`InvalidInput`、`ProjectNotFound`

#### Scenario: 前端成功获取分配列表

- **WHEN** 前端执行 `LIST_PROJECT_END_USER_ROLE_USERS` 查询
- **THEN** 系统 SHALL 返回 `connection.nodes` 数组，每项包含 endUser、role、assignedAt 字段
- **THEN** 若后端返回错误，前端 SHALL 读取 `error.__typename` 并展示对应错误信息
