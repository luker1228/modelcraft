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
