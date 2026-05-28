# Unified User System — Plan 3: 前端双登录页

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现前端双登录页：`/tenant/login` 独立管理员登录页（不再跳转）、`/end-user/[orgName]/login` 支持 `is_admin` 跳转管理后台、统一 cookie 名（`mc_enduser_refresh_token` → `mc_refresh_token`）、更新中间件。

**Architecture:** 最小变更原则——复用现有 `useLogin` hook 实现 `/tenant/login`，在 `useEndUserOrgLoginForm` 加 `is_admin` 分支，将所有 `mc_enduser_refresh_token` 引用改为 `mc_refresh_token`，更新中间件保护 end-user 路由使用统一 cookie。

**Tech Stack:** Next.js 15, React, Zustand, TypeScript

**当前状态：**
- `/tenant/login` — 重定向到 `/login`（待改为独立页面）
- `/end-user/[orgName]/login` — 已存在，成功后跳 workspace（无 is_admin 分支）
- `mc_enduser_refresh_cookie` — 单独 cookie（待合并为 `mc_refresh_token`）
- 中间件 — end-user 检查 `mc_enduser_refresh_token`（待改为 `mc_refresh_token`）

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **修改** | `src/app/tenant/login/page.tsx` — 从 redirect 改为真实登录页 |
| **修改** | `src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts` — 加 is_admin 跳转分支 |
| **修改** | `src/middleware.ts` — `END_USER_REFRESH_COOKIE` 改为 `mc_refresh_token` |
| **修改** | `src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts` — logout 清除 `mc_refresh_token` |
| **修改** | `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts` — 读取 `mc_refresh_token` |

---

## Task 1: `/tenant/login` 改为独立管理员登录页

**Files:**
- Modify: `modelcraft-front/src/app/tenant/login/page.tsx`

当前内容只是 `redirect('/login')`。改为真实登录页（与 `/login` 相同 UI，但用户名登录 tab 在前）。

- [ ] **Step 1: 读取当前 tenant/login/page.tsx 和 login/page.tsx**

```bash
cat modelcraft-front/src/app/tenant/login/page.tsx
cat modelcraft-front/src/app/login/page.tsx
```

- [ ] **Step 2: 将 tenant/login/page.tsx 替换为独立登录页**

复制 `/login/page.tsx` 的内容，改为：
- 默认 tab 为 `USERNAME`（管理员用用户名登录）
- 去掉 "还没有账号？立即注册" 链接（管理员由平台分配，无自助注册入口从这里进）
- 标题改为 "管理员登录"

```tsx
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { loginFormSchema, type LoginFormValues } from '@/shared/validation/auth'
import { useLogin } from '@/web/hooks/auth/use-auth-form'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { PasswordInput } from '@/web/components/common/password-input'
import { Tabs, TabsList, TabsTrigger } from '@web/components/ui/tabs'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@web/components/ui/form'
import type { IdentifierType } from '@/types/auth'

export default function TenantLoginPage() {
  const { login, isLoading, error, identifierType, setIdentifierType } = useLogin()

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: { identifier: '', identifierType: 'USERNAME', password: '' },
  })

  const handleTabChange = (value: string) => {
    const type = value as IdentifierType
    setIdentifierType(type)
    form.setValue('identifierType', type)
    form.setValue('identifier', '')
    form.clearErrors('identifier')
  }

  const handleSubmit = form.handleSubmit(async (values) => {
    await login(values)
  })

  return (
    <AuthLayout title="管理员登录" subtitle="登录管理控制台">
      <Form {...form}>
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          {error && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <Tabs
            value={identifierType}
            onValueChange={handleTabChange}
            className="w-full"
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="USERNAME">用户名登录</TabsTrigger>
              <TabsTrigger value="PHONE">手机号登录</TabsTrigger>
            </TabsList>
          </Tabs>

          <FormField
            control={form.control}
            name="identifier"
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  {identifierType === 'PHONE' ? '手机号' : '用户名'}
                </FormLabel>
                <FormControl>
                  <Input
                    placeholder={
                      identifierType === 'PHONE' ? '请输入手机号' : '请输入用户名'
                    }
                    autoComplete={identifierType === 'PHONE' ? 'tel' : 'username'}
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem>
                <FormLabel>密码</FormLabel>
                <FormControl>
                  <PasswordInput
                    placeholder="请输入密码"
                    autoComplete="current-password"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button type="submit" className="mt-1 h-10 w-full" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
            登录
          </Button>
        </form>
      </Form>
    </AuthLayout>
  )
}
```

- [ ] **Step 3: TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "tenant/login" | head -10
```

预期：无错误。

- [ ] **Step 4: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/app/tenant/login/page.tsx
git commit -m "feat: implement /tenant/login as standalone admin login page"
```

---

## Task 2: 统一 cookie 名 — mc_enduser_refresh_token → mc_refresh_token

**Files:**
- Modify: `modelcraft-front/src/middleware.ts`
- Modify: `modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts`

Backend 已统一用 `mc_refresh_token` cookie，前端需要跟进。

- [ ] **Step 1: 修改 middleware.ts**

将：
```typescript
export const END_USER_REFRESH_COOKIE = 'mc_enduser_refresh_token'
```
改为：
```typescript
export const END_USER_REFRESH_COOKIE = 'mc_refresh_token'
```

注意：`DEV_REFRESH_COOKIE` 已经是 `'mc_refresh_token'`，两者合并后中间件对 end-user 路由的检查也用 `mc_refresh_token`。这意味着登录后只有一个 cookie，tenant 和 end-user 共享同一个 refresh cookie（符合设计）。

- [ ] **Step 2: 验证 middleware.ts 中 DEV_REFRESH_COOKIE 和 END_USER_REFRESH_COOKIE 现在相同**

```bash
grep -n "REFRESH_COOKIE\|mc_refresh" modelcraft-front/src/middleware.ts
```

预期：两个常量都是 `'mc_refresh_token'`。

考虑是否合并为一个常量：
```typescript
export const REFRESH_COOKIE = 'mc_refresh_token'
const DEV_REFRESH_COOKIE = REFRESH_COOKIE
export const END_USER_REFRESH_COOKIE = REFRESH_COOKIE
```

（保持 `END_USER_REFRESH_COOKIE` export 以兼容 `_proxy.ts` 的 import）

- [ ] **Step 3: 确认 _proxy.ts 的 logout cookie 清除**

```bash
grep -n "END_USER_REFRESH_COOKIE\|mc_enduser\|mc_refresh" \
  modelcraft-front/src/app/api/bff/org/\[orgName\]/end-user/auth/_proxy.ts
```

`_proxy.ts` 从 `middleware.ts` import `END_USER_REFRESH_COOKIE`，所以改了常量值后 logout 也自动清除正确的 cookie。无需额外修改。

- [ ] **Step 4: 检查 refresh route**

```bash
cat "modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts"
```

如果 refresh route 硬编码了 `mc_enduser_refresh_token`，更新为 `mc_refresh_token` 或使用 `END_USER_REFRESH_COOKIE` 常量。

- [ ] **Step 5: TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "middleware\|_proxy\|refresh" | head -10
```

- [ ] **Step 6: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/middleware.ts \
        "modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts" \
        "modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts"
git commit -m "feat: unify end-user refresh cookie name to mc_refresh_token"
```

---

## Task 3: useEndUserOrgLoginForm 加 is_admin 跳转分支

**Files:**
- Modify: `modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts`

登录成功后，从 JWT payload 读取 `is_admin`：
- `is_admin=true` → 跳转 `/org/[orgName]/projects`（管理后台）
- `is_admin=false` → 保持原来的 workspace 跳转

- [ ] **Step 1: 读取当前文件**

```bash
cat "modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts"
```

- [ ] **Step 2: 在登录成功分支中加 is_admin 检查**

在 `if (data.accessToken)` 分支中，解析 JWT payload 读取 `is_admin`，然后分支跳转：

```typescript
// 解析 JWT payload（不验签，仅读取字段）
function parseJwtPayload(token: string): Record<string, unknown> {
  try {
    const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64)) as Record<string, unknown>
  } catch {
    return {}
  }
}
```

在 `setAccessToken` 之后加入：
```typescript
// 管理员从 end-user 入口登录 → 跳转管理后台
const payload = parseJwtPayload(data.accessToken)
if (payload.is_admin === true) {
  router.push(`/org/${orgName}/projects`)
  return
}
```

- [ ] **Step 3: 处理无 accessToken 但 is_admin=true 的情况**

（管理员账号总有 accessToken，此情况不会发生，无需处理）

- [ ] **Step 4: TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "useEndUserOrgLoginForm" | head -5
```

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add "modelcraft-front/src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts"
git commit -m "feat: redirect admin users to management dashboard after end-user login"
```

---

## Task 4: 中간件 — end-user 보호 路由 检查

- [ ] **Step 1: 确认统一 cookie 后中间件逻辑仍然正确**

读取更新后的 middleware.ts：
```bash
cat modelcraft-front/src/middleware.ts
```

验证：
1. End-user workspace/projects 路由检查 `mc_refresh_token` cookie
2. Tenant 路由也检查 `mc_refresh_token` cookie
3. 两个路由都指向同一个 cookie → 用户登录一次即可

- [ ] **Step 2: 确认中间件公开路径包含 /tenant/login**

```bash
grep -n "TENANT_LOGIN_PATH\|tenant.*login\|public" modelcraft-front/src/middleware.ts
```

`/tenant/login` 应该在 `DEV_PUBLIC_PATHS` 中（或通过 `TENANT_LOGIN_PATH` 常量包含）。

如果不包含，加入：
```typescript
const DEV_PUBLIC_PATHS = [TENANT_LOGIN_PATH, TENANT_REGISTER_PATH, '/tenant/login']
```

检查 `TENANT_LOGIN_PATH` 是否已经是 `/tenant/login`：
```bash
grep -rn "TENANT_LOGIN_PATH" modelcraft-front/src/shared/constants/routes.ts 2>/dev/null
```

- [ ] **Step 3: lint 检查**

```bash
cd modelcraft-front && npx eslint src/middleware.ts --max-warnings 0 2>&1 | head -10
```

- [ ] **Step 4: Commit（如有变更）**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/middleware.ts
git commit -m "fix: ensure /tenant/login is in public paths and uses unified cookie"
```

---

## Task 5: 全局验证

- [ ] **Step 1: TypeScript 全量检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -v "^$" | head -30
```

预期：0 errors（或仅存量 errors，无本次引入的新错误）。

- [ ] **Step 2: Lint 检查**

```bash
cd modelcraft-front && npx next lint 2>&1 | tail -10
```

- [ ] **Step 3: 确认三条路径的 cookie 逻辑**

```bash
grep -n "mc_refresh_token\|mc_enduser" modelcraft-front/src/middleware.ts
grep -n "mc_refresh_token\|mc_enduser" "modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts"
```

预期：无 `mc_enduser_refresh_token` 残留（仅注释中可以保留历史说明）。

- [ ] **Step 4: 最终 commit**

```bash
cd /data/home/lukemxjia/modelcraft
git status
git add -A && git commit -m "chore: Plan 3 complete — dual login pages, unified cookie, admin redirect"
```
