## Why

`/graphql/runtime/org/{orgName}/meta/user` 这个固定路由承载的是来自 `end_user` 的元数据用户记录。该数据稳定但敏感，如果复用不受限的通用 runtime 查询能力，容易带来不必要的数据库压力，并削弱元数据访问的权限边界。

## What Changes

- 在固定 user runtime 路由上引入专属的**系统用户查询协议**。
- 路由命名统一为 `meta` 资源空间（如 `/graphql/runtime/org/{orgName}/meta/user`），用于承载 org 级共享元数据实体；后续可扩展 `/meta/org` 等端点。
- `meta` 与 `model` 资源空间语义分离，但共用同一套 runtime 鉴权框架与租户注入机制。
- 仅支持三个查询操作：`me`、`findOne`、`findMany`。
- 保留 `findMany` 协议形状以复用前端筛选组件，但在服务端做严格白名单约束。
- 租户范围（`orgName`）仅由服务端上下文注入，不在 GraphQL 类型与过滤条件中暴露。
- 增加明确的查询限制与输入校验，避免高成本或无界元数据查询。

## Capabilities

### New Capabilities
- `runtime-user-system-model-query`: 为 `meta/user` 提供专属 runtime GraphQL 协议，采用受限的 `me/findOne/findMany` 语义。

### Modified Capabilities
- `enduser-org-account`: 扩展面向 runtime 的用户查询行为，在固定 runtime `meta/user` 路由上支持受限版 `me/findOne/findMany` 协议。

## Impact

- 影响 API 面：runtime `meta/user` 固定路由上的 End-user GraphQL schema。
- 影响后端模块：`internal/interfaces/graphql/enduser/*` 的 resolver 与生成文件。
- 影响前端集成：runtime 列表/筛选查询构建器可继续复用 `findMany + Tuser*` 形状。
- 安全与性能：通过字段/操作符白名单、分页上限、强制租户注入提升防护能力。
