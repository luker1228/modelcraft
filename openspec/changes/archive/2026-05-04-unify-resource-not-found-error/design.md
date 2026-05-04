## Context

当前后端在 Org / Project / EndUser 三套 GraphQL Schema 中定义了大量资源不存在错误类型，并在各 resolver 适配层按业务错误码分别映射为不同 `__typename`。前端与 BDD 因此存在大量 `... on XxxNotFound` 片段和字符串断言。该改造属于跨模块收敛（Schema、后端适配、前端文档与测试），并且明确允许破坏性更新，不要求兼容旧错误类型。

## Goals / Non-Goals

**Goals:**
- 将所有“资源不存在”语义统一为单一 GraphQL 错误类型 `ResourceNotFound`。
- 保留资源区分能力，通过结构化字段 `resourceType`（枚举）表达，而不是通过多个 `__typename`。
- 统一后端错误映射策略：所有 `NOT_FOUND.*` 业务错误都映射到 `ResourceNotFound`。
- 让前端和 BDD 只依赖一个 not-found 错误类型，降低后续演进成本。

**Non-Goals:**
- 不调整非 not-found 错误（如 `InvalidInput`、`Conflict`、`Unauthorized`）的建模。
- 不改动业务领域层错误定义与抛错位置（仍可使用 `NOT_FOUND.MODEL` 等细粒度 code）。
- 不提供兼容窗口或双写方案。

## Decisions

### 决策 1：以 `ResourceNotFound` 取代所有 `XxxNotFound`
- 方案：在三套 Schema 中统一使用 `type ResourceNotFound implements Error { message: String!, resourceType: ResourceType! }`。
- 原因：调用方只需处理一个错误类型，显著减少 union 成员数量与分支复杂度。
- 备选方案：保留旧类型并新增统一接口/父类型。
- 不采用原因：保留旧类型会延续复杂度，不能达到“破坏性收敛”的目标。

### 决策 2：`resourceType` 使用枚举而非字符串
- 方案：定义统一 `enum ResourceType`（如 MODEL/PROJECT/CLUSTER/USER/PROFILE/GROUP/ENUM/END_USER/RBAC_PERMISSION/RBAC_BUNDLE/RBAC_ROLE 等）。
- 原因：枚举有更强约束与可读性，前后端类型检查更稳定。
- 备选方案：`resourceType: String!`。
- 不采用原因：字符串缺乏约束，易出现拼写漂移。

### 决策 3：保留 `extensions.code` 的细粒度错误码
- 方案：继续透传业务错误码（如 `NOT_FOUND.MODEL`）到 GraphQL `extensions.code`，仅把 GraphQL union 类型统一为 `ResourceNotFound`。
- 原因：运维排障和日志定位仍可利用细粒度 code，不与类型收敛冲突。
- 备选方案：统一 code 为 `NOT_FOUND`。
- 不采用原因：会丢失调试信息，增加排查成本。

### 决策 4：适配层集中实现映射，不在业务层改造
- 方案：主要修改 `internal/interfaces/graphql/**/adapter/*.go`，按 `bizErr.Info().GetCode()` 的前缀 `NOT_FOUND` 归一映射并填充 `resourceType`。
- 原因：最小化业务层改动，降低回归风险。
- 备选方案：在领域层直接改成新错误对象。
- 不采用原因：会扩大改造面并破坏当前分层职责。

## Risks / Trade-offs

- [风险] 前端/BDD 大量依赖旧 `... on XxxNotFound` 片段，迁移遗漏会导致编译或测试失败。
  → Mitigation：按模块批量替换并在 codegen 后统一跑类型检查与 BDD 回归。

- [风险] `resourceType` 枚举覆盖不全，某些历史错误无法正确映射。
  → Mitigation：先基于现有 `NOT_FOUND.*` 清单定义枚举；适配器默认兜底为 `UNKNOWN` 并记录日志。

- [权衡] 破坏性更新可一次收敛，但会要求前后端同时升级。
  → Mitigation：在同一变更内完成后端 schema/adapter、前端文档与测试一起提交，避免中间态。

## Migration Plan

1. 在 Org / Project / EndUser Schema 增加 `ResourceType` 与 `ResourceNotFound`，并替换所有错误 union 的 `*NotFound` 成员。
2. 执行 GraphQL 代码生成，修复后端编译错误。
3. 修改后端各错误适配器，把 `NOT_FOUND.*` 映射为 `ResourceNotFound{resourceType}`。
4. 批量修改前端 `graphql-docs.ts` 中 `... on XxxNotFound` 为 `... on ResourceNotFound`，并更新运行时分支逻辑。
5. 更新 BDD feature/step 中错误断言为 `ResourceNotFound + resourceType`。
6. 运行 lint、类型检查、后端测试与 BDD 验证；修复剩余不一致点。

## Open Questions

- `ResourceType` 是否需要在 v1 即覆盖 `API_KEY`、`MEMBERSHIP` 等当前低频 not-found 资源？
- 对于历史 `NotFound`（无子类型）应映射到哪个 `resourceType`（建议 `UNKNOWN`）？
- 是否要求在错误 message 中继续包含资源标识（id/name），还是仅依赖 `extensions.code` 与上下文日志？