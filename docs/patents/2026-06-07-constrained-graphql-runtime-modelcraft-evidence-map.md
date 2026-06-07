# Constrained GraphQL Runtime ModelCraft Evidence Map

**Date:** 2026-06-07

| Claimed Concept | ModelCraft Evidence | Why It Matters |
|---|---|---|
| Design-time / runtime separation | `ai-metadata/backend/design/core-principles.md "### 两个核心阶段"` and `ai-metadata/backend/design/core-principles.md "### 5. 设计态与运行态解耦，运行态可独立部署"` | Direct project evidence that design-time and runtime are separated and can evolve independently. |
| Runtime GraphQL as the access carrier | `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "## 3. 整体架构"` and `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "## 7. Execute() 权限注入"` | Direct project evidence that end-user runtime requests travel through GraphQL execution and context injection rather than a separate access layer. |
| Caller-/subject-specific boundary generation | `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "## 2. V1 边界"` and `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "## 7. Execute() 权限注入"` | Direct project evidence for per-request, end-user-scoped permission resolution; the boundary is shaped at request time from `endUserID`, `orgName`, `projectSlug`, and `model.ID`. |
| Subject-specific permission shaping and row pruning | `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "### 5.1 domain/modelruntime 新增"` and `docs/superpowers/specs/2026-05-06-enduser-runtime-data-permission-design.md "## 9. Row Filter（IsSelf WHERE 注入）"` | Direct project evidence that permissions are represented as a request snapshot with `Allowed` / `IsSelf`, and that `SELF` scope injects row filters or owner writes. |
| Schema / capability discovery before execution | `ai-metadata/cli/README.md "## 核心接口一：\`describe\` — 获取 GraphQL Schema"` and `ai-metadata/cli/README.md "## 资源发现（辅助）"` and `ai-metadata/cli/README.md "## CLI 自省"` | Direct project evidence that callers can discover schema, catalog resources, and command capabilities before running queries. |
| Structured feedback / error packaging | `ai-metadata/cli/README.md "## 输出格式"` and `ai-metadata/cli/README.md "### 退出码"` | Direct project evidence that failures are wrapped in machine-readable JSON with `code`, `message`, `retryable`, and `suggestion`; this supports correction workflows, but the query-correction use is indirect. |

## Review Notes

1. This file is not the patent itself.
2. This file exists to help later reviewers ground the application in documented project behavior.
3. If later code anchors are needed, add them after targeted source inspection rather than guessing.
