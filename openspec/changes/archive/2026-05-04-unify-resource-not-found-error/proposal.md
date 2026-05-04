## Why

当前 GraphQL 在多个域中定义了大量 `*NotFound` 错误类型（如 `ModelNotFound`、`ClusterNotFound`、`UserNotFound`），调用方需要维护过多分支判断，导致接口认知成本和演进成本偏高。该能力尚未正式落地到外部稳定版本，适合现在进行一次破坏性收敛。

## What Changes

- **BREAKING**：将 GraphQL Schema 中分散的 `*NotFound` 错误类型统一为 `ResourceNotFound`。
- **BREAKING**：将各 Query/Mutation 错误 union 中的 `*NotFound` 成员替换为 `ResourceNotFound`。
- 新增 `resourceType`（枚举）用于区分资源类别（如 MODEL、PROJECT、CLUSTER、USER 等）。
- 后端错误适配器统一把 `NOT_FOUND.*` 业务错误映射为 `ResourceNotFound`。
- 前端 GraphQL 文档、类型与运行时错误处理改为统一消费 `ResourceNotFound`。
- BDD 场景与 step 断言从具体 `*NotFound` 类型迁移到 `ResourceNotFound + resourceType`。

## Capabilities

### New Capabilities
- `graphql-resource-not-found-unification`: 统一 GraphQL 资源不存在错误建模、映射与调用方消费方式。

### Modified Capabilities
- （无）

## Impact

- 影响后端 GraphQL Schema（`api/graph/org|project|end_user/schema/*.graphql`）与代码生成结果。
- 影响后端 GraphQL 错误适配层（`internal/interfaces/graphql/**/adapter/*.go`）的错误映射逻辑。
- 影响前端 GraphQL 文档与 codegen 产物（`modelcraft-front/src/api-client/**/graphql-docs.ts`、`src/generated/graphql.ts`）。
- 影响 BDD 用例与步骤定义（`tests-bdd/features/**`、`tests-bdd/step-definitions/**`）。
- 对调用方属于破坏性升级：旧 `__typename === XxxNotFound` 分支将失效。