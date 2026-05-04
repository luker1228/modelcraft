## Context

当前 runtime 与相关前端工作区仍主要以 `developer / end-user` 为入口心智。这个划分对登录页和 cookie 路由是有效的，但在数据面已经暴露出明显局限：

- 数据访问能力本身并不天然属于“开发者”或“终端用户”，而是属于不同层级的主体在不同作用范围内访问 data plane。
- 管理员经常需要访问用户级能力，例如查看 catalog、查看某个 model 的运行态数据、验证某个 `EndUserRef` 字段的候选项。
- 部分操作要求“当前是某个具体用户”才有业务语义（例如 `me`、owner/RLS 驱动的查询），而另一些操作则要求“当前是全权限运行态管理者”。
- 未来若同时支持用户级 API key 与管理员 API key，继续按“入口系统”复制接口和鉴权会放大重复实现与权限歧义。

因此，这次设计不再把 `developer / end-user` 当作核心权限术语，而是改为：

- **组织级身份**
- **项目级身份**
- **用户级身份**

同时把 runtime 访问明确放入 **data plane**，并为 data plane 定义统一 principal 和访问模式。

## Goals / Non-Goals

**Goals:**

- 为 runtime/data-plane 能力建立统一的三层身份模型：组织级、项目级、用户级。
- 定义统一 principal 结构，使 JWT、API key、gateway trusted identity 最终都收敛到同一 access model。
- 定义三种 data-plane 访问模式：用户本人访问、管理员扮演用户访问、管理员全权限访问。
- 明确“能被用户级访问的能力，在满足上下文条件时也能被上层管理员访问”的兼容规则。
- 把 database catalog、model catalog、model schema、runtime data、`meta/user` 等能力统一纳入同一 data-plane 授权语义。
- 将 catalog 可见性提升为用户级访问的基础能力，而不是运行态数据查询之外的附属接口。

**Non-Goals:**

- 本次不实现完整的 API key 生命周期管理与签发流程。
- 本次不重写现有登录页、cookie 中间件或前端所有路由结构。
- 本次不定义组织级/项目级 RBAC 的完整后台 UI。
- 本次不直接移除现有 `/graphql/end-user/...` 或 `/graphql/org/...` 外部路径。
- 本次不要求一次性完成所有 runtime endpoint 的协议统一，只要求先统一 access model。

## Decisions

### 1) 用“三层身份”替代 `developer / end-user` 作为 runtime 权限核心术语

**Decision:** 在 data plane 中使用三层 actor class：

- `TENANT_LEVEL`（当前实现由 `orgName` 承载）
- `PROJECT_LEVEL`
- `USER_LEVEL`

其中 `USER_LEVEL` 表示真正的数据使用主体；上两层表示可进入数据面的管理主体。

**Why:**

- 这比“开发者 vs 终端用户”更贴近实际业务语义。
- 组织/项目两层管理者都可能使用 runtime 能力，但它们并不是“终端用户”。
- 未来无论 JWT 还是 API key，都应先解析为 actor class，再决定具体权限。

**Alternatives considered:**

- 继续沿用 `developer / end-user`
  - 放弃原因：会把控制面入口概念错误地延伸到数据面权限语义。
- 仅保留 `admin / user`
  - 放弃原因：无法表达组织级与项目级的作用范围差异。

### 2) 将 runtime principal 拆成主体、范围、凭证、模式四个维度

**Decision:** 统一 data-plane principal 至少包含以下维度：

- actor class：组织级 / 项目级 / 用户级
- scope：org / project 可访问范围
- credential type：JWT / API key / trusted gateway identity
- access mode：`SELF_USER`、`IMPERSONATED_USER`、`RUNTIME_ADMIN`

必要时再附加：

- `effectiveUserId`
- `bypassRLS`
- `permissions`

**Why:**

- “是谁”与“拿什么进来”是两个正交问题。
- “是不是某个具体用户”与“是不是管理员”也是两个正交问题。
- 这能避免后续为 API key 再复制一套 end-user / developer runtime middleware。

**Alternatives considered:**

- 直接用 token 类型推断全部语义
  - 放弃原因：JWT 与 API key 会导致同一业务能力分叉出多套鉴权逻辑。
- 只做一个 `role` 字段
  - 放弃原因：无法清楚表达 impersonation 与 runtime admin 的差异。

### 3) data plane 统一三种访问模式

**Decision:** data plane 统一只认三种访问模式：

- `SELF_USER`：用户级主体以自己身份访问
- `IMPERSONATED_USER`：组织级/项目级主体扮演某个具体用户访问
- `RUNTIME_ADMIN`：组织级/项目级主体以全权限模式访问

**Why:**

- 这三种模式已经覆盖当前讨论的全部 runtime 语义。
- `me`、owner/RLS、preset policy 等能力都需要区分“是否存在具体用户主体”。
- 全权限访问不应伪装成某个特殊用户账号，否则会污染审计与 `me` 语义。

**Alternatives considered:**

- 用“超级用户”伪装为特殊用户级账号
  - 放弃原因：`me`、RLS、owner 等都将出现语义冲突。
- 为每类接口硬编码单独权限分支
  - 放弃原因：后续 API key 和新接口会导致重复判断爆炸。

### 4) 上层管理员访问用户级能力，不靠复制接口，而靠 access mode 兼容

**Decision:** 统一规则定义为：

- 任何 `USER_LEVEL + SELF_USER` 可访问的 data-plane 能力，在满足上下文条件时也必须允许：
  - `TENANT_LEVEL / PROJECT_LEVEL + IMPERSONATED_USER`
  - `TENANT_LEVEL / PROJECT_LEVEL + RUNTIME_ADMIN`（仅限不依赖具体用户主体的能力）

其中：

- `me`、owner/RLS 驱动查询等**必须绑定具体用户主体**的能力，仅允许 `SELF_USER` 或 `IMPERSONATED_USER`
- catalog、`findOne`、`findMany`、全量调试查询等**不依赖具体用户主体**的能力，可允许 `RUNTIME_ADMIN`

**Why:**

- 这满足“能被用户级访问的能力，理论上都能被上层管理员访问”的规则。
- 同时避免把管理员强行伪装成普通用户。

**Alternatives considered:**

- 将管理员与用户能力完全拆成两套平行 API
  - 放弃原因：会破坏能力一致性，并放大未来 API key 的重复实现。

### 5) catalog 可见性是用户级数据访问的基础能力

**Decision:** 用户级主体若可查看某个 project 下的数据，则必须先能在其授权范围内查看：

- database catalog
- model catalog
- model schema subset

在此基础上才允许进入具体 runtime 数据查询/写入。

**Why:**

- 真实产品流程里，用户不可能先知道要查哪张表、哪个 model，再去查询数据。
- catalog visibility 本身就是 data-plane permission 的一部分，而不是额外的便利接口。

**Alternatives considered:**

- 仅对数据查询做权限控制，catalog 统一开放或统一隐藏
  - 放弃原因：前者泄露结构信息，后者又使用户无法自助进入数据面。

### 6) 先统一 access model，再逐步收敛路由 facade

**Decision:** 本次先把 principal 与 access mode 模型固定下来；现有外部路径可暂时保留，但后续都必须能映射到统一 data-plane principal。

这意味着：

- `/graphql/org/.../db/.../model/...`
- `/graphql/end-user/...`
- `/graphql/runtime/.../meta/user`

都不再各自定义一套“这是 developer 接口 / 这是 end-user 接口”的独立权限语义，而是挂靠同一个 access model。

**Why:**

- 路由统一是后续工作，但如果 access model 不先统一，路由永远会继续漂移。
- 先收敛语义层，可以避免边改 API 边改概念造成的反复返工。

## Risks / Trade-offs

- [Risk] 现有 `developer / end-user` 文档、代码、变量命名大量存在，短期内会产生双术语并存。 → Mitigation：先在 proposal/spec/design 中固定新术语，并在实现阶段提供映射说明。
- [Risk] `RUNTIME_ADMIN` 与 `IMPERSONATED_USER` 的边界若定义不清，容易导致 `me`、RLS、审计行为混乱。 → Mitigation：将“是否存在具体用户主体”设为硬规则，并在 spec 中明确限制。
- [Risk] catalog visibility 被纳入 data-plane permission 后，现有 end-user catalog 接口可能需要补权限判断。 → Mitigation：先把 requirement 明确，再分步迁移 catalog 相关 resolver。
- [Risk] 后续 API key 落地时，若绕开统一 principal 模型，仍可能重新长出四套凭证分支。 → Mitigation：把 credential type 纳入规范要求，禁止新接口直接以 token 类型写业务分支。

## Migration Plan

1. 在 spec 层先确立三层身份、principal 维度与三种 data-plane access mode。
2. 梳理现有 runtime、catalog、`meta/user` 能力，标注哪些能力依赖具体用户主体，哪些可开放给 `RUNTIME_ADMIN`。
3. 在后端引入统一 principal 解析边界，使现有 JWT / trusted identity 最终映射到统一 runtime principal。
4. 补齐前端与 BFF 的术语映射，将现有 `developer / end-user` workspace 入口逐步解释为“控制面入口”与“数据面入口”，而不是最终权限语义。
5. 后续新增 API key 时直接接入统一 principal 模型，不再复制一套按入口系统分裂的 runtime 接口。

回滚策略：

- 若实现期发现统一 principal 迁移面过大，可先保留旧路由 facade，只在内部建立 principal 适配层；不需要回退 spec 层术语与能力定义。

## Open Questions

- 组织级与项目级身份的正式英文命名，是否在本项目中固定为 `tenant-level / project-level`，还是继续保留 `org-level / project-level` 作为实现对齐术语。
- `RUNTIME_ADMIN` 是否允许写操作默认放开，还是仍需额外 capability/permission 明确授权。
- `IMPERSONATED_USER` 的触发方式最终是显式传入 `effectiveUserId`、会话态切换，还是调试上下文专用 header。
- 用户级 API key 是否按 org 级账号签发后再附着 project scope，还是直接签发 project-scoped credential。
