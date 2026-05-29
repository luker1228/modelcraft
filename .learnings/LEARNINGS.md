# Learnings

Corrections, insights, and knowledge gaps captured during development.

**Categories**: correction | insight | knowledge_gap | best_practice

---

## [LRN-20260429-001] best_practice

**Logged**: 2026-04-29T02:48:33Z
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
RBAC 权限包添加资源策略弹窗应基于 `modelDatabaseCatalog` 先选数据库，再查询该库的数据表。

### Details
在 `/roles/bundles` 场景中，若直接加载所有数据库的全部模型，会导致请求量过大且交互不清晰。权限配置并不需要数据库全集浏览，应先让用户确定数据库，再请求该数据库下的数据表列表，减少无效请求并让流程更符合“权限按库/表定位”的操作习惯。

### Suggested Action
统一该弹窗流程：Step1 选择数据库（来自 `modelDatabaseCatalog`）→ Step2 展示并选择该库数据表（`GET_MODELS_FOR_RELATION` 按库查询）。避免使用 `listDatabases` 作为此流程数据源。

### Metadata
- Source: user_feedback
- Related Files: modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage.ts; modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx
- Tags: rbac, permissions, database-selection, dialog-flow

---

## [LRN-20260429-002] insight

**Logged**: 2026-04-29T02:57:57Z
**Priority**: medium
**Status**: pending
**Area**: backend

### Summary
后端已具备“模型预设策略纯计算”能力，但当前未通过 GraphQL 对外暴露。

### Details
`EndUserPermissionAppService.ListVirtualPresetsByModel` 可根据模型 owner 字段只读计算可用预设集合（不落库），但 GraphQL schema 仅暴露了 `applyEndUserPresetPolicy` / `addEndUserPresetToBundle` 这类写库 mutation，缺少对应 preview/query 接口。

### Suggested Action
若前端需要“不过 DB，直接显示默认预设策略”，新增 Query（如 `availableEndUserPresets(modelId: ID!): [EndUserPermissionPreset!]!`）调用 `ListVirtualPresetsByModel`，并在接口层明确该能力只读、无副作用。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/internal/app/rbac/permission_app.go; modelcraft-backend/internal/interfaces/graphql/project/rbac.resolvers.go; modelcraft-backend/api/graph/project/schema/rbac.graphql
- Tags: rbac, preset, preview, graphql

---

## [LRN-20260429-003] best_practice

**Logged**: 2026-04-29T03:20:00Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
RBAC 预设策略应按“虚拟模板 → 关联触发物化实例”来建模和沟通，而不是“系统启动即存在固定行数据”。

### Details
当前后端已体现双阶段语义：未关联时，预设以可计算能力存在（virtual，不落库）；关联权限包/模型时，才生成可持久化的权限实例（materialized，带具体 rowPolicy 和 permissionId）。这种解释能同时覆盖“预设初始不存在”与“关联后存在”，并与当前鉴权链（按 permissionId 展开）和快照机制保持一致。

### Suggested Action
在 PRD/接口文案统一术语：Template（虚拟）/ Binding（关联意图）/ Instance（物化权限）。对外说明“关联动作是实例化动作”，避免把预设误解为必须预写入 DB 的静态数据。

### Metadata
- Source: user_feedback
- Related Files: modelcraft-backend/internal/app/rbac/permission_app.go; modelcraft-backend/internal/app/rbac/bundle_app.go; modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql
- Tags: rbac, preset, lazy-materialization, domain-modeling

---

## [LRN-20260503-A1N] best_practice

**Logged**: 2026-05-03T00:00:00Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
GraphQL NotFound 错误统一改造时，真正高成本在 Schema Union/Adapter/BDD，不在前端运行时分支。

### Details
调研发现前端手写运行时代码仅少量依赖具体 `__typename`（如 `ProfileNotFound`），但 `src/api-client/*/graphql-docs.ts` 与 BDD step/feature 中大量使用 `... on XxxNotFound` 片段和字符串断言。若后端直接将多种 `XxxNotFound` 合并为单一 `ResourceNotFound`，主要改造压力集中在 GraphQL schema union 成员、Go error adapter 映射、以及 BDD 断言同步。

### Suggested Action
优先采用分阶段迁移：先引入 `ResourceNotFound`（包含 `resourceType`）并保持兼容窗口，再逐步替换各处 inline fragment 与测试断言，最后收敛旧类型。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/api/graph/project/schema/model.graphql; modelcraft-backend/internal/interfaces/graphql/project/adapter/model_error_adapter.go; modelcraft-front/src/api-client/rbac/graphql-docs.ts; tests-bdd/step-definitions/profile.steps.ts
- Tags: graphql, error-modeling, notfound, compatibility, migration

---

## [LRN-20260503-A2O] best_practice

**Logged**: 2026-05-03T00:00:00Z
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
OpenSpec 的 spec-driven schema 中，`tasks` 虽是 applyRequires，但仍可能被 `specs` 依赖阻塞，不能只生成 proposal/design/tasks 三件套。

### Details
在新建 change 后查看 `openspec status --json`，`tasks` 处于 blocked，缺失依赖为 `design` 与 `specs`。若跳过 `specs`，即使用户口头只提三类工件，也无法达到 apply-ready。

### Suggested Action
对所有变更先执行 `openspec status --change <name> --json`，按 artifact 依赖顺序生成，直到 `applyRequires` 中项目状态为 done。

### Metadata
- Source: investigation
- Related Files: openspec/changes/unify-resource-not-found-error/proposal.md; openspec/changes/unify-resource-not-found-error/specs/graphql-resource-not-found-unification/spec.md; openspec/changes/unify-resource-not-found-error/tasks.md
- Tags: openspec, dependency, workflow, spec-driven

---

## [LRN-20260503-A2P] best_practice

**Logged**: 2026-05-03T10:40:00Z
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
前端 GraphQL codegen 的 schema 真相源是 `modelcraft-front/contract/graph/*`，后端仅修改 `api/graph/*` 不会自动让前端 codegen 通过。

### Details
本次后端已完成 `ResourceNotFound` 收敛并可 `generate-gql + go build`，但前端 `npm run codegen` 仍报 `Unknown type "ResourceNotFound"`。根因是前端 codegen 配置读取 `contract/graph/org|project/schema/*.graphql`，当前 contract 未同步到包含新类型的版本，导致文档校验阶段整体失败。

### Suggested Action
涉及 GraphQL schema 变更时，必须把“contract 同步”作为前置步骤（后端 subtree push → 前端 front-contract-pull），再执行前端 codegen 与 lint/typecheck。

### Metadata
- Source: error
- Related Files: modelcraft-front/codegen.ts; modelcraft-front/src/api-client/profile/graphql-docs.ts; modelcraft-front/src/api-client/rbac/graphql-docs.ts
- Tags: graphql, contract, codegen, schema-sync, resource-not-found

---

## [LRN-20260503-A2Q] best_practice

**Logged**: 2026-05-03T10:50:00Z
**Priority**: medium
**Status**: pending
**Area**: tests

### Summary
BDD 回归前必须先验证 `tests-bdd/support/hooks.ts` 的全局登录前置可用，否则场景会在 `BeforeAll` 直接失败。

### Details
在只跑 profile 场景时，Cucumber 未进入业务步骤就于 `BeforeAll` 报错：`AUTHENTICATION_FAILED: phone number not found`。这类失败与当前功能改动无关，会掩盖真正的场景断言结果。

### Suggested Action
执行 BDD 前先校验测试种子账号与环境变量（`.env.test`）可用；若 `BeforeAll` 登录失败，先修复环境再评估功能回归结果。

### Metadata
- Source: error
- Related Files: tests-bdd/support/hooks.ts
- Tags: bdd, test-environment, beforeall, auth

---

## [LRN-20260503-A2R] best_practice

**Logged**: 2026-05-03T11:05:00Z
**Priority**: high
**Status**: pending
**Area**: tests

### Summary
BDD 全局登录应具备“无账号自动注册+重试登录”自愈能力，避免环境未预置账号时整个回归被 `BeforeAll` 阻断。

### Details
`tests-bdd/support/hooks.ts` 之前仅尝试一次登录，若默认账号不存在就直接抛错。已改为：登录失败后自动执行 `register(TEST_LOGIN_PHONE, TEST_LOGIN_PASSWORD)`，并容忍“已存在”错误，再次登录。验证显示在未提供 `TEST_ACCESS_TOKEN` 时，profile 场景可正常启动并通过。

### Suggested Action
保留该自愈逻辑作为 BDD 默认前置；若仍登录失败，错误应明确区分“注册失败”与“登录失败”，便于快速定位环境问题。

### Metadata
- Source: error
- Related Files: tests-bdd/support/hooks.ts; tests-bdd/support/rest-client.ts
- Tags: bdd, auth-bootstrap, self-healing, beforeall

---

## [LRN-20260503-A2S] insight

**Logged**: 2026-05-03T11:20:00Z
**Priority**: medium
**Status**: pending
**Area**: tests

### Summary
出现 `Unknown type "ResourceNotFound"` 时，先做运行时 schema 探针和单场景复现，避免误判为代码未改。

### Details
本次先通过带签名 JWT 的 introspection 查询确认 `/graphql/org/{orgName}/` 运行时已包含 `ResourceNotFound`，再单独跑 `manage-profile.feature:12` 场景，验证 `updateMyProfile` 可通过。说明该报错并非当前代码库 schema 缺失，而是回归批量运行中的环境/上下文干扰。

### Suggested Action
把“schema 探针 + 单场景最小复现”作为 GraphQL 类型不匹配的第一诊断步骤，再决定是否改 schema/代码。

### Metadata
- Source: investigation
- Related Files: tests-bdd/step-definitions/profile.steps.ts; tests-bdd/support/graphql-client.ts
- Tags: graphql, bdd, troubleshooting, schema-validation

---

## [LRN-20260503-A2T] best_practice

**Logged**: 2026-05-03T11:35:00Z
**Priority**: high
**Status**: pending
**Area**: tests

### Summary
测试调用已下线 REST 路由时，不应直接 `res.json()`；需先按文本读取并做 JSON 容错，否则会被 404 文本响应触发 `SyntaxError`。

### Details
`profile` 场景使用的 `handleAuthProviderWebhook` 命中 `/api/webhook/auth_provider`，当前服务已无该路由，返回非 JSON 文本。原实现直接 `await res.json()` 导致 `Unexpected non-whitespace character after JSON`。已改为先 `res.text()` 再 `JSON.parse`，并在非 JSON 时返回结构化错误；同时将场景数据构造从 webhook 迁移为“基准注册用户 + 伪造 userId”的方式，消除该解析崩溃。

### Suggested Action
所有 BDD REST client 方法默认采用“text→try parse JSON”的安全解析模板；对于历史路由，优先检查 openapi 合约是否仍存在。

### Metadata
- Source: error
- Related Files: tests-bdd/support/rest-client.ts; tests-bdd/step-definitions/profile.steps.ts; modelcraft-backend/api/openapi/auth.yaml
- Tags: bdd, rest-client, json-parse, deprecated-endpoint

---

## [LRN-20260504-001] best_practice

**Logged**: 2026-05-04T00:00:00Z
**Priority**: high
**Status**: pending
**Area**: infra

### Summary
前端到后端的业务请求必须强制经过 Gateway，不能允许前端直连 Backend。

### Details
在 Gateway 架构更新中明确了边界：Gateway 负责外层 token 校验、头注入（`X-User-ID`/`X-Internal-Token`）和链路观测（request-id/traceparent）。若前端绕过 Gateway 直连 Backend，会破坏统一认证代理链路与部署联调基线，导致线上排障和安全边界不一致。

### Suggested Action
在部署联调清单加入硬性检查：前端 `BACKEND_URL` 必须指向 Gateway，Browser 网络流量不得出现直连 Backend；Backend 仅对 Gateway 开放访问路径。

### Metadata
- Source: user_feedback
- Related Files: ai-metadata/backend/deployment/README.md; ai-metadata/backend/development/gateway-architecture.md; modelcraft-gateway/internal/proxy/handler.go
- Tags: gateway, deployment, integration, security-boundary

---

## [LRN-20260504-002] best_practice

**Logged**: 2026-05-04T00:00:00Z
**Priority**: high
**Status**: pending
**Area**: docs

### Summary
Developer 与 EndUser 在 ModelCraft 中属于并行双体系，文档应以“对照表 + 链路 + 边界”统一呈现，避免混写。

### Details
用户明确要求将 Developer 和 EndUser 作为两套体系同步到 ai-metadata。单独描述 Gateway 或单侧流程容易遗漏体系差异（登录入口、Token 验证、GraphQL 路径、后端识别头），导致联调与排障时口径不一致。

### Suggested Action
新增双体系文档作为统一入口，并在 gateway/deployment/index/README 建立交叉引用；后续涉及认证或网关变更，优先更新该对照文档。

### Metadata
- Source: user_feedback
- Related Files: ai-metadata/backend/development/developer-enduser-system.md; ai-metadata/backend/development/gateway-architecture.md; ai-metadata/index.md
- Tags: auth, dual-system, docs-consistency, gateway

---

## [LRN-20260504-003] best_practice

**Logged**: 2026-05-04T00:00:00Z
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
前端文档中 `BACKEND_URL` 应明确解释为 Gateway 地址（BFF 上游），否则会误导成 Backend 直连。

### Details
在补齐 BFF 双体系路由时发现，历史表述“转发到 Go 后端”容易让开发者把 `BACKEND_URL` 配成 8080 Backend，进而绕过 Gateway 认证代理链路。该变量在当前实现中用于 BFF 路由上游，应统一口径为 Gateway 地址。

### Suggested Action
在前端架构与 API Client 文档中同时声明：`BACKEND_URL = Gateway`，并配套写明 Developer/EndUser 两套路由映射与“禁止直连 Backend”的硬性规则。

### Metadata
- Source: investigation
- Related Files: ai-metadata/front/development/architecture.md; ai-metadata/front/development/api-client-design.md; modelcraft-front/src/app/api/auth/[...path]/route.ts
- Tags: frontend, bff, gateway, env, docs

---

## [LRN-20260529-001] best_practice

**Logged**: 2026-05-29T05:33:21Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
统一用户体系下，组织员工可通过全局用户名登录；后端可先按全局用户名解析 orgName，再完成 end-user 登录并把 orgName 回传前端。

### Details
本次实现“首页统一登录入口 + 员工登录无需先确认组织”时发现，现有 end-user 登录接口虽然请求体要求 `orgName`，但底层数据模型已经满足两个关键前提：`users.name` 全局唯一，且 `user_orgs` 通过唯一约束保证每个用户只属于一个 Org。因此无需在首页额外收集组织标识，后端可以先按用户名做全局查找，利用返回用户实体上的 `OrgName` 补全后续项目权限查询、refresh session 存储和 access token 签发，再在登录响应中显式返回 `orgName` 给前端用于跳转。

### Suggested Action
后续如果继续扩展 end-user 入口（统一登录页、CLI 登录、邀请链路），优先复用“全局用户名解析 orgName → 登录响应回传 orgName”这条链路，而不是在前端重复增加组织确认步骤。

### Metadata
- Source: conversation
- Related Files: modelcraft-backend/internal/app/enduser/end_user_auth_service.go; modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go; modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserGlobalLoginForm.ts
- Tags: end-user-auth, login, org-resolution, unified-entry

---

## [LRN-20260505-001] correction

**Logged**: 2026-05-05T23:58:37+08:00
**Priority**: high
**Status**: pending
**Area**: docs

### Summary
`doPlan` 不是补齐前置设计产物的位置，`prd-page-splitter`、demo 页和 `backend-api` 契约文档应前置到 `doPropose`。

### Details
本次先把“功能页拆分 + demo 页”误放到了 `doPlan` 方向。用户明确纠正：这些都属于开工前的提案与设计收敛流程，应写入 `doPropose`，并且 `backend-api` 也要前置，先补齐足够的契约文档，再开始后续实施工作。

### Suggested Action
调整命令工作流时，先区分“提案阶段产物”和“计划阶段产物”。凡是需求总览、PRD 子页、页面 demo、API 契约这类前置设计资产，优先归入 `doPropose`。

### Metadata
- Source: user_feedback
- Related Files: .agents/commands/doPropose.md; .agents/commands/doPlan.md; .agents/agents/backend-api.md; .agents/agents/prd-page-splitter.md
- Tags: workflow, docs, correction, contract-first

---

## [LRN-20260511-001] correction

**Logged**: 2026-05-11T00:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
本仓库的本地部署 env 应按服务拆分维护，不要再引入 shared env 层。

### Details
在设计本地 Docker 部署目录时，最初方案包含 `shared.local.env`。用户明确纠正为“没有 shared env”，即部署配置应保持服务级边界，分别维护 `mysql / redis / backend / gateway / frontend / phpmyadmin` 的 env 文件，而不是再做一层共享抽象。

### Suggested Action
后续调整 `deploy/` 结构时，优先保持 service-owned env 文件；即使少量变量重复，也不要为复用而重新引入 shared env。

### Metadata
- Source: user_feedback
- Related Files: deploy/env/mysql.local.env.example; deploy/env/backend.local.env.example; deploy/env/gateway.local.env.example; deploy/env/frontend.local.env.example; deploy/env/redis.local.env.example; deploy/env/phpmyadmin.local.env.example
- Tags: deploy, env, docker, config-boundary

---

## [LRN-20260512-001] best_practice

**Logged**: 2026-05-12T03:10:00+08:00
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
终端用户登录成功但无项目权限时，后端不返回 `accessToken`，前端必须通过此字段是否存在来判断跳转分支。

### Details
后端 `end_user_auth_service.go` 的设计：只有当 `len(accessibleProjects) > 0` 时才颁发 JWT accessToken。若用户无项目权限，响应体缺少 `accessToken` 字段（仅返回 `userId`/`expiresAt`/`requestId`）。前端若不检查该字段直接跳转 workspace，workspace 页的 `useEndUserTokenReady` 会因 token 为空触发 `/auth/refresh`，而 refresh cookie 此时也未设置（因为也没有颁发 token），最终 401 → 跳回登录页，用户看到的是"停留在登录页没反应"。

### Suggested Action
登录 hook 中明确分支：`if (!data.accessToken) { router.push('/end-user/{orgName}/no-project-access'); return }`。无权限页面已存在（`/end-user/[orgName]/no-project-access/page.tsx`），引导用户联系管理员。

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts; modelcraft-backend/internal/app/enduser/end_user_auth_service.go; modelcraft-front/src/app/end-user/[orgName]/no-project-access/page.tsx
- Tags: end-user-auth, no-project-access, accessToken-conditional, login-branch

---

## [LRN-20260512-002] insight

**Logged**: 2026-05-12T03:10:00+08:00
**Priority**: medium
**Status**: pending
**Area**: frontend

### Summary
Zustand 内存 store 在 Next.js `router.push` 客户端导航后通常保留，但 `useState` 初始化时若用 hook 值而非 `getState()` 读取，可能因订阅延迟导致初始值为 false。

### Details
`useEndUserTokenReady` 原来写 `useState(!!accessToken && !isExpired())`，其中 `accessToken` 来自 Zustand hook 订阅。在 Next.js 客户端导航时，组件首次渲染时 hook 订阅可能还未同步，导致 `accessToken` 取到 null，ready 初始化为 false，触发 refresh 请求。修复：改用 `useEndUserAuthStore.getState()` 直接读取。

### Suggested Action
在 Next.js 客户端组件中，若需要在 useState 初始化时同步读取 Zustand store，应使用 `store.getState()` 而不是 hook 值；后者只在 React render cycle 内保证同步。

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/app/end-user/[orgName]/workspace/page.tsx; modelcraft-front/src/shared/stores/end-user-auth-store.ts
- Tags: zustand, nextjs, useState-init, getState, token-refresh

---

## [LRN-20260513-001] insight

**Logged**: 2026-05-13T18:50:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
Runtime GraphQL 的 end-user 权限解析链缺少 implicit role（Step 3），可能导致 admin 仍被判定无 insert 权限。

### Details
`FindPermissionsByEndUserAndModel` 当前仅合并了“用户显式角色绑定的 bundle + 用户直连 bundle”，未调用 `GetBundleIDsByImplicitRoles`。但仓库接口和 authz SQL 已定义 implicit role 为鉴权链 Step 3。若“admin 全权限”依赖 implicit role 或相关系统角色注入，这条遗漏会使 `ResolvedModelPermissions.Insert.Allowed=false`，在 create 路径直接抛出 `OPERATION_FAILED.PERMISSION`。

### Suggested Action
在 `FindPermissionsByEndUserAndModel` 中补齐 implicit bundle 收集与合并去重；并增加回归测试覆盖“仅 implicit role 授权时 create 应通过”的场景。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go; modelcraft-backend/db/queries/rbac/authz.sql; modelcraft-backend/internal/domain/modelruntime/model_resolver.go
- Tags: runtime-graphql, rbac, implicit-role, permission-denied, insert

---

## [LRN-20260513-002] insight

**Logged**: 2026-05-13T19:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
受保护的 `admin` 角色本身不自动拥有数据权限；若未绑定任何 bundle，runtime create 会被判定 `Permission denied: insert`。

### Details
本地 docker 排查显示：请求走的是 end-user runtime 路径（gateway 日志 `graphql/end-user/...`），后端仅查询了 `GetBundleIDsByUserExplicitRoles` 与 `GetBundleIDsByUserDirect`，随后直接报权限不足。数据库中该用户虽已绑定 `admin` 受保护角色，但 `end_user_role_bundles` / `end_user_user_bundles` / 项目级 implicit role bundle 均为 0，且项目下不存在权限包与权限项，因此 insert 必然被拒绝。

### Suggested Action
在项目初始化或角色分配流程中，显式为 admin 角色绑定至少一个包含 insert 的权限 bundle（如 preset READ_WRITE_ALL），并补一条健康检查：admin 角色存在但无 bundle 时给出告警。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/internal/app/project/project_service.go; modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql; modelcraft-backend/db/schema/mysql/15_admin_role.sql
- Tags: rbac, admin-role, bundle-binding, runtime-permission, docker-debug

---

## [LRN-20260513-003] correction

**Logged**: 2026-05-13T19:20:00+08:00
**Priority**: high
**Status**: promoted
**Area**: backend

### Summary
项目规则确认：受保护 `admin` 角色名本身等价于通配权限，Runtime 不应要求该角色额外绑定 bundle。

### Details
在定位 `Permission denied: insert` 后，用户明确纠正业务语义：`admin` 是内置角色名，语义上就是全权限，不需要通过 `end_user_role_bundles` 或 `end_user_user_bundles` 再做数据权限绑定。此前“admin 需绑定 bundle”仅是旧实现行为，不符合当前产品规则。

### Suggested Action
在 runtime 权限解析中加入 `admin` 角色短路：命中 `is_protected=true && name=admin` 时，直接返回该 model 的全动作允许策略（select/insert/update/delete 全开）。

### Metadata
- Source: user_feedback
- Related Files: modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go
- Tags: rbac, admin-wildcard, correction, runtime-graphql
- Promoted: ai-metadata/backend/common-mistakes.md (BM-20260513-0002)

---

## [LRN-20260513-004] best_practice

**Logged**: 2026-05-13T20:02:00+08:00
**Priority**: medium
**Status**: promoted
**Area**: infra

### Summary
本仓库 `just deploy` 默认仅 `compose up -d`，代码改动后若需生效必须使用 `just deploy force`（`--build`）。

### Details
复测 runtime 权限时，第一次在 8080 仍返回旧错误；日志显示仍执行旧的权限查询链。检查 `justfile` 发现 `deploy` 默认动作是 `up|start`，不会重建镜像。执行 `just deploy force` 后 backend 镜像重建并重启，随后同一请求成功写入。

### Suggested Action
涉及后端/网关代码改动后的本地验证，统一使用 `just deploy force`（或 `just deploy build` + `just deploy up`），避免误判“代码未生效”。

### Metadata
- Source: investigation
- Related Files: justfile; modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go
- Tags: deploy, docker, build-cache, verification
- Promoted: ai-metadata/backend/tools/justfile-guide.md

---

## [LRN-20260513-005] best_practice

**Logged**: 2026-05-13T23:30:00+08:00
**Priority**: medium
**Status**: promoted
**Area**: frontend

### Summary
仅靠放大列宽不能根治“提前截断”：还要处理 `w-full` 拉伸和单元格内部固定 `maxWidth` 的双重影响。

### Details
`ModelRecordTable` 初始问题是 `w-full + table-fixed` 在多列场景压窄列宽，导致 `truncate` 很早触发。后续改为 `min-w-max` 虽然避免了被容器压窄，但又引入新问题：当单元格是 `truncate`（`white-space: nowrap`）且内容为长 UUID 时，表格最小宽度会被内容的 max-content 拉大，表现为“最小列宽像是由数据长度决定”。最终可用组合是：表格使用 `!w-auto table-fixed`，并把 `minWidth` 改为“索引列 + 数据列配置宽度 + 操作列”的计算值（配置驱动，而非内容驱动）；同时保留单元格 `truncate` 与可调 `MIN_COLUMN_WIDTH`，实现“可截断且可继续缩小列宽”。

### Suggested Action
涉及长文本（UUID/Token/编码）表格时，避免直接用 `min-w-max` 作为唯一手段；优先用“按列配置计算 table minWidth”的方式，确保最小宽度由配置决定。若交互上要求操作列始终可见，统一用 `sticky right-0` 固定操作列（表头+数据单元格），并补 `bg` + `z-index` + 左边界分隔线。

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/web/components/shared/data-workspace/ModelRecordTable.tsx
- Tags: table-layout, truncate, readability, frontend
- Promoted: ai-metadata/front/development/known-issues.md

---

## [LRN-20260519-001] best_practice

**Logged**: 2026-05-19T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
MySQL 中参与唯一索引（含联合唯一）的字段必须使用 `NOT NULL`，否则会因 `NULL != NULL` 产生重复数据。

### Details
用户明确指出 MySQL 唯一索引陷阱：联合唯一键 `(A, B, C)` 中若 `C` 可空，会允许 `(1,2,NULL)` 重复插入而不冲突。这会让“业务上应唯一”的键在数据库层失效。已在数据库 skill 增加硬性规则，要求 `PRIMARY KEY/UNIQUE KEY` 的参与列全部 `NOT NULL`，并给出错误/正确示例与评审清单。

### Suggested Action
数据库 schema 评审时，将“唯一索引列是否全部 `NOT NULL`”列为必查项；软删除唯一避让统一使用 `delete_token`（`NOT NULL DEFAULT 0`），不依赖 `NULL`。

### Metadata
- Source: user_feedback
- Related Files: .agents/skills/db-develop/SKILL.md
- Tags: mysql, unique-index, null-semantics, schema-review

---

## [LRN-20260519-002] best_practice

**Logged**: 2026-05-19T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
数据库字段应尽量使用默认值，但默认值必须与业务有效值解耦，避免语义冲突。

### Details
用户新增约束：字段设计时优先给出默认值，降低写入遗漏造成的空值风险；同时明确禁止“默认值=真实业务值”的做法，因为会混淆“系统自动填充”与“业务有意赋值”的语义边界。该规则已写入 db-develop skill 的强制约束章节，并补充示例与评审清单。

### Suggested Action
Schema 评审流程增加两步：先确认字段业务值域，再校验默认值是否为保留值；若找不到安全默认值，允许 `NULL` 并在应用层显式处理。

### Metadata
- Source: user_feedback
- Related Files: .agents/skills/db-develop/SKILL.md
- Tags: mysql, default-value, sentinel, schema-review

---

## [LRN-20260519-003] best_practice

**Logged**: 2026-05-19T00:00:00+08:00
**Priority**: medium
**Status**: pending
**Area**: docs

### Summary
delete_token/deleted_at 软删除规则应与唯一索引、默认值规则一起集中维护在 db-develop skill。

### Details
用户要求把 `delete_token` 与 `deleted_at` 规则迁移到同一入口，避免规则分散在不同文档导致实施遗漏。已将核心字段定义、读/改/删路径约束、与唯一索引联动、黑名单例外迁移到 `.agents/skills/db-develop/SKILL.md`。

### Suggested Action
后续所有数据库约束更新优先在 db-develop skill 同步；涉及软删除时，统一从该入口做 schema 评审和实现检查。

### Metadata
- Source: user_feedback
- Related Files: .agents/skills/db-develop/SKILL.md; ai-metadata/backend/development/soft-delete-sqlc.md
- Tags: soft-delete, delete-token, deleted-at, documentation-centralization

---

## [LRN-20260522-001] best_practice

**Logged**: 2026-05-22T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
导航提案链路中，`AgentUiResponse` 需要在前端类型、工具 schema、Agent 提示词和 notebook 测试四处保持同一契约，否则会出现“类型定义存在但行为不一致”。

### Details
本次核对发现：前端 `types.ts` 已定义完整 `AgentUiResponse`（含 `kind/proposalId/proposalType/message/query/candidates`），但 notebook 测试工具仍用宽松 `response: object{}`；`print_nav_proposal` 只读取部分字段；Agent 提示词示例未体现 `query` 字段，前端工具参数里 `query` 也被标成非必填。结果是测试/提示词无法约束模型稳定输出完整结构，导致“设计有类型，运行靠约定”。

### Suggested Action
新增一条契约一致性检查清单：修改 `AgentUiResponse` 时同步更新 1) `types.ts`，2) `SharedCopilotActions` 参数 schema，3) `admin_agent.py` 提示词示例，4) `notebooks/nb_utils.py` NAV_TOOLS 与解析断言。

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/web/components/features/copilot/types.ts; modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx; modelcraft-agent/agents/admin_agent.py; modelcraft-agent/notebooks/nb_utils.py
- Tags: ai-navigator, contract-drift, copilotkit, notebook-test

---

## [LRN-20260522-002] insight

**Logged**: 2026-05-22T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
在“当前有哪些项目”导航场景下，agent 实际只调用 `list_projects` 并返回文本表格，未继续调用 `show_navigation_proposal`，与 notebook 用例期望不一致。

### Details
使用 `make_nav_payload('当前有哪些项目', layer='org')` 实测事件流中仅出现 `TOOL_CALL_START:list_projects`，`parse_nav_proposal(events)` 返回 `None`。这与 `test_agent_server.ipynb` 中 7-A 用例“每个项目一个 action_candidate”的期望存在偏差，说明当前提示词/工具选择策略在“列举+可导航”复合意图上仍可能退回纯文本输出。

### Suggested Action
为 7-A 场景增加自动断言：`show_navigation_proposal` 必须被调用且 `action_candidate` 数量 ≥ 项目数；若失败则标记为回归。并在 `admin_agent.py` 的 B 类意图规则中强化“列举后必须跟 proposal”的约束示例。

### Metadata
- Source: investigation
- Related Files: modelcraft-agent/notebooks/test_agent_server.ipynb; modelcraft-agent/notebooks/nb_utils.py; modelcraft-agent/agents/admin_agent.py
- Tags: ai-navigator, regression, proposal-missing, tool-selection

---

## [LRN-20260522-003] best_practice

**Logged**: 2026-05-22T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
AG-UI 导航/高亮能力应采用“单一 frontend tool + 人类确认后执行”的模式：仅暴露 `ui.presentProposal`，不向 Agent 直接暴露 `ui.navigate/ui.highlight/ui.guide`。

### Details
用户确认了最终架构：Agent 只负责理解意图并返回候选（`action_candidate` / `clarification_candidate`）；前端展示候选卡片，用户点击后由前端本地白名单执行 `guide/navigate/highlight`，或将澄清选择回传给 Agent 继续推理。该决策将“页面控制权”从 Agent 侧收敛到前端执行层，确保 human-in-the-loop。

### Suggested Action
后续实现统一围绕 `ui.presentProposal`：
1) 类型升级为 `PresentProposalArgs/ProposalCandidate/LocalUiAction`；
2) Agent 提示词明确“禁止直接导航/高亮”；
3) 前端执行器只接受注册过的 `routeCatalog` 与 `AiTargetRegistry`；
4) notebook/测试断言聚焦 `ui.presentProposal` 调用与候选结构。

### Metadata
- Source: user_feedback
- Related Files: docs/superpowers/plans/2026-05-21-ai-navigator-assistant.md; modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx; modelcraft-agent/agents/admin_agent.py
- Tags: ag-ui, human-in-the-loop, proposal-first, whitelist-actions

---

## [LRN-20260522-004] best_practice

**Logged**: 2026-05-22T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
当通过 OpenAI-compatible function calling 暴露前端工具时，工具名需要使用无点格式（如 `ui_present_proposal`），并在文案层映射到语义名 `ui.presentProposal`。

### Details
本次 AG-UI 迁移目标语义名是 `ui.presentProposal`，但实际工具调用通道对函数名有字符约束（通常仅允许字母/数字/下划线/连字符）。继续使用带点命名会带来工具注册或调用失败风险。最终采用“双命名策略”：协议语义保持 `ui.presentProposal`，实际函数名统一为 `ui_present_proposal`，并同步更新前端 action 注册、agent prompt、notebook 测试与解析器。

### Suggested Action
后续新增工具时，优先确定“语义名”和“wire-level 函数名”的映射，并在同一变更中同步更新：
1) 前端 `useCopilotAction` 名称
2) agent 提示词中的工具名
3) notebook 的 NAV_TOOLS 与事件解析断言
4) 测试断言中的 frontend action name

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx; modelcraft-agent/agents/admin_agent.py; modelcraft-agent/notebooks/nb_utils.py; modelcraft-agent/tests/agents/test_admin_agent.py
- Tags: tool-naming, openai-compatible, ag-ui, proposal-first

---

## [LRN-20260522-005] insight

**Logged**: 2026-05-22T00:00:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
即使已迁移到 `ui_present_proposal`，A/B/C 实跑中 agent 仍可能完全不触发前端 proposal 工具，导航意图会退化为纯文本回复。

### Details
本次通过 `nb_utils` 对 `test_agent_server.ipynb` 的 A/B/C 场景实跑（APISIX 路径）得到：
- A（当前有哪些项目）：仅调用 `list_projects`
- B（帮我去数据模型管理）：无任何工具调用
- C（帮我配置权限）：无任何工具调用
三组均未出现 `TOOL_CALL_START: ui_present_proposal`，`parse_nav_proposal` 均为空。这说明仅完成工具改名与提示词替换不足以保证 proposal-first 行为，模型仍会选择直接文本回答。

### Suggested Action
把 notebook A/B/C 变成回归门禁：
1) B/C 必须出现 `ui_present_proposal`
2) A 在列举后必须出现 `ui_present_proposal`（而非仅 `list_projects`）
3) 若未触发则判定回归失败，并迭代 `admin_agent.py` 的系统提示（增加强制规则 + 正反例）

### Metadata
- Source: investigation
- Related Files: modelcraft-agent/notebooks/test_agent_server.ipynb; modelcraft-agent/notebooks/nb_utils.py; modelcraft-agent/agents/admin_agent.py
- Tags: ai-navigator, proposal-missing, regression, tool-selection
- See Also: LRN-20260522-002

---

## [LRN-20260522-006] best_practice

**Logged**: 2026-05-22T15:01:00Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
proposal-first 强约束要在工具层做“可调用 + 必调用”双保险：仅提示词约束不够，需在 direct/list-nav 场景强制 `tool_choice=ui_present_proposal`。

### Details
A/B/C 回归中，虽然前端工具已注入且提示词声明“必须调用 ui_present_proposal”，模型仍会输出纯文本或仅调用 `list_projects`。修复采用两层策略：
1) `admin_agent.py` 在 direct-nav 与“list 后第二轮”场景启用强制模式，仅绑定 `ui_present_proposal` 并设置 `tool_choice`；
2) 保留 retry guard 兜底，首轮未调用 proposal 时再次强制调用。
同时发现 SSE 参数事件从 `TOOL_CALL_ARGS_DELTA` 演进为 `TOOL_CALL_ARGS`（且 `toolCallName` 可能为空），`nb_utils.parse_nav_proposal` 需按 `toolCallId` 关联解析，否则会误判“未返回 proposal 内容”。

### Suggested Action
将 A/B/C 场景固定为回归门禁：
- A 必须出现 `list_projects` 后再出现 `ui_present_proposal`
- B/C 必须直接出现 `ui_present_proposal`
- notebook 解析器持续兼容 `TOOL_CALL_ARGS_DELTA` 与 `TOOL_CALL_ARGS`

### Metadata
- Source: investigation
- Related Files: modelcraft-agent/agents/admin_agent.py; modelcraft-agent/notebooks/nb_utils.py; modelcraft-agent/notebooks/test_agent_server.ipynb
- Tags: proposal-first, tool-choice, sse-events, regression-guard
- See Also: LRN-20260522-005

---

## [LRN-20260524-001] best_practice

**Logged**: 2026-05-24T16:06:18Z
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
CLI GitHub Release 流程要直接使用完整 tag（`cli-vX.Y.Z`）注入 `main.version`，避免版本显示与发布标签不一致。

### Details
`modelcraft-cli/main.go` 的版本信息默认值是 `version=dev/commit=none/buildTime=unknown`，需要通过 workflow 的 `-ldflags` 覆盖。如果在 workflow 里把 tag 做前缀裁剪（例如去掉 `cli-`），会导致 `mc version` 输出和 release tag 不一致，文档命令与实际产物也更容易混淆。

### Suggested Action
在 `.github/workflows/release-cli.yml` 里统一用完整 `RELEASE_TAG`（并校验正则 `^cli-v\d+\.\d+\.\d+$`）驱动：
1) `main.version` 注入值
2) GitHub `tag_name`
3) Release 标题展示

### Metadata
- Source: investigation
- Related Files: .github/workflows/release-cli.yml; modelcraft-cli/main.go
- Tags: github-actions, cli-release, ldflags, versioning

---

## [LRN-20260524-002] best_practice

**Logged**: 2026-05-24T16:30:24Z
**Priority**: high
**Status**: pending
**Area**: infra

### Summary
APISIX 自定义 access_log_format 不能直接使用未注册变量（如 `$route_id`、`$opentelemetry_trace_id`），否则网关会启动失败并反复重启。

### Details
在 `apisix/config.yaml` 增加日志关联字段时，首次使用 `$route_id` 与 `$opentelemetry_trace_id`，容器出现 `nginx: [emerg] unknown "..." variable`，导致 `modelcraft-apisix` 进入 Restarting。改为稳定可用变量后恢复：`$uri`、`$http_x_request_id`、`$http_x_client_request_id`、`$http_traceparent`、`$http_tracestate`。

### Suggested Action
给 APISIX 增加日志字段时，先用一组“HTTP 头/基础 Nginx 变量”做最小可行版本，再重启验证；不要假设 APISIX 运行时变量在 Nginx 启动阶段一定可见。

### Metadata
- Source: error
- Related Files: apisix/config.yaml; deploy/compose/docker-compose.local.yml
- Tags: apisix, nginx-variable, access-log, restart-loop, observability

---

## [LRN-20260524-003] best_practice

**Logged**: 2026-05-24T16:40:57Z
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
BFF 代理路由若手动新建 `Headers()`，必须显式透传观测头（`X-Client-Request-Id`、`X-Request-Id`、`traceparent`、`tracestate`），否则 APISIX/Backend 链路关联会丢失。

### Details
`/api/bff/graphql/**` 路由此前只透传了 `Content-Type/Authorization/X-Action`，导致 APISIX access log 出现 `client_request_id:""`，并破坏前后端日志串联。修复是在 6 个 BFF GraphQL route 里统一增加观测头透传。

### Suggested Action
将“观测头透传”作为 BFF 代理模板的固定项，后续新增 route 时默认包含：
1) `X-Request-Id`
2) `X-Client-Request-Id`
3) `traceparent`
4) `tracestate`

### Metadata
- Source: investigation
- Related Files: modelcraft-front/src/app/api/bff/graphql/org/[orgName]/route.ts; modelcraft-front/src/app/api/bff/graphql/org/[orgName]/project/[projectSlug]/route.ts; modelcraft-front/src/app/api/bff/graphql/org/[orgName]/project/[projectSlug]/db/[db]/model/[model]/route.ts; modelcraft-front/src/app/api/bff/graphql/end-user/org/[orgName]/route.ts; modelcraft-front/src/app/api/bff/graphql/end-user/org/[orgName]/project/[projectSlug]/route.ts; modelcraft-front/src/app/api/bff/graphql/end-user/org/[orgName]/project/[projectSlug]/db/[db]/model/[model]/route.ts
- Tags: bff-proxy, request-id, traceparent, observability

---

## [LRN-20260524-004] best_practice

**Logged**: 2026-05-24T16:49:20Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
在 HTTP 接口层如果已通过 `ChiLoggerMiddleware` 注入 request-scoped logger，就不要在 handler struct 内持有独立 logger；应统一使用 `logfacade.GetLogger(ctx)` 记录日志以携带 request_id。

### Details
`auth/handler.go` 中 `h.logger.Error(...)` 生成的错误日志缺少 `request_id`，而同请求的 middleware 日志带有 `request_id/trace_id/span_id`。原因是 request-scoped logger 已放入 context（`logfacade.WithLogger`），但 handler 使用了结构体成员 logger，绕过了上下文字段。

### Suggested Action
接口层日志统一改为 `logfacade.GetLogger(r.Context()).<Level>(...)`；并移除 handler 中无用的 `logger` 字段与构造参数，避免后续再次误用。

### Metadata
- Source: user_feedback
- Related Files: modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go; modelcraft-backend/internal/interfaces/http/routes.go; modelcraft-backend/internal/middleware/chi_logger.go
- Tags: request-scoped-logger, request-id, auth-handler, interfaces-layer

---

## [LRN-20260524-005] best_practice

**Logged**: 2026-05-24T16:57:54Z
**Priority**: medium
**Status**: pending
**Area**: config

### Summary
该仓库启用写后 lint hook，文件写入后可能被自动改写，后续 Edit 前必须重新 Read 最新内容。

### Details
在调整 `workspace/cli/page.tsx` 时，首次 `Write` 后继续 `Edit` 出现 `File has been modified since read`。原因是仓库 hook 会在写入后自动执行 lint/格式化，并可能直接改动刚写入文件（包括文本内容与 class 排序）。如果继续基于旧快照执行替换，Edit 会失败或误改。

### Suggested Action
对本仓库执行“写入后再编辑”流程时固定采用：
1) `Write` 或首轮 `Edit` 后立刻 `Read` 目标文件刷新快照
2) 再做下一次 `Edit`
3) 最后跑目标文件 lint 验证

### Metadata
- Source: error
- Related Files: modelcraft-front/src/app/end-user/[orgName]/workspace/cli/page.tsx; .agents/hooks
- Tags: write-hook, auto-lint, stale-snapshot, edit-workflow

---

## [LRN-20260525-002] knowledge_gap

**Logged**: 2026-05-25T00:00:00Z
**Priority**: medium
**Status**: pending
**Area**: infra

### Summary
CLI release 当前仅产出 macOS arm64（Mach-O）二进制；Linux x86_64 环境下载后会触发 `Exec format error`。

### Details
用户在 Linux x86_64 机器执行 `./mc version` 报错。排查结果：`file ./mc` 显示 `Mach-O 64-bit arm64 executable`，与本机 `uname -s/-m => Linux x86_64` 不兼容。发布流程中 `release-cli.yml` 仅设置 `GOOS=darwin`、`GOARCH=arm64`，未生成 Linux 资产，且 `cli-v0.1.1` 下 `mc-linux-amd64` 返回 404。

### Suggested Action
1) 发布流水线增加 `linux/amd64`（必要时再加 `linux/arm64`）产物；
2) CLI 指南按平台给下载命令，或在 Linux 场景改为 `go build` 安装路径；
3) 保留“当前仅支持 macOS arm64”提示直到 Linux 产物上线。

### Metadata
- Source: conversation
- Related Files: .github/workflows/release-cli.yml; modelcraft-front/src/app/end-user/[orgName]/workspace/cli/page.tsx
- Tags: cli-release, binary-format, linux, macos-arm64, exec-format-error

---

## [LRN-20260526-001] best_practice

**Logged**: 2026-05-26T20:05:00+08:00
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
CopilotKit 前端工具触发的自动 rerun 应按“latest user message 之后是否已有 proposal”去重，并直接返回终态消息。

### Details
场景 E 中第 1 轮已调用 `ui_present_proposal`，但 CopilotKit/ag-ui 渲染前端工具后会用同一 thread 触发新一轮 run；此时 `_latest_user_text` 仍是原始导航文本。仅检查全历史或仅移除工具不够：orphan frontend tool call 会被 `sanitize_messages` 从发给 LLM 的上下文中移除，而系统提示仍强约束“导航必须调用 proposal”，可能再次产生工具调用或其它工具调用。更稳的做法是：只要 latest user message 之后已经出现 `ui_present_proposal`，该自动 rerun 直接返回无 tool_calls 的终态消息；若用户之后又发新消息，则允许重新 proposal。

### Suggested Action
导航 proposal 类工具都使用 `tool_call_since_latest_user` 模式做回环保护；回归测试应覆盖：1) proposal 后自动 rerun 不再调用任何工具，2) proposal 后有新 HumanMessage 时不被旧 proposal 抑制。Notebook 场景还必须让 `reset_graph()` 同时 `importlib.reload(agents.admin_agent)` 并刷新 `globals()["admin_graph"]`，否则 `from agents.admin_agent import admin_graph` 会继续持有旧 LazyGraph 实例，导致误判修复未生效。

### Metadata
- Source: investigation
- Related Files: modelcraft-agent/agents/admin_agent.py; modelcraft-agent/tests/agents/test_admin_agent.py; modelcraft-agent/notebooks/test_navigation_proposal.ipynb
- Tags: copilotkit, ag-ui, proposal-loop, latest-user-message, regression

---

## [LRN-20260529-001] best_practice

**Logged**: 2026-05-29T00:00:00Z
**Priority**: medium
**Status**: pending
**Area**: backend

### Summary
`gqlgen` 重新生成 resolver 后，文件尾部的 helper 可能被挪进 WARNING 注释块，若仍被当前 resolver 引用会直接触发 undefined 编译错误。

### Details
本次 `modelcraft-backend/internal/interfaces/graphql/project/database.resolvers.go` 编译失败，表面上是 `derefStringOrEmpty`、`gqlDatabaseModeToDomain`、`modelDatabaseToGQL`、`domainDatabaseModeFromGQLPtr` 未定义。根因不是 schema 或 service 逻辑缺失，而是 gqlgen 在更新 resolver 时把这些辅助函数移动到了文件尾部的 `// !!! WARNING !!!` 注释保留区，导致源码中实际没有可编译定义。

### Suggested Action
遇到 gqlgen 生成后的 resolver `undefined` 错误时，先检查目标文件末尾是否存在 WARNING 注释块；若缺失符号就在其中，优先把 helper 恢复为正常函数或迁移到独立 helpers 文件，再补齐所需 import，而不是先去改 schema 或 app/domain 层。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/internal/interfaces/graphql/project/database.resolvers.go
- Tags: gqlgen, resolver, compile-error, helper-functions

---

## [LRN-20260529-A1B] correction

**Logged**: 2026-05-29T00:00:00Z
**Priority**: high
**Status**: pending
**Area**: frontend

### Summary
管理员登录的规范入口已切换为 `/tenant/login`，`/login` 不能再作为兼容主入口继续使用。

### Details
用户明确纠正了登录路由语义：管理员登录应直接使用 `/tenant/login`，旧的 `/login` 已不兼容。仅修改首页按钮文案或局部链接不够，还需要把前端统一常量 `TENANT_LOGIN_PATH` 切到 `/tenant/login`，并把旧 `/login` 处理为跳转到新地址，避免认证中间件、登出跳转、Apollo 鉴权失败回跳等逻辑继续落到旧路由。

### Suggested Action
后续涉及管理员登录入口时，统一以 `TENANT_LOGIN_PATH = '/tenant/login'` 为单一真相源；若需要兼容历史链接，仅保留 `/login -> /tenant/login` 重定向，不再在 `/login` 维护独立登录页实现。

### Metadata
- Source: user_feedback
- Related Files: modelcraft-front/src/shared/constants/routes.ts; modelcraft-front/src/middleware.ts; modelcraft-front/src/app/login/page.tsx
- Tags: login-route, tenant-login, redirect, auth-entry

---

## [LRN-20260529-A1C] insight

**Logged**: 2026-05-29T00:00:00Z
**Priority**: high
**Status**: pending
**Area**: backend

### Summary
租户登录返回 `failed to issue access token` 时，先检查 membership 查询是否失败导致 `orgName` 为空，而不只是盯着 JWT 签发本身。

### Details
本次 `/api/tenant/auth/login` 的 requestId 日志显示，用户查询与 refresh token 落库都成功，真正异常来自 `ListMembershipsWithOrgDetails` 查询 `user_orgs` 表时报 `Table 'modelcraft.user_orgs' doesn't exist`。`TokenService.Login` 当前会吞掉 membership 查询错误，仅在成功时填充 `orgName`；随后 `JWTSigner.IssueAccessToken` 要求 `orgName` 非空，因此最终对外只暴露通用错误 `failed to issue access token`。这会把根因伪装成 JWT 问题，但实际是组织绑定表缺失/查询失败。

### Suggested Action
排查该错误时优先按 requestId 查看同请求前序 SQL 日志，重点看 `ListMembershipsWithOrgDetails` 是否成功；若失败，先修数据库 schema/数据源（尤其是 `user_orgs` 表）或改进日志/错误传播，不要先怀疑签名密钥。

### Metadata
- Source: investigation
- Related Files: modelcraft-backend/internal/app/auth/token_service.go; modelcraft-backend/internal/domain/auth/jwt_signer.go; modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go; modelcraft-backend/db/schema/mysql/06_users.sql
- Tags: tenant-login, jwt, membership, user_orgs, requestid-debug

---
