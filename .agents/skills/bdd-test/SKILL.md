---
name: bdd-test
description: >
  运行 ModelCraft 后端 BDD 验收测试（Cucumber.js + TypeScript + Gherkin）。
  触发此 skill 必须同时满足两个条件：
  (1) 明确是后端上下文（如：后端、API、GraphQL、resolver、服务端、modelcraft-backend），
  (2) 明确是 BDD 测试上下文（如：BDD、Cucumber、Gherkin、feature、验收测试、smoke）。
  只有“后端 + BDD”同时出现时才触发。
  例如："跑后端 BDD 测试"、"执行 backend cucumber"、"后端 auth feature 跑一下"。
  不应触发示例："测试 auth 流程"（未明确后端/BDD）、"跑前端测试"、"跑 e2e"、"跑单元测试"。
  BDD 测试目录位于 tests-bdd/，使用 Cucumber.js v11，测试场景用中文 Gherkin 编写。
  当前测试认证约定是：优先复用 `TEST_ACCESS_TOKEN`，没有时才由 `tests-bdd/support/hooks.ts` + `tests-bdd/support/jwt.ts` 做测试启动兜底。
  注意：生产链路使用 ES256 JWT；`tests-bdd/support/jwt.ts` 里的 HMAC-SHA256 只用于 BDD 启动兜底，不代表正式认证架构。
---

# BDD 测试 Skill

运行 ModelCraft 的 Cucumber.js BDD 验收测试，覆盖 Auth、Model、Field、Enum、Logical Foreign Key 五个领域。

## 领域映射表

用户输入 → npm script 的映射。**必须用此表解析用户意图**：

| 用户可能说的 | 领域 | 命令 |
|-------------|------|------|
| `auth` / `登录` / `注册` / `token` / `认证` | Auth | `npm run test:auth` |
| `model` / `模型` | Model | `npm run test:model` |
| `field` / `字段` | Field | `npm run test:field` |
| `enum` / `枚举` | Enum | `npm run test:enum` |
| `lfk` / `外键` / `logical-foreign-key` / `逻辑外键` | LFK | `npm run test:lfk` |
| `smoke` / `冒烟` | Smoke | `npm run test:smoke` |
| `all` / `全部` / `全量` / 未指定领域 | All | `npm test` |
| `报告` / `report` | Report | `npm run test:report` |

如果用户只说了领域名（如 "enum test"、"跑 model"），直接映射到对应命令执行，不需要确认。

## 执行流程

### Step 1 — 解析领域

从用户消息中提取领域关键词，对照上方映射表确定命令。如果无法确定，跑全量测试。

### Step 2 — 前置检查（静默执行，有问题才报）

```bash
# 并行检查三项
ls ./tests-bdd/node_modules/.package-lock.json 2>/dev/null || (cd ./tests-bdd && npm install)
curl -sf http://localhost:8080/health > /dev/null 2>&1 || echo "FAIL:backend"
test -f ./tests-bdd/.env.test || echo "FAIL:envtest"
```

- `node_modules` 不存在 → 自动 `npm install`
- 后端不可达 → **停止执行**，告诉用户：`后端未运行，请先执行 just run`
- `.env.test` 不存在 → **停止执行**，告诉用户需要创建

如果全部正常，不输出任何检查信息，直接进 Step 3。

### Step 3 — 运行测试

```bash
cd ./tests-bdd && <对应 npm 命令> 2>&1
```

### Step 4 — 分析结果并汇报

#### Step 4.1 — 日志优先定位（新增硬性规则）

在判断失败归因前，**必须先用日志定位**，再看代码：

1. 从失败输出中提取 `requestId`（GraphQL 通常在 `extensions.requestId`，REST 通常在响应体 `requestId`）。
2. 在后端目录按 requestId 查整条链路：

```bash
cd ./modelcraft-backend && just log-cat <requestId>
```

3. 日志中确认“第一现场错误”（最早出现的业务/SQL错误），再决定分类 `[ENV]/[BACKEND]/[TEST]/[DATA]`。
4. 只有日志无法判断时，才进入代码阅读。

**汇报格式**（所有场景都使用此格式）：

```
## <领域> BDD 测试结果

通过 X / 共 Y 个场景

[如果全部通过，到此结束]

### 失败场景（Z 个）

| # | 场景 | 失败步骤 | 错误分类 | 错误摘要 |
|---|------|---------|---------|---------|
| 1 | <场景名> | <Given/When/Then ...> | <分类> | <一句话摘要> |

### 失败详情

**1. <场景名>**（<feature 文件路径>）
- 失败步骤：`<step 文本>`
- 错误分类：<分类标签>
- 错误信息：`<完整错误消息>`
- 判断依据：<为什么归入此分类>
```

## 错误分类规则（核心）

**每个失败必须归入以下分类之一**。分类决定了后续行动。

### 分类 1：环境问题 `[ENV]`

连接失败、token 过期、权限不足、服务不可用。

标志性错误：
- `FetchError` / `ECONNREFUSED` / `fetch failed`
- `401 Unauthorized` / `403 Forbidden`
- `Permission denied: requires '...' permission`
- `SYSTEM_ERROR` + DB 相关（`Unknown column`、`Table doesn't exist`）

**行动：只报告，不修复。** 建议用户检查后端/数据库/权限配置。

### 分类 2：后端行为变更 `[BACKEND]`

后端 API 的响应结构、错误码、校验规则发生变化，导致测试断言不匹配。

标志性错误：
- `expect(received).toBe(expected)` — 期望的错误码/类型名与实际不同
- `expect(received).not.toBeNull()` + `Received: null` — 期望有错误但后端没返回
- `Cannot query field "..." on type "..."` — GraphQL schema 变更
- `GRAPHQL_VALIDATION_FAILED` — 测试用的 GraphQL 查询与当前 schema 不兼容

**行动：只报告，不修复。** 说明：后端 schema 或业务逻辑已变更，测试需要同步更新。列出期望值 vs 实际值。

### 分类 3：测试代码 Bug `[TEST]`

step definition 逻辑错误、selector 写错、fixture 数据问题。

标志性错误：
- `TypeError` / `ReferenceError` 在 step-definitions/ 文件中
- `Cannot read property '...' of undefined/null` — 步骤间状态传递断裂
- `前置条件：创建 ... 失败` — Given 步骤本身失败（非环境问题）
- step 匹配不到（`Undefined step`）

**行动：只报告，不修复。** 指出具体的 step 文件和行号。

### 分类 4：测试数据冲突 `[DATA]`

唯一约束冲突、资源已存在、ID 重复。

标志性错误：
- `Duplicate entry`
- `already exists` / `已存在`（出现在 Given 前置步骤中）

**行动：只报告。** 建议用户清理残留测试数据或重新运行。

## 权限边界（最高优先级）

### 绝对禁止

- **永远不要修改后端代码** — 任何 `modelcraft-backend/` 下的文件都不可触碰，无论什么理由
- **永远不要在汇报测试结果后自动修改测试代码** — 默认行为是**只做报告**

### 测试代码修改的前置条件

测试代码（`tests-bdd/` 下的 feature 文件、step definition、support、fixtures）只有**同时满足以下两个条件**时才能修改：

1. 用户**明确要求**修复（如"帮我修"、"修复测试"、"更新测试"）
2. 失败分类为 `[TEST]`（测试代码自身 Bug）

**不满足条件时的行为：**

| 情况 | 行为 |
|------|------|
| 用户没要求修复 | 只报告，不修改任何代码 |
| 用户要求修复，但分类是 `[ENV]` | 只给建议（检查环境），不修改代码 |
| 用户要求修复，但分类是 `[BACKEND]` | 只给建议（后端需同步），不修改代码 |
| 用户要求修复，但分类是 `[DATA]` | 只给建议（清理数据），不修改代码 |
| 用户要求修复，且分类是 `[TEST]` | **可以修改**测试代码 |

## 测试领域详情

| 领域 | Feature 文件 | npm script | 场景数 |
|------|-------------|------------|-------|
| Auth | `features/auth/{login,register,token}.feature` | `test:auth` | ~21 |
| Model | `features/model/manage-model.feature` | `test:model` | ~9 |
| Field | `features/field/manage-field.feature` | `test:field` | ~6 |
| Enum | `features/enum/manage-enum.feature` | `test:enum` | ~6 |
| LFK | `features/logical-foreign-key/manage-lfk.feature` | `test:lfk` | ~2 |

## 目录结构

```
tests-bdd/
├── features/           # Gherkin 场景（中文）
├── step-definitions/   # TypeScript 步骤实现
├── support/            # World / 客户端 / Hooks / JWT
├── fixtures/           # uniqueName() 工厂
├── reports/            # 生成的 HTML 报告
├── cucumber.js         # Cucumber 配置
├── .env.test           # 环境变量（gitignored）
└── package.json        # 依赖与 npm scripts
```

## 前置条件

1. `cd ./tests-bdd && npm install`（`node_modules` 不存在时自动执行）
2. 后端运行中（`http://localhost:8080`）
3. `./tests-bdd/.env.test` 存在，至少包含 `TEST_ORG_NAME` 和 `TEST_PROJECT_SLUG`
   - `TEST_ACCESS_TOKEN` 推荐提供（优先级最高）
   - 若未提供 token，当前 BDD 会先用 `TEST_LOGIN_PHONE` + `TEST_LOGIN_PASSWORD` 登录；必要时由 `tests-bdd/support/jwt.ts` 生成测试用 token
   - 默认不应把“注册新用户”当作所有 BDD 场景的通用前置；只有测试注册链路本身时才依赖注册流程

## 架构要点

- **World 对象**（`support/world.ts`）：Scenario 级隔离，存储 token/response/tracking
- **JWT 测试兜底**（`support/jwt.ts`）：这是 BDD 专用辅助文件，只负责在测试启动时生成最小可用 token；生产链路使用 ES256，这里的 HMAC-SHA256 只是测试兜底，不代表正式认证架构
- **双客户端**：GraphQL（项目领域）+ REST（auth 接口）
- **自动清理**：After hook 删 model/enum；`@smoke` 跳过清理
- **唯一命名**：`uniqueName("User")` → `"Usera3f2b1c0"` 防冲突

## 注册耦合约束（新增）

- 默认不要把“注册用户”作为 BDD 通用前置（尤其是 model/field/enum/lfk/smoke）。
- 只有在明确测试注册能力时（如 auth/register）才应走注册链路。
- 若非注册场景出现 `CONFLICT.USER`，优先归类为前置设计问题或测试数据冲突，而不是业务能力失败。
- 注册流程已保证 init-org 时，不应把 `org/init` 场景作为常规回归阻断项。
