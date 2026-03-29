# BDD 验收测试 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `tests-bdd/` 目录建立 Cucumber.js + TypeScript BDD 验收测试，覆盖 Project GraphQL 的 model、field、enum、logical_foreign_key 业务域。

**Architecture:** 独立 Node.js 项目（`tests-bdd/`），用 `@cucumber/cucumber` v11 + TypeScript 编写 Gherkin 场景。每个 Scenario 独立 World 实例，登录 → 操作 → 断言 → After 钩子清理。org 和 project 均为预设 fixture，从 `.env.test` 读取。

**Tech Stack:** `@cucumber/cucumber` v11, TypeScript 5, `tsx`, `graphql-request`, `@jest/globals` (expect), `dotenv`

---

## 关键 API 细节（实现前必读）

### 认证
```
POST /api/auth/token
Body: { "code": "<oauth-code>" }
Response: { "accessToken": "...", "tokenType": "Bearer", ... }
```
> ⚠️ 不是 username/password，是 OAuth code 流程。测试账号的 code 从环境变量 `TEST_AUTH_CODE` 读取，或使用已配置的测试用户直接获取 token（见 `modelcraft-backend/tests/conftest.py` 中的 test user setup）。实际上测试中通过 `just test-user-setup` 创建测试用户，用法参考 `tests/conftest.py`。

### GraphQL Endpoint
```
Project GraphQL: POST ${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/
Header: Authorization: Bearer <accessToken>
```

### GraphQL 错误判别
错误在 payload 的 `error` 字段，通过 `__typename` 区分类型（**无** `errorCode` 字段）：
```json
{
  "data": {
    "createModel": {
      "model": null,
      "error": {
        "__typename": "ModelAlreadyExists",
        "message": "Model 'User' already exists"
      }
    }
  }
}
```

`common.steps.ts` 中 `Then 应该返回错误类型 "ModelAlreadyExists"` 检查 `__typename`。

### 主要 Mutation 名称
| 域 | 创建 | 删除 |
|---|---|---|
| Model | `createModel` | `deleteModel` |
| Field | `addFields` | `removeField` |
| Enum | `createEnum` | `deleteEnum` |
| LFK | `createLogicalForeignKey` | `deleteLogicalForeignKey` |

### LFK 特殊结构
LFK 错误在 `result` 字段（不是 `error`），是联合类型：
```json
{ "data": { "createLogicalForeignKey": { "result": { "__typename": "FKColumnsNotFoundError" } } } }
```

---

## 文件结构

```
tests-bdd/
├── features/
│   ├── model/manage-model.feature
│   ├── field/manage-field.feature
│   ├── enum/manage-enum.feature
│   └── logical-foreign-key/manage-lfk.feature
├── step-definitions/
│   ├── auth.steps.ts
│   ├── model.steps.ts
│   ├── field.steps.ts
│   ├── enum.steps.ts
│   ├── lfk.steps.ts
│   └── common.steps.ts
├── support/
│   ├── world.ts
│   ├── hooks.ts
│   ├── graphql-client.ts
│   └── rest-client.ts
├── fixtures/
│   └── factory.ts
├── cucumber.js
├── tsconfig.json
├── package.json
└── .env.test
```

---

## Chunk 1: 项目脚手架

### Task 1: 初始化 `tests-bdd/` 项目

**Files:**
- Create: `tests-bdd/package.json`
- Create: `tests-bdd/tsconfig.json`
- Create: `tests-bdd/cucumber.js`
- Create: `tests-bdd/.env.test`
- Create: `tests-bdd/.gitignore`

- [ ] **Step 1: 创建目录结构**

```bash
cd /home/luke/modelcraft_project
mkdir -p tests-bdd/{features/{model,field,enum,logical-foreign-key},step-definitions,support,fixtures,reports}
```

- [ ] **Step 2: 创建 `package.json`**

```json
// tests-bdd/package.json
{
  "name": "modelcraft-bdd-tests",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "test": "cucumber-js",
    "test:model": "cucumber-js features/model",
    "test:field": "cucumber-js features/field",
    "test:enum": "cucumber-js features/enum",
    "test:lfk": "cucumber-js features/logical-foreign-key",
    "test:smoke": "cucumber-js --tags @smoke",
    "test:report": "cucumber-js --format @cucumber/html-formatter:reports/test-report.html"
  },
  "dependencies": {
    "@cucumber/cucumber": "^11.0.0",
    "@cucumber/html-formatter": "^21.0.0",
    "@jest/globals": "^29.0.0",
    "dotenv": "^16.0.0",
    "graphql-request": "^6.0.0",
    "tsx": "^4.0.0",
    "typescript": "^5.0.0"
  }
}
```

- [ ] **Step 3: 创建 `tsconfig.json`**

```json
// tests-bdd/tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "CommonJS",
    "moduleResolution": "node",
    "strict": true,
    "esModuleInterop": true,
    "resolveJsonModule": true,
    "outDir": "dist",
    "rootDir": "."
  },
  "include": ["**/*.ts"],
  "exclude": ["node_modules", "dist"]
}
```

- [ ] **Step 4: 创建 `cucumber.js`**

```javascript
// tests-bdd/cucumber.js
// 使用 CJS 模式：requireModule + require（不混用 ESM loader）
module.exports = {
  default: {
    requireModule: ['tsx/cjs'],
    require: [
      'support/**/*.ts',
      'step-definitions/**/*.ts',
    ],
    paths: ['features/**/*.feature'],
    format: [
      'progress-bar',
      '@cucumber/html-formatter:reports/test-report.html',
    ],
    publishQuiet: true,
  },
}
```

- [ ] **Step 5: 创建 `.env.test`**

```env
# tests-bdd/.env.test
API_BASE_URL=http://localhost:8080
TEST_ACCESS_TOKEN=<填入测试用户的 accessToken>
TEST_ORG_NAME=test-org
TEST_PROJECT_SLUG=test-project
```

> 💡 `TEST_ACCESS_TOKEN`：通过 `just test-user-setup`（在 `modelcraft-backend/` 中运行）创建测试用户后，运行一次登录流程获取 token 填入。

- [ ] **Step 6: 创建 `.gitignore`**

```
# tests-bdd/.gitignore
node_modules/
dist/
reports/
.env.test
```

- [ ] **Step 7: 安装依赖**

```bash
cd tests-bdd && npm install
```

Expected: `node_modules/` 出现，无错误。

- [ ] **Step 8: 验证 Cucumber 可执行**

```bash
cd tests-bdd && npx cucumber-js --version
```

Expected: 打印版本号（如 `11.x.x`）。

- [ ] **Step 9: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/package.json tests-bdd/tsconfig.json tests-bdd/cucumber.js tests-bdd/.gitignore
git commit -m "chore(bdd): scaffold tests-bdd project structure"
```

---

### Task 2: 实现 `support/` 层

**Files:**
- Create: `tests-bdd/fixtures/factory.ts`
- Create: `tests-bdd/support/rest-client.ts`
- Create: `tests-bdd/support/graphql-client.ts`
- Create: `tests-bdd/support/world.ts`
- Create: `tests-bdd/support/hooks.ts`

- [ ] **Step 1: 创建 `fixtures/factory.ts`**

```typescript
// tests-bdd/fixtures/factory.ts
import { randomUUID } from 'crypto'

/**
 * 生成并发安全的唯一名称，避免多次运行之间冲突。
 * 不使用连字符分隔，因为 model/enum 名称不允许包含连字符。
 * @example uniqueName('User') → 'Usera3f2b1c0'
 */
export const uniqueName = (prefix: string): string =>
  `${prefix}${randomUUID().replace(/-/g, '').slice(0, 8)}`
```

- [ ] **Step 2: 创建 `support/rest-client.ts`**

```typescript
// tests-bdd/support/rest-client.ts
import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export interface TokenResponse {
  accessToken: string
  tokenType: string
}

export class RestClient {
  /**
   * 用 OAuth code 换取 accessToken。
   * 在测试中，通过环境变量 TEST_ACCESS_TOKEN 直接提供 token，
   * 本方法留作 OAuth 完整流程备用。
   */
  async getTokenByCode(code: string): Promise<TokenResponse> {
    const res = await fetch(`${API_BASE_URL}/api/auth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
    if (!res.ok) {
      throw new Error(`Auth failed: ${res.status} ${await res.text()}`)
    }
    return res.json() as Promise<TokenResponse>
  }
}
```

- [ ] **Step 3: 创建 `support/graphql-client.ts`**

```typescript
// tests-bdd/support/graphql-client.ts
import 'dotenv/config'
import { GraphQLClient as GqlClient } from 'graphql-request'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export class GraphQLClient {
  private client: GqlClient

  constructor(orgName: string, projectSlug: string) {
    const url = `${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/`
    this.client = new GqlClient(url)
  }

  setAuth(token: string): void {
    this.client.setHeader('Authorization', `Bearer ${token}`)
  }

  async query<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    return this.client.request<T>(document, variables)
  }

  async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
    return this.client.request<T>(document, variables)
  }
}
```

- [ ] **Step 4: 创建 `support/world.ts`**

```typescript
// tests-bdd/support/world.ts
import 'dotenv/config'
import { World, setWorldConstructor, IWorldOptions } from '@cucumber/cucumber'
import { GraphQLClient } from './graphql-client'
import { RestClient } from './rest-client'

export class ModelCraftWorld extends World {
  // 客户端
  readonly restClient: RestClient
  readonly projectClient: GraphQLClient

  // 认证
  token: string | null = null

  // 固定 fixture（从环境变量读取）
  readonly orgName: string
  readonly projectSlug: string

  // 当前 Scenario 创建的资源（After 钩子清理用）
  createdModelIds: string[] = []
  createdEnumNames: string[] = []

  // 当前操作的 model ID（跨 step 传递，替代模块级变量）
  currentModelId: string | null = null

  // 模型 baseName → 实际 ID 映射（供 lfk.steps.ts 使用）
  modelMap: Record<string, string> = {}

  // 最近一次 Given 创建的实际名称（供 "再次创建" step 使用）
  lastModelName: string | null = null
  lastEnumName: string | null = null

  // 最近操作结果（When → Then 传递）
  lastResponse: Record<string, unknown> | null = null
  lastError: Error | null = null

  constructor(options: IWorldOptions) {
    super(options)

    this.orgName = process.env.TEST_ORG_NAME ?? 'test-org'
    this.projectSlug = process.env.TEST_PROJECT_SLUG ?? 'test-project'

    this.restClient = new RestClient()
    this.projectClient = new GraphQLClient(this.orgName, this.projectSlug)

    // 若提供了预设 token，直接使用（跳过 OAuth 流程）
    const token = process.env.TEST_ACCESS_TOKEN
    if (token) {
      this.token = token
      this.projectClient.setAuth(token)
    }
  }
}

setWorldConstructor(ModelCraftWorld)
```

- [ ] **Step 5: 创建 `support/hooks.ts`**

```typescript
// tests-bdd/support/hooks.ts
import { After, Before } from '@cucumber/cucumber'
import { ModelCraftWorld } from './world'

const DELETE_MODEL = `
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      error { __typename }
    }
  }
`

const DELETE_ENUM = `
  mutation DeleteEnum($name: String!) {
    deleteEnum(name: $name) {
      error { __typename }
    }
  }
`

// 每个 Scenario 前重置追踪列表
Before(function (this: ModelCraftWorld) {
  this.createdModelIds = []
  this.createdEnumNames = []
  this.currentModelId = null
  this.modelMap = {}
  this.lastModelName = null
  this.lastEnumName = null
  this.lastResponse = null
  this.lastError = null
})

// 每个 Scenario 后通过 API 清理创建的数据（@smoke 除外保留数据方便调试）
After({ tags: 'not @smoke' }, async function (this: ModelCraftWorld) {
  // 逆序删除 model（field 随 model 级联删除）
  for (const id of [...this.createdModelIds].reverse()) {
    try {
      await this.projectClient.mutate(DELETE_MODEL, { id })
    } catch {
      // 清理失败不影响测试结果，静默处理
    }
  }

  // 删除 enum
  for (const name of this.createdEnumNames) {
    try {
      await this.projectClient.mutate(DELETE_ENUM, { name })
    } catch {
      // 静默处理
    }
  }
})
```

- [ ] **Step 6: 验证 TypeScript 类型检查通过**

```bash
cd tests-bdd && npx tsc --noEmit
```

Expected: 无错误输出。

- [ ] **Step 7: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/fixtures/ tests-bdd/support/
git commit -m "feat(bdd): add World, GraphQL client, REST client, hooks"
```

---

## Chunk 2: Auth Step + 第一个 Feature（Model）

### Task 3: Auth Step Definition

**Files:**
- Create: `tests-bdd/step-definitions/auth.steps.ts`
- Create: `tests-bdd/step-definitions/common.steps.ts`

- [ ] **Step 1: 创建 `step-definitions/auth.steps.ts`**

```typescript
// tests-bdd/step-definitions/auth.steps.ts
import { Given } from '@cucumber/cucumber'
import { ModelCraftWorld } from '../support/world'

/**
 * 以管理员身份登录。
 * 优先使用 .env.test 中的 TEST_ACCESS_TOKEN（已在 World 构造函数中设置）。
 * 若未设置则抛出明确错误，提示配置方法。
 */
Given('我以管理员身份登录', function (this: ModelCraftWorld) {
  if (!this.token) {
    throw new Error(
      '未找到 TEST_ACCESS_TOKEN。请在 tests-bdd/.env.test 中设置:\n' +
      'TEST_ACCESS_TOKEN=<your-token>\n' +
      '获取方式：在 modelcraft-backend/ 目录运行 just test-user-setup'
    )
  }
  // token 已在 World 构造函数中设置到 projectClient，此处只做验证
})
```

- [ ] **Step 2: 创建 `step-definitions/common.steps.ts`**

```typescript
// tests-bdd/step-definitions/common.steps.ts
import { Then } from '@cucumber/cucumber'
import { expect } from '@jest/globals'
import { ModelCraftWorld } from '../support/world'

/**
 * 断言最近操作返回了指定 __typename 的错误。
 *
 * 错误位于 payload 的 error 字段（不在顶层 errors 数组）：
 *   { data: { createModel: { error: { __typename: "ModelAlreadyExists" } } } }
 *
 * 用法：Then 应该返回错误类型 "ModelAlreadyExists"
 */
Then('应该返回错误类型 {string}', function (this: ModelCraftWorld, expectedTypename: string) {
  expect(this.lastResponse).not.toBeNull()

  // lastResponse 的结构：{ <mutationName>: { error: { __typename: "..." } } }
  // 取第一个 key（只有一个 mutation）
  const payload = Object.values(this.lastResponse!)[0] as Record<string, unknown>
  const error = payload?.error as Record<string, unknown> | null

  expect(error).not.toBeNull()
  expect(error?.__typename).toBe(expectedTypename)
})

/**
 * 断言最近操作成功（error 为 null）。
 */
Then('操作应该成功', function (this: ModelCraftWorld) {
  expect(this.lastResponse).not.toBeNull()
  const payload = Object.values(this.lastResponse!)[0] as Record<string, unknown>
  expect(payload?.error).toBeNull()
})
```

- [ ] **Step 3: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/step-definitions/auth.steps.ts tests-bdd/step-definitions/common.steps.ts
git commit -m "feat(bdd): add auth step and common error assertion step"
```

---

### Task 4: Model Feature + Step Definitions

**Files:**
- Create: `tests-bdd/features/model/manage-model.feature`
- Create: `tests-bdd/step-definitions/model.steps.ts`

- [ ] **Step 1: 创建 `features/model/manage-model.feature`**

```gherkin
# tests-bdd/features/model/manage-model.feature
Feature: 模型管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建模型
    When 我创建名为 "SmokeModel" 的模型
    Then 模型应该创建成功
    And 模型名称应该是 "SmokeModel"

  Scenario: 成功删除模型
    Given 已创建名为 "DeleteMe" 的模型
    When 我删除该模型
    Then 操作应该成功

  Scenario: 创建重名模型时报错
    Given 已创建名为 "DupModel" 的模型
    When 我再次创建名为 "DupModel" 的模型
    Then 应该返回错误类型 "ModelAlreadyExists"

  Scenario Outline: 创建非法名称模型时报错
    When 我创建名为 "<name>" 的模型
    Then 应该返回错误类型 "InvalidModelInput"

    Examples:
      | name             |
      | 123startsWithNum |
      | has space        |
      | has-hyphen       |
      | _startsUnderscore |
```

- [ ] **Step 2: 创建 `step-definitions/model.steps.ts`**

```typescript
// tests-bdd/step-definitions/model.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from '@jest/globals'
import { ModelCraftWorld } from '../support/world'
import { uniqueName } from '../fixtures/factory'

const CREATE_MODEL = `
  mutation CreateModel($input: CreateModelInput!) {
    createModel(input: $input) {
      model {
        id
        name
        title
        databaseName
      }
      error {
        __typename
        ... on ModelAlreadyExists { message }
        ... on InvalidModelInput { message }
        ... on ProjectNotFound { message }
      }
    }
  }
`

const DELETE_MODEL = `
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      error { __typename }
    }
  }
`

Given('已创建名为 {string} 的模型', async function (this: ModelCraftWorld, baseName: string) {
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createModel: { model: { id: string; name: string } | null; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }

  const model = res.createModel.model
  if (!model) throw new Error(`前置条件：创建模型 ${name} 失败`)

  // 存储到 World（不用模块级变量，避免跨 Scenario 污染）
  this.currentModelId = model.id
  this.lastModelName = model.name
  this.createdModelIds.push(model.id)
  // 供 lfk.steps.ts 使用：baseName → 实际 ID
  this.modelMap[baseName] = model.id
})

When('我创建名为 {string} 的模型', async function (this: ModelCraftWorld, baseName: string) {
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createModel: { model: { id: string } | null; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }

  if (res.createModel.model?.id) {
    this.currentModelId = res.createModel.model.id
    this.createdModelIds.push(res.createModel.model.id)
  }
})

When('我再次创建名为 {string} 的模型', async function (this: ModelCraftWorld, _baseName: string) {
  // 使用 Given 步骤保存的实际名称（已包含唯一后缀）
  const name = this.lastModelName
  if (!name) throw new Error('没有记录到上一个模型名称，请先用 Given 创建')

  const res = await this.projectClient.mutate<{
    createModel: { model: unknown; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }
})

When('我删除该模型', async function (this: ModelCraftWorld) {
  const id = this.currentModelId
  if (!id) throw new Error('没有可删除的模型（请先用 Given 创建）')

  const res = await this.projectClient.mutate<{ deleteModel: { error: unknown } }>(
    DELETE_MODEL, { id }
  )
  this.lastResponse = { deleteModel: res.deleteModel }

  // 从清理列表中移除（已手动删除）
  this.createdModelIds = this.createdModelIds.filter(mid => mid !== id)
  this.currentModelId = null
})

Then('模型应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { createModel: { model: unknown; error: unknown } }).createModel
  expect(payload.error).toBeNull()
  expect(payload.model).not.toBeNull()
})

Then('模型名称应该是 {string}', function (this: ModelCraftWorld, baseName: string) {
  const payload = (this.lastResponse as { createModel: { model: { name: string } | null } }).createModel
  // 名称带了 uniqueName 后缀，只检查包含原始 baseName
  expect(payload.model?.name).toContain(baseName)
})
```

- [ ] **Step 3: 运行 model feature，验证可执行（不需要后端）**

```bash
cd tests-bdd && npx cucumber-js features/model --dry-run 2>&1 | head -30
```

Expected: 所有 step 被识别，无 `Undefined` 警告。

- [ ] **Step 4: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/features/model/ tests-bdd/step-definitions/model.steps.ts
git commit -m "feat(bdd): add model feature and step definitions"
```

---

## Chunk 3: Field、Enum Feature

### Task 5: Field Feature + Step Definitions

**Files:**
- Create: `tests-bdd/features/field/manage-field.feature`
- Create: `tests-bdd/step-definitions/field.steps.ts`

- [ ] **Step 1: 创建 `features/field/manage-field.feature`**

```gherkin
# tests-bdd/features/field/manage-field.feature
Feature: 字段管理

  Background:
    Given 我以管理员身份登录
    And 已创建名为 "FieldTestModel" 的模型

  Scenario: 成功添加字段
    When 我为该模型添加名为 "email" 格式为 "STRING" 的字段
    Then 字段应该添加成功
    And 模型应该包含名为 "email" 的字段

  Scenario: 成功删除字段
    Given 模型已有名为 "age" 格式为 "NUMBER" 的字段
    When 我删除名为 "age" 的字段
    Then 字段应该删除成功

  Scenario Outline: 添加非法名称字段时报错
    When 我为该模型添加名为 "<name>" 格式为 "STRING" 的字段
    Then 应该返回错误类型 "InvalidModelInput"

    Examples:
      | name         |
      | 123invalid   |
      | has space    |
      | has-hyphen   |
```

- [ ] **Step 2: 创建 `step-definitions/field.steps.ts`**

```typescript
// tests-bdd/step-definitions/field.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from '@jest/globals'
import { ModelCraftWorld } from '../support/world'
import { uniqueName } from '../fixtures/factory'

const ADD_FIELDS = `
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
      id
      fields {
        name
        format
      }
      error {
        __typename
        ... on InvalidModelInput { message }
      }
    }
  }
`

const REMOVE_FIELD = `
  mutation RemoveField($modelID: ID!, $fieldName: String!) {
    removeField(modelID: $modelID, fieldName: $fieldName) {
      id
      fields { name }
    }
  }
`

// 当前 Scenario 的 model ID（从 world 的 createdModelIds 取最后一个）
function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID（请先用 Given 创建模型）')
  return id
}

Given('模型已有名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: { id: string; fields: Array<{ name: string }> } | null }>(
    ADD_FIELDS,
    { modelID: modelId, input: [{ name: fieldName, title: fieldName, format }] }
  )
  this.lastResponse = { addFields: res.addFields }
  if (!res.addFields) throw new Error(`前置条件：添加字段 ${fieldName} 失败`)
})

When('我为该模型添加名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: { id: string; fields: Array<{ name: string }> } | null }>(
    ADD_FIELDS,
    { modelID: modelId, input: [{ name: fieldName, title: fieldName, format }] }
  )
  this.lastResponse = { addFields: res.addFields }
})

When('我删除名为 {string} 的字段', async function (this: ModelCraftWorld, fieldName: string) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ removeField: { id: string } | null }>(
    REMOVE_FIELD,
    { modelID: modelId, fieldName }
  )
  this.lastResponse = { removeField: res.removeField }
})

Then('字段应该添加成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { addFields: { id: string } | null }).addFields
  expect(payload).not.toBeNull()
})

Then('模型应该包含名为 {string} 的字段', function (this: ModelCraftWorld, fieldName: string) {
  const payload = (this.lastResponse as { addFields: { fields: Array<{ name: string }> } | null }).addFields
  const fieldNames = payload?.fields.map(f => f.name) ?? []
  expect(fieldNames).toContain(fieldName)
})

Then('字段应该删除成功', function (this: ModelCraftWorld) {
  // removeField 直接返回更新后的 Model 对象（无 error 字段）
  // graphql-request 在请求失败时会 throw，能到这一步说明成功
  const payload = (this.lastResponse as { removeField: { id: string } | null }).removeField
  expect(payload).not.toBeNull()
})
```

- [ ] **Step 3: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/features/field/ tests-bdd/step-definitions/field.steps.ts
git commit -m "feat(bdd): add field feature and step definitions"
```

---

### Task 6: Enum Feature + Step Definitions

**Files:**
- Create: `tests-bdd/features/enum/manage-enum.feature`
- Create: `tests-bdd/step-definitions/enum.steps.ts`

- [ ] **Step 1: 创建 `features/enum/manage-enum.feature`**

```gherkin
# tests-bdd/features/enum/manage-enum.feature
Feature: 枚举管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建枚举
    When 我创建名为 "Status" 的枚举，选项为 "ACTIVE,INACTIVE"
    Then 枚举应该创建成功
    And 枚举名称应该包含 "Status"

  Scenario: 创建重名枚举时报错
    Given 已创建名为 "Priority" 的枚举，选项为 "HIGH,LOW"
    When 我再次创建名为 "Priority" 的枚举，选项为 "HIGH,LOW"
    Then 应该返回错误类型 "EnumAlreadyExists"

  Scenario Outline: 创建非法名称枚举时报错
    When 我创建名为 "<name>" 的枚举，选项为 "A,B"
    Then 应该返回错误类型 "InvalidEnumInput"

    Examples:
      | name         |
      | 123invalid   |
      | has space    |
      | has-hyphen   |
```

- [ ] **Step 2: 创建 `step-definitions/enum.steps.ts`**

```typescript
// tests-bdd/step-definitions/enum.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from '@jest/globals'
import { ModelCraftWorld } from '../support/world'
import { uniqueName } from '../fixtures/factory'

const CREATE_ENUM = `
  mutation CreateEnum($input: CreateEnumInput!) {
    createEnum(input: $input) {
      enum {
        id
        name
        displayName
        options { code label order }
      }
      error {
        __typename
        ... on EnumAlreadyExists { message }
        ... on InvalidEnumInput { message }
      }
    }
  }
`

function buildOptions(optionCodes: string) {
  return optionCodes.split(',').map((code, idx) => ({
    code: code.trim(),
    label: code.trim(),
    order: idx + 1,
  }))
}

Given(
  '已创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, baseName: string, optionCodes: string) {
    const name = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEnum: { enum: { id: string; name: string } | null; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }

    const enumDef = res.createEnum.enum
    if (!enumDef) throw new Error(`前置条件：创建枚举 ${name} 失败`)

    this.createdEnumNames.push(enumDef.name)
    this.lastEnumName = enumDef.name
  }
)

When(
  '我创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, baseName: string, optionCodes: string) {
    const name = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEnum: { enum: { id: string; name: string } | null; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }

    if (res.createEnum.enum?.name) {
      this.createdEnumNames.push(res.createEnum.enum.name)
    }
  }
)

When(
  '我再次创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, _baseName: string, optionCodes: string) {
    // 使用与 Given 步骤相同的实际名称
    const name = this.lastEnumName
    if (!name) throw new Error('没有记录到上一个枚举名称，请先用 Given 创建')
    const res = await this.projectClient.mutate<{
      createEnum: { enum: unknown; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }
  }
)

Then('枚举应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { createEnum: { enum: unknown; error: unknown } }).createEnum
  expect(payload.error).toBeNull()
  expect(payload.enum).not.toBeNull()
})

Then('枚举名称应该包含 {string}', function (this: ModelCraftWorld, baseName: string) {
  const payload = (this.lastResponse as { createEnum: { enum: { name: string } | null } }).createEnum
  expect(payload.enum?.name).toContain(baseName)
})
```

- [ ] **Step 3: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/features/enum/ tests-bdd/step-definitions/enum.steps.ts
git commit -m "feat(bdd): add enum feature and step definitions"
```

---

## Chunk 4: LFK Feature + 端到端验证

### Task 7: Logical Foreign Key Feature + Step Definitions

**Files:**
- Create: `tests-bdd/features/logical-foreign-key/manage-lfk.feature`
- Create: `tests-bdd/step-definitions/lfk.steps.ts`

- [ ] **Step 1: 创建 `features/logical-foreign-key/manage-lfk.feature`**

```gherkin
# tests-bdd/features/logical-foreign-key/manage-lfk.feature
Feature: 逻辑外键管理

  Background:
    Given 我以管理员身份登录
    And 已创建名为 "OrderModel" 的模型
    And 已创建名为 "UserModel" 的模型
    And "OrderModel" 已有名为 "userId" 格式为 "STRING" 的字段
    And "UserModel" 已有名为 "id" 格式为 "STRING" 的字段

  Scenario: 成功创建逻辑外键
    When 我创建从 "OrderModel.userId" 到 "UserModel.id" 的逻辑外键
    Then 逻辑外键应该创建成功

  Scenario: 字段不存在时报错
    When 我创建从 "OrderModel.nonExistent" 到 "UserModel.id" 的逻辑外键
    Then 应该返回 LFK 错误类型 "FKColumnsNotFoundError"
```

- [ ] **Step 2: 创建 `step-definitions/lfk.steps.ts`**

```typescript
// tests-bdd/step-definitions/lfk.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from '@jest/globals'
import { ModelCraftWorld } from '../support/world'

const ADD_FIELDS = `
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
      id
      name
    }
  }
`

const CREATE_LFK = `
  mutation CreateLogicalForeignKey($input: CreateLogicalForeignKeyInput!) {
    createLogicalForeignKey(input: $input) {
      result {
        __typename
        ... on LogicalForeignKey {
          id
          pairId
          sourceFields
          targetFields
        }
        ... on FKColumnsNotFoundError { message }
        ... on FKFieldCountMismatchError { message }
      }
    }
  }
`

Given('{string} 已有名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, modelBaseName: string, fieldName: string, format: string
) {
  const modelId = this.modelMap[modelBaseName]
  if (!modelId) throw new Error(`模型 ${modelBaseName} 尚未创建，请先用 Given 创建`)

  await this.projectClient.mutate(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: fieldName, title: fieldName, format }],
  })
})

When('我创建从 {string} 到 {string} 的逻辑外键', async function (
  this: ModelCraftWorld, source: string, target: string
) {
  // source/target 格式: "ModelBaseName.fieldName"
  const [srcModel, srcField] = source.split('.')
  const [tgtModel, tgtField] = target.split('.')

  const srcId = this.modelMap[srcModel]
  const tgtId = this.modelMap[tgtModel]

  if (!srcId || !tgtId) {
    throw new Error(`模型 ID 未找到: ${srcModel}=${srcId}, ${tgtModel}=${tgtId}`)
  }

  const res = await this.projectClient.mutate<{
    createLogicalForeignKey: { result: { __typename: string; id?: string; pairId?: string } }
  }>(CREATE_LFK, {
    input: {
      modelId: srcId,
      refModelId: tgtId,
      sourceFields: [srcField],
      targetFields: [tgtField],
    },
  })
  this.lastResponse = { createLogicalForeignKey: res.createLogicalForeignKey }
})

Then('逻辑外键应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    createLogicalForeignKey: { result: { __typename: string } }
  }).createLogicalForeignKey
  expect(payload.result.__typename).toBe('LogicalForeignKey')
})

Then('应该返回 LFK 错误类型 {string}', function (this: ModelCraftWorld, expectedTypename: string) {
  const payload = (this.lastResponse as {
    createLogicalForeignKey: { result: { __typename: string } }
  }).createLogicalForeignKey
  expect(payload.result.__typename).toBe(expectedTypename)
})
```

> 注：LFK 的 Background 需要同时创建两个模型并各添加字段。`Given 已创建名为 "X" 的模型` step 需要在 `model.steps.ts` 中维护一个 `modelMap`，将 baseName 映射到实际 ID。

- [ ] **Step 3: 更新 `step-definitions/lfk.steps.ts` 以使用 `this.modelMap`**

`model.steps.ts` 的 `Given 已创建名为 {string} 的模型` 已经在 Task 4 中完整实现了 `this.modelMap[baseName] = model.id`（见上文）。`lfk.steps.ts` 直接读取 `this.modelMap[modelBaseName]` 即可，无需额外改动。

验证两个文件的 `modelMap` 引用一致：
- `model.steps.ts`：写入 `this.modelMap[baseName] = model.id`
- `lfk.steps.ts`：读取 `this.modelMap[modelBaseName]`

两者的 key 都是传入的 `baseName`（如 `"OrderModel"`），一致。

- [ ] **Step 4: Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/features/logical-foreign-key/ tests-bdd/step-definitions/lfk.steps.ts
git commit -m "feat(bdd): add logical-foreign-key feature and step definitions"
```

---

### Task 8: Dry-run 验证所有 Feature

- [ ] **Step 1: 在 `tests-bdd/` 中运行全量 dry-run**

```bash
cd tests-bdd && npx cucumber-js --dry-run 2>&1
```

Expected: 所有 step 被识别，**无 `Undefined` 或 `Pending`**，输出类似：
```
38 scenarios (38 skipped)
xxx steps (xxx skipped)
```

- [ ] **Step 2: 修复任何 Undefined step**

若有 `Undefined` step，添加对应实现后重新运行 dry-run。

- [ ] **Step 3: TypeScript 类型检查**

```bash
cd tests-bdd && npx tsc --noEmit
```

Expected: 无错误。

- [ ] **Step 4: Commit（如有改动）**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/
git commit -m "fix(bdd): resolve undefined steps from dry-run"
```

---

### Task 9: 配置测试环境并运行真实测试

- [ ] **Step 1: 在 `modelcraft-backend/` 启动后端服务**

```bash
cd modelcraft-backend && just run force=true
```

Expected: 服务在 8080 端口启动。

- [ ] **Step 2: 确认测试用户和 token**

```bash
cd modelcraft-backend && just test-user-setup
```

将输出的 accessToken 填入 `tests-bdd/.env.test` 的 `TEST_ACCESS_TOKEN`。

- [ ] **Step 3: 运行 @smoke 场景验证连通性**

```bash
cd tests-bdd && npx cucumber-js --tags @smoke
```

Expected: 2 个 scenario 通过（model + enum smoke）。

- [ ] **Step 4: 运行全量测试**

```bash
cd tests-bdd && npm test
```

Expected: 所有 scenario 通过，无失败。

- [ ] **Step 5: 生成 HTML 报告**

```bash
cd tests-bdd && npm run test:report
# 打开 reports/test-report.html 查看结果
```

- [ ] **Step 6: 最终 Commit**

```bash
cd /home/luke/modelcraft_project
git add tests-bdd/
git commit -m "feat(bdd): complete BDD test suite - all scenarios passing"
```

---

## 附录：常见问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| `Cannot find module 'tsx'` | 依赖未安装 | `cd tests-bdd && npm install` |
| `TEST_ACCESS_TOKEN not set` | `.env.test` 未配置 | 运行 `just test-user-setup` 获取 token |
| `GraphQL error: Not authenticated` | Token 过期或无效 | 重新获取 token |
| `Undefined step` | Step 定义文件未被加载 | 检查 `cucumber.js` 的 `require` 路径 |
| `Cannot read property of null` (lastResponse) | `When` step 未设置 `lastResponse` | 检查 step 实现 |
| LFK 测试失败：模型 ID 未找到 | `modelMap` 未正确填充 | 确认 `model.steps.ts` 的 Given 步骤中 `this.modelMap[baseName] = model.id` 已执行 |
