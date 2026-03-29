# BDD 验收测试方案设计

**日期**: 2026-03-29
**状态**: 草稿

---

## 1. 背景与目标

### 问题

- 现有 Python 集成测试（`modelcraft-backend/tests/`）与业务故事脱节，维护成本高
- 缺乏完整用户故事的端到端验收覆盖
- 边界值测试散落在多个文件，难以维护

### 目标

用 **Cucumber.js（TypeScript）+ Gherkin** 替代现有 Python 集成测试，建立完整的用户故事验收测试体系：

- 覆盖从登录到业务操作的完整链路
- 边界值通过 `Scenario Outline + Examples` 表达，不依赖 Python
- 场景文件可读性强，非技术人员可参与编写

### 范围

- ✅ API 级别（Project GraphQL + REST 登录），**不涉及浏览器**
- ✅ 完整链路：认证 → Project GraphQL 业务操作 → 结果验证
- ✅ 完全替代 Python 集成测试（包括边界值场景）
- ❌ 不测试 Org/Project 的 CRUD（org 和 project 均为预设 fixture）
- ❌ 不替代 Go 单元测试（Domain 层 95% 覆盖率保持不变）

### 固定 Fixture 约定

| Fixture | 说明 |
|---------|------|
| `TEST_ORG_NAME` | 测试环境预建的组织，测试不创建/删除 |
| `TEST_PROJECT_SLUG` | 测试环境预建的项目，测试不创建/删除 |

所有 Feature 文件的 `Background` 只做 **登录** 操作，org 和 project 直接从环境变量读取使用。

---

## 2. 技术选型

| 项目 | 选择 | 理由 |
|------|------|------|
| BDD 框架 | `@cucumber/cucumber` v11 | Gherkin 原生支持，内置 TypeScript 类型 |
| 语言 | TypeScript | 可复用 `modelcraft-front/contract/` 的 GraphQL schema 类型 |
| TS 执行器 | `tsx` | 零配置，比 `ts-node` 启动快 3-5 倍 |
| GraphQL 客户端 | `graphql-request` | 轻量，适合测试场景 |
| 断言库 | `expect`（来自 `@jest/globals`） | 熟悉的 API，错误信息清晰 |
| 报告 | `@cucumber/html-formatter` | 开箱即用的 HTML 测试报告 |

---

## 3. 目录结构

```
tests-bdd/                           # 根项目级别，独立于前后端
├── features/                        # Gherkin 场景文件（业务可读）
│   ├── model/
│   │   └── manage-model.feature
│   ├── field/
│   │   └── manage-field.feature
│   ├── enum/
│   │   └── manage-enum.feature
│   └── logical-foreign-key/
│       └── manage-lfk.feature
├── step-definitions/                # Step 实现（TypeScript）
│   ├── auth.steps.ts                # Given 我以管理员身份登录
│   ├── model.steps.ts
│   ├── field.steps.ts
│   ├── enum.steps.ts
│   ├── lfk.steps.ts
│   └── common.steps.ts              # 通用 Then（如断言错误码）
├── support/
│   ├── world.ts                     # 共享状态（每 Scenario 独立实例）
│   ├── hooks.ts                     # After 钩子（清理测试创建的数据）
│   ├── graphql-client.ts            # Project GraphQL 客户端
│   └── rest-client.ts               # REST 客户端（仅用于登录）
├── fixtures/
│   └── factory.ts                   # 测试数据工厂（生成唯一名称）
├── cucumber.js                      # Cucumber 配置
├── tsconfig.json
├── package.json
└── .env.test                        # 测试环境配置
```

---

## 4. 核心设计

### 4.1 World（共享状态）

每个 Scenario 开始时 Cucumber 创建一个独立的 `World` 实例，结束时销毁，天然隔离。

```typescript
// support/world.ts
export interface ModelCraftWorld extends World {
  // 客户端
  restClient: RestClient
  projectClient: GraphQLClient    // /graphql/org/{orgName}/project/{projectSlug}/

  // 认证状态（登录后填入）
  token: string | null

  // 从环境变量读取的固定 fixture（不由测试创建）
  orgName: string       // process.env.TEST_ORG_NAME
  projectSlug: string   // process.env.TEST_PROJECT_SLUG

  // 当前 Scenario 创建的资源（用于 After 钩子清理）
  createdModelNames: string[]

  // 最近操作结果（供 Then 断言）
  lastResponse: unknown
  lastError: unknown
}
```

**设计原则**：
- `orgName` / `projectSlug` 从环境变量读取，不由测试管理
- `createdModelNames` 等列表在 After 钩子中逐一清理
- `lastResponse` / `lastError` 解耦 `When` 和 `Then`，不需要在 Step 之间传参
- `projectClient` 在登录后调用 `setAuth(token)` 完成初始化

### 4.2 客户端设计

当前阶段只需两个客户端：

| 客户端 | 用途 | endpoint |
|--------|------|----------|
| `restClient` | 登录，获取 token | `POST /api/auth/login` |
| `projectClient` | 所有业务操作 | `/graphql/org/{orgName}/project/{projectSlug}/` |

#### 认证流程

```
POST /api/auth/login
Body: { username: string, password: string }
Response: { token: string }

所有后续请求均附带：
Authorization: Bearer <token>
```

#### GraphQL 错误位置

后端使用 **payload 级别的联合错误类型**，错误不在顶层 `errors` 数组，而在 mutation 返回的 payload 中：

```json
{
  "data": {
    "createModel": {
      "error": {
        "__typename": "ModelAlreadyExists",
        "errorCode": "CONFLICT.MODEL",
        "message": "..."
      },
      "model": null
    }
  }
}
```

`common.steps.ts` 中的 `Then 应该返回错误 "..."` 步骤从 `lastResponse` 的 payload `error.errorCode` 提取，**不检查顶层 `errors` 数组**。

```typescript
// support/graphql-client.ts
export class GraphQLClient {
  // URL 在构造时从环境变量组装：
  // ${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/
  constructor(orgName: string, projectSlug: string) {}

  setAuth(token: string) { /* 设置 Bearer token，登录后调用一次 */ }

  async query<T>(document: string, variables?: Record<string, unknown>): Promise<T>
  async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T>
}
```

### 4.3 Feature 文件示例

```gherkin
# features/model/manage-model.feature
Feature: 模型管理

  Background:
    Given 我以管理员身份登录
    # org 和 project 从环境变量读取，无需在此声明

  Scenario: 成功创建模型
    When 我创建名为 "User" 的模型
    Then 模型 "User" 应该存在于项目中

  Scenario: 创建重名模型时报错
    Given 项目中已存在名为 "User" 的模型
    When 我再次创建名为 "User" 的模型
    Then 应该返回错误 "CONFLICT.MODEL"

  Scenario Outline: 创建非法名称的模型时报错
    When 我创建名为 "<name>" 的模型
    Then 应该返回错误 "PARAM_INVALID.MODEL"

    Examples:
      | name            |
      |                 |
      | 123-invalid     |
      | name with space |
      | <script>        |
```

**边界值策略**：`Scenario Outline + Examples` 一张表覆盖所有非法输入，替代 Python 中重复的测试函数。

### 4.4 数据清理策略

org 和 project 是固定 fixture，**不清理**。每个 Scenario 只清理自己创建的资源（model、field、enum、lfk）。

```typescript
// support/hooks.ts
After({ tags: 'not @smoke' }, async function (this: ModelCraftWorld) {
  // 清理当前 Scenario 创建的 model（逆序删除，field/enum 随 model 级联删除）
  for (const modelName of this.createdModelNames.reverse()) {
    await this.projectClient.mutate(DELETE_MODEL_MUTATION, {
      modelName,
    })
  }
  this.createdModelNames = []
})
```

**原则**：
- 只通过 API 清理，不直接操作数据库
- 每个 Scenario 完全独立，可任意顺序执行
- `@smoke` 标记的场景保留数据，方便调试

```typescript
// fixtures/factory.ts
import { randomUUID } from 'crypto'
// 生成并发安全的唯一名称，避免多次运行冲突
export const uniqueName = (prefix: string) => `${prefix}-${randomUUID().slice(0, 8)}`
// 示例：uniqueName('User') → 'User-a3f2b1c0'
```

---

## 5. 与现有测试的关系

| 测试类型 | 现状 | BDD 引入后 |
|----------|------|------------|
| Go 单元测试（Domain 层） | 282 个，95% 覆盖率 | **保留**，职责不变 |
| Python 集成测试 | `modelcraft-backend/tests/` | **逐步废弃**，由 BDD 完全替代 |
| BDD 验收测试 | 无 | **新建**，覆盖 Project GraphQL 所有用户故事 + 边界值 |

**迁移策略**：先建 BDD，场景覆盖 Python 测试后，再删除对应 Python 测试文件。

---

## 6. 运行方式

> **注意**：`loader: ['tsx']` 是 Cucumber v11 的 loader 写法，`package.json` 不设置 `"type": "module"`（CJS 模式）。

```javascript
// cucumber.js
module.exports = {
  default: {
    loader: ['tsx'],
    require: [
      'step-definitions/**/*.ts',
      'support/**/*.ts',
    ],
    paths: ['features/**/*.feature'],
    format: [
      'progress',
      '@cucumber/html-formatter:reports/test-report.html',
    ],
  },
}
```

```json
// package.json scripts
{
  "scripts": {
    "test": "cucumber-js",
    "test:report": "cucumber-js --format @cucumber/html-formatter:reports/test-report.html",
    "test:smoke": "cucumber-js --tags @smoke"
  }
}
```

```bash
# 安装依赖
cd tests-bdd && npm install

# 运行所有场景（后端需已启动）
npm test

# 运行指定 feature
npm test -- features/model/manage-model.feature

# 只运行冒烟测试
npm run test:smoke

# 生成 HTML 报告
npm run test:report
```

---

## 7. 环境配置

```env
# .env.test
API_BASE_URL=http://localhost:8080
TEST_ADMIN_USERNAME=test-admin
TEST_ADMIN_PASSWORD=test-password
TEST_ORG_NAME=test-org          # 预建组织，测试不修改
TEST_PROJECT_SLUG=test-project  # 预建项目，测试不修改
```

---

## 8. 成功标准

- [ ] 所有 Project GraphQL 业务域（model、field、enum、logical_foreign_key）有 Feature 文件
- [ ] 边界值通过 `Scenario Outline` 覆盖，不需要额外测试代码
- [ ] 每个 Scenario 独立运行，无顺序依赖
- [ ] Python 集成测试可完全废弃
- [ ] CI 中独立运行，不阻塞 Go 单元测试（前提：后端服务已在 `API_BASE_URL` 就绪）

---

## 9. 不在范围内

- Org / Project CRUD（由固定 fixture 覆盖，不测试）
- Org GraphQL endpoint
- 浏览器 E2E 测试（Playwright/Cypress）
- 前端组件测试
- 性能测试
- Go 单元测试的替代
