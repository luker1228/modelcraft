# Constrained GraphQL Runtime ModelCraft Evidence Map

**Date:** 2026-06-07

| Claimed Concept | ModelCraft Evidence | Why It Matters |
|---|---|---|
| Design-time / runtime separation | `ai-metadata/backend/design/core-principles.md "### 两个核心阶段"` and `ai-metadata/backend/design/core-principles.md "### 5. 设计态与运行态解耦，运行态可独立部署"` | Direct project evidence that design-time and runtime are separated and can evolve independently. |
| Runtime GraphQL as the access protocol | `ai-metadata/backend/design/core-principles.md "### 1. 运行态 API 只提供 GraphQL，不提供 REST"` and `ai-metadata/backend/design/core-principles.md "### 2. 双 GraphQL 入口：设计态静态 Schema + 运行态动态 Schema"` | Direct project evidence that runtime access is carried through GraphQL rather than REST. |
| Developer / EndUser access split | `ai-metadata/backend/development/developer-enduser-system.md "## 1. 总览"` and `ai-metadata/backend/development/developer-enduser-system.md "### 3. 路由隔离与视图切换"` | Direct project evidence for a role-separated access model and separate user-facing routes. |
| AI-usable structured access path | `ai-metadata/cli/README.md "## 核心设计"` and `ai-metadata/cli/README.md "## 资源发现（辅助）"` | Direct project evidence that the CLI exposes schema discovery and structured query execution for agents. |
| Dynamic runtime schema generation per model or data object | `ai-metadata/backend/design/core-principles.md "### 2. 双 GraphQL 入口：设计态静态 Schema + 运行态动态 Schema"` | Direct project evidence that runtime schema generation is dynamic rather than fixed at design time. |
| Agent-side capability discovery before query execution | `ai-metadata/cli/README.md "## 核心接口一：describe — 获取 GraphQL Schema"` and `ai-metadata/cli/README.md "## 核心接口二：run — 执行 GraphQL 查询"` | Direct project evidence that agents inspect schema before running GraphQL queries. |

## Review Notes

1. This file is not the patent itself.
2. This file exists to help later reviewers ground the application in documented project behavior.
3. If later code anchors are needed, add them after targeted source inspection rather than guessing.
