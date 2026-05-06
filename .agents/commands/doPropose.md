---
name: doPropose
description: "从模糊想法到结构化方案：依次完成 idea-refine、pm overview、PRD 子页与 demo、backend-api 契约设计，再沉淀 spec"
argument-hint: "[idea or topic]"
---

将用户的模糊想法或主题逐步收敛为可执行的规范化方案，并在开工前补齐关键设计产物。

## 流程

### Step 1 — `idea-refine` 发散与精炼

- 输入：用户提供的想法或主题（`$ARGUMENTS`）
- 使用 `idea-refine` skill 对想法进行结构化发散和收敛思考
- 产出：精炼后的核心需求和方向
- 人工确认后继续

### Step 2 — `pm` 需求总览沉淀

- 输入：Step 1 精炼后的需求
- 先参考 `./.agents/agents/pm.md`
- 内容结构参考 `./ai-metadata/prd/dual-ui-architecture/00-overview.md`
- 使用 `pm` agent 从客户问题、用户价值、业务目标和功能范围角度整理需求
- 产出：需求总览文档 `overview.md`
- 人工确认后继续

### Step 3 — `prd-page-splitter` 功能页能力拆分

- 输入：Step 2 确认后的 `overview.md`
- 先参考 `./.agents/agents/prd-page-splitter.md`
- 使用 `prd-page-splitter` agent 将总览文档拆分为前后端都可消费的功能页子文档
- 如果涉及前端功能页，需要同步补充对应设计的 demo 页，用于表达大概的页面结构、关键交互和流程走向
- demo 页目的只是快速对齐思路与流程，不要求完整视觉设计，也不是生产实现
- 产出：
  - 功能页子文档
  - 对应前端功能页的 demo 页
- 人工确认后继续

### Step 4 — `backend-api` 契约文档设计

- 输入：Step 3 确认后的 `overview.md`、功能页子文档和 demo 页
- 先参考 `./.agents/agents/backend-api.md`
- 使用 `backend-api` agent 补齐足够的契约文档，包括 GraphQL / OpenAPI、错误模型、领域接口和应用层 command/result 结构
- 产出：可供前后端并行实现的契约文档集合
- 人工确认后继续

### Step 5 — `spec-driven-development` 规范化

- 输入：Step 4 确认后的需求总览、功能页子文档、demo 页和契约文档
- 使用 `spec-driven-development` skill 将上述产物沉淀为正式规范文档（spec）
- 产出：可供开发使用的 spec 文档

## 约束

- 必须先完成 Step 1 并经用户确认，再进入 Step 2
- 必须先完成 Step 2 并确认 `overview.md`，再进入 Step 3
- 必须先完成 Step 3 并确认功能页子文档与 demo 页，再进入 Step 4
- 必须先完成 Step 4 并确认契约文档足够完整，再进入 Step 5
- 在契约文档补齐前，不得进入“开始干活”的实施阶段
- 如果用户未提供任何输入（`$ARGUMENTS` 为空），先提示用户描述想法，不继续执行
