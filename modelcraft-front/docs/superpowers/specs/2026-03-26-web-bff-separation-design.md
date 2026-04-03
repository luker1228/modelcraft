# Web/BFF 分离架构设计

**日期**: 2026-03-26  
**状态**: 已确认  
**范围**: 同仓库目录分离，模块边界约束，不改任何功能

---

## 背景与动机

ModelCraft 前端是 Next.js 14 项目，目前 Web 层（React 组件/hooks）和 BFF 层（API Routes、Apollo 客户端初始化、认证）混合在 `src/lib/` 和 `src/app/` 中，没有强制边界。随着项目增长，两层代码互相依赖的风险增大。

**目标**：
- 清晰区分 BFF 层（服务端运行逻辑）和 Web 层（客户端交互逻辑）
- 用工具强制依赖方向，避免 Web 层直接 import BFF 内部实现
- 迁移成本尽量低，不改任何功能，不调整 GraphQL 接口

---

## 当前结构问题

```
src/
├── app/
│   ├── api/          ← BFF（API Routes，纯代理）
│   └── (routes)/     ← Web 页面（全部 use client）
├── lib/              ← 混合层（问题所在）
│   ├── apollo-clients.ts     # Apollo 客户端工厂（BFF 逻辑）
│   ├── apollo-wrapper.tsx    # ApolloProvider（Web 逻辑，待拆分）
│   ├── memberships-cache.ts  # 三层缓存（Web 逻辑，待重分类）
│   ├── auth/casdoor.ts       # OAuth（BFF 逻辑）
│   ├── auth/token-utils.ts   # Token 工具（BFF 逻辑）
│   ├── cms/runtime-query-builder.ts  # 动态查询（BFF 逻辑）
│   ├── cms/schema-transformer.ts     # 纯转换（shared）
│   ├── cms/field-linkage.ts          # 字段联动（Web 逻辑）
│   ├── cms/validation.ts             # 纯验证（shared）
│   ├── routing/smart-redirect.ts     # 客户端重定向（Web 逻辑）
│   ├── query-wrapper.tsx             # TanStack Query Provider（Web 逻辑）
│   ├── theme-colors.ts               # 主题配置（shared）
│   ├── typography.ts                 # 排版配置（shared）
│   ├── organization-name-validator.ts # 名称验证（shared）
│   ├── organization-name-generator.ts # 名称生成（shared）
│   ├── utils.ts                      # 通用工具（mixed，待审查）
│   └── uuid-utils.ts                 # UUID 工具（shared）
├── components/       ← Web 层
├── hooks/            ← Web 层
├── stores/           ← Web 层
└── graphql/          ← Web 层
```

`lib/` 混合了三类代码（BFF 服务端逻辑、Web 客户端交互、shared 纯工具），没有工具约束，开发者无法快速判断某个模块属于哪一层。

---

## 目标结构

```
src/
├── bff/                            ← BFF 层（服务端运行，不依赖 React）
│   ├── api/                        ← 原 app/api/ 内容迁移
│   │   ├── auth/
│   │   ├── org/
│   │   ├── user/
│   │   └── copilotkit/
│   ├── apollo/                     ← Apollo 客户端配置（纯 JS，无 JSX）
│   │   ├── clients.ts              ← 原 lib/apollo-clients.ts
│   │   └── links.ts                ← ErrorLink/AuthLink 拆分出来
│   ├── auth/                       ← 认证相关
│   │   ├── casdoor.ts              ← 原 lib/auth/casdoor.ts
│   │   └── token-utils.ts          ← 原 lib/auth/token-utils.ts
│   └── cms/
│       └── runtime-query-builder.ts ← 原 lib/cms/runtime-query-builder.ts
│
├── web/                            ← Web 层（客户端运行，可依赖 React）
│   ├── components/                 ← 原 components/
│   ├── hooks/                      ← 原 hooks/
│   ├── stores/                     ← 原 stores/
│   ├── graphql/                    ← 原 graphql/
│   ├── providers/
│   │   ├── apollo-wrapper.tsx      ← 原 lib/apollo-wrapper.tsx（React Provider 部分）
│   │   └── query-wrapper.tsx       ← 原 lib/query-wrapper.tsx
│   ├── cache/
│   │   └── memberships-cache.ts    ← 原 lib/memberships-cache.ts（客户端三层缓存）
│   ├── cms/
│   │   └── field-linkage.ts        ← 原 lib/cms/field-linkage.ts
│   └── routing/
│       └── smart-redirect.ts       ← 原 lib/routing/smart-redirect.ts
│
├── shared/                         ← 两层共用（纯函数/类型/常量，无副作用，无环境依赖）
│   ├── cms/
│   │   ├── schema-transformer.ts   ← 原 lib/cms/schema-transformer.ts
│   │   └── validation.ts           ← 原 lib/cms/validation.ts
│   ├── utils/
│   │   └── uuid.ts                 ← 原 lib/uuid-utils.ts
│   ├── theme-colors.ts             ← 原 lib/theme-colors.ts
│   ├── typography.ts               ← 原 lib/typography.ts
│   ├── organization-name-validator.ts ← 原 lib/organization-name-validator.ts
│   ├── organization-name-generator.ts ← 原 lib/organization-name-generator.ts
│   └── utils.ts                    ← 原 lib/utils.ts（审查后确认为纯函数）
│
└── app/                            ← Next.js 路由（只做路由编排）
    ├── api/                        ← 路由文件保留，内部 import @bff/
    └── (routes)/                   ← import from @web/*
```

---

## 模块归属判断标准

| 归属 | 判断依据 |
|------|---------|
| `@bff/` | 在 Node.js 服务端运行；涉及 Token、网络代理、Apollo 初始化；不依赖 React |
| `@web/` | 在浏览器客户端运行；依赖 React hooks/组件；处理 UI 交互逻辑 |
| `@shared/` | 纯函数，无副作用，无环境依赖；可在服务端和客户端共用 |

### `lib/` 拆分明细

| 原文件 | 新位置 | 理由 |
|--------|--------|------|
| `apollo-clients.ts` | `@bff/apollo/clients.ts` | Apollo 客户端初始化，无 React 依赖 |
| `apollo-wrapper.tsx` | `@web/providers/apollo-wrapper.tsx` | React Provider 组件，依赖 JSX/hooks |
| `auth/casdoor.ts` | `@bff/auth/casdoor.ts` | OAuth 认证，服务端逻辑 |
| `auth/token-utils.ts` | `@bff/auth/token-utils.ts` | Token 管理，服务端逻辑 |
| `cms/runtime-query-builder.ts` | `@bff/cms/runtime-query-builder.ts` | 动态构建 GraphQL query |
| `memberships-cache.ts` | `@web/cache/memberships-cache.ts` | 三层缓存含 localStorage，运行在客户端 |
| `cms/schema-transformer.ts` | `@shared/cms/schema-transformer.ts` | 纯数据转换 |
| `cms/field-linkage.ts` | `@web/cms/field-linkage.ts` | 字段联动，UI 交互逻辑 |
| `cms/validation.ts` | `@shared/cms/validation.ts` | 纯验证逻辑 |
| `routing/smart-redirect.ts` | `@web/routing/smart-redirect.ts` | 客户端重定向 |
| `query-wrapper.tsx` | `@web/providers/query-wrapper.tsx` | TanStack Query Provider |
| `typography.ts` | `@shared/typography.ts` | 纯配置常量 |
| `theme-colors.ts` | `@shared/theme-colors.ts` | 纯配置常量 |
| `organization-name-validator.ts` | `@shared/organization-name-validator.ts` | 纯验证 |
| `organization-name-generator.ts` | `@shared/organization-name-generator.ts` | 纯字符串生成 |
| `utils.ts` | `@shared/utils.ts` | 纯工具函数（迁移前需确认无副作用） |
| `uuid-utils.ts` | `@shared/utils/uuid.ts` | 纯工具函数 |

---

## 边界约束实现

### 1. tsconfig.json paths

在现有 `@/*` alias 基础上追加，保持向后兼容：

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@bff/*": ["./src/bff/*"],
      "@web/*": ["./src/web/*"],
      "@shared/*": ["./src/shared/*"]
    }
  }
}
```

### 2. ESLint depguard 规则

安装依赖：
```bash
pnpm add -D eslint-plugin-depend
```

在 `.eslintrc.cjs` 中添加：
```javascript
module.exports = {
  plugins: ['depend'],
  rules: {
    'depend/depguard': ['warn', {  // 初期 warn，迁移完成后改为 error
      rules: [
        {
          selector: 'src/web/**',
          deny: ['src/bff'],
          message: 'Web 层不能直接 import BFF 层内部实现',
        },
        {
          selector: 'src/bff/**',
          deny: ['src/web'],
          message: 'BFF 层不能依赖 Web 层',
        },
      ],
    }],
  },
}
```

### 3. 依赖方向

```
app/routes  ──import──>  @web/*
                              │
                        @shared/*
                              │
app/api     ──import──>  @bff/*
```

Web 层和 BFF 层只能通过 `@shared/` 共享类型，不能互相依赖。

---

## app/api 路由处理

Next.js 要求 API Routes 必须在 `app/api/` 目录下。采用**方式一（推荐）**：

`app/api/` 文件只做路由注册（不超过 5 行），业务逻辑全部在 `@bff/` 中：

```typescript
// src/app/api/auth/token/route.ts（保持原位）
import { POST } from '@bff/api/auth/token'
export { POST }
```

| 方式 | 优点 | 缺点 |
|------|------|------|
| **方式一（本方案）** | 保持 Next.js 原生路由结构；迁移量最小；易调试 | 需要 re-export 中间层 |
| 方式二（移动文件） | 边界更清晰 | 需要 symlink 或构建脚本，增加复杂性 |

---

## 迁移策略

### 原则
- 不改任何功能，只移动文件和调整 import 路径
- 分批迁移，每批可独立上线，可独立回滚
- 先建目录结构和 alias，再逐步移文件

### 分批顺序

**第一批：基础设施（风险最低）**
1. 新建 `src/bff/`、`src/web/`、`src/shared/` 目录（含占位 `.gitkeep`）
2. 更新 `tsconfig.json` 新增 path alias（保留原有 `@/*`）
3. 安装 `eslint-plugin-depend`，配置 depguard 为 `warn` 模式
4. 运行 `pnpm lint` 确认无构建错误

**第二批：shared 层迁移**
- `cms/schema-transformer.ts` → `@shared/cms/`
- `cms/validation.ts` → `@shared/cms/`
- `typography.ts` → `@shared/`
- `theme-colors.ts` → `@shared/`
- `organization-name-validator.ts` → `@shared/`
- `organization-name-generator.ts` → `@shared/`
- `uuid-utils.ts` → `@shared/utils/`
- `utils.ts` → `@shared/`（迁移前确认无副作用）

**第三批：BFF 层迁移**
- `apollo-clients.ts` → `@bff/apollo/clients.ts`（并拆出 `links.ts`）
- `auth/` → `@bff/auth/`
- `cms/runtime-query-builder.ts` → `@bff/cms/`
- `app/api/` 各 route.ts 内部逻辑 → `@bff/api/`，route.ts 保留为 re-export

**第四批：Web 层迁移**
- `apollo-wrapper.tsx` → `@web/providers/apollo-wrapper.tsx`
- `query-wrapper.tsx` → `@web/providers/query-wrapper.tsx`
- `memberships-cache.ts` → `@web/cache/`
- `components/` → `@web/components/`
- `hooks/` → `@web/hooks/`
- `stores/` → `@web/stores/`
- `graphql/` → `@web/graphql/`
- `cms/field-linkage.ts` → `@web/cms/`
- `routing/smart-redirect.ts` → `@web/routing/`

**第五批：收尾**
- 将 depguard 从 `warn` 改为 `error`
- 运行 `pnpm lint` 确认 0 errors
- 删除原 `src/lib/` 目录
- 更新所有剩余 `@/lib/*` import 路径

### ESLint warn → error 切换时机

满足以下条件后执行切换：
- [ ] `src/lib/` 目录已清空
- [ ] `pnpm lint`（warn 模式）输出 0 errors
- [ ] `pnpm build` 正常
- [ ] 手动回归验证登录、组织切换、模型编辑功能正常

---

## 风险与回退策略

### 主要风险

| 风险 | 应对 |
|------|------|
| import 路径更新遗漏导致 TS 报错 | TypeScript 编译会立即暴露，每批迁移后运行 `pnpm type-check` |
| `utils.ts` 含有副作用被放入 shared | 迁移前先读文件，确认无浏览器 API/网络调用 |
| depguard 规则误报合理的跨层 import | 在 eslintrc 中添加 ignore 注释 + PR 描述说明原因 |
| Apollo Provider 拆分引入的 React 层级问题 | `apollo-wrapper.tsx` 迁移后运行回归测试，确认 Provider 包裹顺序 |

### 回退策略

迁移**只改文件位置和 import 路径，无功能变更**，任意一批可以独立 `git revert`：

```bash
# 回退某一批迁移
git revert <commit-hash>
```

由于没有数据结构变更，回退不存在数据风险。

---

## 不在本次范围内

- 不修改任何 GraphQL Schema 或接口
- 不改变 Apollo Client 的运行逻辑
- 不引入 monorepo 或独立 package
- 不做服务端渲染（SSR）改造
- 不改 next.config.mjs 的代理配置

---

## 成功标准

1. `src/lib/` 目录被完全删除
2. ESLint depguard 规则以 `error` 模式运行，CI 通过
3. `pnpm build` 正常，无 TypeScript 类型错误
4. 手动回归验证通过：
   - [ ] 登录/OAuth 回调流程正常
   - [ ] 组织切换（Apollo cache reset）正常
   - [ ] 项目列表、模型编辑器正常加载
   - [ ] 数据查询/变更（运行态 GraphQL）正常
   - [ ] CopilotKit AI 功能正常
