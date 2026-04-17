---
name: workflow
description: "统一工作流命令，支持 spec / front / backend / all 四个子命令"
argument-hint: "<spec|front|backend|all> [需求或计划文件路径]"
---

启动四段式研发工作流。必须先识别第一个参数作为子命令：

- `spec`
- `front`
- `backend`
- `all`

如果未提供子命令，先提示用户补充，不继续执行。

---

## 1) `workflow spec`

`spec` 负责原流程前 5 步，最终只产出计划文档，不进入实现。

### Step 0 — `/pm` 结构化需求
- 输入：1-2 句话需求
- 输出：`<prd-file>`
- 人工确认后继续

### Step 1 — `/domain-modeler` 生成领域模型
- 基于 `<prd-file>`
- 输出：`<domain-model-file>` (`.puml`)
- 人工确认后继续

### Step 2 — `prd-page-splitter` 生成 PRD 子页
- 输出：`<prd-subpages-dir-or-file>`
- 人工确认后继续

### Step 3 — 资源校验与上下文构建
- 校验 `<prd-file>` / `<domain-model-file>` / `<prd-subpages-dir-or-file>` 都存在且可读
- 缺失立即停止并提示补齐

### Step 4 — 后端协议设计
- 使用 `backend-api`（或等价能力）设计 GraphQL/OpenAPI 合约
- 输出：`<plan-module-dir>/api-contract.md`
- 人工确认后继续

### Step 5 — 生成前后端计划文档（只写方案，不写代码）
- 在 `plans/` 下按模块创建子目录：`<plan-module-dir>`（例如 `plans/end-user-auth/`）
- 前后端设计文档统一写入 `<plan-module-dir>`
- 同时补齐以下必备文件（均写入 `<plan-module-dir>`）：
  - 协议文件：`api-contract.md`
  - 数据库 schema 终态文件：`db-schema-final.md`（或等价命名）
- 输出：
  - `<plan-module-dir>/<backend-plan-file>`
  - `<plan-module-dir>/<front-architecture-file>`
  - `<plan-module-dir>/api-contract.md`
  - `<plan-module-dir>/db-schema-final.md`

> `spec` 结束条件：`<plan-module-dir>` 下前后端文档生成完成并通过人工确认。

---

## 2) `workflow front`

`front` 只推进前端线，后端接口调用先走 mock。

## 输入
- 默认前端方案文件固定为：`plans/end-user-auth/end-user-auth-front-architecture.md`（相对仓库根目录）
- 若用户显式传入其他 `<front-architecture-file>`，可覆盖默认值

## 流程
1. `front-architect`
   - 基于 `<front-architecture-file>` 完善前端框架和模块边界
   - 明确 BFF mock 契约和类型
2. `front-reviewer`
   - 审计架构与实现计划一致性、可维护性、边界清晰度
   - 输出问题清单和修正建议
3. `front-worker`
   - 根据审计后方案实现前端
   - **强制使用 mock，不调用真实后端**

## 任务拆分与并行派发（强制）
- 前端实现前必须先拆任务，再派发多个 `front-worker` 并行处理
- 拆分原则：
  - 按模块/页面拆分（如 auth、apikey、profile）
  - 写入集合互斥（避免多个 worker 同时改同一文件）
  - 公共类型与共享组件单独一个 worker 负责
- 并行派发要求：
  - 同一波次在**一条消息**中发起多个 subagent 调用
  - 每个 worker prompt 必须包含明确 ownership（文件/目录边界）
- 前端建议切片：
  1. `front-worker-a`：页面容器与路由
  2. `front-worker-b`：BFF mock 与接口适配
  3. `front-worker-c`：共享组件与类型收敛

## 约束
- front 阶段禁止切换到真实后端
- 仅允许改前端与 BFF mock 层

---

## 3) `workflow backend`

`backend` 只推进后端线，采用「并行开发 + 串行收口」。

## 输入
- 必须有 `<backend-plan-file>`（建议来自 `workflow spec`，位于某个 `<plan-module-dir>`）

## 流程
1. 并行开发阶段
   - `backend-worker`：按 `<backend-plan-file>` 实现后端（DB migration / Repository / App / Resolver）
   - `bdd-test`：并行开发/补充相关 BDD 用例
2. 串行收口阶段
   - `backend-reviewer`：统一审计行为正确性、分层约束、回归风险
   - `bdd-test`：最终执行验收并输出报告

## 并行规则（强约束）
- 并行阶段（Wave 1）：
  - 多个 `backend-worker` 按子域并行实现（如 apikey、role、auth）
  - `bdd-test` 开发/补充可与 backend 并行
- 收口阶段（串行）：
  - `backend-reviewer` 统一审计
  - `bdd-test` 最终执行与报告

## 后端任务拆分与并行派发（强制）
- 后端实现前必须先拆任务，再派发多个 `backend-worker` 并行处理
- 拆分原则：
  - 按 PRD 子页或业务子域拆分
  - 每个 worker 绑定明确写入边界（migration / repository / app / resolver）
  - 避免两个 worker 修改同一核心文件
- 并行派发要求：
  - 同一波次在**一条消息**中发起多个 subagent 调用
  - 每个 worker prompt 必须包含：输入计划文件、目标子域、文件 ownership、验收标准

---

## 4) `workflow all`

`all` 只负责联调收口，不重复做 spec/front/backend 的前置工作。

## 前置条件
- 前端 mock 版本已可运行
- 后端实现与验收已通过

## 流程
1. 将前端/BFF 从 mock 切换到真实后端接口
2. 执行联调验证（接口连通、关键路径、错误处理）
3. 输出联调结论与问题清单

## 交接规则（强约束）
- `all` 阶段完成“mock -> real”切换后，**流程交给用户（你）**做最终确认与发布决策
- Agent 不擅自继续下一阶段

---

## 通用 Guardrails

- 未完成 `spec` 产物时，不得直接进入 `all`
- `front` 与 `backend` 可并行，但各自必须遵守本阶段边界
- 每个子命令都要输出：
  - 输入文件路径
  - 产出文件路径
  - 当前状态（进行中/完成/阻塞）
- 若缺少必要输入文件，立即停止并提示具体缺失路径
- `plans` 目录必须按模块使用子目录组织；禁止将前后端设计文档直接平铺在 `plans/` 根目录
- 协议文件与数据库 schema 终态文件必须放在 `plans/<module>/` 下，不再放到 `ai-metadata/`
