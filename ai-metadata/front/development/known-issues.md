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

## ~~Issue: Hook 绕过 BFF 直接调用后端 GraphQL → 401 Unauthorized~~ ✅ 已解决

**日期**: 2026-04-25 | **解决日期**: 2026-06-12

**结果**: 终端用户前端页面已移除，`useOrgEndUsers.ts` 及相关 BFF 路由一并删除，此 issue 已无影响范围。

---

## Issue: ModelRecordTable 列宽看起来被长 ID“锁死”，拖拽像在整体缩放

**日期**: 2026-05-14

**症状**:

- 长 ID（UUID）列很难继续缩小
- 拖动某列宽度时，视觉上像整表一起变化
- 期望是单列可缩小，内容显示 `xxx...` 截断

**根因**:

1. 仅用 `min-w-max` 兜底时，表格最小宽度会受内容 max-content 影响，长文本会把最小宽度“顶住”
2. 单元格容器缺少 `min-w-0`/`overflow-hidden` 时，`truncate` 不稳定，长字符串会影响列宽行为

**涉及文件**:

- `src/web/components/shared/data-workspace/ModelRecordTable.tsx`

**修复方案**:

- 表格宽度改为**按列配置计算**：`width = minWidth = 索引列 + 数据列宽总和 + 操作列`
- 数据单元格加 `overflow-hidden`
- 内容容器加 `w-full min-w-0 max-w-full`
- 主值保持 `truncate`，确保长 ID 显示为 `xxx...`
- 操作列（关联/编辑/删除）使用 `sticky right-0` 固定在右侧，数据列横向滚动

**经验教训**:

1. **不要把 `min-w-max` 当作唯一方案** — 对长文本表格应优先“配置驱动宽度”，避免内容反向控制布局。
2. **`truncate` 生效有前提** — 父容器要允许收缩（`min-w-0`）并裁剪（`overflow-hidden`）。
3. **固定操作列要成对处理** — 表头和单元格都要 `sticky right-0`，并补 `bg + z-index + border-l`。
