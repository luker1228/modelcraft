---
name: workflow
description: "全功能开发工作流 — 从 1-2 句话需求出发，先引导生成 PRD/领域模型/PRD子页，再串联后端协议设计、前后端开发、验收与联调"
argument-hint: "<1-2句话需求描述>"
---

启动 ModelCraft **全功能开发工作流**。你只需要输入 1-2 句话描述需求，工作流会先引导你补齐三个必要资源，再进入完整研发流程。

---

## 工作流步骤

### Step 0 — 引导使用 `/pm` 结构化需求（串行）

输入是用户的 1-2 句话需求描述。

执行要求：
- 引导用户运行 `/pm`，将原始需求整理为 PRD 文档
- 明确产出 PRD 文件路径（记为 `<prd-file>`）
- 展示 PRD 摘要并请求人工审核

审核提示：

```
PRD 已生成。请先审核 PRD 内容，确认后继续？
[继续] [修改PRD]
```

使用 **AskUserQuestion tool** 获取确认。若用户选择「修改PRD」，继续停留在本步骤直到确认。

---

### Step 1 — 引导使用 `/domain-modeler` 生成领域模型（串行）

前置条件：Step 0 的 PRD 已确认。

执行要求：
- 引导用户运行 `/domain-modeler`，基于已确认 PRD 生成 PlantUML 领域模型
- 明确产出 `.puml` 文件路径（记为 `<domain-model-file>`）
- 展示模型摘要并请求人工审核

审核提示：

```
领域模型已生成。请先审核 .puml 内容，确认后继续？
[继续] [修改模型]
```

使用 **AskUserQuestion tool** 获取确认。若用户选择「修改模型」，继续停留在本步骤直到确认。

---

### Step 2 — 引导使用 `prd-page-splitter` 生成 PRD 子页（串行）

前置条件：Step 0/1 已确认。

执行要求：
- 引导用户调用 `prd-page-splitter` agent，按功能拆分 PRD 子页
- 明确产出子页目录或文件路径（记为 `<prd-subpages-dir-or-file>`）
- 展示子页列表并请求人工审核

审核提示：

```
PRD 子页已生成。请先审核拆分结果，确认后继续？
[继续] [修改拆分]
```

使用 **AskUserQuestion tool** 获取确认。若用户选择「修改拆分」，继续停留在本步骤直到确认。

> 至此，三个必要资源齐备：`<prd-file>`、`<domain-model-file>`、`<prd-subpages-dir-or-file>`。

---

### Step 3 — 验证输入资源（串行）

检查三个必要资源是否存在且可读：
- PRD 文件 `<prd-file>`
- 领域模型 `.puml` 文件 `<domain-model-file>`
- PRD 子页 `<prd-subpages-dir-or-file>`

如有缺失，**停止执行**并告知用户补充。

读取三个文件内容，建立上下文摘要，供后续 agent 使用。

从 `<prd-file>` 推导 `<prd-module-dir>`（即 PRD 文件所在目录），并定义以下计划文件（统一写入 `plans/` 目录，文件名由工作流自动推导）：
- `<backend-plan-file>`
- `<front-architecture-file>`

除 Step 5 计划文档外，其他工作流产物默认写入 `<prd-module-dir>`。

---

### Step 4 — 后端协议设计（串行）

使用 **Agent tool**，`subagent_type: backend-api`，设计 GraphQL Schema 和/或 OpenAPI 接口：

**Prompt 要点：**
- 传入 PRD 摘要 + 领域模型内容 + PRD 子页内容
- 产出：GraphQL `.graphql` Schema 文件变更建议 或 OpenAPI YAML 片段
- 要求：遵循 Error Interface + Union 模式，明确归属 Org GraphQL 还是 Project GraphQL
- 产出物写入 `<prd-module-dir>/api-contract.md`

等待完成后，**展示协议设计结果**，询问用户：

```
后端协议设计完成。请审核 `<prd-module-dir>/api-contract.md`，确认后继续？
[继续] [修改协议]
```

使用 **AskUserQuestion tool** 获取确认。

---

### Step 5 — 并行启动：前端架构 + 后端方案落地（并行）

用户确认协议后，前端可立即开始；同时推进后端具体方案落地：

#### 5A — 前端架构设计

`subagent_type: front-architect`，`run_in_background: true`

**Prompt 要点：**
- 传入 PRD 子页内容 + 已确认的 API Contract
- 产出：模块目录结构、BFF mock 接口定义、TypeScript 类型骨架、组件分层规划
- 产出物写入 `<front-architecture-file>`（位于 `plans/` 目录）
- 前端使用 **BFF mock**，不依赖真实后端

#### 5B — 后端具体方案落地（不写代码）

`subagent_type: general-purpose`，`run_in_background: true`

**Prompt 要点：**
- 传入 PRD 子页内容 + 领域模型 + 已确认的 API Contract
- 产出：`<backend-plan-file>`（位于 `plans/` 目录，文件名自动推导）
- 重点覆盖：数据库设计（核心表/字段/关系、索引、约束、迁移策略）
- 重点覆盖：核心流程思路、事务边界、分层落点（Repository/App/Resolver）
- 补充：实现顺序、验收口径、主要风险与回滚思路
- **禁止提交具体代码实现**，只输出可执行技术方案

> **CRITICAL**: 5A 和 5B 在**同一条消息**中用多个 Agent tool call 并行启动。

> **CRITICAL**: 从 Step 6 开始，后续开发任务默认以 Step 5 产出的两份计划文件为主：
> - `<front-architecture-file>`（前端主依据）
> - `<backend-plan-file>`（后端主依据）
> PRD、领域模型、API Contract 作为一致性校验和边界约束的辅助资料。

等待 5B 完成后，展示并要求用户审核 `<backend-plan-file>`：

```
后端落地方案已生成。请先审核 `<backend-plan-file>`，确认后再进入后端实现？
[继续] [修改方案]
```

使用 **AskUserQuestion tool** 获取确认。若用户选择「修改方案」，继续停留在 Step 5B；前端线（5A/Step 6）可继续推进。

---

### Step 6 — 前端 Worker 并行开发（等待 5A 完成后启动）

5A（前端架构）完成后，根据架构产出，对每个前端模块派发独立的 `front-worker` agent，`run_in_background: true`

**Prompt 要点（每个 worker）：**
- 以 `<front-architecture-file>` 中对应模块的规划为主依据
- 传入 BFF mock 接口定义
- PRD/API Contract 仅用于校验功能边界与接口一致性
- **使用 BFF mock 数据**，不调用真实后端
- 完成后标记 TaskUpdate: completed

---

### Step 7 — 后端开发与部署（等待 5B 方案确认后串行）

前置条件：`<backend-plan-file>` 已经用户确认。

先派发 `backend-worker` agent 按方案实现（可按子页并行）：

**Prompt 要点（每个 worker）：**
- 必须基于 `<backend-plan-file>` 实现，不得偏离数据库设计与事务边界
- 传入对应子页需求 + 领域模型 + 已确认 API Contract + `<backend-plan-file>`
- 以 `<backend-plan-file>` 为主，PRD 子页/领域模型/API Contract 作为辅助校验依据
- 任务：实现 DB migration → Repository → App Service → Resolver
- 完成后标记 TaskUpdate: completed

所有 `backend-worker` 完成后，使用 **general-purpose agent** 执行：

```bash
just build && just run
```

检查服务是否正常启动（health check），如有构建或启动错误，**重新派发 backend-worker** 修复对应模块。

---

### Step 8 — 后端验收测试（串行，循环直到通过）

使用 `backend-reviewer` agent 执行验收测试：

**Prompt 要点：**
- 运行 BDD 测试 + Integration 测试
- 输出失败的测试用例列表和错误信息
- 写入 `<prd-module-dir>/review-result.md`

**如果测试失败** → 根据失败报告，重新派发对应的 `backend-worker` 修复 → 重新部署（Step 7）→ 重新验收（Step 8）

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

### Step 9 — BFF 接入真实后端，前后端联调（串行）

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

### Step 10 — 收尾

展示完整工作流摘要：

```
## 工作流完成

### 产出物
- `<prd-module-dir>/api-contract.md`       — 后端协议定义
- `<backend-plan-file>`                    — 后端落地方案（含数据库设计与思路）
- `<front-architecture-file>`              — 前端架构规划
- `<prd-module-dir>/review-result.md`      — 后端验收报告

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

✅ Step 0  /pm 结构化需求
✅ Step 1  /domain-modeler 生成领域模型
✅ Step 2  prd-page-splitter 生成子页
✅ Step 3  输入验证
✅ Step 4  后端协议设计
⟳  Step 5A 前端架构设计（进行中）
✅ Step 5B 后端方案落地（已确认）
○  Step 6  前端开发（等待架构完成）
○  Step 7  后端开发与部署
○  Step 8  后端验收
○  Step 9  联调
```

---

## Guardrails

- **Step 0/1/2 必须完成并经人工确认**，才可进入后续开发流程
- **三个输入资源必须全部存在**，缺一不可，否则立即停止
- **Step 4 必须等用户确认协议**才能进入 Step 5
- **Step 6 必须等 5A 完成**才能启动（worker 需要架构蓝图）
- **Step 7 必须等 5B 方案经人工确认**，确认前不得启动后端代码实现
- **Step 8 失败必须修复再重测**，不能跳过
- **Step 9 必须等用户手动确认 UI**，不自动进入联调
- **每个 agent prompt 必须自包含**，包含所有必要上下文（路径、内容摘要、约束）
- **联调只修改 BFF 层**，不修改前端组件
- **Step 5 计划产物（前端架构 + 后端方案）固定写入 `plans/` 目录**
- **除 Step 5 计划产物外，其他工作流产物必须写入 `<prd-module-dir>`**
- **Step 6 起后续开发默认以 `<front-architecture-file>` 与 `<backend-plan-file>` 为主依据，其他资料为辅助校验**
