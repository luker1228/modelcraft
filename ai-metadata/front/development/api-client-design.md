# API Client 层设计

`src/api-client/` 是前端与 Go 后端之间的适配层，封装 GraphQL 文档、Apollo 客户端、认证逻辑和 HTTP 代理处理器。

---

## 目录结构

```
src/api-client/
├── apollo/
│   ├── clients.ts               # 三种 Apollo Client 工厂函数
│   └── public.ts                # 对外暴露的公开 API（门面）
│
├── auth/
│   ├── auth-client.ts           # 自建认证封装：Token 生命周期管理
│   ├── go-auth-client.ts        # Go 后端内部认证客户端
│   ├── token-utils.ts           # Token 读写工具
│   └── public.ts                # 对外暴露的公开 API（门面）
│
├── runtime-query/
│   ├── runtime-query-builder.ts # 动态 GraphQL 查询/变更构建器
│   └── public.ts
│
├── api/
│   ├── org/init.ts              # POST /api/org/init 处理器
│   └── user/memberships.ts      # GET /api/user/memberships 处理器
│
├── end-user/                    # 终端用户模块
├── model-enum/                  # 枚举字段特化模块
│
│ ── 各业务模块（GraphQL 文档 + mock handlers）──
├── model/
│   ├── graphql-docs.ts          # model 相关 gql 查询和变更文档
│   └── index.ts                 # barrel export
├── enum/
│   ├── graphql-docs.ts
│   └── index.ts
├── project/
│   ├── graphql-docs.ts
│   └── index.ts
├── cluster/
│   ├── graphql-docs.ts
│   └── index.ts
├── rbac/
│   ├── graphql-docs.ts
│   └── index.ts
├── rls/
│   ├── graphql-docs.ts
│   └── index.ts
├── user/
│   ├── graphql-docs.ts
│   └── index.ts
├── profile/
│   ├── graphql-docs.ts
│   └── index.ts
│
└── noop.ts                      # NOOP_QUERY / NOOP_MUTATION 占位符
```

---

## GraphQL 文档组织

每个业务模块在 `src/api-client/{module}/graphql-docs.ts` 中存放所有的 gql 查询和变更文档，不再集中放在 `src/web/graphql/`。

```ts
// src/api-client/model/graphql-docs.ts
import { gql } from '@apollo/client'

export const GET_MODELS = gql`...`
export const CREATE_MODEL = gql`...`
// ...
```

Web 层通过模块路径导入：

```ts
// ✅ 正确
import { GET_MODELS, CREATE_MODEL } from '@/api-client/model'
import { GET_ENUMS } from '@/api-client/enum'

// ❌ 旧路径，已废弃
import { GET_MODELS } from '@web/graphql'
import { GET_MODELS } from '@web/graphql/queries/model'
```

### codegen 扫描范围

`codegen.ts` 的 `documents` 配置指向 `src/api-client/**/*.ts`，codegen 会自动扫描所有模块的 gql 文档生成类型。

```ts
// codegen.ts
const config: CodegenConfig = {
  documents: 'src/api-client/**/*.ts',  // 扫描所有 api-client 模块
  // ...
}
```

---

## GraphQL 查询开发规范

### 以 contract 为唯一参照物

编写或修改 `graphql-docs.ts` 中的查询时，**必须以 `contract/graph/` 目录下的 Schema 为唯一真相源**，不得凭记忆或旧查询推测字段是否存在。

```
contract/graph/
├── org/schema/      # Org 域类型定义
└── project/schema/  # Project 域类型定义（Model、Field、Enum 等）
```

**开发流程：**

1. 打开 `contract/graph/project/schema/` 或 `contract/graph/org/schema/` 确认目标类型的字段列表
2. 只查询 Schema 中**明确存在**的字段
3. 编写完成后运行 `npm run codegen` 验证，codegen 报 `Cannot query field` 即说明查询了不存在的字段

```bash
# 验证查询合法性
npm run codegen
```

### 常见错误

```
Cannot query field "xxx" on type "YYY"
```

原因：查询了 Schema 中不存在的字段，或字段已被重命名/删除。

**排查步骤：**
1. 在 `contract/graph/` 中搜索该类型的定义
2. 对照实际字段列表，删除或修正查询中的非法字段
3. 重跑 `npm run codegen` 直到通过

### 错误类型字段规范

错误 union 中的各类型（`InvalidInput`、`ModelNotFound`、`FieldFormatImmutable` 等）**只能查询 Schema 中定义的字段**。当前 Project Schema 中常见错误类型的可用字段：

| 错误类型 | 可用字段 |
|---------|---------|
| `InvalidInput` | `message`, `suggestion` |
| `ModelNotFound` | `message` |
| `ProjectNotFound` | `message` |
| `ModelAlreadyExists` | `message`, `suggestion` |
| `ModelTableAlreadyExists` | `message`, `suggestion` |
| `CannotDeleteDeployedModel` | `message`, `suggestion` |
| `FieldFormatImmutable` | `message` |
| `FieldReferenceInUse` | `message`, `suggestion` |

> 以上信息以 `contract/graph/project/schema/` 为准，如有更新以 contract 为准。

### Contract 更新后的处理

当后端修改了 Schema（字段增删、类型变更），需通过 `front-contract-pull` skill 同步 contract，再重跑 codegen：

```bash
# 1. 同步后端最新 contract
# (使用 front-contract-pull skill)

# 2. 重新生成 TypeScript 类型
npm run codegen

# 3. 修复因 schema 变更导致的查询不合法问题
```

---

## 门面模式（Public Facade）

每个 api-client 子模块通过 `index.ts` 或 `public.ts` 作为**唯一对外出口**，Web Layer 只能从门面导入，禁止访问内部实现文件。

ESLint 规则强制执行此边界：禁止 Web Layer 跳过 `public.ts`/`index.ts` 直接访问模块内部文件。

---

## GraphQL 端点规范

采用 `/graphql` 前缀模式，统一所有 GraphQL 端点。

| 通道 | 端点 | 客户端实例 | 用途 |
|------|------|-----------|------|
| Org-Scoped | `/graphql/org/{orgName}/` | 单例 | 项目、集群、用户、角色管理 |
| Project-Scoped | `/graphql/org/{orgName}/project/{projectSlug}/` | 每次新建 | 模型、字段、枚举 CRUD |
| Model Runtime | `/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}` | 每次新建 | 运行时数据查询/变更 |

```tsx
// ❌ 禁止：组件中硬编码端点
const url = `/graphql/org/${orgName}/project/${slug}/`

// ✅ 正确：通过 api-client 导出的函数获取
import { createProjectScopedClient } from '@api-client/apollo/public'
const client = createProjectScopedClient(orgName, slug)
```

---

## Auth 模块

负责自建用户名/密码认证和 JWT Token 的完整生命周期管理。

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
| `refreshAccessToken()` | 静默刷新 Token（单例模式） |
| `redirectToLogin()` | 跳转登录页 |

**Token 刷新采用单例 Promise**，确保并发场景下只发起一次刷新请求。

---

## Apollo 模块

### 三种客户端实例

| 客户端 | 实例策略 | 原因 |
|--------|----------|------|
| Org-Scoped | 单例 | org 内操作共享缓存，减少重复请求 |
| Project-Scoped | 每次新建 | 不同 project 的模型数据不应互相污染缓存 |
| Model Runtime | 每次新建 | 动态 Schema，无法共享缓存 |

每种客户端均包含：Auth Link（注入 Bearer Token + x-request-id）。

---

## API Handlers 模块

Next.js API Routes 作为代理层，将请求先转发至 Gateway，再由 Gateway 转发至 Backend。

### BFF 双体系路由约定（强制）

| 体系 | 前端 BFF 路径 | Gateway 路径 | 说明 |
|------|---------------|--------------|------|
| Developer | `/api/auth/*` | `/auth/*` | 管理端登录/刷新/登出等认证链路 |
| Developer | `/api/bff/graphql/org/{orgName}[/...]` | `/graphql/org/{orgName}[/...]` | 设计态 Org/Project GraphQL |
| EndUser | `/api/bff/org/{orgName}/end-user/auth/*` | `/api/end-user/auth/*` | 终端用户登录/刷新/select-project |
| EndUser | `/api/bff/graphql/org/{orgName}/project/{projectSlug}` | `/graphql/org/{orgName}/project/{projectSlug}` | 终端用户 GraphQL（与 Developer 共用同一端点） |

- 前端 `BACKEND_URL` 必须配置为 Gateway 地址。
- 禁止浏览器侧或前端服务侧直接请求 Backend 业务端口。

| 路由 | 方法 | 说明 |
|------|------|------|
| `/api/auth/token` | POST | username/password → JWT |
| `/api/auth/refresh` | POST | refresh token → 新 access token |
| `/api/user/memberships` | GET | 获取用户所属组织列表 |
| `/api/org/init` | POST | 初始化组织（幂等） |

---

## Mock 开发策略（MSW + 页面级控制）

前端开发阶段使用 **MSW（Mock Service Worker）** 在网络层拦截请求，支持**页面级 mock 控制**——只有指定页面走 mock，其余页面调用真实 API，实现开发流程原子化。

### 页面级 mock 控制

通过环境变量 `NEXT_PUBLIC_MOCK_PAGES` 控制哪些页面启用 mock：

```bash
# .env.local
NEXT_PUBLIC_MOCK_PAGES=model-editor,enum-list
```

多个页面用逗号分隔。未列出的页面走真实 API，无需修改任何代码。

```ts
// src/mocks/page-mock-config.ts
import { isMockPage } from '@/mocks/page-mock-config'

isMockPage('model-editor')  // true / false
```

### 目录结构

```
src/mocks/
├── browser.ts               # 浏览器端 MSW worker 启动入口
├── node.ts                  # Node 端（测试）MSW server 启动入口
├── page-mock-config.ts      # 读取 NEXT_PUBLIC_MOCK_PAGES，导出 isMockPage()
├── MSWProvider.tsx          # React Provider，懒加载 MSW worker
│
├── handlers/
│   ├── index.ts             # 汇总 handlers（按 isMockPage 动态组合）
│   ├── model/
│   │   └── handlers.ts      # model 模块 mock handlers
│   ├── enum/
│   │   └── handlers.ts      # enum 模块 mock handlers
│   ├── end-user/
│   │   └── auth-handlers.ts # 终端用户认证 handlers（始终激活）
│   ├── project/
│   │   ├── rbac-handlers.ts # RBAC handlers
│   │   └── generated.ts     # codegen 自动生成，禁止手动编辑
│   └── org/
│       └── generated.ts     # codegen 自动生成，禁止手动编辑
│
└── data/
    ├── org/                 # Org 域 mock 数据工厂
    └── project/             # Project 域 mock 数据工厂
```

### handlers/index.ts 动态组合

```ts
// src/mocks/handlers/index.ts
import { isMockPage } from '../page-mock-config'
import { modelHandlers } from './model/handlers'
import { enumHandlers } from './enum/handlers'
import { rbacHandlers } from './project/rbac-handlers'

function buildHandlers() {
  const active = [...profileHandlers, ...endUserAuthHandlers]  // 始终激活

  if (isMockPage('model-editor')) active.push(...modelHandlers)
  if (isMockPage('enum-list') || isMockPage('enum-detail')) active.push(...enumHandlers)
  if (isMockPage('rbac')) active.push(...rbacHandlers)

  return active
}

export const handlers = buildHandlers()
```

### 页面 key 列表

| 页面 key | 对应 handlers | 说明 |
|----------|--------------|------|
| `model-editor` | modelHandlers | 模型编辑器 |
| `enum-list` | enumHandlers | 枚举列表页 |
| `enum-detail` | enumHandlers | 枚举详情页 |
| `rbac` | rbacHandlers | RBAC 权限管理 |

### MSW 启动控制

```tsx
// src/mocks/MSWProvider.tsx
// 由 NEXT_PUBLIC_API_MOCKING=enabled 控制是否启动 worker
```

```bash
NEXT_PUBLIC_API_MOCKING=enabled    # 开发阶段（启动 MSW worker）
NEXT_PUBLIC_MOCK_PAGES=model-editor,enum-list  # 只有这两个页面走 mock
```

### 开发 → 联调切换

```
开发阶段
  NEXT_PUBLIC_API_MOCKING=enabled
  NEXT_PUBLIC_MOCK_PAGES=model-editor
  model-editor 页面 → MSW 拦截 → mock 数据
  其他页面 → 真实 API

后端接口就绪，准备联调
  从 NEXT_PUBLIC_MOCK_PAGES 中移除该页面 key
  该页面请求 → 真实后端（零代码修改）
```

### 注意事项

- `src/mocks/handlers/*/generated.ts` 由 codegen 生成，**禁止手动编辑**
- mock 数据工厂放在 `src/mocks/data/`，使用 `@faker-js/faker` 生成随机数据
- 新增页面 mock 时：在 `handlers/` 下新建模块 `handlers.ts`，并在 `index.ts` 的 `buildHandlers()` 中注册对应的 `isMockPage('your-page-key')` 分支

---

## 参考文档

| 主题 | 文档 |
|------|------|
| 前端架构总览 | [architecture.md](./architecture.md) |
| ESLint 层边界规则 | [eslint-rules.md](./eslint-rules.md) |
| 代码规范 | [code-conventions.md](./code-conventions.md) |
