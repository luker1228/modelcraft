# ModelCraft Front - Code Snippets 速查表

## 1️⃣ 页面文件位置与导入

### 完整文件路径
```
modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx
```

### 页面导入项
```tsx
import * as React from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ArrowLeft, ShieldCheck, Loader2, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@web/components/ui/alert-dialog'
import { PageLayout } from '@web/components/features/layout'

import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type { EndUserPermission } from '@/types'
```

---

## 2️⃣ 权限点显示组件代码

### 权限点列表容器
```tsx
<div className="rounded-md border border-border">
  {/* Header row */}
  <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
    <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
      关联的权限点
    </span>
    {!loading && (
      <span className="ml-2 rounded-full bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
        {permissions.length}
      </span>
    )}
  </div>

  {/* Content */}
  {/* ... loading / empty / list states ... */}
</div>
```

### 权限点项目卡片完整代码
```tsx
{permissions.map((perm) => (
  <div
    key={perm.id}
    className="group flex items-center gap-3 px-4 py-3 hover:bg-muted/20"
  >
    <div className="min-w-0 flex-1">
      {/* Permission name section */}
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-foreground">
          {perm.displayName || perm.modelDisplayName || perm.modelId}
        </span>
        {perm.displayName && (perm.modelDisplayName || perm.modelId) && (
          <span className="font-mono text-[11px] text-muted-foreground/60">
            {perm.modelDisplayName ?? perm.modelId}
          </span>
        )}
      </div>
      
      {/* Permission badges */}
      <div className="mt-1 flex items-center gap-1.5">
        {/* Action badge */}
        <Badge
          variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
          className="h-4 px-1.5 py-0 text-[10px]"
        >
          {ACTION_LABEL[perm.action] ?? perm.action}
        </Badge>
        
        {/* Row scope badge */}
        <Badge variant="outline" className="h-4 px-1.5 py-0 text-[10px]">
          {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
        </Badge>
      </div>
    </div>

    {/* Delete button with dialog */}
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-destructive"
          disabled={removingId === perm.id}
        >
          {removingId === perm.id ? (
            <Loader2 className="size-3.5 animate-spin" />
          ) : (
            <Trash2 className="size-3.5" />
          )}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认解绑权限点</AlertDialogTitle>
          <AlertDialogDescription>
            确定要从该权限包中移除「{perm.displayName ?? perm.modelId}」吗？
            此操作不会删除权限点本身。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            onClick={() => void handleRemove(perm)}
          >
            确认移除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
))}
```

### 空状态代码
```tsx
permissions.length === 0 ? (
  <div className="flex flex-col items-center justify-center py-12">
    <ShieldCheck className="mb-3 size-8 text-muted-foreground/30" />
    <p className="text-sm text-muted-foreground">暂无关联权限点</p>
  </div>
) : (
  // list content
)
```

### 加载状态代码
```tsx
{loading ? (
  <div className="space-y-px">
    {Array.from({ length: 4 }).map((_, i) => (
      <div key={i} className="flex items-center gap-3 px-4 py-3">
        <Skeleton className="h-4 w-28" />
        <Skeleton className="h-5 w-14" />
        <Skeleton className="h-5 w-16" />
      </div>
    ))}
  </div>
) : (
  // content
)}
```

---

## 3️⃣ 权限点删除处理逻辑

### 页面层删除处理器
```tsx
const [removingId, setRemovingId] = React.useState<string | null>(null)

const handleRemove = async (perm: EndUserPermission) => {
  setRemovingId(perm.id)
  try {
    const result = await removePermission(perm.id)
    if (result.success) {
      toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
    } else {
      toast.error(result.errorMessage ?? '移除失败，请重试')
    }
  } catch {
    toast.error('移除权限点时发生错误')
  } finally {
    setRemovingId(null)
  }
}
```

### Hook 层删除逻辑
```tsx
const removePermission = useCallback(
  async (permissionId: string): Promise<MutationResult> => {
    if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
    const result = await removePermissionMutation({
      variables: { input: { bundleId, permissionId } },
    })
    const payload = result.data?.removeEndUserPermissionFromBundle
    if (payload?.error) {
      return { success: false, errorMessage: payload.error.message ?? '移除权限点失败' }
    }
    return { success: true }
  },
  [removePermissionMutation, bundleId],
)
```

---

## 4️⃣ Hook 使用示例

### useBundleManage Hook 完整使用
```tsx
export default function BundleDetailPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`

  // 使用 hook 获取数据
  const { bundle, loading, error, removePermission } = useBundleManage({
    orgName,
    projectSlug,
    bundleId,
  })

  const permissions = bundle?.permissions ?? []

  // ... rest of component
}
```

### Hook 内部数据获取
```tsx
export function useBundleManage({
  orgName,
  projectSlug,
  bundleId,
}: UseBundleManageProps): UseBundleManageReturn {
  const client = useProjectScopedClient(projectSlug, orgName)
  const skip = !orgName || !projectSlug || !bundleId

  // Query 1: 获取权限包详情（包含权限点）
  const {
    data: bundleData,
    loading: bundleLoading,
    error: bundleError,
  } = useQuery(GET_END_USER_BUNDLE, {
    client,
    variables: { id: bundleId ?? '' },
    skip,
  })

  // Query 2: 获取所有可用权限点
  const {
    data: permissionsData,
    loading: permissionsLoading,
    error: permissionsError,
  } = useQuery(GET_END_USER_PERMISSIONS, {
    client,
    skip: !orgName || !projectSlug,
  })

  // Mutation 1: 添加权限点
  const [addPermissionMutation] = useMutation(ADD_PERMISSION_TO_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES],
  })

  // Mutation 2: 移除权限点
  const [removePermissionMutation] = useMutation(REMOVE_PERMISSION_FROM_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES],
  })

  // ... callback definitions ...

  const bundle: EndUserPermissionBundle | null = bundleData?.endUserPermissionBundle ?? null
  const allPermissions: EndUserPermission[] = permissionsData?.endUserPermissions?.edges?.map(
    (edge: any) => edge.node,
  ) ?? []

  return {
    bundle,
    allPermissions,
    loading: bundleLoading || permissionsLoading,
    error: bundleError ?? permissionsError,
    addPermission,
    removePermission,
  }
}
```

---

## 5️⃣ GraphQL 查询代码

### 导入所有 GraphQL 操作
```tsx
import {
  GET_END_USER_BUNDLE,
  GET_END_USER_PERMISSIONS,
  GET_END_USER_BUNDLES,
  ADD_PERMISSION_TO_BUNDLE,
  REMOVE_PERMISSION_FROM_BUNDLE,
} from '@/api-client/rbac'
```

### GET_END_USER_BUNDLE 查询
```graphql
export const GET_END_USER_BUNDLE = gql`
  query GetEndUserBundle($id: ID!) {
    endUserPermissionBundle(id: $id) {
      id
      name
      description
      createdAt
      updatedAt
      permissions {
        sortOrder
        permission {
          id
          modelId
          action
          rowScope
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
      }
    }
  }
`
```

### REMOVE_PERMISSION_FROM_BUNDLE Mutation
```graphql
export const REMOVE_PERMISSION_FROM_BUNDLE = gql`
  mutation RemovePermissionFromBundle($input: RemoveEndUserPermissionFromBundleInput!) {
    removeEndUserPermissionFromBundle(input: $input) {
      bundle {
        id
        name
        permissions {
          sortOrder
          permission {
            id
            modelId
            action
            rowScope
          }
        }
      }
      error {
        __typename
        message
      }
    }
  }
`
```

---

## 6️⃣ 类型定义代码

### EndUserPermission 接口
```typescript
export interface EndUserPermission {
  id: string
  modelId: string
  action: EndUserPermissionAction              // 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
  rowScope: EndUserRowScope                    // 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'
  columnPolicy: ColumnPolicy
  displayName?: string
  modelDisplayName?: string
  description?: string
  createdAt: string
  updatedAt: string
}
```

### EndUserPermissionBundle 接口
```typescript
export interface EndUserPermissionBundle {
  id: string
  name: string
  description?: string
  permissions: EndUserPermission[]
  createdAt: string
  updatedAt: string
}
```

### 枚举类型
```typescript
export type EndUserPermissionAction = 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
export type EndUserRowScope = 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'
export type ColumnAccessMode = 'VISIBLE' | 'READONLY' | 'MASKED' | 'HIDDEN'
```

---

## 7️⃣ 标签映射代码

### 完整标签映射对象
```tsx
const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '更新',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部',
  SELF: '本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '部门及子部门',
}

const ACTION_VARIANT: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  SELECT: 'secondary',
  INSERT: 'default',
  UPDATE: 'outline',
  DELETE: 'destructive',
  EXPORT: 'outline',
}
```

### 使用标签映射
```tsx
// 获取操作类型的中文标签
const actionLabel = ACTION_LABEL[permission.action] ?? permission.action

// 获取操作类型的徽章颜色
const actionVariant = ACTION_VARIANT[permission.action] ?? 'secondary'

// 获取数据范围的中文标签
const scopeLabel = ROW_SCOPE_LABEL[permission.rowScope] ?? permission.rowScope

// 渲染徽章
<Badge variant={actionVariant} className="...">
  {actionLabel}
</Badge>
```

---

## 8️⃣ 路由与导航代码

### 返回链接构建
```tsx
const { orgName, projectSlug, bundleId } =
  useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`

<Link href={backHref} className="...">
  <ArrowLeft className="size-4" />
  返回权限包列表
</Link>
```

### 参数提取
```tsx
// 从 URL 参数中提取
const params = useParams<{
  orgName: string
  projectSlug: string
  bundleId: string
}>()

// 作为 hook 参数传递
useBundleManage({
  orgName: params.orgName,
  projectSlug: params.projectSlug,
  bundleId: params.bundleId,
})
```

---

## 9️⃣ 状态管理代码

### 页面级别 state
```tsx
// 追踪正在删除的权限点 ID
const [removingId, setRemovingId] = React.useState<string | null>(null)

// 使用
disabled={removingId === perm.id}

// 更新
setRemovingId(perm.id)    // 开始删除
setRemovingId(null)        // 删除完成
```

### Hook 返回的状态
```tsx
interface UseBundleManageReturn {
  bundle: EndUserPermissionBundle | null        // 权限包数据
  allPermissions: EndUserPermission[]           // 所有可用权限点
  loading: boolean                               // 加载中
  error: Error | undefined                       // 错误对象
  addPermission: (permissionId: string) => Promise<MutationResult>
  removePermission: (permissionId: string) => Promise<MutationResult>
}
```

---

## 🔟 常见用例代码片段

### 检查权限点是否为空
```tsx
const permissions = bundle?.permissions ?? []
const hasPermissions = permissions.length > 0

if (!hasPermissions) {
  // 显示空状态
}
```

### 显示权限点名称
```tsx
// 优先级：displayName > modelDisplayName > modelId
const displayName = perm.displayName || perm.modelDisplayName || perm.modelId

// 如果有两个名称，显示两个
{perm.displayName && (perm.modelDisplayName || perm.modelId) && (
  <span className="font-mono text-[11px]">
    {perm.modelDisplayName ?? perm.modelId}
  </span>
)}
```

### 条件性渲染权限点列表
```tsx
{loading ? (
  // 骨架屏加载器
  <Skeleton />
) : error ? (
  // 错误消息
  <div>加载失败：{error.message}</div>
) : permissions.length === 0 ? (
  // 空状态
  <EmptyState />
) : (
  // 权限点列表
  <PermissionList permissions={permissions} />
)}
```

### 删除权限点完整流程
```tsx
// 1. 声明状态
const [removingId, setRemovingId] = useState<string | null>(null)

// 2. 处理函数
const handleRemove = async (perm: EndUserPermission) => {
  setRemovingId(perm.id)
  try {
    const result = await removePermission(perm.id)
    if (result.success) {
      toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
    } else {
      toast.error(result.errorMessage ?? '移除失败')
    }
  } catch {
    toast.error('发生错误')
  } finally {
    setRemovingId(null)
  }
}

// 3. 在 UI 中使用
<Button
  disabled={removingId === perm.id}
  onClick={() => handleRemove(perm)}
>
  {removingId === perm.id ? <Loader2 /> : <Trash2 />}
</Button>
```

---

