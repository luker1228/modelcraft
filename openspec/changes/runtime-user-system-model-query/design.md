## Context

当前用户元数据存储于 `end_user`，通过固定 runtime 路由访问：
`/graphql/runtime/org/{orgName}/meta/user`。

该路由在概念上属于**系统模型**（`user`）查询入口，而不是通用自定义模型查询入口。我们需要同时满足两点：
- 保持前端查询体验一致（尤其是 `findMany` 风格筛选）；
- 对元数据场景施加严格的安全与性能约束。

## Goals / Non-Goals

**Goals:**
- 为系统 `user` 模型定义专属 runtime GraphQL 协议，支持 `me`、`findOne`、`findMany`。
- 保留 `findMany` 协议形状，复用已有前端筛选构建模式。
- 通过字段/操作符/排序/分页白名单限制查询能力。
- 通过上下文注入 org，强制租户隔离。
- 保持 `user` 作为“可被关联的稳定模型标识”（`id` 与显示字段语义稳定）。

**Non-Goals:**
- 不构建 end-user GraphQL 的通用 runtime 查询引擎。
- 不在公开 GraphQL schema 中暴露 org 字段。
- 不支持任意深度布尔条件树与无限制动态操作符。
- 不在该路由新增写入/变更类 mutation。

## Decisions

### 1) 系统模型专属协议优先于通用模型协议
**Decision:** 将 `user` 作为固定 runtime 路由下的系统模型，采用专用 GraphQL 类型与输入定义。

**Why:**
- `end_user` 属于元数据域，变更慢且安全要求高。
- 固定 schema 更容易落实容量控制和安全策略。
- 与“租户字段仅为实现细节”的约束一致。

**Alternative considered:** 复用完整通用 runtime 模型查询链路。
- 放弃原因：表达能力过宽，带来更高负载与越界风险。

### 2) 保留 `findMany` 形状，但能力受限
**Decision:** 保留 `findMany(where, orderBy, skip, take)` 与 `Tuser*` 命名，同时强制白名单约束。

**Why:**
- 前端筛选 UI 与 query builder 可低成本复用。
- 通过服务端校验和有限 schema 表达面控制风险。

**Alternative considered:** 仅使用 `users(input)` 专属接口。
- 当前不选：会损失协议复用并增加迁移成本。

### 3) 白名单字段与边界限制
**Decision:**
- 可过滤字段：`id`、`username`、`createdAt`。
- 可排序字段：`createdAt`（服务端可追加 `id` 作为稳定次序）。
- 分页限制：`take` 默认 20、最大 50；`skip` 默认 0、最大 1000。

**Why:**
- 满足当前业务需求。
- 避免高成本扫描和无界请求。

**Alternative considered:** 直接暴露 `isForbidden` 与更大操作符矩阵。
- 延后处理：只有在明确业务场景需要时再增量开放。

### 4) 租户范围始终隐式且强制
**Decision:** 所有查询都从上下文注入 `orgName`；客户端不能传入或覆盖租户条件。

**Why:**
- 防止跨租户访问与误泄漏。
- 保持 GraphQL 契约面向业务语义，而非存储实现。

### 5) 结果结构沿用 runtime 元信息
**Decision:**
- `me` 与 `findOne` 返回 `TuserFindOneResult { item, timeCost, reqId }`。
- `findMany` 返回 `TuserFindManyResult { items, timeCost, reqId, totalCount? }`。

**Why:**
- 与现有 runtime 协议与可观测性字段保持一致。

## Risks / Trade-offs

- [Risk] 大 offset 的 `skip` 仍可能影响性能。
  → Mitigation: 设定硬上限（`<=1000`），并预留后续迁移 cursor 分页路径。

- [Risk] 白名单过窄导致未来产品诉求无法直接满足。
  → Mitigation: 采用受控扩展流程，按字段与操作符逐项开放。

- [Risk] “系统模型 vs 自定义模型”双心智可能增加维护者理解成本。
  → Mitigation: 在 schema 注释与架构文档中明确系统模型原则。

## Migration Plan

1. 新增系统 user runtime 查询所需 schema 类型与查询字段。
2. 在 resolver 层实现输入校验、租户注入与应用服务映射。
3. 上线时增加不合法过滤/操作符的日志观测。
4. 前端 runtime user 调用切换到 `me/findOne/findMany` 协议。
5. 若存在遗留调用路径，在迁移完成后进入弃用窗口。

## Open Questions

- `findOne` 首版是否同时支持 `id` 与 `username` 唯一查询，还是仅支持 `id`？
- `findMany` 是否必须返回 `totalCount`，还是可选以减少 count 查询开销？
- 是否需要在当前 internal token 之外，为该路由增加独立权限开关？

## 拟定 GraphQL 形态

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

input StringFilter {
  eq: String
  contains: String
  startsWith: String
  in: [String!]
}

input IDFilter {
  eq: ID
  in: [ID!]
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
