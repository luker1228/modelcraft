# ModelCraft Front - Roles/Bundles 页面架构与数据流

## 📐 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│  Browser URL Route                                                  │
│  /org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]     │
└───────────────────────┬─────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Page Component (Client Component)                                  │
│  src/app/.../roles/bundles/[bundleId]/page.tsx                      │
│                                                                      │
│  - Extract route params: orgName, projectSlug, bundleId             │
│  - Use useBundleManage hook                                         │
│  - Render UI with:                                                  │
│    * Back navigation                                                │
│    * Bundle header (name, description)                              │
│    * Permissions list container                                     │
│    * Empty state or permission items                                │
└───────────────────────┬─────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Data Fetching Layer                                                │
│  useBundleManage Hook                                               │
│  src/app/.../rbac/bundles/_hooks/useBundleManage.ts                │
│                                                                      │
│  Queries:                                                            │
│  ├─ GET_END_USER_BUNDLE (variables: bundleId)                      │
│  │  └─ Returns: bundle info + permissions nested                   │
│  │                                                                  │
│  └─ GET_END_USER_PERMISSIONS (no vars)                             │
│     └─ Returns: all permissions with pagination                    │
│                                                                      │
│  Mutations:                                                          │
│  ├─ ADD_PERMISSION_TO_BUNDLE                                       │
│  │  └─ Variables: { bundleId, permissionId }                       │
│  │  └─ Refetch: GET_END_USER_BUNDLE, GET_END_USER_BUNDLES         │
│  │                                                                  │
│  └─ REMOVE_PERMISSION_FROM_BUNDLE                                  │
│     └─ Variables: { bundleId, permissionId }                       │
│     └─ Refetch: GET_END_USER_BUNDLE, GET_END_USER_BUNDLES         │
└───────────────────────┬─────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│  GraphQL Client Layer                                               │
│  Apollo Client (Project-scoped)                                     │
│  src/api-client/rbac/graphql-docs.ts                               │
└───────────────────────┬─────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Backend GraphQL API                                                │
│  Resolvers:                                                          │
│  ├─ Query.endUserPermissionBundle(id)                              │
│  ├─ Query.endUserPermissions(input)                                │
│  ├─ Query.endUserPermissionBundles(input)                          │
│  ├─ Mutation.addEndUserPermissionToBundle(input)                   │
│  └─ Mutation.removeEndUserPermissionFromBundle(input)              │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 数据流向

### 页面初始化流程

```
BundleDetailPage Component Loads
    │
    ├─ useParams() → Extract bundleId
    │
    ├─ useBundleManage() Hook
    │   │
    │   ├─ useQuery(GET_END_USER_BUNDLE, { id: bundleId })
    │   │   └─ Parallel Query 1
    │   │
    │   └─ useQuery(GET_END_USER_PERMISSIONS, {})
    │       └─ Parallel Query 2
    │
    ├─ Return { bundle, permissions, loading, error }
    │
    └─ Render
        ├─ If loading: Show skeleton loaders
        ├─ If error: Show error message
        └─ If success:
            ├─ Show bundle name & description
            └─ Show permissions list:
                ├─ If permissions.length === 0 → Empty state
                └─ If permissions.length > 0 → List items
```

### 删除权限点流程

```
User Interaction: Click Trash Icon on Permission Item
    │
    ├─ Display AlertDialog confirmation
    │
    └─ User Confirms Delete
        │
        ├─ Call removePermission(permissionId)
        │   │
        │   ├─ Execute REMOVE_PERMISSION_FROM_BUNDLE mutation
        │   │   │
        │   │   ├─ Variables: { bundleId, permissionId }
        │   │   │
        │   │   └─ Backend deletes association
        │   │
        │   ├─ Auto-refetch: GET_END_USER_BUNDLE
        │   │
        │   ├─ Auto-refetch: GET_END_USER_BUNDLES
        │   │
        │   └─ Return: { success: true/false, errorMessage? }
        │
        ├─ Update UI state: setRemovingId(null)
        │
        └─ Show Toast notification
            ├─ Success: "已移除「xxx」"
            └─ Error: Error message
```

---

## 📊 权限点(Permission)数据结构

### 数据转换流程

```
GraphQL Response (GET_END_USER_BUNDLE)
└─ endUserPermissionBundle
   └─ permissions [] (Array of PermissionAssociation)
      └─ permission (EndUserPermission)
         ├─ id: string
         ├─ modelId: string  (e.g., "user_model")
         ├─ action: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
         ├─ rowScope: 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'
         ├─ displayName: string (optional, e.g., "查看用户")
         ├─ modelDisplayName: string (optional, e.g., "用户管理")
         ├─ description: string (optional)
         └─ columnPolicy: { defaultMode, rules[] }

         │
         ▼ (Page transforms to display format)

         Display in UI
         ├─ Primary Name: displayName || modelDisplayName || modelId
         ├─ Secondary Name: modelDisplayName (if displayName exists)
         ├─ Action Badge: ACTION_LABEL[action] (Chinese label)
         │  └─ Color variant: ACTION_VARIANT[action]
         │     ├─ 'SELECT' → secondary (gray)
         │     ├─ 'INSERT' → default (primary)
         │     ├─ 'UPDATE' → outline
         │     ├─ 'DELETE' → destructive (red)
         │     └─ 'EXPORT' → outline
         ├─ Scope Badge: ROW_SCOPE_LABEL[rowScope] (Chinese label)
         └─ Delete Button: Interactive with loading state
```

---

## 🎯 关键组件关系

```
BundleDetailPage
├─ PageLayout (Container)
│
├─ Back Navigation Link
│  └─ href: `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`
│
├─ Header Section
│  ├─ h1: Bundle Name
│  └─ p: Bundle Description
│
├─ Error Display (conditional)
│
└─ Permissions Container (with border)
   ├─ Header Row
   │  ├─ "关联的权限点" Label
   │  └─ Permission Count Badge
   │
   └─ Content Area (3 states)
      ├─ Loading State
      │  └─ 4x Skeleton loaders
      │
      ├─ Empty State
      │  ├─ ShieldCheck Icon (lucide-react)
      │  └─ "暂无关联权限点" Text
      │
      └─ List State
         └─ Permission Items (map)
            ├─ Left Section: Name & Badges
            │  ├─ Primary Name Text
            │  ├─ Secondary Model ID (if applicable)
            │  ├─ Action Badge
            │  └─ Row Scope Badge
            │
            └─ Right Section: Delete Button
               ├─ Hover visibility (opacity-0 → opacity-100)
               ├─ Loading spinner
               └─ AlertDialog confirmation
                  ├─ Title: "确认解绑权限点"
                  ├─ Description: Confirm message
                  ├─ Cancel button
                  └─ Confirm button (destructive)
```

---

## 🔌 使用的 UI 组件库

来源: `@web/components/ui` 和 `lucide-react`

| 组件 | 用途 |
|-----|------|
| `Button` | 返回链接、删除按钮、对话框操作 |
| `Badge` | 权限类型和数据范围显示 |
| `Skeleton` | 加载状态占位符 |
| `AlertDialog` | 确认删除对话框 |
| `PageLayout` | 页面容器和最大宽度限制 |
| `ShieldCheck` icon | 空状态图标 |
| `ArrowLeft` icon | 返回导航箭头 |
| `Trash2` icon | 删除按钮图标 |
| `Loader2` icon | 加载中动画 |

---

## 📝 状态管理

### Component State (useState)

```tsx
const [removingId, setRemovingId] = React.useState<string | null>(null)
// 用途: 标记正在删除的权限点，用于按钮加载状态
```

### Hook State (useBundleManage)

```tsx
// From useQuery(GET_END_USER_BUNDLE)
bundle: EndUserPermissionBundle | null
bundleLoading: boolean
bundleError: Error | undefined

// From useQuery(GET_END_USER_PERMISSIONS)
allPermissions: EndUserPermission[]
permissionsLoading: boolean
permissionsError: Error | undefined

// Combined
loading: boolean (bundleLoading || permissionsLoading)
error: Error | undefined (bundleError ?? permissionsError)
```

### No Redux/Context

- ✅ Local component state only
- ✅ Apollo cache for queries
- ✅ Automatic refetch after mutations

---

## 🌍 国际化与标签映射

```tsx
// 操作类型映射
SELECT  → "查询"      (secondary badge)
INSERT  → "新增"      (default badge)
UPDATE  → "更新"      (outline badge)
DELETE  → "删除"      (destructive badge)
EXPORT  → "导出"      (outline badge)

// 数据范围映射
ALL                 → "全部"
SELF                → "本人"
DEPT                → "本部门"
DEPT_AND_CHILDREN   → "部门及子部门"
```

---

## 🚀 性能考虑

1. **并行数据加载**
   - `GET_END_USER_BUNDLE` 和 `GET_END_USER_PERMISSIONS` 并行执行
   - Reduces total loading time

2. **自动缓存与重新获取**
   ```tsx
   refetchQueries: [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES]
   // After mutation success, both queries automatically refetch
   ```

3. **条件查询跳过**
   ```tsx
   skip: !orgName || !projectSlug || !bundleId
   // Queries don't execute until all params are available
   ```

4. **loading 骨架屏**
   - 4 个预设行的骨架加载器
   - 替代空白加载状态

---

## ⚠️ 错误处理

### 查询错误
- 显示红色错误提示: "加载失败：{error.message}"

### Mutation 错误
- 检查 `result.data?.removeEndUserPermissionFromBundle?.error`
- Toast 显示: "移除权限点时发生错误"

### 无效状态保护
- `if (!bundleId)` → 返回错误消息

---

