## 1. Schema 契约

- [x] 1.1 新增系统 user runtime 所需 GraphQL 类型与输入：`Tuser`、`TuserUniqueWhereInput`、`TuserWhereInput`、`TuserOrderByInput`、`TuserFindOneResult`、`TuserFindManyResult`。
- [x] 1.2 在固定 runtime `meta/user` 路由 schema 上新增 `me`、`findOne`、`findMany` 查询字段，并在注释中明确"租户字段不对外暴露"。

## 2. Resolver 与校验

- [x] 2.1 实现 `me` resolver，基于认证用户身份与租户上下文返回结果。
- [x] 2.2 实现 `findOne` resolver，完成唯一条件（`id`/`username`）校验与租户内查询执行。
- [x] 2.3 实现 `findMany` resolver，完成白名单校验（字段/操作符/排序字段/分页边界）。
- [x] 2.4 强制 `take`/`skip` 默认值与硬上限（`take<=50`、`skip<=1000`），并返回结构化输入错误。

## 3. 数据访问与租户注入

- [x] 3.1 将查询输入映射到现有 end-user 应用服务调用链，且不暴露租户字段。
- [x] 3.2 确保 org 始终来自 middleware 上下文注入，客户端不得覆盖。

## 4. 生成与验证

- [x] 4.1 通过 `gqlgen.end_user.yml` 重新生成 end-user GraphQL 代码。
- [x] 4.2 新增/调整 `me/findOne/findMany` 成功与拒绝路径测试。
- [x] 4.3 在 `/graphql/runtime/org/{orgName}/meta/user` 路由上验证构建与运行时行为。

## 5. 上线与兼容

- [x] 5.1 前端 runtime user 查询切换到新协议（`me/findOne/findMany`）。
  - 经调研：前端未直接调用旧的 `/graphql/runtime/org/{orgName}/project/{projectSlug}/user` 路由，
    EndUserRef 组件通过 `end-user` 端点（`/graphql/end-user/org/{orgName}/project/{projectSlug}`）
    的 `users` 查询运行，该端点未变更，无需前端迁移。
- [x] 5.2 若存在旧用户列表查询路径，补充临时兼容说明并跟踪弃用窗口。
  - 旧路由 `/graphql/runtime/org/{orgName}/project/{projectSlug}/user` 已从后端移除，
    替换为 `/graphql/runtime/org/{orgName}/meta/user`（org 级）。
  - `end-user` 端点的 `user/users` 旧查询保持不变，不影响现有前端功能。
