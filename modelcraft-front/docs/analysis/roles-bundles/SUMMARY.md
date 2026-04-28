# 📋 ModelCraft Front - Roles/Bundles 页面探索总结

## 📍 快速定位

| 内容 | 位置 |
|-----|-----|
| **页面文件** | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx` |
| **数据 Hook** | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage.ts` |
| **GraphQL 定义** | `modelcraft-front/src/api-client/rbac/graphql-docs.ts` |
| **类型定义** | `modelcraft-front/src/types/rbac.ts` |

---

## 🎯 4 个核心问题的答案

### 1️⃣ 页面对应的文件路径

**完整路径**：
```
modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/roles/bundles/[bundleId]/page.tsx
```

**文件特性**：
- 使用 `'use client'` - 客户端组件
- 大小: 8.3 KB
- 框架: Next.js App Router + React Hooks
- 状态管理: Apollo Client + Local State (React.useState)

---

### 2️⃣ 权限点显示的相关组件

#### 权限点显示组件架构

```
┌─ Permissions Container (rounded-md border)
│
├─ Header Row
│  ├─ Label: "关联的权限点"
│  └─ Count Badge: {permissions.length}
│
└─ Content (3 states)
   ├─ Loading State
   │  └─ 4x Skeleton Loaders
   │
   ├─ Empty State ← 重点！
   │  ├─ ShieldCheck Icon (lucide-react)
   │  ├─ Text: "暂无关联权限点"
   │  └─ Layout: flex flex-col items-center justify-center py-12
   │
   └─ List State
      └─ Permission Items (div.map)
         ├─ Left: Name + Badges (flex-1)
         │  ├─ Primary Name
         │  ├─ Secondary Model ID (optional)
         │  ├─ Action Badge (colored)
         │  └─ Scope Badge (outlined)
         │
         └─ Right: Delete Button
            ├─ Hover visibility
            ├─ Loading spinner
            └─ AlertDialog confirmation
```

#### 权限点显示使用的 UI 组件

| 组件 | 来源 | 用途 |
|-----|------|------|
| `Badge` | `@web/components/ui/badge` | 权限类型和数据范围徽章 |
| `Button` | `@web/components/ui/button` | 删除按钮 |
| `Skeleton` | `@web/components/ui/skeleton` | 加载占位符 |
| `AlertDialog` | `@web/components/ui/alert-dialog` | 删除确认对话框 |
| `PageLayout` | `@web/components/features/layout` | 页面容器 |
| Icons | `lucide-react` | 图标 (ShieldCheck, Trash2, Loader2, ArrowLeft) |

#### 权限点属性显示逻辑

```tsx
// 权限点显示属性优先级
displayName || modelDisplayName || modelId

// 显示在徽章中的属性
- action: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
  → 映射到中文: 查询 | 新增 | 更新 | 删除 | 导出
  
- rowScope: 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'
  → 映射到中文: 全部 | 本人 | 本部门 | 部门及子部门
```

---

### 3️⃣ 空状态处理逻辑

#### 空状态渲染位置
**第 109-115 行** - 权限点列表内容区

```tsx
permissions.length === 0 ? (
  <div className="flex flex-col items-center justify-center py-12">
    <ShieldCheck className="mb-3 size-8 text-muted-foreground/30" />
    <p className="text-sm text-muted-foreground">暂无关联权限点</p>
  </div>
) : (
  // 权限点列表
)
```

#### 空状态特点

| 特性 | 值 |
|-----|---|
| **触发条件** | `permissions.length === 0` |
| **图标** | `ShieldCheck` (lucide-react) |
| **图标大小** | `size-8` (32px) |
| **图标颜色** | `text-muted-foreground/30` (半透明灰色) |
| **文本** | "暂无关联权限点" |
| **文本大小** | `text-sm` |
| **垂直间距** | `py-12` (48px) |
| **对齐方式** | 水平竖直居中 (`flex-col items-center justify-center`) |
| **含义** | 权限包中没有关联任何权限点 |

#### 其他两种状态

**加载状态** (Lines 104-108)
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
)}
```
- 显示 4 行骨架加载器
- 用于 API 数据加载中

**列表状态** (Lines 116+)
- 显示权限点项目列表
- 每行一个权限点
- 支持交互（删除）

---

### 4️⃣ 权限点相关 GraphQL 查询

#### 查询操作总览

| 操作 | 类型 | 目的 | 变量 |
|-----|------|------|------|
| `GET_END_USER_BUNDLE` | Query | 获取权限包详情 + 权限点 | `{id: bundleId}` |
| `GET_END_USER_PERMISSIONS` | Query | 获取所有可用权限点 | 无 |
| `ADD_PERMISSION_TO_BUNDLE` | Mutation | 添加权限点到权限包 | `{bundleId, permissionId}` |
| `REMOVE_PERMISSION_FROM_BUNDLE` | Mutation | 从权限包移除权限点 | `{bundleId, permissionId}` |

#### GET_END_USER_BUNDLE 查询详解

**作用**：获取权限包的完整信息（包括其中的权限点）

**返回字段**：
```graphql
{
  id                          # 权限包 ID
  name                        # 权限包名称
  description                 # 权限包描述
  createdAt                   # 创建时间
  updatedAt                   # 更新时间
  permissions {               # 权限点数组
    sortOrder                 # 排序顺序
    permission {              # 权限点详情
      id                      # 权限点 ID
      modelId                 # 关联的 Model ID
      action                  # 操作类型 (SELECT|INSERT|UPDATE|DELETE|EXPORT)
      rowScope                # 行数据范围 (ALL|SELF|DEPT|DEPT_AND_CHILDREN)
      displayName             # 权限点显示名称
      description             # 权限点描述
      columnPolicy {          # 列访问策略
        defaultMode
        rules { fieldName, mode, maskPattern }
      }
    }
  }
}
```

#### REMOVE_PERMISSION_FROM_BUNDLE Mutation 详解

**作用**：从权限包中移除一个权限点

**输入**：
```typescript
input: {
  bundleId: string        # 权限包 ID
  permissionId: string    # 权限点 ID
}
```

**返回**：
```graphql
{
  bundle {                # 更新后的权限包
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
  error {                 # 错误信息（如果有）
    __typename
    message
  }
}
```

**自动 Refetch**：
- `GET_END_USER_BUNDLE` - 刷新权限包详情
- `GET_END_USER_BUNDLES` - 刷新权限包列表

#### GraphQL 文件位置与导入

**文件**：`modelcraft-front/src/api-client/rbac/graphql-docs.ts`

**导入方式**：
```tsx
import {
  GET_END_USER_BUNDLE,
  GET_END_USER_PERMISSIONS,
  GET_END_USER_BUNDLES,
  ADD_PERMISSION_TO_BUNDLE,
  REMOVE_PERMISSION_FROM_BUNDLE,
} from '@/api-client/rbac'
```

---

## 📊 数据流向总结

### 初始化流程

```
1. URL 路由参数解析
   └─ orgName, projectSlug, bundleId

2. useBundleManage Hook 执行
   ├─ Query: GET_END_USER_BUNDLE (获取权限包 + 权限点)
   └─ Query: GET_END_USER_PERMISSIONS (获取所有权限点)

3. 并行加载
   └─ 返回 { bundle, allPermissions, loading, error, removePermission, addPermission }

4. 页面渲染
   ├─ 如果 loading → 显示骨架屏
   ├─ 如果 error → 显示错误信息
   ├─ 如果 permissions 为空 → 显示空状态
   └─ 否则 → 显示权限点列表
```

### 删除权限点流程

```
1. 用户点击权限点的删除按钮
   └─ removingId 设为该权限点 ID

2. 显示 AlertDialog 确认对话框

3. 用户确认删除
   └─ 调用 removePermission(permissionId)

4. Hook 执行 REMOVE_PERMISSION_FROM_BUNDLE mutation
   └─ 向后端发送 { bundleId, permissionId }

5. 后端处理后返回结果
   └─ 检查 error 字段

6. 自动 refetch 两个查询
   └─ GET_END_USER_BUNDLE 和 GET_END_USER_BUNDLES

7. 显示 Toast 通知
   └─ 成功: "已移除「xxx」"
   └─ 失败: 显示错误信息

8. 页面自动更新
   └─ 权限点从列表中移除
```

---

## 🔑 关键实现要点

### 权限点名称显示优先级

```tsx
perm.displayName || perm.modelDisplayName || perm.modelId
```

- **优先级 1**：`displayName` - 权限点的友好显示名称
- **优先级 2**：`modelDisplayName` - Model 的显示名称
- **优先级 3**：`modelId` - Model 的 ID

### 标签映射和颜色

```tsx
ACTION_LABEL: { SELECT: '查询', INSERT: '新增', UPDATE: '更新', DELETE: '删除', EXPORT: '导出' }
ROW_SCOPE_LABEL: { ALL: '全部', SELF: '本人', DEPT: '本部门', DEPT_AND_CHILDREN: '部门及子部门' }
ACTION_VARIANT: { SELECT: 'secondary', INSERT: 'default', UPDATE: 'outline', DELETE: 'destructive', EXPORT: 'outline' }
```

### 状态管理

- **页面级**：`removingId` - 正在删除的权限点 ID
- **Hook 级**：`bundle`, `allPermissions`, `loading`, `error`
- **缓存**：Apollo Client 自动缓存查询结果
- **无需 Redux/Context**：本页面自包含

---

## ✨ 核心亮点

✅ **完整的状态管理**
- Loading、Error、Empty、Success 四种状态完整处理

✅ **优雅的空状态设计**
- 使用有意义的图标 (Shield) 和友好文本

✅ **安全的删除确认**
- AlertDialog 确认对话框防止误删

✅ **并行数据加载**
- 两个查询并行执行，减少总加载时间

✅ **自动缓存更新**
- Mutation 后自动 refetch，无需手动刷新

✅ **类型安全**
- 完整的 TypeScript 类型定义

✅ **国际化支持**
- 所有标签都使用中文映射

---

## 🚀 扩展可能性

### 后续可以添加的功能

1. **添加权限点**
   - 使用 `ADD_PERMISSION_TO_BUNDLE` mutation
   - 显示可用权限点的列表/对话框

2. **权限点排序**
   - 使用 GraphQL 的 `sortOrder` 字段
   - 实现拖拽排序 UI

3. **批量操作**
   - 多选权限点
   - 批量删除

4. **权限点搜索**
   - 在列表中搜索权限点
   - 过滤按 action、rowScope 等

5. **权限点详情面板**
   - 展开权限点详情
   - 显示 columnPolicy 等高级信息

---

## 📚 相关文件索引

```
modelcraft-front/
├── src/
│   ├── app/
│   │   └── org/[orgName]/project/[projectSlug]/
│   │       ├── roles/
│   │       │   └── bundles/
│   │       │       └── [bundleId]/
│   │       │           └── page.tsx ⭐ 主页面
│   │       └── rbac/
│   │           └── bundles/
│   │               └── _hooks/
│   │                   └── useBundleManage.ts ⭐ 数据 Hook
│   ├── api-client/
│   │   └── rbac/
│   │       └── graphql-docs.ts ⭐ GraphQL 定义
│   ├── types/
│   │   └── rbac.ts ⭐ 类型定义
│   └── components/
│       ├── ui/
│       │   ├── badge.tsx
│       │   ├── button.tsx
│       │   ├── skeleton.tsx
│       │   └── alert-dialog.tsx
│       └── features/
│           └── layout/
│               └── PageLayout.tsx
└── node_modules/
    └── lucide-react/ (图标库)
```

---

