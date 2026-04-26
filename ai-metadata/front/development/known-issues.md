# 已知问题 (Known Issues)

记录前端项目中发现的问题及修复方案，供后续开发参考。

---

## Issue: `useProjectScopedClient` 每次渲染创建新 ApolloClient 实例导致无限重渲染

**日期**: 2026-03-27

**症状**: 访问 `/org/{orgName}/projects/{projectSlug}/model-editor` 等使用 project-scoped Apollo Client 的页面时，浏览器报错：

```
Error: Too many re-renders. React limits the number of renders to prevent an infinite loop.
```

**根因**: `src/bff/apollo/clients.ts` 中的 `useProjectScopedClient` hook 在每次组件渲染时都调用 `createProjectScopedClient(currentOrg, projectSlug)`，返回一个新的 ApolloClient 实例。当这个新实例被传给 `useQuery` 的 `client` 参数时，Apollo 检测到 client 引用变化，触发重新获取数据 → 结果返回 → 组件重渲染 → 又创建新 client → 无限循环。

**涉及文件**:
- `src/bff/apollo/clients.ts` — `useProjectScopedClient` 函数
- `src/web/hooks/useDatabases.ts` — 调用方之一（使用 `useProjectScopedClient` 作为 `useQuery` 的 client）
- 其他所有使用 `useProjectScopedClient` 的 hook/组件

**修复方案**: 用 `useMemo` 缓存 client 实例，仅在 `projectSlug` 或 `currentOrg` 变化时才重建。

```tsx
// ❌ 修复前：每次渲染创建新实例
export function useProjectScopedClient(projectSlug?: string): ApolloClient<any> {
  const currentOrg = useOrganizationStore((s) => s.currentOrg)
  if (projectSlug && currentOrg) {
    return createProjectScopedClient(currentOrg, projectSlug)
  }
  return getOrgScopedClient()
}

// ✅ 修复后：useMemo 缓存
export function useProjectScopedClient(projectSlug?: string): ApolloClient<any> {
  const currentOrg = useOrganizationStore((s) => s.currentOrg)
  return useMemo(() => {
    if (projectSlug && currentOrg) {
      return createProjectScopedClient(currentOrg, projectSlug)
    }
    return getOrgScopedClient()
  }, [projectSlug, currentOrg])
}
```

**经验教训**:

1. **Hook 中返回的对象必须稳定** — 任何在 hook 中动态创建并返回给 `useQuery`/`useMutation` 的对象（client、context、cache 等），都必须用 `useMemo` 或 `useRef` 确保引用稳定，否则会导致无限重渲染。
2. **`useQuery({ client })` 对 client 引用敏感** — Apollo Client 的 `useQuery` hook 会比较 client 引用，引用变化等同于 query 配置变化，会触发重新获取。
3. **创建 ApolloClient 的开销很大** — 每次创建包含完整的 cache、link chain，不应该在渲染路径中反复执行。

---

## Issue: Hook 绕过 BFF 直接调用后端 GraphQL → 401 Unauthorized

**日期**: 2026-04-25

**症状**: GraphQL 请求返回 HTTP 401，response body 为 `Unauthorized`，请求没有进入 resolver，后端日志中 `request_id` 为空。

**根因**: Web 层 hook（`useOrgEndUsers`）自己写了 `postOrgGraphQL` 函数，直接 `fetch('/graphql/org/{orgName}/')` 绕过了 BFF，而 BFF route（`/api/bff/org/[orgName]/end-user/users`）已经存在且正确处理了认证转发。

正确的架构是：**Web 层 → BFF route → 后端**。BFF route 负责把客户端的 cookie/authorization header 转发给后端，Web 层完全不用关心认证细节。

```
❌ 错误路径（hook 绕过 BFF）
useOrgEndUsers.postOrgGraphQL()
    → fetch('/graphql/org/{orgName}/')     ← 直接打后端，无 Authorization → 401

✅ 正确路径（经过 BFF）
useOrgEndUsers
    → fetch('/api/bff/org/{orgName}/end-user/users')   ← 打 BFF route
    → BFF route 转发 authorization header
    → fetch('/graphql/org/{orgName}/')                  ← BFF 打后端 → 正常
```

**涉及文件**:
- `src/web/hooks/end-users/useOrgEndUsers.ts` — 自写 `postOrgGraphQL` 绕过了已有的 BFF route

**修复方案**: 删除 hook 内的 `postOrgGraphQL`，改为调用 BFF REST 接口：

```ts
// ❌ 错误：hook 自己绕过 BFF 打 GraphQL
fetch(`/graphql/org/${orgName}/`, { method: 'POST', ... })

// ✅ 正确：hook 调用 BFF，由 BFF 负责认证和后端通信
fetch(`/api/bff/org/${orgName}/end-user/users`, { cache: 'no-store' })
fetch(`/api/bff/org/${orgName}/end-user/users`, { method: 'POST', ... })
fetch(`/api/bff/org/${orgName}/end-user/users/${userId}/status`, { method: 'PATCH', ... })
fetch(`/api/bff/org/${orgName}/end-user/users/${userId}`, { method: 'DELETE' })
```

**经验教训**:

1. **先检查是否已有 BFF route** — 写 hook 前先看 `/api/bff/...` 目录是否已有对应路由，直接 `fetch('/api/bff/...')` 即可，不要自己造 raw fetch 调用链。
2. **浏览器端可以直接用 Apollo Client 调 Org/Project GraphQL** — `listEndUsers`、`listProjects` 等接口，在 hook 里直接 `useQuery` 即可，auth link 自动注入 JWT。
3. **服务端 BFF route 用 X-Internal-Token 认证，不转发用户 JWT** — BFF 应该用服务端持有的 `X-Internal-Token`（配合 `X-Org-Name`/`X-Project-Slug` header）自主完成认证，优先用 `callGoXxx()` 封装，也可以直接打 GraphQL（带 Internal Token）。
4. **症状识别** — 后端 `request_id` 为空 + HTTP 401 = 请求被 JWT 中间件拦截，检查认证 header 是否正确传递。
