# Dual UI Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将前后端切换到双 UI 架构——租户端 `/tenant/login` + `/org/...`，用户端 `/end-user/[orgName]/workspace`，后端 auth 路径迁移至 `/api/tenant/auth/*`。

**Architecture:**
- 后端 OpenAPI `auth.yaml` 路径从 `/api/auth/*` 迁移到 `/api/tenant/auth/*`，重新生成代码
- 前端新增 `/tenant/login` 页面（复用现有登录组件），删除旧 `/login` 路由
- 前端 end-user 侧：删除 `select-project` 中间页，登录后统一跳转新建的 `/end-user/[orgName]/workspace` 页面

**Tech Stack:** Go/Chi (oapi-codegen), Next.js App Router, TypeScript, react-hook-form, shadcn/ui, Zustand

---

## 文件变更清单

### 后端

| 操作 | 文件 |
|------|------|
| Modify | `modelcraft-backend/api/openapi/auth.yaml` — 4 条路径 `/api/auth/*` → `/api/tenant/auth/*` |
| Modify | `modelcraft-backend/api/openapi/openapi.yaml` — 确认 $ref 无需变更（路径在 auth.yaml 中） |
| Regenerate | `modelcraft-backend/internal/interfaces/http/generated/server.gen.go` — `just generate-oapi` |
| Modify | `modelcraft-backend/internal/interfaces/http/chi_setup.go` — `publicPaths` 白名单 4 条路径更新 |

### 前端

| 操作 | 文件 |
|------|------|
| Create | `modelcraft-front/src/app/tenant/login/page.tsx` — 租户登录页（复用现有 LoginPage 逻辑） |
| Modify | `modelcraft-front/src/app/api/auth/[...path]/route.ts` — proxy 目标从 `/auth/*` 改为 `/api/tenant/auth/*` |
| Modify | `modelcraft-front/src/web/hooks/auth/use-auth-form.ts` — fetch 路径 `/api/auth/login` → `/api/tenant/auth/login`，register 同理 |
| Modify | `modelcraft-front/src/web/components/features/organization/user-menu.tsx` — logout fetch 路径更新 |
| Modify | `modelcraft-front/src/middleware.ts` — `DEV_PUBLIC_PATHS` 改为 `/tenant/login`、`/register`；end-user workspace 路径加入 protected |
| Create | `modelcraft-front/src/app/end-user/[orgName]/workspace/page.tsx` — Workspace 主页 |
| Create | `modelcraft-front/src/app/end-user/[orgName]/workspace/layout.tsx` — Workspace layout（顶部栏） |
| Create | `modelcraft-front/src/app/end-user/[orgName]/workspace/[projectSlug]/data/page.tsx` — 数据页占位（路由迁移） |
| Modify | `modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts` — 登录后跳 `/workspace`，去掉 select-project 分支 |
| Modify | `modelcraft-front/src/types/end-user-auth.ts` — 清理废弃的 select-project 类型 |
| Delete | `modelcraft-front/src/app/end-user/[orgName]/select-project/page.tsx` |
| Delete | `modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/select-project/route.ts` |
| Delete | `modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserProjectSelector.ts` |

---

## Task 1: 后端 — auth.yaml 路径迁移

**Files:**
- Modify: `modelcraft-backend/api/openapi/auth.yaml`

- [ ] **Step 1: 修改 auth.yaml 中的 4 条路径**

将文件中所有 `/api/auth/` 替换为 `/api/tenant/auth/`：

```yaml
# auth.yaml — 修改前 → 修改后

# /api/auth/register  →  /api/tenant/auth/register
# /api/auth/login     →  /api/tenant/auth/login
# /api/auth/logout    →  /api/tenant/auth/logout
# /api/auth/refresh   →  /api/tenant/auth/refresh
```

具体改动（文件 `modelcraft-backend/api/openapi/auth.yaml`，将 4 个顶层 path key 全部重命名）：

```yaml
paths:
  /api/tenant/auth/register:
    post:
      operationId: register
      # ... 以下内容不变
  /api/tenant/auth/login:
    post:
      operationId: login
      # ... 以下内容不变
  /api/tenant/auth/logout:
    post:
      operationId: logout
      # ... 以下内容不变
  /api/tenant/auth/refresh:
    post:
      operationId: refreshToken
      # ... 以下内容不变
```

- [ ] **Step 2: 重新生成后端代码**

```bash
cd modelcraft-backend
just generate-oapi
```

预期：`internal/interfaces/http/generated/server.gen.go` 中路由注册行变为：
```go
r.Post(options.BaseURL+"/api/tenant/auth/login", wrapper.Login)
r.Post(options.BaseURL+"/api/tenant/auth/register", wrapper.Register)
r.Post(options.BaseURL+"/api/tenant/auth/refresh", wrapper.RefreshToken)
r.Post(options.BaseURL+"/api/tenant/auth/logout", wrapper.Logout)
```

- [ ] **Step 3: 确认生成结果**

```bash
grep -n "tenant/auth\|/api/auth/" internal/interfaces/http/generated/server.gen.go
```

预期：只出现 `tenant/auth`，不再有 `/api/auth/`。

- [ ] **Step 4: 编译验证**

```bash
just build
```

预期：编译成功，0 error。

- [ ] **Step 5: Commit**

```bash
cd modelcraft-backend
git add api/openapi/auth.yaml internal/interfaces/http/generated/server.gen.go
git commit -m "feat: migrate tenant auth paths from /api/auth/* to /api/tenant/auth/*"
```

---

## Task 2: 后端 — chi_setup.go publicPaths 更新

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/chi_setup.go`

- [ ] **Step 1: 更新 publicPaths 白名单**

在 `chi_setup.go` 的 `conditionalAuthMiddleware` 函数中，找到 `publicPaths` map，将租户 auth 路径更新：

```go
publicPaths := map[string]bool{
    "/api/tenant/auth/register": true,  // 原 /api/auth/register
    "/api/tenant/auth/login":    true,  // 原 /api/auth/login
    "/api/tenant/auth/logout":   true,  // 原 /api/auth/logout
    "/api/tenant/auth/refresh":  true,  // 原 /api/auth/refresh
    // End-user auth: 不变
    "/api/end-user/auth/login":          true,
    "/api/end-user/auth/register":       true,
    "/api/end-user/auth/refresh":        true,
    "/api/end-user/auth/logout":         true,
    "/api/end-user/auth/me":             true,
    "/api/end-user/auth/select-project": true,
}
```

- [ ] **Step 2: 编译验证**

```bash
just build
```

预期：0 error。

- [ ] **Step 3: 启动服务做冒烟验证**

```bash
just run &
sleep 3
# 测试新路径可以访问（无 Token 不应 401，因为是 public）
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/tenant/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"test","password":"wrong"}'
# 预期：401（凭证错误）而非 404（路由不存在）
```

- [ ] **Step 4: 验证旧路径已失效**

```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"test","password":"wrong"}'
# 预期：404（路由不存在）
```

- [ ] **Step 5: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/http/chi_setup.go
git commit -m "fix: update publicPaths whitelist for /api/tenant/auth/* routes"
```

---

## Task 3: 前端 — BFF proxy 路径迁移

**Files:**
- Modify: `modelcraft-front/src/app/api/auth/[...path]/route.ts`

- [ ] **Step 1: 更新 BFF proxy 转发目标**

打开 `src/app/api/auth/[...path]/route.ts`，将 upstreamUrl 构造行从：

```typescript
const upstreamUrl = `${GATEWAY_URL}/auth/${pathSegments.join('/')}`
```

改为：

```typescript
const upstreamUrl = `${GATEWAY_URL}/api/tenant/auth/${pathSegments.join('/')}`
```

- [ ] **Step 2: 验证文件完整内容**

文件完整内容应为：

```typescript
/**
 * Auth Proxy Route — /api/auth/[...path]
 *
 * Forwards all tenant auth requests to the gateway /api/tenant/auth/*.
 * Transparently passes back Set-Cookie headers so mc_refresh_token is stored
 * under the localhost domain.
 */

import { NextRequest, NextResponse } from 'next/server'

const GATEWAY_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

async function handler(req: NextRequest, { params }: { params: { path: string[] } }) {
  const pathSegments = (await params).path
  const upstreamUrl = `${GATEWAY_URL}/api/tenant/auth/${pathSegments.join('/')}`

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')
  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const body = req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  const upstreamRes = await fetch(upstreamUrl, {
    method: req.method,
    headers,
    body,
  })

  const resBody = await upstreamRes.arrayBuffer()

  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
export const PUT = handler
export const PATCH = handler
export const DELETE = handler
```

- [ ] **Step 3: Commit**

```bash
cd modelcraft-front
git add src/app/api/auth/\[...path\]/route.ts
git commit -m "fix: update tenant auth BFF proxy target to /api/tenant/auth/*"
```

---

## Task 4: 前端 — middleware.ts 路由规则更新

**Files:**
- Modify: `modelcraft-front/src/middleware.ts`

当前 `middleware.ts` 有两处需要更新：
1. `DEV_PUBLIC_PATHS` 从 `['/login', '/register']` 改为 `['/tenant/login', '/register']`
2. end-user workspace 路径（`/end-user/[orgName]/workspace`）加入受保护路径

- [ ] **Step 1: 更新 DEV_PUBLIC_PATHS**

找到：
```typescript
const DEV_PUBLIC_PATHS = ['/login', '/register']
```

改为：
```typescript
const DEV_PUBLIC_PATHS = ['/tenant/login', '/register']
```

- [ ] **Step 2: 更新 END_USER_PUBLIC_SUFFIXES_RE**

当前公开路径正则包含 `login|select-project|no-project-access`，需改为只保留 `login`（workspace 是受保护路径）：

找到：
```typescript
const END_USER_PUBLIC_SUFFIXES_RE = /^\/end-user\/[^/]+\/(login|select-project|no-project-access)(\/.*)?$/
```

改为：
```typescript
const END_USER_PUBLIC_SUFFIXES_RE = /^\/end-user\/[^/]+\/login(\/.*)?$/
```

- [ ] **Step 3: 更新 END_USER_PROTECTED_RE**

当前保护路径匹配 `/end-user/{orgName}/{projectSlug}/*`，需扩展为也保护 workspace 路径：

找到：
```typescript
const END_USER_PROTECTED_RE = /^\/end-user\/([^/]+)\/([^/]+)(\/.*)?$/
```

这个正则本身可以匹配 `/end-user/acme/workspace`（workspace 作为第二段），**无需修改正则**。但 redirect 逻辑中 `orgName = match[1]` 依然正确。

验证：`/end-user/acme/workspace` 不匹配公开路径正则 → 进入保护路径检查 → 无 cookie → redirect 到 `/end-user/acme/login`。正确。

- [ ] **Step 4: 更新注释说明**

将文件顶部注释中 `select-project` 和 `no-project-access` 的说明更新为 `workspace`：

```typescript
/**
 * ...
 * End-User Auth:
 *  - All /end-user/* routes are handled separately before developer auth.
 *  - Public end-user paths (login) are allowed through.
 *  - Protected end-user paths (/end-user/[orgName]/workspace, /end-user/[orgName]/[projectSlug]/*)
 *    require the mc_enduser_refresh_token HttpOnly cookie.
 *    If missing, redirect to /end-user/[orgName]/login.
 */
```

- [ ] **Step 5: Commit**

```bash
cd modelcraft-front
git add src/middleware.ts
git commit -m "fix: update middleware routes for tenant/login and end-user workspace protection"
```

---

## Task 5: 前端 — 新建 /tenant/login 页面

**Files:**
- Create: `modelcraft-front/src/app/tenant/login/page.tsx`

租户登录页复用现有 `/login/page.tsx` 的全部 UI 和逻辑，仅调整页面标题和底部说明文字。

- [ ] **Step 1: 创建目录和文件**

```bash
mkdir -p modelcraft-front/src/app/tenant/login
```

- [ ] **Step 2: 创建 page.tsx**

```typescript
// src/app/tenant/login/page.tsx
// 租户端登录页。复用 /login 的登录表单逻辑，路由迁移至 /tenant/login。
import type { Metadata } from 'next'
import LoginPage from '@/app/login/page'

export const metadata: Metadata = {
  title: '租户登录 — ModelCraft',
  robots: {
    index: false,
    follow: false,
  },
}

// 复用 LoginPage 组件（username/password 表单，调用 /api/auth/login BFF）
export default LoginPage
```

> 注意：这里直接 re-export `LoginPage`。LoginPage 本身是 `'use client'` 组件，Server Component 直接导出给 App Router 即可。

- [ ] **Step 3: 验证页面可访问**

启动前端开发服务器：
```bash
cd modelcraft-front
npm run dev
```

访问 `http://localhost:3000/tenant/login`，预期：显示登录表单（手机号/用户名 Tab + 密码）。

- [ ] **Step 4: Commit**

```bash
cd modelcraft-front
git add src/app/tenant/login/page.tsx
git commit -m "feat: add /tenant/login route for tenant authentication"
```

---

## Task 6: 前端 — 删除旧 /login 页面，更新登录后跳转

**Files:**
- Modify: `modelcraft-front/src/web/hooks/auth/use-auth-form.ts`
- Modify: `modelcraft-front/src/web/components/features/organization/user-menu.tsx`

登录后的重定向目标：旧代码登录成功后跳 `/org/[orgName]/projects`（不变），但登录失败时跳回的是 `/login`，需确认无硬编码。

- [ ] **Step 1: 检查 use-auth-form.ts 中的重定向路径**

```bash
grep -n "router.push\|redirect\|/login" modelcraft-front/src/web/hooks/auth/use-auth-form.ts
```

找到所有 `/login` 引用，确认是否有登出后跳回 `/login` 的逻辑，如有则改为 `/tenant/login`。

- [ ] **Step 2: 检查 user-menu.tsx 中的 logout 逻辑**

```bash
grep -n "router.push\|redirect\|/login" modelcraft-front/src/web/components/features/organization/user-menu.tsx
```

当前 `user-menu.tsx` 中 logout 后会跳转登录页。将任何 `/login` 改为 `/tenant/login`：

找到（如存在）：
```typescript
router.push('/login')
```

改为：
```typescript
router.push('/tenant/login')
```

- [ ] **Step 3: 全局搜索其他 /login 硬编码**

```bash
grep -rn "router.push.*'/login'\|redirect.*'/login'\|href.*\"/login\"" \
  modelcraft-front/src \
  --include="*.ts" --include="*.tsx" \
  | grep -v node_modules | grep -v generated | grep -v end-user
```

对每条结果，将 `/login` 改为 `/tenant/login`。

- [ ] **Step 4: 中间件中的 loginUrl 更新**

`middleware.ts` 开发者路由 redirect 中：

找到：
```typescript
const loginUrl = new URL('/login', request.url)
```

改为：
```typescript
const loginUrl = new URL('/tenant/login', request.url)
```

- [ ] **Step 5: Commit**

```bash
cd modelcraft-front
git add src/middleware.ts src/web/hooks/auth/use-auth-form.ts \
  src/web/components/features/organization/user-menu.tsx
git commit -m "fix: update all /login redirects to /tenant/login"
```

---

## Task 7: 前端 — 新建 end-user Workspace 页面

**Files:**
- Create: `modelcraft-front/src/app/end-user/[orgName]/workspace/page.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/workspace/layout.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/workspace/_components/WorkspaceProjectsTab.tsx`

- [ ] **Step 1: 创建目录结构**

```bash
mkdir -p modelcraft-front/src/app/end-user/\[orgName\]/workspace/_components
```

- [ ] **Step 2: 创建 layout.tsx（顶部栏）**

```typescript
// src/app/end-user/[orgName]/workspace/layout.tsx
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Workspace',
  robots: { index: false, follow: false },
}

export default function WorkspaceLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
```

- [ ] **Step 3: 创建 WorkspaceProjectsTab 组件**

```typescript
// src/app/end-user/[orgName]/workspace/_components/WorkspaceProjectsTab.tsx
'use client'

import { useRouter } from 'next/navigation'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspaceProjectsTabProps {
  orgName: string
  projects: EndUserAccessibleProject[]
}

export function WorkspaceProjectsTab({ orgName, projects }: WorkspaceProjectsTabProps) {
  const router = useRouter()

  if (projects.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <span className="text-4xl mb-4">🔒</span>
        <p className="text-base font-medium text-foreground">您暂无项目访问权限</p>
        <p className="mt-2 text-sm text-muted-foreground">请联系管理员授权</p>
      </div>
    )
  }

  return (
    <div>
      <p className="mb-6 text-sm text-muted-foreground">选择要进入的项目</p>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <div
            key={project.slug}
            className="cursor-pointer rounded-lg border bg-background p-5 transition-shadow hover:shadow-md hover:border-primary/40"
            onClick={() => router.push(`/end-user/${orgName}/workspace/${project.slug}/data`)}
          >
            <p className="font-semibold text-foreground">{project.title}</p>
            <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
              {project.slug}
            </p>
            <button
              className="mt-4 text-sm font-medium text-primary hover:underline"
              onClick={(e) => {
                e.stopPropagation()
                router.push(`/end-user/${orgName}/workspace/${project.slug}/data`)
              }}
            >
              进入 →
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 创建 workspace page.tsx**

```typescript
// src/app/end-user/[orgName]/workspace/page.tsx
'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: Promise<{ orgName: string }>
}

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const router = useRouter()
  const [orgName, setOrgName] = useState('')
  const [projects, setProjects] = useState<EndUserAccessibleProject[]>([])
  const [activeTab, setActiveTab] = useState<'projects'>('projects')
  const accessToken = useEndUserAuthStore((s) => s.accessToken)

  useEffect(() => {
    params.then(({ orgName: name }) => {
      setOrgName(name)
      // 从 sessionStorage 读取登录时写入的 project 列表
      const raw = sessionStorage.getItem(`eu_accessible_projects_${name}`)
      if (raw) {
        try {
          setProjects(JSON.parse(raw) as EndUserAccessibleProject[])
        } catch {
          setProjects([])
        }
      }
    })
  }, [params])

  const handleLogout = async () => {
    await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    })
    router.push(`/end-user/${orgName}/login`)
  }

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      {/* 顶部栏 */}
      <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-background px-6">
        <span className="text-base font-semibold text-foreground">{orgName}</span>
        <div className="flex items-center gap-4">
          <button
            onClick={handleLogout}
            className="text-sm text-destructive hover:underline"
          >
            登出
          </button>
        </div>
      </header>

      {/* Tab 导航 */}
      <nav className="flex border-b bg-background px-6">
        <button
          className={`border-b-2 px-4 py-3 text-sm font-medium transition-colors ${
            activeTab === 'projects'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('projects')}
        >
          Projects
        </button>
        <button
          className="cursor-not-allowed border-b-2 border-transparent px-4 py-3 text-sm text-muted-foreground/50"
          disabled
          title="即将推出"
        >
          （待定）
        </button>
      </nav>

      {/* 主内容 */}
      <main className="flex-1 p-6">
        {activeTab === 'projects' && (
          <WorkspaceProjectsTab orgName={orgName} projects={projects} />
        )}
      </main>
    </div>
  )
}
```

- [ ] **Step 5: Commit**

```bash
cd modelcraft-front
git add \
  src/app/end-user/\[orgName\]/workspace/page.tsx \
  src/app/end-user/\[orgName\]/workspace/layout.tsx \
  src/app/end-user/\[orgName\]/workspace/_components/WorkspaceProjectsTab.tsx
git commit -m "feat: add end-user workspace page at /end-user/[orgName]/workspace"
```

---

## Task 8: 前端 — 迁移 data 页路由至 workspace 子路径

当前 data 页在 `/end-user/[orgName]/[projectSlug]/data`，新路由应为 `/end-user/[orgName]/workspace/[projectSlug]/data`。

**Files:**
- Create: `modelcraft-front/src/app/end-user/[orgName]/workspace/[projectSlug]/data/page.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/workspace/[projectSlug]/data/layout.tsx`

- [ ] **Step 1: 创建新路由目录**

```bash
mkdir -p "modelcraft-front/src/app/end-user/[orgName]/workspace/[projectSlug]/data"
```

- [ ] **Step 2: 创建 layout.tsx（直接复用现有 layout）**

```typescript
// src/app/end-user/[orgName]/workspace/[projectSlug]/data/layout.tsx
// 复用旧路由 layout 逻辑
export { default } from '@/app/end-user/[orgName]/[projectSlug]/data/layout'
```

- [ ] **Step 3: 创建 page.tsx（直接复用现有 page）**

```typescript
// src/app/end-user/[orgName]/workspace/[projectSlug]/data/page.tsx
// 复用旧路由 page 逻辑
export { default } from '@/app/end-user/[orgName]/[projectSlug]/data/page'
```

- [ ] **Step 4: 验证路由可访问**

启动开发服务器，访问 `http://localhost:3000/end-user/acme/workspace/sales-system/data`（需要有效 cookie）。预期：数据页正常渲染（或正确重定向到登录页）。

- [ ] **Step 5: Commit**

```bash
cd modelcraft-front
git add "src/app/end-user/[orgName]/workspace/[projectSlug]/data/"
git commit -m "feat: add data page at /end-user/[orgName]/workspace/[projectSlug]/data"
```

---

## Task 9: 前端 — 更新 login hook 跳转逻辑，删除 select-project

**Files:**
- Modify: `modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts`
- Delete: `modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserProjectSelector.ts`
- Delete: `modelcraft-front/src/app/end-user/[orgName]/select-project/page.tsx`
- Delete: `modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/select-project/route.ts`

- [ ] **Step 1: 更新 useEndUserOrgLoginForm.ts — 统一跳转 workspace**

找到登录成功后的跳转分支，删除 `select-project` 中间跳转，改为**无论多少 Project 均跳转 workspace**：

将原来的三分支逻辑：
```typescript
if (projects.length === 0) {
  router.push(`/end-user/${orgName}/no-project-access`)
  return
}
if (projects.length === 1) {
  // 直接跳数据页
  ...
  router.push(`/end-user/${orgName}/${projectSlug}/data`)
  return
}
// N 个 → select-project
sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
router.push(`/end-user/${orgName}/select-project`)
```

改为统一跳 workspace：
```typescript
// 无论 0/1/N 个 Project，均跳转 workspace
// workspace 页面自行读取 sessionStorage 中的 project 列表
sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
router.push(`/end-user/${orgName}/workspace`)
```

完整更新后的 `onSubmit` 成功路径：
```typescript
const projects: EndUserAccessibleProject[] = data.projects ?? []

// 存入 sessionStorage 供 workspace 页面使用（含空数组）
sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
router.push(`/end-user/${orgName}/workspace`)
```

- [ ] **Step 2: 删除 select-project 相关文件**

```bash
cd modelcraft-front
rm src/app/end-user/\[orgName\]/select-project/page.tsx
rm src/app/api/bff/org/\[orgName\]/end-user/auth/select-project/route.ts
rm src/web/hooks/end-user-auth-v2/useEndUserProjectSelector.ts
```

- [ ] **Step 3: 检查是否还有对 select-project 的引用**

```bash
grep -rn "select-project\|useEndUserProjectSelector\|EndUserProjectSelector" \
  modelcraft-front/src \
  --include="*.ts" --include="*.tsx" \
  | grep -v node_modules
```

预期：无结果（或只剩 mock handler，可暂留）。如有残留引用则逐一删除。

- [ ] **Step 4: Commit**

```bash
cd modelcraft-front
git add src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts
git rm src/app/end-user/\[orgName\]/select-project/page.tsx
git rm src/app/api/bff/org/\[orgName\]/end-user/auth/select-project/route.ts
git rm src/web/hooks/end-user-auth-v2/useEndUserProjectSelector.ts
git commit -m "refactor: remove select-project flow, all end-user logins redirect to workspace"
```

---

## Task 10: 前端 — 清理废弃类型和最终 lint 验证

**Files:**
- Modify: `modelcraft-front/src/types/end-user-auth.ts`

- [ ] **Step 1: 删除 select-project 相关类型**

在 `src/types/end-user-auth.ts` 中，删除以下类型定义：
- `EndUserSelectProjectRequest`
- `EndUserSelectProjectResponse`
- `EndUserSelectProjectError`
- `EndUserPendingSessionPayload`
- `EndUserLoginResponse`（旧 union 类型，已被新 hook 替代）

保留：`EndUserAccessibleProject`（workspace 页仍在使用）。

- [ ] **Step 2: 运行 lint**

```bash
cd modelcraft-front
npm run lint
```

预期：0 error。如有 lint 错误按提示修复。

- [ ] **Step 3: TypeScript 类型检查**

```bash
cd modelcraft-front
npx tsc --noEmit
```

预期：0 error。

- [ ] **Step 4: Commit**

```bash
cd modelcraft-front
git add src/types/end-user-auth.ts
git commit -m "chore: remove deprecated select-project types from end-user-auth.ts"
```

---

## Self-Review

### Spec 覆盖检查

| Spec 要求 | 对应 Task |
|-----------|-----------|
| 后端 `/api/auth/*` → `/api/tenant/auth/*` | Task 1, 2 |
| 前端 BFF proxy 路径更新 | Task 3 |
| `/tenant/login` 新页面 | Task 5 |
| 旧 `/login` 跳转更新为 `/tenant/login` | Task 6 |
| middleware 规则更新 | Task 4 |
| end-user workspace 页面 | Task 7 |
| data 页路由迁移至 workspace 子路径 | Task 8 |
| select-project 流程删除，统一跳 workspace | Task 9 |
| 废弃类型清理 | Task 10 |

### 潜在风险点

1. **Task 8 re-export**：`export { default } from '...'` 在 Next.js App Router 中对 `'use client'` 组件有效，但 Server Component 需要确认。如果旧 data/page.tsx 是 Server Component，改为直接 import 并 re-render。

2. **`/end-user/[orgName]/no-project-access` 路由**：Task 9 移除了跳转该页的逻辑，但页面文件还存在。可暂留，不影响流程；后续单独清理。

3. **sessionStorage 在 workspace 刷新后失效**：workspace 页面刷新后 sessionStorage 可能丢失 project 列表。v1 接受此行为（刷新后显示空状态），后续可在 workspace 加 GraphQL 查询补充。
