---
name: workflow
description: "全功能开发工作流 — 从 PRD + 领域模型 + PRD子页 出发，串联后端协议设计、前后端并行开发、后端验收、前端联调的完整流程"
argument-hint: "<prd-file> <domain-model-file> <prd-subpages-dir-or-file>"
---

启动 ModelCraft **全功能开发工作流**。输入三个必要资源，自动编排后续所有步骤。

## 前置条件（人工完成，不自动化）

以下步骤需用户在执行本命令前**手动完成并审核**：

| 步骤 | 命令 | 说明 |
|------|------|------|
| ① 生成 PRD | `/pm` | 将需求整理为 PRD 文档 |
| ② 生成领域模型 | `/domain-modeler` | 从 PRD 提取 PlantUML 领域模型 |
| ③ 生成 PRD 子页 | `prd-page-splitter` agent | 将 PRD 拆分为子功能页 |

**三个资源缺一不可**，请在参数中提供：

```
/workflow <prd文件路径> <领域模型.puml文件路径> <prd子页目录或文件>
```

---

## 工作流步骤

### Step 0 — 验证输入

检查三个必要资源是否存在且可读：
- PRD 文件
- 领域模型 `.puml` 文件
- PRD 子页（目录或文件）

如有缺失，**停止执行**并告知用户补充。

读取三个文件内容，建立上下文摘要，供后续 agent 使用。

---

### Step 1 — 后端协议设计（串行）

使用 **Agent tool**，`subagent_type: backend-api`，设计 GraphQL Schema 和/或 OpenAPI 接口：

**Prompt 要点：**
- 传入 PRD 摘要 + 领域模型内容 + PRD 子页内容
- 产出：GraphQL `.graphql` Schema 文件变更建议 或 OpenAPI YAML 片段
- 要求：遵循 Error Interface + Union 模式，明确归属 Org GraphQL 还是 Project GraphQL
- 产出物写入 `ai-metadata/workflow-output/` 目录下的 `api-contract.md`

等待完成后，**展示协议设计结果**，询问用户：

```
后端协议设计完成。请审核 api-contract.md，确认后继续？
[继续] [修改协议]
```

使用 **AskUserQuestion tool** 获取确认。

---

### Step 2 — 并行启动：前端架构 + 后端开发（并行）

用户确认协议后，**同时**启动两路并行：

#### 2A — 前端架构设计

`subagent_type: front-architect`，`run_in_background: true`

**Prompt 要点：**
- 传入 PRD 子页内容 + 已确认的 API Contract
- 产出：模块目录结构、BFF mock 接口定义、TypeScript 类型骨架、组件分层规划
- 产出物写入 `ai-metadata/workflow-output/front-architecture.md`
- 前端使用 **BFF mock**，不依赖真实后端

#### 2B — 后端并行开发

对每个 PRD 子页，派发独立的 `backend-worker` agent，`run_in_background: true`

**Prompt 要点（每个 worker）：**
- 传入该子页需求 + 领域模型 + 已确认的 API Contract
- 任务：实现 DB migration → Repository → App Service → Resolver
- 完成后标记 TaskUpdate: completed

> **CRITICAL**: 2A 和所有 2B 在**同一条消息**中用多个 Agent tool call 并行启动。

---

### Step 3 — 前端 Worker 并行开发（等待 2A 完成后启动）

2A（前端架构）完成后，根据架构产出，对每个前端模块派发独立的 `front-worker` agent，`run_in_background: true`

**Prompt 要点（每个 worker）：**
- 传入 front-architecture.md 中对应模块的规划
- 传入 BFF mock 接口定义
- **使用 BFF mock 数据**，不调用真实后端
- 完成后标记 TaskUpdate: completed

---

### Step 4 — 后端部署与调试（等待所有 2B 完成后串行）

所有 `backend-worker` 完成后，使用 **general-purpose agent** 执行：

```bash
just build && just run
```

检查服务是否正常启动（health check），如有构建或启动错误，**重新派发 backend-worker** 修复对应模块。

---

### Step 5 — 后端验收测试（串行，循环直到通过）

使用 `backend-reviewer` agent 执行验收测试：

**Prompt 要点：**
- 运行 BDD 测试 + Integration 测试
- 输出失败的测试用例列表和错误信息
- 写入 `ai-metadata/workflow-output/review-result.md`

**如果测试失败** → 根据失败报告，重新派发对应的 `backend-worker` 修复 → 重新部署（Step 4）→ 重新验收（Step 5）

**循环**，直到所有测试通过。通过后展示：

```
✅ 后端验收通过！所有测试绿灯。

现在请人工测试前端 UI：
- 前端使用 BFF mock 数据运行中
- 请访问前端页面，检查 UI 是否符合 PRD 预期
- 确认 UI 无误后，输入"联调"继续
```

使用 **AskUserQuestion tool** 等待用户手动测试 UI 并确认。

---

### Step 6 — BFF 接入真实后端，前后端联调（串行）

用户确认 UI 后，派发 `front-worker` agent 将 BFF 从 mock 切换到真实后端接口：

**Prompt 要点：**
- 将 BFF layer 中的 mock 实现替换为真实 GraphQL / REST 调用
- 参考已确认的 API Contract
- 保持前端组件层不变，只修改 BFF 层
- 完成后运行前端 lint 检查

完成后展示：

```
✅ 联调完成！BFF 已接入真实后端。

建议人工验证端到端流程：
1. 启动后端：just run
2. 启动前端：npm run dev（在 modelcraft-front/）
3. 访问前端，验证完整功能
```

---

### Step 7 — 收尾

展示完整工作流摘要：

```
## 工作流完成

### 产出物
- api-contract.md      — 后端协议定义
- front-architecture.md — 前端架构规划
- review-result.md      — 后端验收报告

### 建议后续步骤
1. git commit 后端代码（modelcraft-backend/）
2. git subtree push API Contract
3. git subtree pull 前端同步 Contract
4. git commit 前端代码（modelcraft-front/）
5. 根项目 git add submodule 引用并提交
```

---

## 进度跟踪

在整个流程中，用 TaskCreate/TaskUpdate 跟踪任务状态，并在自然节点展示进度：

```
## 工作流进度

✅ Step 0  输入验证
✅ Step 1  后端协议设计
⟳  Step 2A 前端架构设计（进行中）
✅ Step 2B 后端开发（3/3 worker 完成）
○  Step 3  前端开发（等待架构完成）
○  Step 4  后端部署
○  Step 5  后端验收
○  Step 6  联调
```

---

## Guardrails

- **三个输入必须全部存在**，缺一不可，否则立即停止
- **Step 1 必须等用户确认协议**才能进入 Step 2
- **Step 3 必须等 2A 完成**才能启动（worker 需要架构蓝图）
- **Step 5 失败必须修复再重测**，不能跳过
- **Step 6 必须等用户手动确认 UI**，不自动进入联调
- **每个 agent prompt 必须自包含**，包含所有必要上下文（路径、内容摘要、约束）
- **联调只修改 BFF 层**，不修改前端组件
