---
name: doBackend
description: "后端四步流：Plan 模式 → backend-worker 并行开发 → review 前环境准备（db up + just run --force）→ backend-reviewer 串行收口（失败回流 backend-worker）"
argument-hint: "[plan file path or requirement description]"
---

执行后端开发四步流程：Plan 模式、backend-worker 并行开发、review 前环境准备、backend-reviewer 串行收口。

## 测试分工说明

本流程采用两层测试策略，职责严格分离：

| 测试类型 | 工具 | 覆盖范围 |
|----------|------|----------|
| **高价值逻辑测试** | `test-driven-development` | Schema 生成、SQL 生成、Adapter 转换（含 Fuzz 测试） |
| **流程验收测试** | `bdd-test` | 用户级业务流程（Use Case / 集成场景） |

> `test-driven-development` **不写 Use Case 测试**，Use Case 测试统一由 `bdd-test` 负责。

## 流程

### Step 0 — 进入 Plan 模式（必需）

- 在开始任何实现、生成、测试或调试动作前，必须先进入 Plan 模式
- 先在 Plan 中明确任务边界、影响文件、测试策略与验收标准
- 只有在 Plan 获批后，才能继续执行 Step 1~Step 4

### Step 1 — 输入校验

- 检查 `$ARGUMENTS` 是否提供了计划文件路径或需求描述
- 如果是文件路径，校验文件存在且可读；缺失则立即停止并提示
- 如果未提供任何输入，提示用户补充计划文件路径或需求说明

### Step 2 — 多 Agent 并行实现（必需）

- 开始实现前必须先拆分任务，并采用多 Agent 并行开发
- 必须在同一条消息中并行派发多个 `backend-worker` agent
- 每个 Agent 必须有明确 ownership（文件/目录边界），避免并发修改同一文件
- 推荐按分层拆分：migration / repository / app / resolver / adapter
- 并行开发后串行收口：使用 `backend-reviewer` 做统一代码审查、冲突处理、回归验证

根据任务是否涉及 `internal/domain/modelruntime/` 或 `internal/app/modelruntime/`，选择对应 skill：

#### 情况 A — 涉及 modelruntime 模块

- 使用 `modelruntime-dev` skill
- 该模块有严格架构约束，**必须**遵守：
  - `graphqlModelResolver` 是无状态 Schema 构建器，不得持有任何 `context.Context` 字段
  - Resolver 闭包内必须通过 `getGraphqlRequestContext(p.Context)` 获取请求级状态
  - 关联关系 resolver 必须返回 Thunk（`func() (interface{}, error)`），不得直接返回结果
  - Schema 构建方法通过参数透传 `ctx`，禁止从 `p` 取 ctx
- 涵盖：Schema 动态生成 / Resolver / Dataloader / graphqlRequestContext / RuntimeModel

#### 情况 B — 常规后端开发

- 优先使用 `backend-worker` agent 并行实现
- 如需技能辅助，可在 worker 内调用 `backend-develop` skill
- 按照 DDD 分层架构实现
- 涵盖：DB migration / Repository / App / Resolver / GraphQL Schema

- 产出：可运行的后端实现代码
- 人工确认后继续

### Step 3 — Review 前环境准备（必需）

- 在进入 `backend-reviewer` 之前，必须先调用 `db-develop` skill
- 按 `db-develop` 规范同步数据库：在 `modelcraft-backend/` 执行 `just db up`（如有需要可指定 env 文件）
- 数据库同步完成后，在 `modelcraft-backend/` 执行 `just run --force`，强制重启后端服务
- 建议执行健康检查（如 `curl -sf http://localhost:8080/health`），确认服务可用后再进入审查

推荐执行顺序（必须在 `modelcraft-backend/` 目录下）：

```bash
# 1) 先调 db-develop skill（理解当前 env 和迁移策略）
# 2) 同步数据库
just db up

# 3) 强制重启后端
just run --force

# 4) 健康检查
curl -sf http://localhost:8080/health
```

任一命令失败时：立即停止流程，修复后重新执行 Step 3，不得直接进入 `backend-reviewer`。

### Step 4 — `backend-reviewer` 串行收口（必需）

- 使用 `backend-reviewer` 对 Step 2 产出进行统一审查与验收
- 审查范围包含：分层约束、行为正确性、回归风险、测试结果
- 审核通过：流程完成
- 审核失败：**回到 Step 2**，由 `backend-worker` 按审查意见继续迭代修复

## 约束

- 开始工作前必须先进入 Plan 模式并获得批准
- 必须采用多 Agent 并行开发（使用 `backend-worker` 并行实现，`backend-reviewer` 串行收口）
- 必须按 DDD 分层架构实现，不得跨层调用
- **modelruntime 模块**：涉及 `internal/domain/modelruntime/` 或 `internal/app/modelruntime/` 时必须使用 `modelruntime-dev` skill，不得用 `backend-develop` 替代
- **modelruntime 架构红线**：`graphqlModelResolver` 禁止持有 `context.Context` 字段；关联关系 resolver 必须返回 Thunk；请求级状态只能通过 `getGraphqlRequestContext(p.Context)` 获取
- backend-worker 内部应完成 schema 同步、必要逻辑测试、调试与 bdd 验收
- domain 单测与 bdd 验收为必需；adapter/converter 需补充 Fuzz；repository 默认不新增测试
- **进入 backend-reviewer 前，必须先调用 `db-develop` skill 并完成 `just db up`**
- **进入 backend-reviewer 前，必须执行 `just run --force` 重启后端服务**
- 禁止跳过 backend-reviewer 串行收口直接宣布完成
- backend-reviewer 审核失败时，必须回流到 backend-worker 流程继续修复
- 如果缺少必要输入文件，立即停止并提示具体缺失路径
