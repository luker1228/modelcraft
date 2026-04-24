# RBAC 行列级权限系统 — 前端架构方案

> **基准文档**：`ai-metadata/prd/rbac/` (00~04)
> **技术栈**：Next.js 14 (App Router) · TypeScript · Apollo Client · shadcn/ui · Tailwind CSS · MSW
> **GraphQL 端点**：Project-Scoped (`/graphql/org/{orgName}/project/{projectSlug}/`)
> **日期**：2026-04-24

---

## 0. 上下文说明

### 现有 API 现状

当前 `permission.graphql` 使用 Casbin 风格的 `obj/act` 权限模型（`PermissionRole`、`PermissionDef`），
与 PRD 中描述的 RBAC 模型（`Permission` 权限点 + `PermissionBundle` 权限包 + `rowScope/columnPolicy`）**尚未对齐**。

前端架构方案基于 **PRD 所描述的目标 GraphQL 接口** 设计，后端需按本文 §3 列出的 operation 列表补全 Schema，
前端在 Wave 1 阶段先使用 MSW mock，联调时按 §7 清单逐项切换。

### 知识图谱关键节点（来自 graphify）

- `createProjectScopedClient()` — 所有 RBAC 操作走 **Project-Scoped Apollo Client**（非 Org-Scoped）
  - 端点：`/graphql/org/{orgName}/project/{projectSlug}/`
  - 所有 operation 变量中需要传入 `projectSlug`
- `mockCreateRoleMutation()` — Community 0 已有角色 mock 基础，可复用骨架
- `RoleTable.tsx` — Community 84，已有基础角色表格，RBAC 角色管理页在此基础上扩展
- `normalizeDomainError()` / `createUnknownError()` — 错误处理体系，RBAC 错误统一走此管道

---

## 1. 页面结构与路由

```
/org/[orgName]/projects/[projectSlug]/settings/rbac/
├── page.tsx                     → redirect → /bundles（默认子页）
├── layout.tsx                   → RBACSettingsLayout（含侧边 Tab 导航）
├── bundles/
│   ├── page.tsx                 → 权限包列表页
│   └── [bundleId]/
│       └── page.tsx             → 权限包详情/编辑页（含权限点关联）
├── permissions/
│   ├── page.tsx                 → 权限点列表页（按 Model 分组）
│   └── new/
│       └── page.tsx             → 创建权限点向导页
├── roles/
│   ├── page.tsx                 → 角色列表页
│   └── [roleId]/
│       └── page.tsx             → 角色详情/编辑页（含权限包关联）
└── users/
    └── page.tsx                 → 用户授权页（分配角色/权限包、查看有效权限）
```

### 路由表

| 路径 | 组件文件 | 功能说明 |
|------|----------|----------|
| `/settings/rbac` | `rbac/page.tsx` | 重定向到 `/settings/rbac/bundles` |
| `/settings/rbac` (layout) | `rbac/layout.tsx` | RBAC 设置侧边 Tab 导航 |
| `/settings/rbac/bundles` | `bundles/page.tsx` | 权限包列表；支持搜索、创建、删除 |
| `/settings/rbac/bundles/[bundleId]` | `bundles/[bundleId]/page.tsx` | 权限包编辑；基本信息 + 关联权限点多选 |
| `/settings/rbac/permissions` | `permissions/page.tsx` | 权限点列表；按 Model 折叠分组 |
| `/settings/rbac/permissions/new` | `permissions/new/page.tsx` | 创建权限点三步向导 |
| `/settings/rbac/roles` | `roles/page.tsx` | 角色列表；区分普通角色/隐式角色徽标 |
| `/settings/rbac/roles/[roleId]` | `roles/[roleId]/page.tsx` | 角色编辑；关联权限包 |
| `/settings/rbac/users` | `users/page.tsx` | 用户授权；分配角色/权限包；查看有效权限合并视图 |

> **注意**：`/settings/rbac/` 整体在 Design Mode 下可见，对应 `workspaceMode: "design"` 约束。
> RBAC 是 Project 维度的配置，路由挂载在 `/org/[orgName]/projects/[projectSlug]/settings/rbac/`。

---

## 2. 组件架构

### 2.1 页面级组件树

```
app/org/[orgName]/projects/[projectSlug]/settings/rbac/
├── layout.tsx
│   └── RBACSettingsLayout            ← Tab 导航（bundles / permissions / roles / users）
│
├── bundles/page.tsx
│   └── BundleListPage                ← 容器：加载列表、触发创建弹窗
│       ├── BundleTable               ← DataTable：名称、描述、权限点数量、操作列
│       ├── CreateBundleDialog        ← 弹窗：名称 + 描述 + 权限点多选
│       └── DeleteBundleAlertDialog   ← 二次确认删除
│
├── bundles/[bundleId]/page.tsx
│   └── BundleEditPage                ← 容器：加载详情
│       ├── BundleBasicInfoForm       ← React Hook Form：name / description
│       └── BundlePermissionSelector  ← 多选权限点（按 Model 分组的 CheckboxTree）
│
├── permissions/page.tsx
│   └── PermissionListPage            ← 容器
│       ├── PermissionGroupedList     ← 按 Model 折叠展示权限点
│       │   └── PermissionRow         ← 单行：Model名/动作/行策略/列策略摘要/操作
│       └── DeletePermissionAlertDialog
│
├── permissions/new/page.tsx
│   └── CreatePermissionWizard        ← 3 步向导容器
│       ├── Step1ModelActionSelect    ← 选择 Model + 动作（select/insert/update/delete/export）
│       ├── Step2RowScopeSelector     ← 行策略选择（含动态校验）← 共享组件
│       └── Step3ColumnPolicyEditor  ← 列策略编辑器 ← 共享组件
│
├── roles/page.tsx
│   └── RoleListPage
│       ├── RoleTable                 ← 区分 isImplicit；隐式角色行灰显、无删除按钮
│       ├── CreateRoleDialog          ← 名称 + 描述（仅普通角色）
│       └── DeleteRoleAlertDialog     ← 仅对非隐式角色展示
│
├── roles/[roleId]/page.tsx
│   └── RoleEditPage
│       ├── RoleBasicInfoForm         ← 隐式角色禁用名称/描述编辑
│       └── RoleBundleSelector        ← 多选权限包（Table with checkbox）
│
└── users/page.tsx
    └── UserAuthPage
        ├── EndUserTable              ← 用户列表（userName / email / 已分配角色摘要）
        ├── UserDetailDrawer          ← 右侧 Sheet：选中用户的授权详情
        │   ├── UserRoleAssignPanel   ← 分配/撤销角色（过滤隐式角色）
        │   ├── UserBundleAssignPanel ← 直接关联/解除权限包
        │   └── EffectivePermView     ← 有效权限合并展示（三通道 → 按 Model 折叠）
        └── (无独立弹窗，全用 Sheet 侧边栏)
```

### 2.2 共享组件（`web/components/features/rbac/`）

| 组件 | 路径 | 说明 |
|------|------|------|
| `RowScopeSelector` | `rbac/RowScopeSelector.tsx` | 行策略 RadioGroup + 动态前提校验（见 §5） |
| `ColumnPolicyEditor` | `rbac/ColumnPolicyEditor.tsx` | 字段多选 + 策略模式（见 §5） |
| `BundlePermissionSelector` | `rbac/BundlePermissionSelector.tsx` | 权限点多选树（按 Model 分组 CheckboxTree） |
| `RoleBundleSelector` | `rbac/RoleBundleSelector.tsx` | 权限包多选表格（Table + checkbox） |
| `ImplicitRoleBadge` | `rbac/ImplicitRoleBadge.tsx` | "隐式角色" Badge，用于角色列表/详情页 |
| `EffectivePermView` | `rbac/EffectivePermView.tsx` | 有效权限合并展示（三通道），按 Model 折叠 |
| `PermissionRowSummary` | `rbac/PermissionRowSummary.tsx` | 单个权限点的摘要行（Model/动作/行策略/列策略标签） |

### 2.3 私有 Hooks（`_hooks/` 页面内）

| Hook | 所在页面 | 职责 |
|------|----------|------|
| `useBundleList` | bundles page | 加载权限包列表、删除、搜索过滤 |
| `useBundleEdit` | bundles/[id] | 加载详情、保存基本信息、管理权限点关联 |
| `usePermissionList` | permissions page | 加载权限点列表、按 Model 分组 |
| `useCreatePermissionWizard` | permissions/new | 向导状态机、行策略前提校验、提交 |
| `useRoleList` | roles page | 加载角色列表、创建、删除（隐式角色保护） |
| `useRoleEdit` | roles/[id] | 加载角色详情、管理权限包关联 |
| `useUserAuth` | users page | 加载用户列表、管理角色/权限包分配、有效权限展开 |

---

## 3. BFF 层设计

> 所有 RBAC 操作走 **Project-Scoped Apollo Client**（`createProjectScopedClient`）。
> GraphQL 操作定义文件：`src/web/graphql/queries/rbac.ts` 和 `src/web/graphql/mutations/rbac.ts`。

### 3.1 后端需新增的 GraphQL 类型（Schema 扩展）

当前 `permission.graphql` 使用 Casbin `obj/act` 模型，需在 `api/graph/project/schema/rbac.graphql` 新增：

```graphql
# 新文件：api/graph/project/schema/rbac.graphql

# ---- 权限点（EndUserPermission）----
enum RBACAction {
  SELECT
  INSERT
  UPDATE
  DELETE
  EXPORT
}

enum RowScope {
  ALL
  SELF
  DEPT
  DEPT_AND_CHILDREN
}

enum ColumnAccessMode {
  VISIBLE
  READONLY
  MASKED
  HIDDEN
}

type ColumnRule {
  fieldName: String!
  mode: ColumnAccessMode!
  maskPattern: String
}

type ColumnPolicy {
  defaultMode: ColumnAccessMode!
  rules: [ColumnRule!]!
}

type EndUserPermission {
  id: ID!
  modelId: String!
  modelDisplayName: String!   # 冗余字段，方便前端展示
  action: RBACAction!
  rowScope: RowScope!
  columnPolicy: ColumnPolicy!
  description: String
  createdAt: Time!
}

# ---- 权限包（EndUserPermissionBundle）----
type EndUserPermissionBundle {
  id: ID!
  name: String!
  description: String
  permissions: [EndUserPermission!]!
  createdAt: Time!
  updatedAt: Time!
}

# ---- 角色（EndUserRole）----
# 建议在 project schema 中：
#   type EndUserRole { ... isImplicit: Boolean! ... bundles: [EndUserPermissionBundle!]! }

type EndUserRole {
  id: ID!
  name: String!
  description: String
  isImplicit: Boolean!
  bundles: [EndUserPermissionBundle!]!
}

# ---- 有效权限视图（合并三通道）----
type EffectivePermission {
  modelId: String!
  modelDisplayName: String!
  action: RBACAction!
  rowScope: RowScope!       # 多来源取并集后最宽范围
  columnPolicy: ColumnPolicy!
  sources: [String!]!       # e.g. ["direct_bundle", "role:admin", "implicit:SYSTEM_USER"]
}

# ---- Queries ----
extend type Query {
  # 权限点
  endUserPermissions(projectSlug: String!): [EndUserPermission!]!
  endUserPermission(id: ID!): EndUserPermission

  # 权限包
  endUserBundles(projectSlug: String!): [EndUserPermissionBundle!]!
  endUserBundle(id: ID!): EndUserPermissionBundle

  # 角色（含隐式角色）
  endUserRoles(projectSlug: String!, includeImplicit: Boolean): [EndUserRole!]!
  endUserRole(id: ID!): EndUserRole

  # 用户有效权限（三通道合并）
  endUserEffectivePermissions(userId: String!, projectSlug: String!): [EffectivePermission!]!
}

# ---- Mutations ----
extend type Mutation {
  # 权限点 CRUD
  createEndUserPermission(input: CreateEndUserPermissionInput!): CreateEndUserPermissionPayload!
  deleteEndUserPermission(id: ID!): DeleteEndUserPermissionPayload!

  # 权限包 CRUD
  createEndUserBundle(input: CreateEndUserBundleInput!): CreateEndUserBundlePayload!
  updateEndUserBundle(id: ID!, input: UpdateEndUserBundleInput!): UpdateEndUserBundlePayload!
  deleteEndUserBundle(id: ID!): DeleteEndUserBundlePayload!

  # 权限包 ↔ 权限点关联
  addPermissionToBundle(bundleId: ID!, permissionId: ID!): BundlePermissionPayload!
  removePermissionFromBundle(bundleId: ID!, permissionId: ID!): BundlePermissionPayload!

  # 角色 ↔ 权限包关联
  addBundleToRole(roleId: ID!, bundleId: ID!): RoleBundlePayload!
  removeBundleFromRole(roleId: ID!, bundleId: ID!): RoleBundlePayload!

  # 用户 ↔ 角色分配（非隐式角色）
  assignRoleToEndUser(userId: String!, roleId: ID!, projectSlug: String!): AssignRolePayload!
  revokeRoleFromEndUser(userId: String!, roleId: ID!, projectSlug: String!): RevokeRolePayload!

  # 用户直接关联权限包
  addBundleToUser(userId: String!, bundleId: ID!, projectSlug: String!): UserBundlePayload!
  removeBundleFromUser(userId: String!, bundleId: ID!, projectSlug: String!): UserBundlePayload!
}
```

### 3.2 前端 GraphQL Operation 清单

**查询（`src/web/graphql/queries/rbac.ts`）**

```typescript
export const GET_END_USER_BUNDLES = gql`
  query GetEndUserBundles($projectSlug: String!) {
    endUserBundles(projectSlug: $projectSlug) {
      id name description
      permissions { id modelId modelDisplayName action rowScope }
    }
  }
`

export const GET_END_USER_BUNDLE = gql`
  query GetEndUserBundle($id: ID!) {
    endUserBundle(id: $id) {
      id name description
      permissions {
        id modelId modelDisplayName action rowScope
        columnPolicy {
          defaultMode
          rules { fieldName mode maskPattern }
        }
      }
    }
  }
`

export const GET_END_USER_PERMISSIONS = gql`
  query GetEndUserPermissions($projectSlug: String!) {
    endUserPermissions(projectSlug: $projectSlug) {
      id modelId modelDisplayName action rowScope description
      columnPolicy {
        defaultMode
        rules { fieldName mode maskPattern }
      }
    }
  }
`

export const GET_END_USER_ROLES = gql`
  query GetEndUserRoles($projectSlug: String!, $includeImplicit: Boolean) {
    endUserRoles(projectSlug: $projectSlug, includeImplicit: $includeImplicit) {
      id name description isImplicit
      bundles { id name }
    }
  }
`

export const GET_END_USER_ROLE = gql`
  query GetEndUserRole($id: ID!) {
    endUserRole(id: $id) {
      id name description isImplicit
      bundles { id name description permissions { id } }
    }
  }
`

export const GET_END_USER_EFFECTIVE_PERMISSIONS = gql`
  query GetEndUserEffectivePermissions($userId: String!, $projectSlug: String!) {
    endUserEffectivePermissions(userId: $userId, projectSlug: $projectSlug) {
      modelId modelDisplayName action rowScope sources
      columnPolicy {
        defaultMode
        rules { fieldName mode maskPattern }
      }
    }
  }
`

# 已有 org members，复用
export const GET_ORGANIZATION_MEMBERS = gql`
  query GetOrganizationMembersForRBAC {
    organizationMembers {
      id userID userName role { id name isImplicit }
    }
  }
`
```

**变更（`src/web/graphql/mutations/rbac.ts`）**

```typescript
export const CREATE_END_USER_BUNDLE = gql`...`
export const UPDATE_END_USER_BUNDLE = gql`...`
export const DELETE_END_USER_BUNDLE = gql`...`
export const ADD_PERMISSION_TO_BUNDLE = gql`...`
export const REMOVE_PERMISSION_FROM_BUNDLE = gql`...`
export const CREATE_END_USER_PERMISSION = gql`...`
export const DELETE_END_USER_PERMISSION = gql`...`
export const CREATE_END_USER_ROLE = gql`...` # 非隐式角色
export const DELETE_END_USER_ROLE = gql`...` # 前端保护：isImplicit=true 不调用
export const ADD_BUNDLE_TO_ROLE = gql`...`
export const REMOVE_BUNDLE_FROM_ROLE = gql`...`
export const ASSIGN_ROLE_TO_END_USER = gql`...`
export const REVOKE_ROLE_FROM_END_USER = gql`...`
export const ADD_BUNDLE_TO_USER = gql`...`
export const REMOVE_BUNDLE_FROM_USER = gql`...`
```

### 3.3 Mock 数据格式（Wave 1）

Mock 数据工厂放在 `src/mocks/data/org/rbac-factory.ts`：

```typescript
import { faker } from '@faker-js/faker'

// 权限包 mock
export function createMockBundle(override = {}) {
  return {
    id: faker.string.uuid(),
    name: `${faker.commerce.department()}查看包`,
    description: faker.lorem.sentence(),
    permissions: Array.from({ length: faker.number.int({ min: 1, max: 4 }) },
      () => createMockPermission()),
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

// 权限点 mock
export function createMockPermission(override = {}) {
  const actions = ['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'EXPORT']
  const rowScopes = ['ALL', 'SELF', 'DEPT', 'DEPT_AND_CHILDREN']
  const modelNames = ['orders', 'customers', 'products', 'invoices']
  const modelName = faker.helpers.arrayElement(modelNames)
  return {
    id: faker.string.uuid(),
    modelId: modelName,
    modelDisplayName: modelName.charAt(0).toUpperCase() + modelName.slice(1),
    action: faker.helpers.arrayElement(actions),
    rowScope: faker.helpers.arrayElement(rowScopes),
    columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
    description: faker.lorem.sentence(),
    ...override,
  }
}

// 隐式角色 mock（固定，不可删）
export const MOCK_IMPLICIT_ROLE: EndUserRole = {
  id: 'implicit-authenticated-user',
  name: 'SYSTEM_AUTHENTICATED_USER',
  description: '所有已登录用户自动获得的基础权限',
  isImplicit: true,
  bundles: [],
}

// 普通角色 mock
export function createMockRole(override = {}): EndUserRole {
  return {
    id: faker.string.uuid(),
    name: `${faker.person.jobTitle()}角色`,
    description: faker.lorem.sentence(),
    isImplicit: false,
    bundles: [],
    ...override,
  }
}

// 有效权限合并视图 mock
export function createMockEffectivePermission(override = {}) {
  return {
    modelId: 'orders',
    modelDisplayName: 'Orders',
    action: 'SELECT',
    rowScope: 'ALL',
    columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
    sources: ['role:admin', 'implicit:SYSTEM_AUTHENTICATED_USER'],
    ...override,
  }
}
```

---

## 4. 状态管理方案

### 4.1 Apollo Client Cache 策略

所有 RBAC 操作走 **Project-Scoped Apollo Client**（`getProjectScopedClient(orgName, projectSlug)`，已有单例）。

**缓存失效策略**（mutation 后）：

| 操作 | 需要 refetch 的 query |
|------|----------------------|
| 创建/删除权限包 | `GetEndUserBundles` |
| 编辑权限包（基本信息/权限点） | `GetEndUserBundle(id)` + `GetEndUserBundles` |
| 创建/删除权限点 | `GetEndUserPermissions` |
| 创建/删除角色 | `GetEndUserRoles` |
| 编辑角色（权限包关联） | `GetEndUserRole(id)` + `GetEndUserRoles` |
| 用户角色/权限包变更 | `GetEndUserEffectivePermissions(userId)` |

**实现方式**：在 mutation options 中使用 `refetchQueries`（不使用乐观更新，RBAC 权限数据一致性优先）：

```typescript
const [createBundle] = useMutation(CREATE_END_USER_BUNDLE, {
  refetchQueries: [
    { query: GET_END_USER_BUNDLES, variables: { projectSlug } },
  ],
})
```

**例外**：`addPermissionToBundle` / `removePermissionFromBundle` 操作频繁时，
可用 Apollo `cache.modify` 局部更新权限包的 `permissions` 数组，避免整列表重刷。

### 4.2 表单状态（React Hook Form + Zod）

所有表单使用 React Hook Form + Zod 验证，不引入额外状态库。

```typescript
// 权限包基本信息表单 schema
const bundleFormSchema = z.object({
  name: z.string().min(1, '名称不能为空').max(50, '名称最多 50 字符'),
  description: z.string().max(200, '描述最多 200 字符').optional(),
})

// 权限点创建 schema（含行策略前提校验）
const createPermissionSchema = z.object({
  modelId: z.string().min(1),
  action: z.enum(['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'EXPORT']),
  rowScope: z.enum(['ALL', 'SELF', 'DEPT', 'DEPT_AND_CHILDREN']),
  columnPolicy: z.object({
    defaultMode: z.enum(['VISIBLE', 'READONLY', 'MASKED', 'HIDDEN']),
    rules: z.array(z.object({
      fieldName: z.string(),
      mode: z.enum(['VISIBLE', 'READONLY', 'MASKED', 'HIDDEN']),
      maskPattern: z.string().optional(),
    })),
  }),
})
  .superRefine((data, ctx) => {
    // 行策略前提校验在 useCreatePermissionWizard 中处理，
    // 此处只做基础字段验证
  })
```

### 4.3 向导状态（useCreatePermissionWizard）

3 步向导使用 `useState` 管理步骤状态，不用 URL 路由（避免用户刷新丢失中间状态）：

```typescript
type WizardStep = 'model-action' | 'row-scope' | 'column-policy'

interface WizardState {
  step: WizardStep
  modelId: string | null
  modelFields: ModelField[]     // 从 Model 接口拉取的字段列表
  action: RBACAction | null
  rowScope: RowScope | null
  columnPolicy: ColumnPolicy    // { defaultMode, rules: ColumnRule[] }
  // 前提校验结果（Step 2 拉取）
  hasOwnerField: boolean
  hasDeptIdField: boolean
}
```

### 4.4 乐观更新

本期 RBAC 配置属于**低频管理操作**，以正确性优先，一律不做乐观更新。
唯一例外是 `UserDetailDrawer` 中角色/权限包的 toggle 交互，可在 `cache.modify` 局部更新
避免 Drawer 关闭重开，但仍以 `refetchQueries` 兜底。

---

## 5. 关键 UI 组件设计

### 5.1 列策略编辑器（`ColumnPolicyEditor`）

**用途**：创建权限点 Step 3，或编辑权限点时配置哪些字段可见/可编辑/脱敏。

**Props**：

```typescript
interface ColumnRule {
  fieldName: string
  mode: ColumnAccessMode  // VISIBLE / READONLY / MASKED / HIDDEN
  maskPattern?: string    // 脱敏模式时的掩码规则
}

interface ColumnPolicy {
  defaultMode: ColumnAccessMode  // 默认列访问模式
  rules: ColumnRule[]            // 逐字段覆盖规则
}

interface ColumnPolicyEditorProps {
  modelId: string
  fields: ModelField[]          // 从后端 Schema 获取的字段列表（Wave 1 mock）
  value: ColumnPolicy           // 受控（含 defaultMode + rules）
  onChange: (policy: ColumnPolicy) => void
}
```

**布局**（shadcn/ui `Table`）：

```
┌────────────────────────────────────────────────────────────────────┐
│ 默认模式: ○ 可见  ○ 只读  ○ 脱敏  ○ 隐藏                           │
├────────────────────────────────────────────────────────────────────┤
│ [全选/全不选] 字段名              列策略（覆盖默认）                 │
├────────────────────────────────────────────────────────────────────┤
│ ☑  id              ○ 可见  ○ 只读  ● 脱敏  ○ 隐藏                  │
│ ☑  created_at      ● 可见  ○ 只读  ○ 脱敏  ○ 隐藏                  │
│ ☑  status          ○ 可见  ● 只读  ○ 脱敏  ○ 隐藏                  │
│ □  phone           (未选中的字段 = 沿用默认模式，不写入 rules)      │
└────────────────────────────────────────────────────────────────────┘
```

**逻辑规则**：
- `defaultMode` 为整个权限点的默认列访问模式，未被 `rules` 覆盖的字段沿用此值
- 被 checkbox 选中的字段可在 `rules` 中覆盖其 `mode`
- `MASKED`（脱敏）时可配置 `maskPattern`（如 `"***-****"` 掩码规则）
- `HIDDEN` 表示该列对此权限点完全不可见
- 若 `rules` 为空，所有字段遵从 `defaultMode`

**样式约束**：
- 使用 `text-foreground` / `text-muted-foreground`（禁止 `text-gray-*`）
- 标题行 `font-semibold`（禁止 `font-bold`）
- RadioGroup 使用 `@radix-ui/react-radio-group` via shadcn

### 5.2 行策略选择器（`RowScopeSelector`）

**用途**：创建权限点 Step 2，选择行过滤范围并做前提校验。

**Props**：

```typescript
interface RowScopeSelectorProps {
  modelId: string
  value: RowScope | null
  onChange: (scope: RowScope) => void
  // 前提校验结果（由父组件 useCreatePermissionWizard 提前查询）
  hasOwnerField: boolean
  hasDeptIdField: boolean
}
```

**渲染逻辑**：

```
┌──────────────────────────────────────────────────────────┐
│ ● ALL              所有行，不过滤                          │
│                                                          │
│ ○ SELF             仅当前用户的行                         │
│   ⚠ 需要 owner (EndUserRef) 字段     [字段不存在时禁用]   │
│                                                          │
│ ○ DEPT             当前用户所在部门的行                   │
│   ⚠ 需要 dept_id 字段               [字段不存在时禁用]   │
│                                                          │
│ ○ DEPT_AND_CHILDREN  当前部门及所有下级部门的行           │
│   ⚠ 需要 dept_id 字段               [字段不存在时禁用]   │
└──────────────────────────────────────────────────────────┘
```

**校验规则**（严格遵循 PRD）：
- `SELF` 选项：`hasOwnerField === false` → `disabled` + Tooltip 说明"该 Model 缺少 `owner` 字段"
- `DEPT` / `DEPT_AND_CHILDREN`：`hasDeptIdField === false` → `disabled` + Tooltip
- 前端阻断优先，后端仍做后置校验（联调时对齐错误码）

**提前查询**：进入 Step 2 时，`useCreatePermissionWizard` 调用 Model 字段查询接口，
判断是否存在 `owner`（type = `EndUserRef`）和 `dept_id` 字段，结果缓存到向导状态。

```typescript
// 查询 Model 字段（Project-Scoped GraphQL，Wave 1 mock）
const { data: modelFieldsData } = useQuery(GET_MODEL_FIELDS, {
  variables: { modelId: wizardState.modelId },
  skip: !wizardState.modelId,
})

const hasOwnerField = modelFieldsData?.model.fields
  .some(f => f.name === 'owner' && f.type === 'EndUserRef') ?? false
const hasDeptIdField = modelFieldsData?.model.fields
  .some(f => f.name === 'dept_id') ?? false
```

> **注意**：`GET_MODEL_FIELDS` 走 Project-Scoped Client（而非 Org-Scoped），
> 需要在向导中额外处理 projectSlug 上下文。Wave 1 阶段用 MSW mock。

---

## 6. 实现顺序（Wave 划分）

### Wave 1 — BFF Mock + 类型定义（~3 天）

**目标**：打通类型体系，所有页面可用 mock 数据跑通，无需后端接口。

**任务清单**（可并行）：

| 任务 | 负责人方向 | 文件 |
|------|-----------|------|
| 1.1 RBAC 类型定义 | types | `src/types/rbac.ts` + `index.ts` 导出 |
| 1.2 GraphQL Operation 定义 | graphql | `src/web/graphql/queries/rbac.ts` + `mutations/rbac.ts` |
| 1.3 MSW mock 数据工厂 | mock | `src/mocks/data/org/rbac-factory.ts` |
| 1.4 MSW mock handler（org 域） | mock | `src/mocks/handlers/org/rbac.ts`（手写，codegen 后删除） |
| 1.5 `RowScopeSelector` 共享组件 | ui | `src/web/components/features/rbac/RowScopeSelector.tsx` |
| 1.6 `ColumnPolicyEditor` 共享组件 | ui | `src/web/components/features/rbac/ColumnPolicyEditor.tsx` |
| 1.7 `RBACSettingsLayout` Tab 导航 | ui | `src/app/org/[orgName]/projects/[projectSlug]/settings/rbac/layout.tsx` |

> 1.1 + 1.2 是串行前置，其余 1.3~1.7 可完全并行。

### Wave 2 — 权限点 + 权限包 CRUD（~5 天）

**前置**：Wave 1 完成。

**任务清单**（大部分可并行）：

| 任务 | 依赖 | 文件 |
|------|------|------|
| 2.1 权限点列表页 | 1.1/1.2/1.3 | `permissions/page.tsx` + `_hooks/usePermissionList.ts` |
| 2.2 创建权限点向导（3步） | 1.5/1.6 | `permissions/new/page.tsx` + `_hooks/useCreatePermissionWizard.ts` |
| 2.3 权限包列表页 | 1.1/1.2/1.3 | `bundles/page.tsx` + `_hooks/useBundleList.ts` |
| 2.4 权限包编辑页 | 2.3 | `bundles/[bundleId]/page.tsx` + `_hooks/useBundleEdit.ts` |
| 2.5 `BundlePermissionSelector` 共享组件 | 2.1 完成后 | `rbac/BundlePermissionSelector.tsx` |

> 2.1 与 2.3 完全并行；2.4 依赖 2.5，2.5 依赖 2.1 数据结构稳定后再做。

### Wave 3 — 角色管理 + 用户授权（~4 天）

**前置**：Wave 2 完成（需要权限包数据结构稳定）。

**任务清单**（大部分可并行）：

| 任务 | 依赖 | 文件 |
|------|------|------|
| 3.1 角色列表页（含隐式角色保护） | 2.x types | `roles/page.tsx` + `_hooks/useRoleList.ts` |
| 3.2 角色编辑页（关联权限包） | 3.1 + `RoleBundleSelector` | `roles/[roleId]/page.tsx` + `_hooks/useRoleEdit.ts` |
| 3.3 `RoleBundleSelector` 共享组件 | Wave 2 bundles 稳定 | `rbac/RoleBundleSelector.tsx` |
| 3.4 用户授权页（列表 + Drawer） | 3.1 | `users/page.tsx` + `_hooks/useUserAuth.ts` |
| 3.5 `EffectivePermView` 有效权限视图 | 3.4 | `rbac/EffectivePermView.tsx` |

> 3.1 + 3.4 可并行；3.2 依赖 3.3；3.5 依赖 3.4 中数据结构定义。

### Wave 并行图

```
Wave 1:  [types] → [ops] → [mock handler]
                         ↘ [RowScopeSelector]     ← 并行
                         ↘ [ColumnPolicyEditor]   ← 并行
                         ↘ [RBACSettingsLayout]   ← 并行

Wave 2:  [permission list] ←──────────────────────── 并行
         [permission wizard (step1→2→3)]
         [bundle list] ←────────────────────────── 并行
         [bundle edit + BundlePermissionSelector]  (等 permission list 稳定)

Wave 3:  [role list] ←──────────────────────────── 并行
         [role edit + RoleBundleSelector]
         [user auth page + EffectivePermView]       (等 bundle/role 稳定)
```

---

## 7. 与后端联调注意事项（Mock → Real 切换清单）

### 7.1 前置条件（联调开始前核对）

- [ ] 后端已新增 `api/graph/project/schema/rbac.graphql` 并实现所有 resolver
- [ ] 后端 `EndUserRole` 类型已增加 `isImplicit` / `bundles` 字段
- [ ] `front-contract-pull` 已拉取最新 GraphQL contract
- [ ] `npm run codegen` 已重新生成 `src/generated/graphql.ts`
- [ ] `src/mocks/handlers/org/rbac.ts` 手写 handler 已移除（或注释），改用 codegen 生成的 handler

### 7.2 逐接口切换清单

**权限点（Permissions）**

- [ ] `endUserPermissions` Query — 确认字段 `modelDisplayName` 是否后端冗余返回，或前端自行 join
- [ ] `createEndUserPermission` — 确认 `columnPolicy.rules` 为空时后端默认"全列可见"语义
- [ ] `deleteEndUserPermission` — 确认被权限包引用时后端返回的错误类型，前端对齐错误展示
- [ ] 行策略前提校验 — 当 Model 缺少 `owner`/`dept_id` 时，后端 `createEndUserPermission` 返回的 error union 类型

**权限包（Bundles）**

- [ ] `endUserBundles` Query — 确认 `permissions` 是否嵌套返回，或需单独 query（N+1 风险）
- [ ] `addPermissionToBundle` / `removePermissionFromBundle` — 确认幂等行为
- [ ] `deleteEndUserBundle` — 确认被角色/用户引用时的错误类型

**角色（Roles）**

- [ ] `endUserRoles(includeImplicit: true)` — 确认隐式角色是否默认隐藏
- [ ] `deleteEndUserRole` — 后端需保护 `isImplicit=true` 角色不可删；前端已有保护，联调时验证后端也拒绝
- [ ] `isImplicit` 字段语义 — RBAC 模型中只有 `isImplicit`（内置隐式角色标志），**`isSystem` 字段不存在于 RBAC 模型中**，勿与 Casbin 系统角色混淆

**用户授权（Users）**

- [ ] `assignRoleToEndUser` vs 现有 `assignRoleToUser` — 确认是复用还是新增（避免重复接口）
- [ ] `addBundleToUser` / `removeBundleFromUser` — 现有 schema 无此接口，需后端新增
- [ ] `endUserEffectivePermissions` — 确认三通道合并逻辑在后端执行，`sources` 字段格式对齐

**Model 字段查询（RowScope 前提校验）**

- [ ] 权限点向导 Step 2 需要查 Model 字段（走 Project-Scoped client），确认字段 type 的枚举值
- [ ] `EndUserRef` 字段的 `type` 值后端如何表示（枚举名 / 字符串），前端 `hasOwnerField` 判断条件需对齐

### 7.3 UI 行为验证清单

- [ ] 隐式角色在"分配给用户"选择器中**不出现**（前端 filter `isImplicit === true`）
- [ ] 隐式角色行**无删除按钮**，且名称/描述**不可编辑**（`isImplicit === true` 时 form disabled）
- [ ] `rowScope = SELF` 且 Model 无 `owner` 字段时，`RowScopeSelector` 该 radio 项灰显且 Tooltip 说明原因
- [ ] `rowScope = DEPT / DEPT_AND_CHILDREN` 且 Model 无 `dept_id` 时，同上
- [ ] 权限包被角色关联后删除权限包，前端需刷新角色关联列表

### 7.4 环境切换方式

```bash
# 开发阶段（使用 MSW mock）
NEXT_PUBLIC_API_MOCKING=enabled  # .env.local

# 联调阶段（使用真实后端）
# 删除或置空 NEXT_PUBLIC_API_MOCKING
# BFF 代码无需修改
```

---

## 附录：目录结构速览

```
src/
├── app/org/[orgName]/projects/[projectSlug]/settings/rbac/
│   ├── layout.tsx                     # RBACSettingsLayout
│   ├── page.tsx                       # redirect
│   ├── bundles/
│   │   ├── page.tsx                   # BundleListPage
│   │   └── [bundleId]/page.tsx        # BundleEditPage
│   ├── permissions/
│   │   ├── page.tsx                   # PermissionListPage
│   │   └── new/page.tsx               # CreatePermissionWizard
│   ├── roles/
│   │   ├── page.tsx                   # RoleListPage
│   │   └── [roleId]/page.tsx          # RoleEditPage
│   └── users/
│       └── page.tsx                   # UserAuthPage
│
├── web/
│   ├── components/features/rbac/
│   │   ├── RowScopeSelector.tsx       # 共享：行策略选择器
│   │   ├── ColumnPolicyEditor.tsx     # 共享：列策略编辑器
│   │   ├── BundlePermissionSelector.tsx
│   │   ├── RoleBundleSelector.tsx
│   │   ├── ImplicitRoleBadge.tsx
│   │   ├── EffectivePermView.tsx
│   │   ├── PermissionRowSummary.tsx
│   │   └── index.ts
│   ├── graphql/
│   │   ├── queries/rbac.ts            # 所有 RBAC query documents
│   │   └── mutations/rbac.ts          # 所有 RBAC mutation documents
│   └── hooks/rbac/                    # 全局复用 hooks（如有）
│
├── types/
│   └── rbac.ts                        # EndUserPermission, EndUserPermissionBundle, EndUserRole 等业务类型
│
└── mocks/
    ├── data/org/rbac-factory.ts       # mock 数据工厂
    └── handlers/org/rbac.ts           # 手写 mock handler（Wave 1 阶段）
```
