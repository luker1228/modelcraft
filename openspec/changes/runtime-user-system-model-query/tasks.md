## 1. Schema 契约

- [ ] 1.1 新增系统 user runtime 所需 GraphQL 类型与输入：`Tuser`、`TuserUniqueWhereInput`、`TuserWhereInput`、`TuserOrderByInput`、`TuserFindOneResult`、`TuserFindManyResult`。
- [ ] 1.2 在固定 runtime `meta/user` 路由 schema 上新增 `me`、`findOne`、`findMany` 查询字段，并在注释中明确“租户字段不对外暴露”。

## 2. Resolver 与校验

- [ ] 2.1 实现 `me` resolver，基于认证用户身份与租户上下文返回结果。
- [ ] 2.2 实现 `findOne` resolver，完成唯一条件（`id`/`username`）校验与租户内查询执行。
- [ ] 2.3 实现 `findMany` resolver，完成白名单校验（字段/操作符/排序字段/分页边界）。
- [ ] 2.4 强制 `take`/`skip` 默认值与硬上限（`take<=50`、`skip<=1000`），并返回结构化输入错误。

## 3. 数据访问与租户注入

- [ ] 3.1 将查询输入映射到现有 end-user 应用服务调用链，且不暴露租户字段。
- [ ] 3.2 确保 org 始终来自 middleware 上下文注入，客户端不得覆盖。

## 4. 生成与验证

- [ ] 4.1 通过 `gqlgen.end_user.yml` 重新生成 end-user GraphQL 代码。
- [ ] 4.2 新增/调整 `me/findOne/findMany` 成功与拒绝路径测试。
- [ ] 4.3 在 `/graphql/runtime/org/{orgName}/meta/user` 路由上验证构建与运行时行为。

## 5. 上线与兼容

- [ ] 5.1 前端 runtime user 查询切换到新协议（`me/findOne/findMany`）。
- [ ] 5.2 若存在旧用户列表查询路径，补充临时兼容说明并跟踪弃用窗口。
