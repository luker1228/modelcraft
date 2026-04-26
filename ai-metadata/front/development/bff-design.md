# BFF 层设计

BFF（Backend For Frontend）层是前端与 Go 后端之间的适配层，封装认证逻辑、GraphQL 客户端实例化和 HTTP 代理处理器。

---

## 目录结构

```
src/bff/
├── auth/
│   ├── casdoor.ts               # Casdoor SDK 封装：Token 生命周期管理
│   ├── token-utils.ts           # 从 localStorage 读取 org 上下文
│   └── public.ts                # 对外暴露的公开 API（门面）
│
├── apollo/
│   ├── clients.ts               # 三种 Apollo Client 工厂函数
│   └── public.ts                # 对外暴露的公开 API（门面）
│
├── cms/
│   ├── runtime-query-builder.ts # 动态 GraphQL 查询/变更构建器
│   └── public.ts                # 对外暴露的公开 API（门面）
│
└── api/
    ├── auth/token.ts            # POST /api/auth/token 处理器
    ├── auth/refresh.ts          # POST /api/auth/refresh 处理器
    ├── user/memberships.ts      # GET /api/user/memberships 处理器
    ├── org/init.ts              # POST /api/org/init 处理器
    └── copilotkit.ts            # POST /api/copilotkit 处理器
```

---

## 门面模式（Public Facade）

每个 BFF 子模块都有一个 `public.ts` 作为**唯一对外出口**，Web Layer 只能从门面导入，禁止访问内部模块。

```mermaid
graph LR
    W["Web Layer"] -->|"@bff/auth/public"| PA["auth/public.ts"]
    W -->|"@bff/apollo/public"| PB["apollo/public.ts"]
    W -->|"@bff/cms/public"| PC["cms/public.ts"]

    PA --> A["auth/casdoor.ts<br/><i>内部实现</i>"]
    PB --> B["apollo/clients.ts<br/><i>内部实现</i>"]
    PC --> C["cms/runtime-query-builder.ts<br/><i>内部实现</i>"]

    style A fill:#f5f5f5,stroke:#ccc
    style B fill:#f5f5f5,stroke:#ccc
    style C fill:#f5f5f5,stroke:#ccc
```

ESLint 规则强制执行此边界：禁止 Web Layer 跳过 `public.ts` 直接访问 BFF 内部文件。

---

## BFF Route 中调用后端的两种方式

BFF route 运行在服务端，不持有浏览器的 Design JWT，但有两种合法路径调用后端：

| 方式 | 认证 | 适用场景 |
|------|------|---------|
| `callGoXxx()` → `/internal/...` | `X-Internal-Token` | 已有封装的管理操作（用户 CRUD、访问控制等） |
| `postGraphQL()` → `/graphql/...` | `X-Internal-Token` + `X-Org-Name`/`X-Project-Slug` | 需要走 GraphQL 但没有对应 Internal API 时 |

**不能做的事**：把浏览器用户的 JWT 从 `req.headers` 转发给 GraphQL——这依赖浏览器主动传 token，架构不稳定。BFF 应该用服务端持有的 `X-Internal-Token` 自己完成认证。

```ts
// ✅ 方式一：用已有的 go-client 封装
import { callGoListOrgEndUsers } from '@/bff/end-user/end-user-go-client'
const result = await callGoListOrgEndUsers({ orgName })

// ✅ 方式二：BFF 用 X-Internal-Token 打 GraphQL
const res = await fetch(`${GO_BACKEND_URL}/graphql/org/${orgName}/`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Internal-Token': INTERNAL_TOKEN,
    'X-Org-Name': orgName,
  },
  body: JSON.stringify({ query, variables }),
})

// ❌ 错误：转发浏览器 JWT（服务端不稳定，且已有更好方式）
const authorization = req.headers.get('authorization')
headers.authorization = authorization   // ← 不要这样做
```

---

## 端点感知集中在 BFF

后端 API URL、路径拼接、认证头注入等细节**只存在于 BFF 层**。组件层禁止硬编码任何端点 URL。

```tsx
// ❌ 禁止：组件中硬编码端点
const url = `/graphql/org/${orgName}/project/${slug}/`

// ✅ 正确：通过 BFF 导出的函数获取
import { createProjectScopedClient } from '@bff/apollo/public'
const client = createProjectScopedClient(orgName, slug)
```

---

## GraphQL 端点规范

采用 `/graphql` 前缀模式，统一所有 GraphQL 端点。客户端只需一个 base URL，通过路径扩展访问不同资源。

| 通道 | 端点 | 客户端实例 | 用途 |
|------|------|-----------|------|
| Org-Scoped | `/graphql/org/{orgName}/` | 单例 | 项目、集群、用户、角色管理 |
| Project-Scoped | `/graphql/org/{orgName}/project/{projectSlug}/` | 每次新建 | 模型、字段、枚举 CRUD |
| Model Runtime | `/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}` | 每次新建 | 运行时数据查询/变更 |

### 为什么用 prefix 模式

项目客户端只使用 GraphQL 协议。prefix 模式下客户端只需知道一个 base URL `/graphql`，通过路径扩展即可覆盖所有资源范围，便于在 nginx/proxy 层统一拦截。

---

## Auth 模块

负责 Casdoor OAuth2 集成和 JWT Token 的完整生命周期管理。

### Token 存储

| Key | 内容 |
|-----|------|
| `localStorage.auth_token` | JWT access token |
| `localStorage.auth_refresh_token` | Refresh token |

### 对外暴露的函数（via `public.ts`）

| 函数 | 说明 |
|------|------|
| `getToken()` | 获取当前 access token |
| `storeToken(access, refresh)` | 持久化 Token |
| `removeToken()` | 清除所有 Token（登出） |
| `isAuthenticated()` | 检查是否已登录且未过期 |
| `isTokenNearExpiry()` | Token 是否在 5 分钟内过期 |
| `getUserInfoFromToken()` | 从 JWT payload 解析用户信息 |
| `exchangeCodeForToken(code)` | OAuth 授权码换 Token |
| `refreshAccessToken()` | 静默刷新 Token（单例模式） |
| `redirectToLogin()` | 跳转 Casdoor 登录页 |

**Token 刷新采用单例 Promise**，确保并发场景下只发起一次刷新请求，其他请求等待同一个 Promise 结果。

---

## Apollo 模块

### 三种客户端实例

```mermaid
graph TD
    F["前端组件"]

    F -->|"org 级操作<br/>项目/集群/用户/角色"| OC["Org-Scoped Client<br/><i>单例</i><br/>/graphql/org/{orgName}/"]
    F -->|"project 级操作<br/>模型/字段/枚举"| PC["Project-Scoped Client<br/><i>每次新建</i><br/>/graphql/org/{orgName}/project/{slug}/"]
    F -->|"运行时数据<br/>查询/变更"| RC["Model Runtime Client<br/><i>每次新建</i><br/>/graphql/org/{orgName}/project/{slug}/db/{db}/model/{model}"]

    OC --> B["Go Backend"]
    PC --> B
    RC --> B
```

| 客户端 | 实例策略 | 原因 |
|--------|----------|------|
| Org-Scoped | 单例 | org 内操作共享缓存，减少重复请求 |
| Project-Scoped | 每次新建 | 不同 project 的模型数据不应互相污染缓存 |
| Model Runtime | 每次新建 | 动态 Schema，无法共享缓存 |

每种客户端均包含：Auth Link（注入 Bearer Token + x-request-id）、Error Link（401 自动跳转登录）。

---

## API Handlers 模块

Next.js API Routes 作为代理层，将请求转发至 Go 后端，统一由 BFF 处理认证 header 的注入。

| 路由 | 方法 | 说明 |
|------|------|------|
| `/api/auth/token` | POST | OAuth code → JWT |
| `/api/auth/refresh` | POST | refresh token → 新 access token |
| `/api/user/memberships` | GET | 获取用户所属组织列表 |
| `/api/org/init` | POST | 初始化组织（幂等） |
| `/api/copilotkit` | POST | AI Copilot 代理 |

---

## CMS 模块（动态查询构建器）

根据模型名和字段列表，在运行时动态生成 GraphQL 查询和变更语句，供 Model Runtime Client 使用。

| 函数 | 生成内容 |
|------|----------|
| `buildFindManyQuery` | 分页列表查询 |
| `buildFindUniqueQuery` | 按主键查询单条 |
| `buildCreateMutation` | 创建记录 |
| `buildUpdateMutation` | 更新记录 |
| `buildDeleteMutation` | 删除记录 |
| `buildCountQuery` | 统计数量 |

---

## Membership 三级缓存

对 `/api/user/memberships` 的结果进行三级缓存，减少组织切换时的重复网络请求。

```mermaid
flowchart TD
    R["请求 memberships"]
    M{"Memory Cache<br/>TTL: 5 min"}
    L{"localStorage<br/>持久化"}
    A["GET /api/user/memberships"]

    R --> M
    M -->|命中| RES["返回结果"]
    M -->|未命中| L
    L -->|命中| RES
    L -->|未命中| A
    A --> RES
    A --> M
    A --> L
```

同一时刻最多只有一个并发 API 请求（单例 Promise），防止 org 切换时重复发起请求。

---

## Mock 开发策略（MSW）

前端开发阶段使用 **MSW（Mock Service Worker）** 在网络层拦截请求，BFF 代码无需感知 mock 存在，联调时关闭 MSW 即可，零修改 BFF 代码。

### 为什么选 MSW

- 在 Service Worker / Node 网络层拦截，BFF 代码保持干净
- mock handler 从 GraphQL Schema 自动生成，保证与后端契约一致
- 可模拟延迟、4xx/5xx、网络超时等真实场景
- 联调时一行代码关闭，无需改动任何 BFF 实现

### 目录结构

```
src/mocks/
├── browser.ts          # 浏览器端 MSW worker 启动入口
├── node.ts             # Node 端（测试）MSW server 启动入口
├── handlers/
│   ├── index.ts        # 汇总所有 handler
│   ├── org/            # Org-Scoped GraphQL handlers
│   │   └── generated.ts  # codegen 自动生成，禁止手动编辑
│   └── project/        # Project-Scoped GraphQL handlers
│       └── generated.ts  # codegen 自动生成，禁止手动编辑
└── data/
    ├── org/            # Org 域 mock 数据工厂
    └── project/        # Project 域 mock 数据工厂
```

### 安装与配置

```bash
# 安装 MSW
npm install msw --save-dev

# 生成 Service Worker 文件（放入 public/）
npx msw init public/ --save
```

`src/mocks/browser.ts`：

```ts
import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

export const worker = setupWorker(...handlers)
```

在 `src/app/layout.tsx`（或 `_app.tsx`）中启动：

```ts
// 只在开发环境 + mock 模式下启动
if (process.env.NEXT_PUBLIC_API_MOCKING === 'enabled') {
  const { worker } = await import('@/mocks/browser')
  await worker.start({ onUnhandledRequest: 'bypass' })
}
```

`.env.local` 中控制开关：

```bash
NEXT_PUBLIC_API_MOCKING=enabled   # 开发阶段
# NEXT_PUBLIC_API_MOCKING=        # 联调 / 生产（留空或删除）
```

### codegen 生成 mock handlers

`contract/` 目录更新后，运行 codegen 自动生成 MSW handler：

```bash
# 生成类型 + mock handlers（一次执行）
npm run codegen
```

`codegen.ts` 配置示例：

```ts
// codegen.ts
const config: CodegenConfig = {
  generates: {
    // 1. 生成 TypeScript 类型
    'src/generated/graphql.ts': {
      plugins: ['typescript', 'typescript-operations'],
    },
    // 2. 生成 MSW handlers（org 域）
    'src/mocks/handlers/org/generated.ts': {
      schema: 'contract/graph/org/schema/**/*.graphql',
      plugins: ['typescript-msw'],
    },
    // 3. 生成 MSW handlers（project 域）
    'src/mocks/handlers/project/generated.ts': {
      schema: 'contract/graph/project/schema/**/*.graphql',
      plugins: ['typescript-msw'],
    },
  },
}
```

生成的 handler 只提供骨架，需在 `data/` 目录补充具体 mock 数据：

```ts
// src/mocks/data/project/model-factory.ts
import { faker } from '@faker-js/faker'

export function createMockModel(override = {}) {
  return {
    id: faker.string.uuid(),
    name: faker.word.noun(),
    displayName: faker.commerce.productName(),
    description: faker.lorem.sentence(),
    createdAt: faker.date.recent().toISOString(),
    ...override,
  }
}
```

### handler 汇总

```ts
// src/mocks/handlers/index.ts
import { orgHandlers } from './org/generated'
import { projectHandlers } from './project/generated'

export const handlers = [
  ...orgHandlers,
  ...projectHandlers,
]
```

### 开发 → 联调切换流程

```
开发阶段
  NEXT_PUBLIC_API_MOCKING=enabled
  前端请求 → MSW 拦截 → mock 数据
        ↓
后端接口就绪，准备联调
  .env.local 中删除或置空 NEXT_PUBLIC_API_MOCKING
  前端请求 → BFF → 真实后端
  （BFF 代码零修改）
```

### 注意事项

- `src/mocks/handlers/*/generated.ts` 由 codegen 生成，**禁止手动编辑**
- mock 数据工厂放在 `src/mocks/data/`，由 worker 负责维护
- `contract/` 更新后必须重新运行 codegen，否则 mock 与 spec 可能不一致
- mock handler 默认返回成功响应；错误场景（4xx/5xx）在 `data/` 中补充专用工厂函数

---

## Mock 开发策略（MSW）

前端开发阶段使用 **MSW（Mock Service Worker）** 在网络层拦截请求，BFF 代码无需感知 mock 存在，联调时关闭 MSW 即可，零修改 BFF 代码。

### 为什么选 MSW

- 在 Service Worker / Node 网络层拦截，BFF 代码保持干净
- mock handler 从 GraphQL Schema 自动生成，保证与后端契约一致
- 可模拟延迟、4xx/5xx、网络超时等真实场景
- 联调时一行代码关闭，无需改动任何 BFF 实现

### 目录结构

```
src/mocks/
├── browser.ts          # 浏览器端 MSW worker 启动入口
├── node.ts             # Node 端（测试）MSW server 启动入口
├── handlers/
│   ├── index.ts        # 汇总所有 handler
│   ├── org/            # Org-Scoped GraphQL handlers
│   │   └── generated.ts  # codegen 自动生成，禁止手动编辑
│   └── project/        # Project-Scoped GraphQL handlers
│       └── generated.ts  # codegen 自动生成，禁止手动编辑
└── data/
    ├── org/            # Org 域 mock 数据工厂
    └── project/        # Project 域 mock 数据工厂
```

### 安装与配置

```bash
# 安装 MSW
npm install msw --save-dev

# 生成 Service Worker 文件（放入 public/）
npx msw init public/ --save
```

`src/mocks/browser.ts`：

```ts
import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

export const worker = setupWorker(...handlers)
```

在 `src/app/layout.tsx`（或 `_app.tsx`）中启动：

```ts
// 只在开发环境 + mock 模式下启动
if (process.env.NEXT_PUBLIC_API_MOCKING === 'enabled') {
  const { worker } = await import('@/mocks/browser')
  await worker.start({ onUnhandledRequest: 'bypass' })
}
```

`.env.local` 中控制开关：

```bash
NEXT_PUBLIC_API_MOCKING=enabled   # 开发阶段
# NEXT_PUBLIC_API_MOCKING=        # 联调 / 生产（留空或删除）
```

### codegen 生成 mock handlers

`contract/` 目录更新后，运行 codegen 自动生成 MSW handler：

```bash
# 生成类型 + mock handlers（一次执行）
npm run codegen
```

`codegen.ts` 配置示例：

```ts
// codegen.ts
const config: CodegenConfig = {
  generates: {
    // 1. 生成 TypeScript 类型
    'src/generated/graphql.ts': {
      plugins: ['typescript', 'typescript-operations'],
    },
    // 2. 生成 MSW handlers（org 域）
    'src/mocks/handlers/org/generated.ts': {
      schema: 'contract/graph/org/schema/**/*.graphql',
      plugins: ['typescript-msw'],
    },
    // 3. 生成 MSW handlers（project 域）
    'src/mocks/handlers/project/generated.ts': {
      schema: 'contract/graph/project/schema/**/*.graphql',
      plugins: ['typescript-msw'],
    },
  },
}
```

生成的 handler 只提供骨架，需在 `data/` 目录补充具体 mock 数据：

```ts
// src/mocks/data/project/model-factory.ts
import { faker } from '@faker-js/faker'

export function createMockModel(override = {}) {
  return {
    id: faker.string.uuid(),
    name: faker.word.noun(),
    displayName: faker.commerce.productName(),
    description: faker.lorem.sentence(),
    createdAt: faker.date.recent().toISOString(),
    ...override,
  }
}
```

### handler 汇总

```ts
// src/mocks/handlers/index.ts
import { orgHandlers } from './org/generated'
import { projectHandlers } from './project/generated'

export const handlers = [
  ...orgHandlers,
  ...projectHandlers,
]
```

### 开发 → 联调切换流程

```
开发阶段
  NEXT_PUBLIC_API_MOCKING=enabled
  前端请求 → MSW 拦截 → mock 数据
        ↓
后端接口就绪，准备联调
  .env.local 中删除或置空 NEXT_PUBLIC_API_MOCKING
  前端请求 → BFF → 真实后端
  （BFF 代码零修改）
```

### 注意事项

- `src/mocks/handlers/*/generated.ts` 由 codegen 生成，**禁止手动编辑**
- mock 数据工厂放在 `src/mocks/data/`，由 worker 负责维护
- `contract/` 更新后必须重新运行 codegen，否则 mock 与 spec 可能不一致
- mock handler 默认返回成功响应；错误场景（4xx/5xx）在 `data/` 中补充专用工厂函数

---

## 参考文档

| 主题 | 文档 |
|------|------|
| 前端架构总览 | [architecture.md](./architecture.md) |
| ESLint 层边界规则 | [eslint-rules.md](./eslint-rules.md) |
| 代码规范 | [code-conventions.md](./code-conventions.md) |
