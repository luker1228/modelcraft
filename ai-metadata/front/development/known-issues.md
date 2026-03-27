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
