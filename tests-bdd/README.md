# ModelCraft BDD Tests

基于 [Cucumber.js](https://cucumber.io/) 的行为驱动测试套件，用 Gherkin 语言描述业务场景，覆盖模型设计时 GraphQL API。

---

## 前置条件

- Node.js 18+
- 后端服务已运行（`http://localhost:8080`）
- 已存在测试用的 Org 和 Project

---

## 快速开始

### 1. 安装依赖

```bash
cd tests-bdd
npm install
```

### 2. 获取测试 Token

在 `modelcraft-backend/` 目录执行：

```bash
just test-user-setup
```

复制输出的 Access Token。

### 3. 配置环境变量

编辑 `tests-bdd/.env.test`：

```bash
API_BASE_URL=http://localhost:8080
TEST_ACCESS_TOKEN=<粘贴上一步获取的 token>
TEST_ORG_NAME=test-org       # 你的 Org 名称
TEST_PROJECT_SLUG=test-project  # 你的 Project Slug
```

### 4. 运行测试

```bash
# 运行全部测试
npm test

# 按功能域运行
npm run test:model   # 模型管理
npm run test:field   # 字段管理
npm run test:enum    # 枚举管理
npm run test:lfk     # 逻辑外键

# 只跑冒烟测试（带 @smoke 标签）
npm run test:smoke

# 生成 HTML 报告
npm run test:report
# 报告输出到 reports/test-report.html
```

---

## 目录结构

```
tests-bdd/
├── features/                  # Gherkin 业务场景（中文）
│   ├── model/                 # 模型 CRUD
│   ├── field/                 # 字段管理
│   ├── enum/                  # 枚举管理
│   └── logical-foreign-key/   # 逻辑外键
├── step-definitions/          # TypeScript Step 实现
│   ├── auth.steps.ts          # 登录步骤
│   ├── common.steps.ts        # 通用断言（操作成功 / 返回错误类型）
│   ├── model.steps.ts         # 模型相关步骤
│   ├── field.steps.ts         # 字段相关步骤
│   ├── enum.steps.ts          # 枚举相关步骤
│   └── lfk.steps.ts           # 逻辑外键步骤
├── support/
│   ├── world.ts               # Cucumber World（跨 Step 共享状态）
│   ├── graphql-client.ts      # GraphQL 请求封装
│   ├── rest-client.ts         # REST 请求封装（OAuth 备用）
│   └── hooks.ts               # Before/After 生命周期（数据清理）
├── fixtures/
│   └── factory.ts             # uniqueName() — 生成无冲突的测试数据名称
├── reports/                   # 生成的 HTML 测试报告
├── cucumber.js                # Cucumber 配置
├── .env.test                  # 环境变量（不提交到 Git）
└── tsconfig.json
```

---

## 测试覆盖范围

| 功能域 | Feature 文件 | 主要场景 |
|--------|-------------|----------|
| 模型管理 | `features/model/manage-model.feature` | 创建、删除、重名报错、非法名称报错 |
| 字段管理 | `features/field/manage-field.feature` | 添加字段、删除字段 |
| 枚举管理 | `features/enum/manage-enum.feature` | 创建枚举、重名报错 |
| 逻辑外键 | `features/logical-foreign-key/manage-lfk.feature` | 创建外键、字段不存在报错 |

---

## 设计说明

### 测试数据隔离

每个 Scenario 的测试数据名称都通过 `uniqueName()` 添加 8 位随机后缀（如 `User` → `Usera3f2b1c0`），避免并发或重复运行时冲突。

### 自动清理

每个 Scenario 执行后，`After` 钩子会通过 API 删除本次创建的所有模型和枚举。带 `@smoke` 标签的场景**跳过清理**，方便调试时查看数据。

### 状态管理

使用 Cucumber **World** 对象（`support/world.ts`）在同一 Scenario 的各 Step 间传递状态（当前模型 ID、最近响应等），避免模块级变量导致的跨 Scenario 污染。

### GraphQL Endpoint

Project 域接口地址格式：

```
/graphql/org/{orgName}/project/{projectSlug}/
```

Token 通过 `Authorization: Bearer <token>` 请求头传递。

---

## 常见问题

**Q: 运行时报 `未找到 TEST_ACCESS_TOKEN`**

在 `.env.test` 中填入有效 token。Token 通过 `just test-user-setup`（后端目录）获取。

**Q: 报 `ProjectNotFound` 或连接错误**

确认 `API_BASE_URL`、`TEST_ORG_NAME`、`TEST_PROJECT_SLUG` 配置正确，且后端服务正在运行。

**Q: 测试数据没有清理**

带 `@smoke` 标签的场景不会自动清理，其余场景如果 After 钩子清理失败会静默跳过（不影响测试结果），可手动登录后台删除。
