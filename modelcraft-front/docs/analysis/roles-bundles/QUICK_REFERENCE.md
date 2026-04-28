# 🚀 ModelCraft Front - Roles/Bundles 快速参考

## ⚡ 一句话总结

这是一个**权限包详情页**，展示权限包中包含的所有权限点（permissions），支持删除操作。

---

## 🎯 4 个问题的直接答案

### 1. 页面文件路径在哪?
```
✅ modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx
```

---

### 2. 权限点显示在哪个组件？
```
✅ 第 99-146 行的 "Permissions Container" div
  ├─ 权限点列表容器 (Lines 99-146)
  ├─ 权限点项目卡片 (Lines 118-145)
  └─ 每个权限点显示: 名称 + 两个徽章(action/rowScope) + 删除按钮
```

**核心显示逻辑**：
```tsx
// 权限点名称优先级
displayName || modelDisplayName || modelId

// 权限点显示内容
├─ 名称: displayName (或 modelDisplayName/modelId)
├─ 操作徽章: ACTION_LABEL[action] (查询/新增/更新/删除/导出)
├─ 范围徽章: ROW_SCOPE_LABEL[rowScope] (全部/本人/本部门/部门及子部门)
└─ 删除按钮: 垂直菜单中显示(hover时), 点击显示确认对话框
```

---

### 3. 空状态怎么处理?
```
✅ Lines 109-115, 当 permissions.length === 0 时:

<div className="flex flex-col items-center justify-center py-12">
  <ShieldCheck className="mb-3 size-8 text-muted-foreground/30" />
  <p className="text-sm text-muted-foreground">暂无关联权限点</p>
</div>

设计特点:
  ├─ 盾牌图标 (ShieldCheck from lucide-react)
  ├─ 友好文本 "暂无关联权限点"
  ├─ 垂直居中布局
  └─ 半透明灰色配色
```

---

### 4. GraphQL 查询在哪?
```
✅ modelcraft-front/src/api-client/rbac/graphql-docs.ts

核心查询:
  ├─ GET_END_USER_BUNDLE: 获取权限包 + 权限点
  ├─ GET_END_USER_PERMISSIONS: 获取所有可用权限点
  ├─ ADD_PERMISSION_TO_BUNDLE: 添加权限点到权限包
  └─ REMOVE_PERMISSION_FROM_BUNDLE: 从权限包移除权限点 ← 页面用到

Hook 中使用:
  └─ useBundleManage() @ src/app/.../rbac/bundles/_hooks/useBundleManage.ts
     ├─ useQuery(GET_END_USER_BUNDLE)
     ├─ useQuery(GET_END_USER_PERMISSIONS)
     ├─ useMutation(REMOVE_PERMISSION_FROM_BUNDLE)
     └─ 返回: { bundle, allPermissions, loading, error, removePermission }
```

---

## 📂 核心文件地图

```
❇️  页面文件 (最重要)
    └─ src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx

📡 数据层
    ├─ Hook: src/app/.../rbac/bundles/_hooks/useBundleManage.ts
    └─ GraphQL: src/api-client/rbac/graphql-docs.ts

🏷️  类型定义
    └─ src/types/rbac.ts (EndUserPermission, EndUserPermissionBundle 等)

🎨 UI 组件
    ├─ @web/components/ui/badge
    ├─ @web/components/ui/button
    ├─ @web/components/ui/skeleton
    ├─ @web/components/ui/alert-dialog
    ├─ @web/components/features/layout (PageLayout)
    └─ lucide-react (icons)
```

---

## 🔗 数据关系

```
权限包 (EndUserPermissionBundle)
├─ id, name, description
├─ createdAt, updatedAt
└─ permissions[] (Array of EndUserPermission)
   ├─ id
   ├─ modelId (如 "user_model")
   ├─ action ('SELECT'|'INSERT'|'UPDATE'|'DELETE'|'EXPORT')
   ├─ rowScope ('ALL'|'SELF'|'DEPT'|'DEPT_AND_CHILDREN')
   ├─ displayName (可选)
   ├─ modelDisplayName (可选)
   ├─ description (可选)
   └─ columnPolicy { defaultMode, rules[] }
```

---

## 💾 Hook 使用方式

```tsx
// 在页面中使用
const { bundle, loading, error, removePermission } = useBundleManage({
  orgName,        // 从 useParams 获取
  projectSlug,    // 从 useParams 获取
  bundleId,       // 从 useParams 获取
})

// 获取权限点列表
const permissions = bundle?.permissions ?? []

// 删除权限点
const handleRemove = async (perm) => {
  const result = await removePermission(perm.id)
  if (result.success) {
    toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
  }
}
```

---

## 📊 4 种渲染状态

| 状态 | 条件 | 显示内容 |
|------|------|--------|
| **加载中** | `loading === true` | 4 行骨架屏加载器 (Skeleton) |
| **错误** | `error !== undefined` | 红色错误提示："加载失败：{error.message}" |
| **空** | `permissions.length === 0` | 盾牌图标 + "暂无关联权限点" 文字 |
| **列表** | `permissions.length > 0` | 权限点列表，每行一项 |

---

## 🎨 标签映射表

```tsx
// 操作类型 → 中文
ACTION_LABEL = {
  'SELECT': '查询',
  'INSERT': '新增',
  'UPDATE': '更新',
  'DELETE': '删除',
  'EXPORT': '导出',
}

// 数据范围 → 中文
ROW_SCOPE_LABEL = {
  'ALL': '全部',
  'SELF': '本人',
  'DEPT': '本部门',
  'DEPT_AND_CHILDREN': '部门及子部门',
}

// 操作类型 → 徽章颜色
ACTION_VARIANT = {
  'SELECT': 'secondary',      // 灰色
  'INSERT': 'default',        // 蓝色/主色
  'UPDATE': 'outline',        // 描边
  'DELETE': 'destructive',    // 红色
  'EXPORT': 'outline',        // 描边
}
```

---

## 🔄 删除权限点的完整流程

```
1️⃣  用户在权限点卡片上悬停
    └─ 删除按钮从隐藏变为可见 (opacity-0 → opacity-100)

2️⃣  用户点击删除按钮
    ├─ setRemovingId(perm.id) 记录正在删除的项
    └─ 显示 AlertDialog 确认对话框

3️⃣  用户确认删除
    └─ 调用 handleRemove(perm)

4️⃣  handleRemove 函数执行
    ├─ 调用 removePermission(perm.id) mutation
    ├─ 向后端发送 REMOVE_PERMISSION_FROM_BUNDLE
    │  └─ 变量: { bundleId, permissionId }
    ├─ 后端处理并返回结果
    └─ 检查是否有错误

5️⃣  删除完成后
    ├─ 自动 refetch GET_END_USER_BUNDLE (刷新权限包详情)
    ├─ 自动 refetch GET_END_USER_BUNDLES (刷新权限包列表)
    ├─ setRemovingId(null) 清空删除状态
    └─ 显示 Toast 通知

6️⃣  页面自动更新
    └─ 权限点从列表中移除 (权限包的 permissions 数组改变)
```

---

## 🆚 GraphQL 查询 vs Mutation

| 操作 | GET_END_USER_BUNDLE | REMOVE_PERMISSION_FROM_BUNDLE |
|-----|-------------------|------------------------------|
| **类型** | Query | Mutation |
| **触发时机** | 页面初始化 | 用户删除权限点 |
| **变量** | `{id: bundleId}` | `{input: {bundleId, permissionId}}` |
| **返回** | Bundle + Permissions | Updated Bundle + Error |
| **自动Refetch** | - | 是 (自动重新查询) |

---

## ⚙️ 核心类型接口

```typescript
// 权限点类型
interface EndUserPermission {
  id: string
  modelId: string
  action: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
  rowScope: 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'
  displayName?: string
  modelDisplayName?: string
  description?: string
  columnPolicy: ColumnPolicy
  createdAt: string
  updatedAt: string
}

// 权限包类型
interface EndUserPermissionBundle {
  id: string
  name: string
  description?: string
  permissions: EndUserPermission[]
  createdAt: string
  updatedAt: string
}

// Hook 返回类型
interface UseBundleManageReturn {
  bundle: EndUserPermissionBundle | null
  allPermissions: EndUserPermission[]
  loading: boolean
  error: Error | undefined
  addPermission: (id: string) => Promise<MutationResult>
  removePermission: (id: string) => Promise<MutationResult>
}
```

---

## 🎪 页面 UI 层级

```
PageLayout (最大宽度 5xl)
├─ Back Navigation
│  └─ Link → /org/{orgName}/project/{projectSlug}/roles?tab=bundles
│
├─ Header Section
│  ├─ h1: Bundle Name
│  └─ p: Bundle Description
│
├─ Error Display (if error)
│
└─ Permissions Container (rounded border)
   ├─ Header Row
   │  ├─ Label: "关联的权限点"
   │  └─ Count Badge
   │
   └─ Content
      ├─ Loading State: 4x Skeleton
      ├─ Empty State: Shield + Text
      └─ List State: Permission Items
         ├─ Permission Item (hoverable group)
         │  ├─ Left: Name + Badges
         │  └─ Right: Delete Button (hidden → visible on hover)
         └─ [... more items ...]
```

---

## 🔑 关键代码片段

### 获取权限点列表
```tsx
const permissions = bundle?.permissions ?? []
```

### 显示权限点名称（3 层优先级）
```tsx
{perm.displayName || perm.modelDisplayName || perm.modelId}
```

### 渲染权限点徽章
```tsx
<Badge variant={ACTION_VARIANT[perm.action] ?? 'secondary'}>
  {ACTION_LABEL[perm.action] ?? perm.action}
</Badge>
```

### 删除权限点
```tsx
const result = await removePermission(perm.id)
if (result.success) {
  toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
}
```

---

## ❓ FAQ

### Q: 权限点的名称从哪里来?
A: 优先级为 `displayName > modelDisplayName > modelId`

### Q: 删除权限点会删除权限点本身吗?
A: 不会，只是从权限包中解除绑定

### Q: 权限点列表为空时页面显示什么?
A: 显示盾牌图标 + "暂无关联权限点" 的空状态

### Q: 删除后需要手动刷新吗?
A: 不需要，mutation 后会自动 refetch 相关查询

### Q: 这个页面有搜索功能吗?
A: 没有，只显示权限包中的权限点列表

### Q: 支持批量删除吗?
A: 不支持，需要逐个删除

---

## 🚀 快速开发指南

### 如果要添加搜索功能:
1. 在 hook 中添加搜索参数
2. 使用 `permissions.filter()` 过滤
3. 添加搜索输入框到 UI

### 如果要添加排序功能:
1. 使用 GraphQL 的 `sortOrder` 字段
2. 实现排序按钮
3. 调用 `permissions.sort()` 排序

### 如果要添加批量操作:
1. 添加 checkbox 选择
2. 实现多选状态
3. 批量调用 removePermission mutation

---

