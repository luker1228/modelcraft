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
