---
paths:
  - "src/bff/apollo/**/*.ts"
  - "src/web/hooks/**/*.ts"
  - "src/hooks/**/*.ts"
  - "src/app/**/*.tsx"
  - "src/components/**/*.tsx"
---

# Apollo Client 实例引用稳定性规则

## 核心规则

在 React Hook 中动态创建并返回给 `useQuery` / `useMutation` 的对象（client、context 等），**必须**用 `useMemo` 或 `useRef` 确保引用稳定。否则 Apollo 会因 client 引用变化而重新获取数据，导致无限重渲染。

## 禁止模式

```tsx
// ❌ 禁止：每次渲染创建新实例
export function useProjectScopedClient(projectSlug?: string): ApolloClient<any> {
  const currentOrg = useOrganizationStore((s) => s.currentOrg)
  if (projectSlug && currentOrg) {
    return createProjectScopedClient(currentOrg, projectSlug)
  }
  return getOrgScopedClient()
}
```

## 必须模式

```tsx
// ✅ 必须：useMemo 缓存实例
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

## 适用范围

此规则适用于所有在 Hook 中动态创建并传递给 Apollo 的对象，包括但不限于：

- `ApolloClient` 实例（通过 `useQuery({ client })` 传递）
- GraphQL context 对象（通过 `useQuery({ context })` 传递）
- 任何作为 Apollo hook 参数的对象

## 参考案例

详见 [ai-metadata/front/development/known-issues.md](../../../ai-metadata/front/development/known-issues.md) 中的 `useProjectScopedClient` 无限重渲染 issue。
