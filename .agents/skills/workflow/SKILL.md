---
name: workflow
description: >
  统一工作流命令，支持 spec / front / backend / all 四个子命令，驱动四段式研发流程。
  当用户输入 `/workflow`、"启动工作流"、"workflow spec"、"workflow front"、"workflow backend"、
  "workflow all"、"开始 spec 流程"、"推进前端线"、"推进后端线"、"联调收口" 时必须立即触发此 skill。
  也适用于：用户说"帮我做需求分析到实现"、"从 PRD 到代码"、"走完整开发流程"、"前后端并行开发"
  等描述完整研发链路的场景。遇到上述情形时主动触发，即便用户没有明确说 "workflow"。
argument-hint: "<spec|front|backend|all> [需求或计划文件路径]"
---

# Workflow — 四段式研发工作流

启动前先识别第一个参数作为子命令：`spec` / `front` / `backend` / `all`。

如果未提供子命令，提示用户补充后停止，不继续执行。

---

## 1) `workflow spec`

负责需求→设计→计划前 5 步，**只产出计划文档，不进入实现**。

### Step 0 — `/pm` 结构化需求
- 输入：1-2 句话需求描述
- 调用 `/pm` skill 结构化需求
- 输出：`<prd-file>`
- **人工确认后继续**

### Step 1 — `/domain-modeler` 生成领域模型
- 基于 `<prd-file>` 调用 `domain-modeler` skill
- 输出：`<domain-model-file>`（`.puml` 类图）
- **人工确认后继续**

### Step 2 — `prd-page-splitter` 生成 PRD 子页
- 将 PRD 按业务子域拆分为子页文件
- 输出：`<prd-subpages-dir-or-file>`
- **人工确认后继续**

### Step 3 — 资源校验与上下文构建
- 校验以下三项均存在且可读：
  - `<prd-file>`
  - `<domain-model-file>`
  - `<prd-subpages-dir-or-file>`
- 任何一项缺失：**立即停止并提示补齐具体路径**

### Step 4 — 后端协议设计
- 使用 `backend-api` agent（或等价能力）设计 GraphQL / OpenAPI 合约
- 输出：`<plan-module-dir>/api-contract.md`
- **人工确认后继续**

### Step 5 — 生成前后端计划文档（只写方案，不写代码）
- 在 `plans/` 下按模块创建子目录：`<plan-module-dir>`（例：`plans/end-user-auth/`）
- **禁止将文档平铺在 `plans/` 根目录**
- 必备文件（均写入 `<plan-module-dir>`）：

  | 文件 | 说明 |
  |------|------|
  | `<backend-plan-file>` | 后端实现方案 |
  | `<front-architecture-file>` | 前端架构方案 |
  | `api-contract.md` | GraphQL/OpenAPI 协议 |
  | `db-schema-final.md` | 数据库 schema 终态 |

> **spec 结束条件**：`<plan-module-dir>` 下四份文件生成完成并通过人工确认。

---

## 2) `workflow front`

**只推进前端线**，后端接口调用先走 mock。

### 输入
- 默认前端方案文件：`plans/end-user-auth/end-user-auth-front-architecture.md`（相对仓库根）
- 用户显式传入其他路径时覆盖默认值

### 流程

1. **`front-architect`** — 基于 `<front-architecture-file>` 完善前端框架和模块边界，明确 BFF mock 契约和类型
2. **`front-reviewer`** — 审计架构与实现计划一致性、可维护性、边界清晰度；输出问题清单和修正建议
3. **`front-worker`（并行）** — 根据审计后方案实现前端，**强制使用 mock，不调用真实后端**

### 任务拆分与并行派发（强制）

实现前必须先拆任务，再在**一条消息**中派发多个 `front-worker` 并行处理：

| 拆分原则 | 说明 |
|----------|------|
| 按模块/页面拆分 | 例：auth、apikey、profile |
| 写入互斥 | 避免多个 worker 同时改同一文件 |
| 公共资源隔离 | 共享类型与组件单独一个 worker |

**建议切片**：
- `front-worker-a`：页面容器与路由
- `front-worker-b`：BFF mock 与接口适配
- `front-worker-c`：共享组件与类型收敛

每个 worker prompt 必须包含明确 **ownership**（文件/目录边界）。

### 约束
- front 阶段**禁止切换到真实后端**
- 仅允许改前端与 BFF mock 层

---

## 3) `workflow backend`

**只推进后端线**，采用「并行开发 + 串行收口」模式。

### 输入
- 必须有 `<backend-plan-file>`（建议来自 `workflow spec`，位于某个 `<plan-module-dir>`）

### 流程

**Wave 1（并行）**
- `backend-worker`：按 `<backend-plan-file>` 实现后端（DB migration → Repository → App → Resolver）
- `test-driven-development`：编写并执行 **domain 单元测试**；并对 **adapter / converter 转换点**补充单测（该类测试必须包含 **Fuzz**）

**Wave 2（串行收口）**
- `backend-reviewer`：统一审计行为正确性、分层约束、回归风险
- `bdd-test`：执行流程验收并输出报告

### 后端测试策略（强约束）

- **必须有 domain 层单元测试**
- **必须执行 bdd-test 验收**
- **Adapter / Converter 需要增加 Fuzz 单测**
- **Repository 层不要求新增测试**（除非用户明确提出）

### 任务拆分与并行派发（强制）

实现前必须先拆任务，再在**一条消息**中派发多个 `backend-worker` 并行处理：

| 拆分原则 | 说明 |
|----------|------|
| 按 PRD 子页或业务子域拆分 | 例：apikey、role、auth |
| 每个 worker 绑定明确写入边界 | migration / repository / app / resolver |
| 避免冲突 | 两个 worker 不修改同一核心文件 |

每个 worker prompt 必须包含：输入计划文件、目标子域、文件 ownership、验收标准。

---

## 4) `workflow all`

**只负责联调收口**，不重复做 spec/front/backend 的前置工作。

### 前置条件
- 前端 mock 版本已可运行
- 后端实现与验收已通过

### 流程
1. 将前端/BFF 从 mock 切换到真实后端接口
2. 执行联调验证（接口连通、关键路径、错误处理）
3. 输出联调结论与问题清单

### 交接规则（强约束）
- `all` 阶段完成「mock → real」切换后，**流程交给用户**做最终确认与发布决策
- Agent 不擅自继续下一阶段

---

## 通用 Guardrails

- 未完成 `spec` 产物时，**不得直接进入 `all`**
- `front` 与 `backend` 可并行，但各自必须遵守本阶段边界
- backend 测试基线：domain 单测 + bdd-test 必须执行；adapter/converter 需补 Fuzz；repository 默认不新增测试
- 每个子命令都要输出：
  - 输入文件路径
  - 产出文件路径
  - 当前状态（进行中 / 完成 / 阻塞）
- 若缺少必要输入文件，**立即停止并提示具体缺失路径**
- `plans/` 目录必须按模块使用子目录组织；禁止将前后端设计文档直接平铺在根目录
- 协议文件与数据库 schema 终态文件必须放在 `plans/<module>/` 下，不再放到 `ai-metadata/`
