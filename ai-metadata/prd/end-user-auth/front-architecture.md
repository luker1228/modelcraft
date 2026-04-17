# 终端用户认证 — 前端架构设计

> **所属模块**：终端用户认证（End-User Auth）
> **父文档**：[00-end-user-auth.md](./00-end-user-auth.md)
> **类型**：架构设计文档（只做研究和设计，不直接产出业务代码）
> **日期**：2026-04-16

---

## 设计原则

本模块遵循「**最大对称，最小侵入**」原则：

1. **对称现有开发者认证**：BFF 文件、Store、Hooks 命名和结构完全对称现有 `auth/` 体系，以 `end-user-` 为前缀区分
2. **零侵入开发者逻辑**：middleware 追加分支，不修改任何开发者 auth 文件
3. **Cookie 硬隔离**：`end_user_refresh_token` vs `refresh_token`，path 绑定到具体 Project
4. **GraphQL 共用 Apollo Client**：开发者侧用户管理直接复用现有 `@apollo/client` 体系，新增 operations 文件即可

---

## 1. 目录结构规划

### 全景概览

```
src/
├── app/
│   ├── api/bff/
│   │   └── end-user/
│   │       ├── auth/
│   │       │   ├── login/
│   │       │   │   └── route.ts          # 终端用户登录 BFF Handler
│   │       │   ├── register/
│   │       │   │   └── route.ts          # 终端用户注册 BFF Handler
│   │       │   ├── logout/
│   │       │   │   └── route.ts          # 终端用户登出 BFF Handler
│   │       │   ├── refresh/
│   │       │   │   └── route.ts          # Token 刷新 BFF Handler
│   │       │   └── me/
│   │       │       └── route.ts          # 获取当前用户信息 BFF Handler
│   │
│   └── org/[orgName]/projects/[projectSlug]/
│       ├── user/
│       │   └── login/
│       │       └── page.tsx              # 终端用户登录页（公开，无需 auth 守卫）
│       ├── data/
│       │   ├── layout.tsx                # 数据管理页布局（守卫：end_user_refresh_token）
│       │   └── page.tsx                  # 数据管理落地页（占位）
│       └── end-users/
│           ├── layout.tsx                # 用户管理页布局（守卫：refresh_token，开发者专属）
│           └── page.tsx                  # 开发者管理终端用户账号
│
├── bff/
│   └── end-user/                         # 对称 bff/auth/
│       ├── end-user-go-client.ts         # 对称 go-auth-client.ts
│       ├── end-user-cookie-utils.ts      # 对称 cookie-utils.ts
│       ├── end-user-jwt-utils.ts         # 对称 jwt-utils.ts（含 role claim）
│       └── end-user-auth-client.ts       # 对称 auth-client.ts（client-side）
│
├── shared/
│   └── stores/
│       └── end-user-auth-store.ts        # 对称 auth-store.ts
│
├── web/
│   ├── hooks/
│   │   └── end-user-auth/               # 对称 hooks/auth/
│   │       ├── use-end-user-auth.ts      # 对称 use-auth.ts（useRequireEndUserAuth / useEndUser）
│   │       └── use-end-user-form.ts      # 对称 use-auth-form.ts（登录表单 Hook）
│   ├── graphql/
│   │   ├── queries/
│   │   │   └── end-user.ts              # GraphQL listEndUsers query
│   │   └── mutations/
│   │       └── end-user.ts              # GraphQL createEndUser / updateEndUserStatus / deleteEndUser
│   └── components/
│       └── features/
│           ├── end-user-auth/           # 终端用户登录业务组件
│           │   ├── EndUserLoginForm.tsx
│           │   └── EndUserLoginLayout.tsx
│           └── end-users/              # 开发者用户管理业务组件
│               ├── EndUsersTable.tsx
│               ├── CreateEndUserDialog.tsx
│               └── EndUserStatusBadge.tsx
│
└── types/
    └── end-user-auth.ts                  # 终端用户认证相关类型定义
```

### 文件职责说明

#### `src/app/` — 路由层

| 文件 | 职责 |
|------|------|
| `org/.../user/login/page.tsx` | **公开路由**。终端用户登录页 Server Component 外壳 + 注入 `orgName`/`projectSlug` 参数。无任何 auth 守卫，SEO 设置 `robots: noindex`。 |
| `org/.../data/layout.tsx` | **终端用户路由守卫**。Client Component，调用 `useRequireEndUserAuth()`，无 cookie 时重定向到 `./user/login`。 |
| `org/.../data/page.tsx` | 数据管理落地页，v1 可为占位页面。 |
| `org/.../end-users/layout.tsx` | **开发者路由守卫**，复用现有 `useRequireAuth()`，确认开发者 cookie 存在。 |
| `org/.../end-users/page.tsx` | 开发者管理终端用户账号，渲染 `EndUsersTable` + `CreateEndUserDialog`，通过 Apollo Client 调用 GraphQL。 |
| `api/bff/end-user/auth/login/route.ts` | BFF POST，接收 `{ orgName, projectSlug, username, password }`，调用 Go 内网 `/internal/end-user/auth/login`，BFF 自签 end-user access token，写 `end_user_refresh_token` Cookie。 |
| `api/bff/end-user/auth/register/route.ts` | BFF POST，接收 `{ orgName, projectSlug, username, password }`，调用 Go 内网 `/internal/end-user/auth/register`，行为同 login（注册后自动登录）。 |
| `api/bff/end-user/auth/logout/route.ts` | BFF POST，读 Cookie → 调 Go revoke → 清 Cookie。 |
| `api/bff/end-user/auth/refresh/route.ts` | BFF POST，读 Cookie → 调 Go token rotation → 重写 Cookie → 返回新 access token。 |
| `api/bff/end-user/auth/me/route.ts` | BFF GET，验证 end-user access token → 解析 payload → 以 `X-End-User-Id` 等 Header 调 Go `/internal/end-user/auth/me` → 返回 `{ id, username, createdAt }`。供数据管理页右上角展示用户名。 |

#### `src/bff/end-user/` — BFF 服务层

| 文件 | 职责 |
|------|------|
| `end-user-go-client.ts` | 封装调用 Go 内网认证接口的 fetch 函数，不依赖 Cookie（服务端）。定义错误类型 `EndUserAuthError`、`EndUserConflictError` 等。 |
| `end-user-cookie-utils.ts` | `set/get/clearEndUserRefreshTokenCookie()`，key = `end_user_refresh_token`，path 绑定到 `/org/{orgName}/projects/{projectSlug}`。 |
| `end-user-jwt-utils.ts` | `signEndUserAccessToken(userId, orgName, projectSlug)`，payload 含 `role: "end_user"` claim，issuer = `modelcraft-end-user`。 |
| `end-user-auth-client.ts` | 客户端工具函数：`getEndUserToken()`、`refreshEndUserAccessToken()`、`isEndUserAuthenticated()`，对称 `auth-client.ts`，内部调用 `useEndUserAuthStore`。 |

#### `src/shared/stores/` — 状态层

| 文件 | 职责 |
|------|------|
| `end-user-auth-store.ts` | Zustand store，存储终端用户 access token 和用户信息，对称 `auth-store.ts`。 |

#### `src/web/` — UI 层

| 文件 | 职责 |
|------|------|
| `hooks/end-user-auth/use-end-user-auth.ts` | `useRequireEndUserAuth()`（页面级守卫 hook）、`useEndUser()`（获取当前终端用户信息 + logout）。 |
| `hooks/end-user-auth/use-end-user-form.ts` | 终端用户登录表单 hook，封装 `react-hook-form` + `zod` 校验 + BFF 调用 + 错误映射。 |
| `graphql/queries/end-user.ts` | `LIST_END_USERS` GraphQL query document。 |
| `graphql/mutations/end-user.ts` | `CREATE_END_USER`、`UPDATE_END_USER_STATUS`、`DELETE_END_USER` mutation documents。 |
| `components/features/end-user-auth/EndUserLoginForm.tsx` | 纯 UI 登录表单组件，受控表单，接收外部 hook props。 |
| `components/features/end-user-auth/EndUserLoginLayout.tsx` | 终端用户登录页专属 Layout，无开发者导航栏，显示 Project 名称 + Logo。 |
| `components/features/end-users/EndUsersTable.tsx` | 用户列表表格，使用 `useListEndUsersQuery` hook，含搜索、分页、操作列。 |
| `components/features/end-users/CreateEndUserDialog.tsx` | 创建用户 Modal，使用 `useCreateEndUserMutation` hook，含字段校验。 |
| `components/features/end-users/EndUserStatusBadge.tsx` | 状态 Badge（启用/禁用），纯 UI 组件。 |

---

## 2. BFF 层设计

### 2.1 `end-user-go-client.ts`

对称 `go-auth-client.ts`，调用 Go Backend 内网 `/internal/end-user/` 接口。

**结果类型**：

```typescript
export interface EndUserLoginResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserRegisterResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserRefreshResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserMeResult {
  id: string
  username: string
  createdAt: string
}
```

**错误类型（对称 `go-auth-client.ts` 错误体系）**：

```typescript
// 凭证错误（用户名/密码错误，或用户不存在）
export class EndUserInvalidCredentialsError extends Error { ... }

// 账号已被禁用
export class EndUserAccountDisabledError extends Error { ... }

// 用户名冲突（注册/创建时）
export class EndUserConflictError extends Error { ... }

// Token 无效或已过期（含 reuse 攻击检测）
export class EndUserTokenError extends Error { ... }

// Project 未关联 DatabaseCluster
export class EndUserClusterNotConfiguredError extends Error { ... }

// 参数校验失败
export class EndUserParamInvalidError extends Error { ... }
```

**核心函数**：

```typescript
// 所有调用均携带 X-Internal-Token Header
export async function callGoEndUserLogin(params: {
  orgName: string; projectSlug: string; username: string; password: string
}): Promise<EndUserLoginResult>

export async function callGoEndUserRegister(params: {
  orgName: string; projectSlug: string; username: string; password: string
}): Promise<EndUserRegisterResult>

export async function callGoEndUserRefresh(params: {
  orgName: string; projectSlug: string; refreshToken: string
}): Promise<EndUserRefreshResult>

export async function callGoEndUserLogout(params: {
  orgName: string; projectSlug: string; refreshToken: string
}): Promise<void>  // best-effort，catch 静默失败

// BFF 已验证 JWT，传 userId/orgName/projectSlug 给 Go（X-End-User-Id 等 Header）
export async function callGoEndUserMe(params: {
  orgName: string; projectSlug: string; userId: string
}): Promise<EndUserMeResult>
```

**错误解析规则**：Go Backend 统一错误格式 `{ error: { code, message } }`，错误码映射：

| Go 错误码 | 抛出错误类 |
|-----------|-----------|
| `INVALID_CREDENTIALS` | `EndUserInvalidCredentialsError` |
| `ACCOUNT_DISABLED` | `EndUserAccountDisabledError` |
| `CONFLICT` | `EndUserConflictError` |
| `INVALID_REFRESH_TOKEN` | `EndUserTokenError` |
| `CLUSTER_NOT_CONFIGURED` | `EndUserClusterNotConfiguredError` |
| `PARAM_INVALID` | `EndUserParamInvalidError` |

### 2.2 `end-user-cookie-utils.ts`

对称 `cookie-utils.ts`，key = `end_user_refresh_token`，**Cookie path 绑定到 Project 级路由**。

```typescript
const COOKIE_NAME = 'end_user_refresh_token'
const COOKIE_MAX_AGE = 7 * 24 * 60 * 60  // 7 days

// Cookie path 绑定到具体 Project，防止不同 Project 的 token 互相读取
// 格式：/org/{orgName}/projects/{projectSlug}
function getProjectCookiePath(orgName: string, projectSlug: string): string

// 写入 Cookie
export function setEndUserRefreshTokenCookie(
  token: string,
  orgName: string,
  projectSlug: string,
): void

// 读取 Cookie（用于 BFF Route Handler 内读取）
export function getEndUserRefreshTokenFromCookie(): string | undefined

// 清除 Cookie（需传 path 才能精确清除）
export function clearEndUserRefreshTokenCookie(
  orgName: string,
  projectSlug: string,
): void
```

> **与 `cookie-utils.ts` 的关键差异**：Cookie path 从 `/`（全局）改为 `/org/{orgName}/projects/{projectSlug}`（Project 级），实现不同 Project 的终端用户 session 硬隔离。

### 2.3 `end-user-jwt-utils.ts`

对称 `jwt-utils.ts`，payload 额外携带角色和项目上下文：

```typescript
// issuer 与开发者不同，天然隔离
const ISSUER = 'modelcraft-end-user'  // 开发者 issuer = 'modelcraft'
const EXPIRY = '1h'

export interface EndUserTokenPayload {
  userId: string
  orgName: string
  projectSlug: string
}

// 签发：payload 含 sub(userId) + org_name + project_slug + role="end_user"
export async function signEndUserAccessToken(
  payload: EndUserTokenPayload,
): Promise<string>

// 验证：检查 issuer，防止开发者 token 被用于终端用户场景
export async function verifyEndUserAccessToken(
  token: string,
): Promise<EndUserTokenPayload>
```

### 2.4 Route Handler 请求/响应类型定义

```typescript
// BFF 登录/注册请求（前端 → BFF）
interface EndUserLoginRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

// BFF 登录/注册响应（BFF → 前端）
// 同时写 end_user_refresh_token HttpOnly Cookie
interface EndUserAuthResponse {
  accessToken: string    // BFF 自签的 end-user JWT（1h）
  expiresIn: number      // 3600
}

// BFF /me 响应
interface EndUserMeResponse {
  id: string
  username: string
  createdAt: string      // ISO 8601
}

// BFF 错误响应（统一格式）
interface EndUserBffError {
  error: {
    code: string
    message: string
  }
}
```

---

## 3. Middleware 扩展方案

在现有 `middleware.ts` 基础上，**追加终端用户路由判断分支**，不修改任何现有开发者逻辑。

```typescript
// src/middleware.ts（扩展后完整伪代码）
import { NextRequest, NextResponse } from 'next/server'

// ============================================
// 开发者认证配置（现有，完整保留，不变）
// ============================================
const DEVELOPER_PUBLIC_PATHS = ['/login', '/register', '/auth/callback']
const DEVELOPER_COOKIE = 'refresh_token'

// ============================================
// 终端用户认证配置（新增）
// ============================================
const END_USER_COOKIE = 'end_user_refresh_token'

/**
 * 判断路径是否属于终端用户数据路由（需要守卫）。
 * 匹配：/org/{orgName}/projects/{projectSlug}/data 及其子路径
 */
function isEndUserDataRoute(pathname: string): boolean {
  return /^\/org\/[^/]+\/projects\/[^/]+\/data(\/.*)?$/.test(pathname)
}

/**
 * 判断路径是否为终端用户登录页（公开，无需守卫）。
 * 匹配：/org/{orgName}/projects/{projectSlug}/user/login
 */
function isEndUserLoginPage(pathname: string): boolean {
  return /^\/org\/[^/]+\/projects\/[^/]+\/user\/login$/.test(pathname)
}

/**
 * 从路径中提取 orgName 和 projectSlug，用于构造重定向 URL。
 */
function extractProjectParams(
  pathname: string,
): { orgName: string; projectSlug: string } | null {
  const match = pathname.match(/^\/org\/([^/]+)\/projects\/([^/]+)/)
  if (!match) return null
  return { orgName: match[1], projectSlug: match[2] }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // 1. 放行所有 /api/* 路由（BFF endpoints，无需守卫）
  if (pathname.startsWith('/api/')) {
    return NextResponse.next()
  }

  // ============================================
  // 2. 终端用户路由守卫（新增分支，在开发者守卫之前判断）
  // ============================================

  // 2a. 终端用户登录页本身：公开，直接放行
  if (isEndUserLoginPage(pathname)) {
    return NextResponse.next()
  }

  // 2b. 终端用户数据管理路由：需要 end_user_refresh_token cookie
  if (isEndUserDataRoute(pathname)) {
    const hasEndUserToken = request.cookies.has(END_USER_COOKIE)
    console.log(`[middleware] end-user route ${pathname} — cookie present: ${hasEndUserToken}`)

    if (!hasEndUserToken) {
      const params = extractProjectParams(pathname)
      if (params) {
        const loginUrl = new URL(
          `/org/${params.orgName}/projects/${params.projectSlug}/user/login`,
          request.url,
        )
        loginUrl.searchParams.set('redirect', pathname)  // 注意：终端用户用 redirect，开发者用 returnUrl
        return NextResponse.redirect(loginUrl)
      }
    }

    return NextResponse.next()
  }

  // ============================================
  // 3. 开发者路由守卫（现有逻辑，完整保留，不做任何修改）
  // ============================================

  // 放行开发者公开页
  if (DEVELOPER_PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next()
  }

  // 检查开发者 refresh_token cookie
  const hasRefreshToken = request.cookies.has(DEVELOPER_COOKIE)
  console.log(`[middleware] ${pathname} — cookie present: ${hasRefreshToken}`)

  if (!hasRefreshToken) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('returnUrl', pathname)
    console.log(`[middleware] No refresh token, redirecting to: ${loginUrl.toString()}`)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  // 与现有 matcher 相同，不变
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
```

**关键设计决策**：

| 决策 | 理由 |
|------|------|
| 新分支在开发者守卫**之前**判断 | 保证 `/data` 路由优先走终端用户逻辑，不会被开发者守卫误捕获 |
| 使用正则匹配（而非 `startsWith`） | 精确匹配 project 路径段，防止误匹配 `/org/x/projects/` 等中间路径 |
| 重定向参数用 `?redirect=`（区别于开发者的 `?returnUrl=`） | 两套体系参数名隔离，防止混淆 |
| middleware 只检测 cookie 存在，不验证 JWT | 与现有开发者策略一致，实际 token 验证由客户端 silent refresh 完成 |
| 不提取 orgName/projectSlug 做路由 log | 避免路径正则复杂度，保持与现有 log 风格一致 |

---

## 4. Store 设计

对称 `auth-store.ts`，设计 `end-user-auth-store.ts`：

```typescript
// src/shared/stores/end-user-auth-store.ts
import { create } from 'zustand'

/** JWT payload 中解析出来的终端用户信息（部分字段来自 /me 接口） */
export interface EndUserInfo {
  id: string           // userId（JWT sub claim）
  username: string     // 从 /me 接口获取后填充（JWT 中无此字段）
  orgName: string      // 来自 JWT org_name claim
  projectSlug: string  // 来自 JWT project_slug claim
}

interface EndUserAuthState {
  // Token 状态
  accessToken: string | null
  expiresAt: number | null       // Unix timestamp（毫秒）

  // 用户信息（login 后从 /me 接口获取，可延迟填充）
  userInfo: EndUserInfo | null

  // Actions
  setAccessToken: (token: string, expiresIn: number) => void
  setUserInfo: (info: EndUserInfo) => void
  clearSession: () => void        // 同时清 token 和 userInfo
  isTokenExpired: () => boolean
}

export const useEndUserAuthStore = create<EndUserAuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,
  userInfo: null,

  setAccessToken: (token: string, expiresIn: number) => {
    set({
      accessToken: token,
      expiresAt: Date.now() + expiresIn * 1000,
    })
  },

  setUserInfo: (info: EndUserInfo) => set({ userInfo: info }),

  clearSession: () => set({ accessToken: null, expiresAt: null, userInfo: null }),

  isTokenExpired: () => {
    const { expiresAt } = get()
    if (!expiresAt) return true
    // 提前 5 分钟触发 silent refresh（与开发者策略一致）
    return Date.now() > expiresAt - 5 * 60 * 1000
  },
}))
```

**与 `auth-store.ts` 的差异**：

| 对比项 | `auth-store.ts` | `end-user-auth-store.ts` |
|--------|-----------------|--------------------------|
| export 名 | `useAuthStore` | `useEndUserAuthStore` |
| 用户信息 | 无（从 token decode） | 含 `userInfo`（username 不在 JWT 中，需从 `/me` 填充） |
| 清除方法 | `clearAccessToken()` | `clearSession()`（同时清 token 和 userInfo） |
| JWT payload | `user_id`/`phone`/`name` | `sub`/`org_name`/`project_slug`/`role` |

> ⚠️ **Store 无持久化**：页面刷新后 `accessToken` 和 `userInfo` 均为 null。`useRequireEndUserAuth()` 会自动用 Cookie 中的 `end_user_refresh_token` 执行 silent refresh 恢复 token；`userInfo` 在 silent refresh 成功后通过调用 `/api/bff/end-user/auth/me` 重新填充。

---

## 4.1 `end-user-auth-client.ts`（client-side 工具）

对称 `bff/auth/auth-client.ts`，提供客户端侧工具函数。

```typescript
// src/bff/end-user/end-user-auth-client.ts
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import type { EndUserInfo } from '@shared/stores/end-user-auth-store'

export type { EndUserInfo }

/** 从 in-memory store 获取 end-user access token */
export function getEndUserToken(): string | null {
  return useEndUserAuthStore.getState().accessToken
}

/** 清除 end-user session（token + userInfo） */
export function removeEndUserSession(): void {
  useEndUserAuthStore.getState().clearSession()
}

// 防并发刷新
let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

/**
 * 使用 Cookie 中的 end_user_refresh_token 做 silent refresh。
 * 并发调用共享同一个请求。
 */
export async function refreshEndUserAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) return _refreshPromise

  _isRefreshing = true
  _refreshPromise = (async () => {
    try {
      const res = await fetch('/api/bff/end-user/auth/refresh', {
        method: 'POST',
        credentials: 'same-origin',
      })
      if (!res.ok) {
        useEndUserAuthStore.getState().clearSession()
        return null
      }
      const data = (await res.json()) as { accessToken?: string; expiresIn?: number }
      if (data.accessToken && data.expiresIn) {
        useEndUserAuthStore.getState().setAccessToken(data.accessToken, data.expiresIn)
        return data.accessToken
      }
      return null
    } catch {
      return null
    } finally {
      _isRefreshing = false
      _refreshPromise = null
    }
  })()

  return _refreshPromise
}

/**
 * 获取终端用户信息并填充 store.userInfo。
 * 在 silent refresh 成功后调用，确保 username 可用于右上角展示。
 */
export async function fetchAndCacheEndUserInfo(): Promise<EndUserInfo | null> {
  try {
    const res = await fetch('/api/bff/end-user/auth/me', {
      credentials: 'same-origin',
    })
    if (!res.ok) return null
    const data = (await res.json()) as { id: string; username: string; createdAt: string }
    // 从当前 token 的 store 中读取 org/project 上下文
    const store = useEndUserAuthStore.getState()
    // org/project 从 JWT payload 解析（decodeJWT 与开发者侧同款工具）
    const info: EndUserInfo = {
      id: data.id,
      username: data.username,
      orgName: store.userInfo?.orgName ?? '',
      projectSlug: store.userInfo?.projectSlug ?? '',
    }
    store.setUserInfo(info)
    return info
  } catch {
    return null
  }
}

/** 检查是否已认证（token 存在且未过期） */
export function isEndUserAuthenticated(): boolean {
  const { accessToken, isTokenExpired } = useEndUserAuthStore.getState()
  return !!accessToken && !isTokenExpired()
}
```

---

## 5. 页面组件分层

### 5.1 终端用户登录页

**路由**：`/org/[orgName]/projects/[projectSlug]/user/login`

#### 组件树

```
page.tsx                              (Server Component — 注入 params 为 orgName/projectSlug)
└── EndUserLoginLayout                (Client Component — 独立 layout，无开发者导航)
    ├── ProjectBranding               (显示 Project 名称，从 params 读取；v2 可调 API 获取 Logo)
    └── EndUserLoginForm              (Client Component — 核心表单)
        ├── ErrorBanner               (表单顶部错误提示，条件显示)
        ├── FormField: username       (shadcn/ui Input，autocomplete="username")
        ├── FormField: password       (PasswordInput，含明文切换按钮，autocomplete="current-password")
        └── SubmitButton              (shadcn/ui Button，loading 状态下禁用 + Loader2 图标)
```

#### 使用的 Hooks

| Hook | 来源 | 用途 |
|------|------|------|
| `useEndUserLoginForm(orgName, projectSlug)` | `src/web/hooks/end-user-auth/use-end-user-form.ts` | 表单状态、提交逻辑、错误映射、登录后跳转 |
| `useRouter` | `next/navigation` | 登录成功后 `router.replace()` 跳转 |
| `useSearchParams` | `next/navigation` | 读取 `?redirect=` 参数，登录成功后回跳 |

#### `useEndUserLoginForm` 核心逻辑（伪代码）

```typescript
// src/web/hooks/end-user-auth/use-end-user-form.ts

export function useEndUserLoginForm(orgName: string, projectSlug: string) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<EndUserLoginFormValues>({
    resolver: zodResolver(endUserLoginSchema),  // username: min(1) + max(64), password: min(1) + max(128)
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = async (values: EndUserLoginFormValues) => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/bff/end-user/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgName, projectSlug, ...values }),
        credentials: 'same-origin',
      })
      if (!res.ok) {
        const data = await res.json()
        setError(mapEndUserErrorCode(data.error?.code, res.status))
        form.resetField('password')   // PRD 要求：错误后清空密码字段
        form.setFocus('password')     // PRD 要求：自动聚焦密码字段
        return
      }
      const { accessToken, expiresIn } = await res.json() as EndUserAuthResponse
      useEndUserAuthStore.getState().setAccessToken(accessToken, expiresIn)

      // 异步填充 userInfo（不阻塞跳转）：登录后立即从 /me 接口获取 username，
      // 供数据管理页右上角展示；失败不影响主流程。
      void fetchAndCacheEndUserInfo()

      // 跳转：优先使用 ?redirect= 参数，否则跳转数据管理落地页
      const redirect = searchParams.get('redirect')
      const target = redirect ?? `/org/${orgName}/projects/${projectSlug}/data`
      router.replace(target)
    } catch {
      setError('登录服务暂时不可用，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }

  return { form, onSubmit, isLoading, error }
}
```

#### `EndUserLoginLayout` 隔离要求

- **不**引入 `AppLayout`、开发者 `AppSidebar` 等组件
- 参考 `src/web/components/features/auth/auth-layout.tsx` 结构，但独立实现
- 使用居中 card 布局，只展示 ProjectBranding + 登录表单
- Server Component 层从 `params` 获取 `orgName`/`projectSlug` 并以 props 传入

#### `data/layout.tsx` — 终端用户守卫（对称 `projects/[projectSlug]/layout.tsx`）

#### `use-end-user-auth.ts` 实现

```typescript
// src/web/hooks/end-user-auth/use-end-user-auth.ts
'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import {
  refreshEndUserAccessToken,
  fetchAndCacheEndUserInfo,
  removeEndUserSession,
} from '@bff/end-user/end-user-auth-client'
import type { EndUserInfo } from '@shared/stores/end-user-auth-store'

/**
 * 页面级守卫 hook。
 * middleware 已保证 end_user_refresh_token cookie 存在才能进入此 layout，
 * 但页面刷新后 in-memory token 会丢失，需要 silent refresh 恢复。
 */
export function useRequireEndUserAuth() {
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()
  const params = useParams<{ orgName: string; projectSlug: string }>()

  useEffect(() => {
    async function restoreSession() {
      const { accessToken, isTokenExpired } = useEndUserAuthStore.getState()

      // token 存在且未过期：直接放行
      if (accessToken && !isTokenExpired()) {
        // userInfo 可能因页面刷新丢失，异步补充（不阻塞渲染）
        if (!useEndUserAuthStore.getState().userInfo) {
          void fetchAndCacheEndUserInfo()
        }
        setIsLoading(false)
        return
      }

      // token 不存在或已过期：尝试 silent refresh
      const newToken = await refreshEndUserAccessToken()
      if (!newToken) {
        // refresh 失败（cookie 过期/revoked），重定向到登录页
        const loginUrl = `/org/${params.orgName}/projects/${params.projectSlug}/user/login`
        router.replace(`${loginUrl}?redirect=${encodeURIComponent(window.location.pathname)}`)
        return
      }

      // refresh 成功：填充 userInfo 供右上角展示
      void fetchAndCacheEndUserInfo()
      setIsLoading(false)
    }

    restoreSession()
  }, [])

  return { isLoading }
}

/**
 * 获取当前终端用户信息 + logout。
 * 不含守卫逻辑，适合右上角 UserMenu 等非守卫场景。
 */
export function useEndUser(): { user: EndUserInfo | null; logout: () => Promise<void> } {
  const userInfo = useEndUserAuthStore((s) => s.userInfo)
  const router = useRouter()
  const params = useParams<{ orgName: string; projectSlug: string }>()

  const logout = async () => {
    await fetch('/api/bff/end-user/auth/logout', {
      method: 'POST',
      credentials: 'same-origin',
    }).catch(() => {})  // best-effort
    removeEndUserSession()
    router.replace(`/org/${params.orgName}/projects/${params.projectSlug}/user/login`)
  }

  return { user: userInfo, logout }
}
```

```typescript
// src/app/org/[orgName]/projects/[projectSlug]/data/layout.tsx
'use client'

export default function DataLayout({ children }: { children: React.ReactNode }) {
  const { isLoading } = useRequireEndUserAuth()  // 对称 useRequireAuth()
  const params = useParams()

  if (isLoading) return <LoadingScreen message="Authenticating..." />

  return (
    // 使用终端用户专属布局（无项目配置导航），v1 可先用简单 wrapper
    <div>{children}</div>
  )
}
```

### 5.2 用户管理页（`/end-users`）

**路由**：`/org/[orgName]/projects/[projectSlug]/end-users`
**守卫**：开发者 `refresh_token`（**直接复用** `ProjectLayout` 和现有 `useRequireAuth()`，无需修改）

#### 组件树

```
end-users/
├── layout.tsx         (不需要新建，end-users 目录在 [projectSlug] 下，自动继承 ProjectLayout 守卫)
└── page.tsx           (Client Component)
    └── EndUsersView
        ├── EndUsersHeader
        │   ├── <h1>用户管理</h1>
        │   └── CreateEndUserButton    (onClick → setIsCreateOpen(true))
        ├── EndUsersSearch             (受控 Input，onChange → setSearch → refetch)
        ├── EndUsersTable              (使用 useListEndUsersQuery)
        │   ├── <thead>用户名 / 状态 / 创建时间 / 操作</thead>
        │   └── <tbody> (nodes.map)
        │       ├── EndUserStatusBadge (isForbidden ? "禁用" : "启用")
        │       └── EndUserActions
        │           ├── DisableButton  → DisableConfirmDialog → useUpdateEndUserStatusMutation
        │           ├── EnableButton   → useUpdateEndUserStatusMutation (无确认)
        │           └── DeleteButton   → DeleteConfirmDialog → useDeleteEndUserMutation
        └── CreateEndUserDialog        (isOpen 受 page.tsx state 控制)
            ├── FormField: username
            ├── FormField: password
            ├── FormField: confirmPassword
            └── SubmitButton           (useCreateEndUserMutation)
```

#### GraphQL Hook 使用对照

| UI 操作 | 使用的 Hook | 触发后行为 |
|---------|------------|-----------|
| 页面加载 / 搜索 | `useListEndUsersQuery` | 展示用户列表，error union 显示相应错误 Toast |
| 点击「+ 创建用户」→ 提交 | `useCreateEndUserMutation` | 成功：关闭 Modal + `refetch()` + Toast「创建成功」；`EndUserAlreadyExists` → Modal 内错误提示 |
| 点击「禁用」→ 确认 | `useUpdateEndUserStatusMutation({ isForbidden: true })` | 成功：`refetch()` |
| 点击「启用」 | `useUpdateEndUserStatusMutation({ isForbidden: false })` | 成功：`refetch()` |
| 点击「删除」→ 确认 | `useDeleteEndUserMutation` | 成功：`refetch()` + Toast「删除成功」 |

---

## 6. GraphQL Codegen

### 6.1 新增 GraphQL Operations 文件

#### `src/web/graphql/queries/end-user.ts`

```typescript
import { gql } from '@apollo/client'

export const LIST_END_USERS = gql`
  query ListEndUsers($input: ListEndUsersInput) {
    listEndUsers(input: $input) {
      connection {
        nodes {
          id
          username
          isForbidden
          createdBy
          createdAt
          updatedAt
        }
        pageInfo {
          hasNextPage
          hasPreviousPage
          startCursor
          endCursor
        }
        totalCount
      }
      error {
        ... on ClusterNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
```

#### `src/web/graphql/mutations/end-user.ts`

```typescript
import { gql } from '@apollo/client'

export const CREATE_END_USER = gql`
  mutation CreateEndUser($input: CreateEndUserInput!) {
    createEndUser(input: $input) {
      endUser {
        id
        username
        isForbidden
        createdAt
      }
      error {
        ... on EndUserAlreadyExists { message }
        ... on EndUserPasswordTooWeak { message suggestion }
        ... on ClusterNotFound { message }
        ... on InvalidInput { message }
        ... on ProjectNotFound { message }
      }
    }
  }
`

export const UPDATE_END_USER_STATUS = gql`
  mutation UpdateEndUserStatus($input: UpdateEndUserStatusInput!) {
    updateEndUserStatus(input: $input) {
      endUser {
        id
        isForbidden
      }
      error {
        ... on EndUserNotFound { message }
        ... on ClusterNotFound { message }
        ... on InvalidInput { message }
        ... on ProjectNotFound { message }
      }
    }
  }
`

export const DELETE_END_USER = gql`
  mutation DeleteEndUser($input: DeleteEndUserInput!) {
    deleteEndUser(input: $input) {
      success
      error {
        ... on EndUserNotFound { message }
        ... on ClusterNotFound { message }
        ... on ProjectNotFound { message }
      }
    }
  }
`
```

### 6.2 Codegen 生成的 Hooks 命名

| GraphQL Operation | 生成 Hook | 类型 |
|-------------------|-----------|------|
| `query ListEndUsers` | `useListEndUsersQuery` | `QueryHookOptions<ListEndUsersQuery, ListEndUsersQueryVariables>` |
| `mutation CreateEndUser` | `useCreateEndUserMutation` | `MutationHookOptions<CreateEndUserMutation, CreateEndUserMutationVariables>` |
| `mutation UpdateEndUserStatus` | `useUpdateEndUserStatusMutation` | `MutationHookOptions<...>` |
| `mutation DeleteEndUser` | `useDeleteEndUserMutation` | `MutationHookOptions<...>` |

> Codegen 运行命令与现有一致，参考 `ai-metadata/backend/development/contract-sync.md` 中的前端 codegen 工作流。后端 Schema 新增 `end_user.graphql` 并 push 后，前端执行 `front-contract-pull` skill 同步，再运行 codegen。

---

## 7. 类型定义

集中在 `src/types/end-user-auth.ts`，不污染现有 `src/types/auth.ts`：

```typescript
// src/types/end-user-auth.ts

// ============================================================
// BFF 请求/响应类型（前端 ↔ BFF Route Handler）
// ============================================================

export interface EndUserLoginRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

export interface EndUserRegisterRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

/** login 和 register 统一响应格式 */
export interface EndUserAuthResponse {
  accessToken: string    // BFF 自签 end-user JWT（1h）
  expiresIn: number      // 3600
}

export interface EndUserMeResponse {
  id: string
  username: string
  createdAt: string      // ISO 8601
}

// ============================================================
// Go Backend 内网请求/响应类型（BFF 内部使用）
// ============================================================

export interface GoEndUserLoginResponse {
  userId: string
  refreshToken: string
  expiresAt: string      // ISO 8601
}

export interface GoEndUserRefreshResponse {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface GoEndUserMeResponse {
  id: string
  username: string
  createdAt: string
}

// ============================================================
// Go Backend 错误结构（BFF 内部解析用）
// ============================================================

export interface GoEndUserError {
  error: {
    code: EndUserErrorCode
    message: string
  }
}

// ============================================================
// 错误码
// ============================================================

export type EndUserErrorCode =
  | 'INVALID_CREDENTIALS'        // 用户名/密码错误（防枚举，统一返回）
  | 'ACCOUNT_DISABLED'           // 账号已禁用
  | 'CONFLICT'                   // 用户名已存在
  | 'INVALID_REFRESH_TOKEN'      // refresh token 无效/过期/已 revoke
  | 'CLUSTER_NOT_CONFIGURED'     // Project 未关联 Cluster
  | 'PARAM_INVALID'              // 参数校验失败
  | 'UNAUTHORIZED'               // JWT 缺失或无效

/** 错误码 → 用户提示文案（PRD §错误提示规则） */
export function mapEndUserErrorCode(
  code: string | undefined,
  httpStatus: number,
): string {
  if (code === 'INVALID_CREDENTIALS') return '用户名或密码错误，请重试'
  if (code === 'ACCOUNT_DISABLED')    return '该账号已被禁用，请联系管理员'
  if (code === 'CLUSTER_NOT_CONFIGURED') return '服务暂时不可用，请联系管理员'
  if (httpStatus >= 500)              return '登录服务暂时不可用，请稍后重试'
  return '登录服务暂时不可用，请稍后重试'
}

// ============================================================
// JWT Payload 类型（客户端 decode 用，仅供展示）
// ============================================================

export interface EndUserJWTPayload {
  sub: string             // userId
  org_name: string
  project_slug: string
  role: 'end_user'
  exp?: number
  iat?: number
}

// ============================================================
// 表单 Schema（Zod）— 注释形式，在 shared/validation/ 中实现
// ============================================================
//
// endUserLoginSchema:
//   username: z.string().min(1, '请输入用户名').max(64)
//   password: z.string().min(1, '请输入密码').max(128)
//
// createEndUserSchema:
//   username: z.string().min(3).max(64).regex(/^[a-zA-Z0-9_-]+$/)
//   password: z.string().min(8).regex(/^(?=.*[a-zA-Z])(?=.*\d)/)
//   confirmPassword: z.string() + .refine(password === confirmPassword)
```

---

## 8. BFF Mock 接口定义

前端开发阶段，在 `NODE_ENV=development` 且后端未就绪时，通过 Route Handler 内联 mock 或 `msw` 拦截返回以下数据。

### POST `/api/bff/end-user/auth/login`

| 场景 | 状态码 | 响应体 |
|------|--------|--------|
| 成功 | 200 | `{ "accessToken": "eyJ...", "expiresIn": 3600 }` |
| 凭证错误 | 401 | `{ "error": { "code": "INVALID_CREDENTIALS", "message": "用户名或密码错误" } }` |
| 账号禁用 | 403 | `{ "error": { "code": "ACCOUNT_DISABLED", "message": "该账号已被禁用" } }` |
| Cluster 未配置 | 503 | `{ "error": { "code": "CLUSTER_NOT_CONFIGURED", "message": "服务暂时不可用" } }` |

### POST `/api/bff/end-user/auth/register`

| 场景 | 状态码 | 响应体 |
|------|--------|--------|
| 成功 | 200 | `{ "accessToken": "eyJ...", "expiresIn": 3600 }` |
| 用户名冲突 | 409 | `{ "error": { "code": "CONFLICT", "message": "该用户名已被使用" } }` |
| 参数无效 | 400 | `{ "error": { "code": "PARAM_INVALID", "message": "密码强度不足" } }` |

### POST `/api/bff/end-user/auth/logout`

| 场景 | 状态码 | 响应体 |
|------|--------|--------|
| 成功 | 204 | 无 body，同时清除 `end_user_refresh_token` Cookie |

### POST `/api/bff/end-user/auth/refresh`

| 场景 | 状态码 | 响应体 |
|------|--------|--------|
| 成功 | 200 | `{ "accessToken": "eyJ...(new)", "expiresIn": 3600 }` |
| Cookie 不存在 / token 无效 | 401 | `{ "error": { "code": "INVALID_REFRESH_TOKEN", "message": "..." } }` |

### GET `/api/bff/end-user/auth/me`

| 场景 | 状态码 | 响应体 |
|------|--------|--------|
| 成功 | 200 | `{ "id": "550e8400-...", "username": "alice", "createdAt": "2026-04-10T08:00:00Z" }` |
| JWT 缺失/无效 | 401 | `{ "error": { "code": "UNAUTHORIZED", "message": "..." } }` |
| 账号禁用 | 403 | `{ "error": { "code": "ACCOUNT_DISABLED", "message": "..." } }` |

### GraphQL `listEndUsers` Mock

```json
{
  "data": {
    "listEndUsers": {
      "connection": {
        "nodes": [
          {
            "id": "user-001",
            "username": "alice",
            "isForbidden": false,
            "createdBy": "dev-user-42",
            "createdAt": "2026-04-10T08:00:00Z",
            "updatedAt": "2026-04-10T08:00:00Z"
          },
          {
            "id": "user-002",
            "username": "bob",
            "isForbidden": true,
            "createdBy": "dev-user-42",
            "createdAt": "2026-04-12T10:30:00Z",
            "updatedAt": "2026-04-14T09:00:00Z"
          }
        ],
        "pageInfo": {
          "hasNextPage": false,
          "hasPreviousPage": false,
          "startCursor": "cursor-001",
          "endCursor": "cursor-002"
        },
        "totalCount": 2
      },
      "error": null
    }
  }
}
```

### GraphQL `createEndUser` Mock

**成功**：
```json
{
  "data": {
    "createEndUser": {
      "endUser": { "id": "user-003", "username": "charlie", "isForbidden": false, "createdAt": "2026-04-16T12:00:00Z" },
      "error": null
    }
  }
}
```

**用户名冲突**：
```json
{
  "data": {
    "createEndUser": {
      "endUser": null,
      "error": { "__typename": "EndUserAlreadyExists", "message": "该用户名已被使用" }
    }
  }
}
```

---

## 9. 环境变量

新增以下环境变量（在 `.env.local` 中配置，与现有变量并列）：

| 变量名 | 示例值 | 说明 |
|--------|--------|------|
| `INTERNAL_SERVICE_TOKEN` | `secret-internal-token` | BFF 调用 Go 内网接口的共享密钥，通过 `X-Internal-Token` Header 携带 |
| `JWT_SECRET` | 已有 | end-user JWT 复用同一个 secret，通过 `issuer: 'modelcraft-end-user'` 与开发者 JWT 隔离 |

> 如需完全独立的 secret，可新增 `END_USER_JWT_SECRET` 并在 `end-user-jwt-utils.ts` 中使用。

---

## 10. 实现顺序建议

| 阶段 | 任务 | 前置依赖 |
|------|------|----------|
| P0-1 | `end-user-cookie-utils.ts` + `end-user-go-client.ts` + `end-user-jwt-utils.ts` | 无 |
| P0-2 | BFF Route Handlers（login / register / refresh / logout / me） | P0-1 |
| P0-3 | `end-user-auth-store.ts` + `end-user-auth-client.ts` | P0-2 |
| P0-4 | middleware 扩展（追加终端用户守卫分支） | P0-1 |
| P0-5 | `use-end-user-form.ts` + 登录页组件 + `user/login/page.tsx` | P0-2, P0-3 |
| P0-6 | `data/layout.tsx`（`useRequireEndUserAuth` 守卫） | P0-3, P0-4 |
| P1-1 | GraphQL operations 文件（前端部分） + Codegen（后端 Schema 就绪后） | 后端 `end_user.graphql` push |
| P1-2 | `EndUsersTable` + `CreateEndUserDialog` + `end-users/page.tsx` | P1-1 |

---

## 附录：关键文件对称映射

| 开发者认证（现有） | 终端用户认证（新增） | 核心差异 |
|-------------------|---------------------|----------|
| `bff/auth/go-auth-client.ts` | `bff/end-user/end-user-go-client.ts` | 调用路径 `/internal/end-user/auth/*`，携带 `X-Internal-Token`，错误码集合不同 |
| `bff/auth/cookie-utils.ts` | `bff/end-user/end-user-cookie-utils.ts` | key=`end_user_refresh_token`，path 绑定 Project（而非全局 `/`） |
| `bff/auth/jwt-utils.ts` | `bff/end-user/end-user-jwt-utils.ts` | issuer=`modelcraft-end-user`，payload 含 `role`/`org_name`/`project_slug` |
| `bff/auth/auth-client.ts` | `bff/end-user/end-user-auth-client.ts` | 额外 `fetchAndCacheEndUserInfo()`，调用 `/api/bff/end-user/auth/refresh` |
| `shared/stores/auth-store.ts` | `shared/stores/end-user-auth-store.ts` | 额外 `userInfo` 字段；清除方法为 `clearSession()`；无持久化，刷新后由 silent refresh 恢复 |
| `web/hooks/auth/use-auth.ts` | `web/hooks/end-user-auth/use-end-user-auth.ts` | 守卫跳转到 `./user/login`；silent refresh 后自动调 `/me` 填充 `userInfo` |
| `web/hooks/auth/use-auth-form.ts` | `web/hooks/end-user-auth/use-end-user-form.ts` | 只有用户名+密码（无手机号/用户名切换 Tab）；登录后异步填充 `userInfo` |
| `app/login/page.tsx` | `app/org/.../user/login/page.tsx` | 独立路由、独立 Layout、`robots: noindex` |
| `app/api/bff/auth/login/route.ts` | `app/api/bff/end-user/auth/login/route.ts` | 调用内网路径不同，cookie key 不同 |
| `app/api/bff/auth/register/route.ts` | `app/api/bff/end-user/auth/register/route.ts` | 终端用户自注册（username+password，无手机号），注册即登录 |
| `app/api/bff/auth/refresh/route.ts` | `app/api/bff/end-user/auth/refresh/route.ts` | 读 `end_user_refresh_token` cookie |
| —（无对等） | `app/api/bff/end-user/auth/me/route.ts` | 新增；验证 JWT → 调 Go `/internal/end-user/auth/me` → 返回 username 供右上角展示 |
| `middleware: DEVELOPER_COOKIE` | `middleware: END_USER_COOKIE` | 同一 middleware.ts 中独立分支，新分支先于开发者守卫执行；重定向参数用 `?redirect=` |
