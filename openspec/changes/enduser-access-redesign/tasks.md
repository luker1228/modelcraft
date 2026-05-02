# 实施任务清单：EndUser 访问授权模型重设计

**变更**：`enduser-access-redesign`

---

## 1. 后端：新增 GraphQL Query

- [ ] 1.1 在 `api/graph/project/schema/rbac.graphql` 追加 `ProjectEndUserRoleUser`、`ProjectEndUserRoleUserConnection`、`ListProjectEndUserRoleUsersPayload`、`ListProjectEndUserRoleUsersError` 类型定义及 `ListProjectEndUserRoleUsersInput` input
- [ ] 1.2 在 `api/graph/project/schema/rbac.graphql` 的 `extend type Query` 块追加 `listProjectEndUserRoleUsers` 字段（`@hasPermission(action: "rbac:read")`）
- [ ] 1.3 运行 `just generate-gql` 生成 Go 接口代码
- [ ] 1.4 实现 `listProjectEndUserRoleUsers` Resolver：从 Context 获取 `projectSlug`，查询 `end_user_role_users WHERE project_slug = ?`，JOIN `end_user_users`，支持 `search`（username 模糊）和 `roleId` 过滤，返回 cursor-based 分页
- [ ] 1.5 为 `listProjectEndUserRoleUsers` 编写 Repository 层接口及 SQLC 查询

## 2. 后端：调整 assignEndUserRole 存在性检查

- [ ] 2.1 确认 `assignEndUserRole` Resolver 中的 EndUser 存在性校验逻辑
- [ ] 2.2 将校验改为「EndUser 是否属于当前 Org」（Org 级查询），移除依赖 `EndUserProjectAccess` 的 Project 级预注册检查
- [ ] 2.3 同步调整 `revokeEndUserRole` 中的 EndUser 存在性校验（保持一致）
- [ ] 2.4 更新 `EndUserNotFoundInProject` 错误类型的注释/描述，说明新语义：「用户不属于当前 Org」

## 3. 后端：废弃 EndUserProjectAccess

- [ ] 3.1 删除 `api/graph/project/schema/end_user_access.graphql` 文件（含 `grantEndUserProjectAccess`、`updateEndUserProjectAccess`、`revokeEndUserProjectAccess`、`listProjectEndUserAccess`）
- [ ] 3.2 运行 `just generate-gql` 删除对应生成代码
- [ ] 3.3 删除 `EndUserProjectAccess` domain entity、`EndUserProjectAccessRepository` 接口及其实现
- [ ] 3.4 编写 DB migration：`DROP TABLE end_user_project_access`（使用 Atlas）
- [ ] 3.5 更新登录流程中「获取可访问 Project 列表」的查询逻辑，改为查 `end_user_role_users`（而非 `end_user_project_access`）
- [ ] 3.6 运行全量测试，确认无编译错误和 BDD 失败

## 4. API Contract 同步

- [ ] 4.1 在后端仓库执行 `git subtree push --prefix=api contracts main`，将更新后的 Schema 推送到共享合约仓库
- [ ] 4.2 在前端仓库执行 `front-contract-pull`，拉取最新 API Contract 到 `contract/` 目录

## 5. 前端：API Client 变更

- [ ] 5.1 在 `src/api-client/rbac/graphql-docs.ts` 删除以下常量：`LIST_PROJECT_END_USER_ACCESS`、`GRANT_PROJECT_END_USER_ACCESS`、`UPDATE_PROJECT_END_USER_ACCESS`、`REVOKE_PROJECT_END_USER_ACCESS`
- [ ] 5.2 在 `src/api-client/rbac/graphql-docs.ts` 新增 `LIST_PROJECT_END_USER_ROLE_USERS` 查询（字段：`endUser { id, username, isForbidden }`、`role { id, name, description }`、`assignedAt`、`pageInfo`、`totalCount`、错误联合）

## 6. 前端：Project 访问控制页重写

- [ ] 6.1 更新 Project `/end-user-access` 页面的列表数据源，改用 `LIST_PROJECT_END_USER_ROLE_USERS`（去除旧的 `LIST_PROJECT_END_USER_ACCESS`）
- [ ] 6.2 更新列表展示列：username、角色名、授权时间、「修改角色」按钮、「撤销」按钮
- [ ] 6.3 重写「添加用户」弹窗：用户名下拉从 Org 用户列表选取，角色下拉默认预选第一个 `isImplicit=false` 的 Role；若无可用 Role 则显示提示并禁用确认
- [ ] 6.4 「添加用户」弹窗确认时调用 `ASSIGN_END_USER_ROLE_TO_USER`（`assignEndUserRole`）
- [ ] 6.5 实现「修改角色」：先调用 `revokeEndUserRole` 再调用 `assignEndUserRole`
- [ ] 6.6 实现「撤销」：调用 `revokeEndUserRole`，成功后刷新列表
- [ ] 6.7 删除或停用已无后端支撑的旧 hook/组件（原依赖 `grantEndUserProjectAccess` 的部分）

## 7. 前端：Org 用户详情页更新

- [ ] 7.1 更新 `/org/{orgName}/end-users/{userId}` 详情页的「项目访问」区域，改用 `GET_END_USER_ROLE_USERS(endUserId)` 获取数据
- [ ] 7.2 将查询结果按 `projectSlug` 分组，展示列：项目名、角色名、授权时间
- [ ] 7.3 每行添加「前往管理 →」跳转链接，目标：`/org/{orgName}/project/{slug}/end-user-access`
- [ ] 7.4 移除「授权」按钮及 Grant 弹窗组件

## 8. 验收

- [ ] 8.1 端到端验证：在 Project 页面为 Org 用户分配 Role，确认用户登录后可见该 Project
- [ ] 8.2 端到端验证：撤销用户在某 Project 下所有 Role 后，用户登录后不再看到该 Project
- [ ] 8.3 端到端验证：Org 详情页「项目访问」区域只读展示，「前往管理」链接正确跳转
- [ ] 8.4 运行前端 lint（`npm run lint`）无报错
- [ ] 8.5 运行后端测试（`just test`）无失败
