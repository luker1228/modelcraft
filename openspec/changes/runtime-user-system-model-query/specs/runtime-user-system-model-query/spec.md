## ADDED Requirements

### Requirement: 系统 user runtime 路由 MUST 暴露专属查询契约
系统 MUST 在 `/graphql/runtime/org/{orgName}/meta/user` 暴露面向系统 `user` 元数据模型的专属 GraphQL 查询契约，并且 MUST 仅支持 `me`、`findOne`、`findMany` 三个查询操作。

#### Scenario: 固定 runtime user 路由仅提供专属查询操作
- **WHEN** 客户端在固定 runtime user 路由执行 schema introspection
- **THEN** 路由暴露 `me`、`findOne`、`findMany` 查询字段

### Requirement: 系统 user runtime 契约 MUST 定义明确的 GraphQL SDL
系统 MUST 为 `/graphql/runtime/org/{orgName}/meta/user` 提供如下最小可用 GraphQL SDL（字段可增不减，且不得引入租户输入字段）：

```graphql
type Tuser {
  id: ID!
  username: String!
  createdAt: String!
}

input TuserUniqueWhereInput {
  id: ID
  username: String
}

input IDFilter {
  eq: ID
  in: [ID!]
}

input StringFilter {
  eq: String
  contains: String
  startsWith: String
  in: [String!]
}

input DateTimeFilter {
  eq: String
  gte: String
  lte: String
}

input TuserWhereInput {
  id: IDFilter
  username: StringFilter
  createdAt: DateTimeFilter
}

enum SortDirection {
  asc
  desc
}

input TuserOrderByInput {
  createdAt: SortDirection
}

type TuserFindOneResult {
  item: Tuser
  timeCost: Int!
  reqId: String!
}

type TuserFindManyResult {
  items: [Tuser!]!
  timeCost: Int!
  reqId: String!
}

extend type Query {
  me: TuserFindOneResult!
  findOne(where: TuserUniqueWhereInput!): TuserFindOneResult!
  findMany(
    where: TuserWhereInput
    orderBy: [TuserOrderByInput!]
    skip: Int
    take: Int
  ): TuserFindManyResult!
}
```

#### Scenario: 契约包含约定类型与查询字段
- **WHEN** 客户端读取该路由 GraphQL schema
- **THEN** schema 至少包含 `Tuser`、`TuserUniqueWhereInput`、`TuserWhereInput`、`TuserOrderByInput`、`TuserFindOneResult`、`TuserFindManyResult` 与 `me/findOne/findMany`

### Requirement: 系统 user 的 findMany MUST 受白名单约束
系统 MUST 将系统 `user` 模型的 `findMany` 过滤能力限制在 `id`、`username`、`createdAt`，并且 MUST 拒绝所有非白名单字段或操作符。

#### Scenario: 白名单过滤字段允许执行
- **WHEN** 客户端使用 `id`、`username` 或 `createdAt` 作为过滤条件调用 `findMany`
- **THEN** 查询成功执行并返回当前租户范围内的 `items`

#### Scenario: 非白名单字段被拒绝
- **WHEN** 客户端在 `findMany` 中使用 `id`、`username`、`createdAt` 之外的过滤字段
- **THEN** 服务端返回输入校验错误，且不得执行数据查询

### Requirement: 系统 user 查询 MUST 强制分页和排序边界
系统 MUST 对 `findMany` 强制执行 `take` 默认 `20`、`take` 最大 `50`、`skip` 最小 `0`、`skip` 最大 `1000`，并且 MUST 仅接受已配置的可排序字段。

#### Scenario: take 超过上限被拒绝
- **WHEN** 客户端提交 `take > 50` 的 `findMany` 请求
- **THEN** 服务端返回输入校验错误

#### Scenario: skip 超过上限被拒绝
- **WHEN** 客户端提交 `skip > 1000` 的 `findMany` 请求
- **THEN** 服务端返回输入校验错误

### Requirement: 系统 user 查询 MUST 强制隐式租户注入
系统 MUST 从请求上下文提取 `orgName`，并 MUST 将其应用到 `me`、`findOne`、`findMany` 的全部查询执行中；GraphQL 输入与输出类型 MUST NOT 暴露租户字段。

#### Scenario: 公开契约不包含租户字段
- **WHEN** 客户端查看 user 查询相关输入输出类型
- **THEN** GraphQL 契约中不出现 `orgName`

#### Scenario: 查询结果按上下文租户隔离
- **WHEN** 客户端在 runtime user 路由执行 `findMany`
- **THEN** 返回记录仅属于当前认证上下文的 org
