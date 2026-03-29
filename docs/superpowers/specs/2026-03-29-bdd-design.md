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

- ✅ API 级别（GraphQL + REST），**不涉及浏览器**
- ✅ 完整链路：认证 → 业务操作 → 结果验证
- ✅ 完全替代 Python 集成测试（包括边界值场景）
- ❌ 不替代 Go 单元测试（Domain 层 95% 覆盖率保持不变）

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
│   ├── auth/
│   │   └── login.feature
│   ├── project/
│   │   └── manage-project.feature
│   ├── model/
│   │   └── manage-model.feature
│   ├── field/
│   │   └── manage-field.feature
│   ├── cluster/
│   │   └── manage-cluster.feature
│   └── enum/
│       └── manage-enum.feature
├── step-definitions/                # Step 实现（TypeScript）
│   ├── auth.steps.ts
│   ├── project.steps.ts
│   ├── model.steps.ts
│   ├── field.steps.ts
│   ├── cluster.steps.ts
│   ├── enum.steps.ts
│   └── common.steps.ts              # 通用 Given/Then（如断言错误码）
├── support/
│   ├── world.ts                     # 共享状态（每 Scenario 独立实例）
│   ├── hooks.ts                     # Before/After 钩子（数据清理）
│   ├── graphql-client.ts            # GraphQL 客户端（Org + Project 两个 endpoint）
│   └── rest-client.ts               # REST 客户端（Auth/Org API）
├── fixtures/
│   └── factory.ts                   # 测试数据工厂（生成唯一名称等）
├── cucumber.js                      # Cucumber 配置
├── tsconfig.json
├── package.json
└── .env.test                        # 测试环境配置（API base URL、测试账号）
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
  orgClient: GraphQLClient        // /graphql/org/{orgName}/
  projectClient: GraphQLClient    // /graphql/org/{orgName}/project/{projectSlug}/

  // 认证状态
  token: string | null
  currentOrgName: string | null

  // 当前操作上下文
  currentProjectSlug: string | null
  currentClusterName: string | null

  // 最近操作结果（供 Then 断言）
  lastResponse: unknown
  lastError: unknown
}
```

**设计原则**：
- `lastResponse` / `lastError` 解耦 `When` 和 `Then`，不需要在 Step 之间传参
- 两个 GraphQL client 对应项目里两个不同 endpoint（Org GraphQL / Project GraphQL）
- 客户端自动从 `world.token` 读取认证信息

### 4.2 客户端设计

项目有三条 API 通道，客户端需分别封装：

| 通道 | 路径 | 封装 |
|------|------|------|
| Org GraphQL | `/graphql/org/{orgName}/` | `orgClient` |
| Project GraphQL | `/graphql/org/{orgName}/project/{projectSlug}/` | `projectClient` |
| REST (OpenAPI) | `/api/auth/*`、`/api/org/*` | `restClient` |

```typescript
// support/graphql-client.ts
export class GraphQLClient {
  constructor(private type: 'org' | 'project') {}

  setAuth(token: string) { /* 设置 Bearer token */ }
  setContext(orgName: string, projectSlug?: string) { /* 设置 URL 参数 */ }

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
    And 存在组织 "test-org"
    And 存在项目 "test-project"

  Scenario: 成功创建模型
    When 我在项目中创建名为 "User" 的模型
    Then 模型 "User" 应该存在于项目中

  Scenario: 创建重名模型时报错
    Given 项目中已存在名为 "User" 的模型
    When 我再次创建名为 "User" 的模型
    Then 应该返回错误 "CONFLICT.MODEL"

  Scenario Outline: 创建非法名称的模型时报错
    When 我在项目中创建名为 "<name>" 的模型
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

```typescript
// support/hooks.ts
After(async function (this: ModelCraftWorld) {
  // 通过 API 逆向操作清理，不直接操作数据库
  if (this.currentProjectSlug) {
    await this.orgClient.deleteProject(this.currentProjectSlug)
  }
})

// @smoke 标记的场景保留数据，方便调试
After({ tags: 'not @smoke' }, async function (this: ModelCraftWorld) {
  // 清理逻辑
})
```

**原则**：
- 只通过 API 清理数据，不直接操作数据库
- 每个 Scenario 完全独立，可任意顺序执行
- `factory.ts` 生成带时间戳的唯一名称，避免并发冲突

---

## 5. 与现有测试的关系

| 测试类型 | 现状 | BDD 引入后 |
|----------|------|------------|
| Go 单元测试（Domain 层） | 282 个，95% 覆盖率 | **保留**，职责不变 |
| Python 集成测试 | `modelcraft-backend/tests/` | **逐步废弃**，由 BDD 完全替代 |
| BDD 验收测试 | 无 | **新建**，覆盖所有用户故事 + 边界值 |

**迁移策略**：先建 BDD，场景覆盖 Python 测试后，再删除对应 Python 测试文件。

---

## 6. 运行方式

```bash
# 安装依赖
cd tests-bdd && npm install

# 运行所有场景
npm test

# 运行指定 feature
npm test -- features/model/manage-model.feature

# 只运行冒烟测试
npm test -- --tags @smoke

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
TEST_ORG_NAME=test-org
```

---

## 8. 成功标准

- [ ] 所有核心业务域（auth、project、model、field、cluster、enum）有 Feature 文件
- [ ] 边界值通过 `Scenario Outline` 覆盖，不需要额外的测试代码
- [ ] 每个 Scenario 独立运行，无顺序依赖
- [ ] Python 集成测试可完全废弃
- [ ] CI 中独立运行，不阻塞 Go 单元测试

---

## 9. 不在范围内

- 浏览器 E2E 测试（Playwright/Cypress）
- 前端组件测试
- 性能测试
- Go 单元测试的替代
