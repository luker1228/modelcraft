## Context

当前开发者认证处于中间态：

- frontend 已经把大部分 developer auth 流量路由到 `/api/auth/*`，并把 design GraphQL 路由到 `/api/bff/graphql/*`
- backend design middleware 仍保留 direct developer JWT verification 的假设
- legacy issuer 与 claims 处理仍然和部分已迁移的 developer identity 概念混在一起
- CLI 路径还没有被明确收敛到同一套 gateway-first trust model

目标拓扑现在已经明确：

- web: `front -> bff -> gateway -> backend`
- cli: `cli -> gateway -> backend`

这意味着 gateway 是开发者认证的唯一权威，而 backend design API 则转为 gateway-trusted resource server。

## Goals / Non-Goals

**Goals:**
- 让 gateway 成为开发者认证的唯一验签者
- 定义一套稳定的 gateway-to-backend trusted developer identity contract
- 移除 backend design API 对 direct developer bearer token validation 的依赖
- 让 web 与 CLI 的 developer request path 统一到同一套 trust model
- 在迁移完成后允许破坏性清理 backend 侧的 legacy developer JWT compatibility

**Non-Goals:**
- 不重设计 end-user authentication
- 不重设计 runtime end-user JWT validation
- 不引入新的 frontend auth UX
- 不定义超出 backend contract 之外的 gateway 内部实现细节

## Decisions

### 1. Gateway 是唯一的 developer auth boundary

Decision:
- 所有 developer token validation、issuer compatibility、session/refresh handling 都在 gateway 里完成

Rationale:
- 把 `modelcraft` 到 `mc-developer` 的迁移逻辑集中在一个位置
- 避免 BFF、gateway、backend 三处同时维护认证真相
- 让浏览器和 CLI 共用同一套 contract

Alternative considered:
- 继续让 backend 直接校验 developer bearer JWT
- 放弃这个方案，因为它会保留混杂的信任边界，并让 legacy compatibility 长期滞留在 backend 代码中

### 2. Backend design API 信任 gateway identity，而不是 direct developer JWT

Decision:
- backend design GraphQL 和受保护的 design REST 消费 gateway 注入的 trusted developer identity
- backend 不再把 direct developer bearer token 当作 canonical design-time auth path

Rationale:
- 与既定请求拓扑一致
- 让 backend 关注 authorization 与业务逻辑本身
- 允许破坏性删除 backend 侧的 legacy JWT compatibility

Alternative considered:
- 长期同时支持 direct backend JWT validation 和 gateway-trusted mode
- 放弃这个方案，因为两种模式长期共存会让迁移始终不完整，也会让故障排查更困难

### 3. Trusted identity contract 必须显式且最小化

Decision:
- gateway 向 backend 转发一组稳定的 trusted developer identity headers
- backend 在将请求视为已认证之前，必须验证这组 contract 的存在性和来源

Expected contract shape:
- 一个承载 user identity 的必需 developer identity header
- 一个表明请求来源于已认证 gateway 的必需 trust marker
- 仅在现有 backend authorization 逻辑需要时，才携带可选的 organization 或 actor context

Rationale:
- 单独依赖 `X-User-ID` 过于隐式
- 显式 trust marker 能让中间件意图更可审计，也更安全地演进

Alternative considered:
- 继续只保留 `X-User-ID`
- 放弃这个方案，因为它会模糊信任边界，也更难识别潜在的伪造假设

### 4. Frontend BFF 在 developer auth 中保持 transport-only 角色

Decision:
- frontend BFF 继续承担 same-origin transport 与 cookie bridge 角色
- BFF 不成为第二个 developer token verifier 或 compatibility layer

Rationale:
- 保持浏览器集成简单
- 防止 frontend 层再长出第二套独立 auth 实现

### 5. Gateway 路径完成后允许做破坏性清理

Decision:
- legacy backend-side direct developer JWT verification 路径可以直接删除，而不是长期保留在 compatibility flag 后面

Rationale:
- 用户已经明确允许 breaking changes
- 这次迁移的目标是收敛为单一路径，而不是永久保留多套并存方案

## Risks / Trade-offs

- [Risk] Gateway 与 backend 在 trusted header 语义上可能在 rollout 期间不一致 -> Mitigation: 先在一个聚焦 middleware 的迁移步骤里定义 contract，再做后续清理。
- [Risk] 现有测试可能隐式依赖 direct backend bearer token validation -> Mitigation: 将测试改为使用 gateway-trusted request fixtures，并显式删除 legacy 假设。
- [Risk] CLI 客户端可能仍有未记录的 direct backend 调用 -> Mitigation: 在移除 legacy backend validation 前，先完整 inventory CLI call sites。
- [Risk] Frontend 代码目前仍可能给 BFF GraphQL 请求附带 bearer token -> Mitigation: 短期内允许 gateway 边界接收它，但必须保证 backend 的信任建立在 gateway contract 上，而不是下游 bearer parsing 上。

## Migration Plan

1. 定义并文档化 gateway 到 backend 的 trusted developer identity contract。
2. 更新 backend design-time middleware，使其通过该 trusted contract 完成认证。
3. 校准 frontend BFF routes 与 GraphQL paths，使其默认走 gateway-authenticated forwarding。
4. 盘点并校准 CLI design-time calls 到 gateway path。
5. 移除 backend design API 上的 direct developer JWT verification 行为。
6. 在 gateway compatibility 上线后，删除 backend 对 legacy developer issuer 的兼容。

Rollback:
- 如果 gateway contract rollout 失败，临时恢复先前的 backend middleware 行为
- rollback 必须保持短期，避免重新引入 split auth truth

## Open Questions

- trusted developer identity contract 应该由哪些精确 header 名称组成？
- gateway 是否已经注入了一个 backend 可验证的 internal trust marker，还是需要新增？
- CLI 当前是否存在任何实际绕过 gateway 的 design-time REST 调用？
