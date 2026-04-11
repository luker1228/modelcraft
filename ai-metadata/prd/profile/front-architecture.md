# Profile 前端架构方案（front-architecture）

> 范围：仅做前端架构与骨架设计，不包含实现细节。  
> 依据：`ai-metadata/front/development/*`、`ai-metadata/front/style/quick-start.md`、`ai-metadata/prd/profile/*`、`ai-metadata/prd/profile/api-contract.md`。

---

## 1. 模块划分

### 1.1 ProfileRoute（页面路由编排）
- 职责：承载 Profile 页面入口与子路由编排（概览/编辑），只做组合与路由跳转
- 所属层：App
- 依赖：Web（页面私有 hooks/components）

### 1.2 ProfilePageModule（Profile 页面私有模块）
- 职责：实现 Profile 概览与编辑页的页面级交互状态编排
- 所属层：Web（页面私有 `_components` / `_hooks`）
- 依赖：Web 全局 hooks、`@/types`

### 1.3 ProfileDomainHooks（全局业务 hooks）
- 职责：封装 `myUserProfile` 查询与 `updateMyProfile` 变更，统一 GraphQL payload 解析
- 所属层：Web
- 依赖：`web/graphql/*`、`@/generated/graphql`、Shared

### 1.4 ProfileGraphQLOperations（GraphQL 操作定义）
- 职责：维护 Profile 域 query/mutation 文档，供 codegen 生成类型
- 所属层：Web
- 依赖：contract schema（只读）

### 1.5 AuthRegisterBffAdapter（注册 BFF 适配）
- 职责：适配 `/api/bff/auth/register`，透传并标准化后端新增的 `profile` 快照字段
- 所属层：BFF
- 依赖：`bff/auth/go-auth-client.ts`、`src/types/auth.ts`

### 1.6 ProfileMockContract（Mock 契约层）
- 职责：基于 codegen 产物补齐 Profile 场景 mock 数据工厂与错误场景样例（含 `ProfileNotFound`）
- 所属层：Web（mocks）
- 依赖：`src/mocks/handlers/org/generated.ts`、`src/generated/graphql.ts`

### 1.7 ProfileTypes（领域类型层）
- 职责：定义前端 Profile 领域 view model / form model / error model 骨架
- 所属层：Shared（`src/types`）
- 依赖：无

---

## 2. 本次涉及目录结构（仅列变更路径）

```text
modelcraft-front/src/
├── app/
│   ├── org/[orgName]/profile/
│   │   ├── page.tsx
│   │   ├── edit/page.tsx
│   │   ├── _components/
│   │   │   ├── ProfileOverviewPanel.tsx
│   │   │   ├── ProfileEditForm.tsx
│   │   │   └── index.ts
│   │   └── _hooks/
│   │       ├── use-profile-page-state.ts
│   │       ├── use-profile-edit-form.ts
│   │       ├── use-profile-page-data.ts
│   │       ├── types.ts
│   │       └── index.ts
│   └── api/bff/auth/register/route.ts                 # 修改
├── web/
│   ├── graphql/
│   │   ├── queries/profile.ts
│   │   └── mutations/profile.ts
│   ├── hooks/user/
│   │   ├── use-my-user-profile.ts
│   │   └── use-update-my-profile.ts
│   └── graphql/{queries,mutations}/index.ts            # 修改导出
├── bff/auth/go-auth-client.ts                          # 修改（register 响应契约）
├── mocks/
│   ├── data/org/profile-factory.ts
│   └── handlers/index.ts                               # 修改（挂载 profile handlers）
└── types/
    ├── profile.ts
    ├── auth.ts                                         # 修改（RegisterResponse 扩展 profile）
    └── index.ts                                        # 修改（re-export profile）
```

---

## 3. BFF 接口与 Mock 数据契约

## 3.1 BFF 接口（注册链路）

### `POST /api/bff/auth/register`
- 请求：`phone + userName + password`
- 响应：`userId + orgName + profile(snapshot)`
- 错误映射：
  - `400` 参数错误（`PARAM_INVALID.AUTH`）
  - `409` 用户冲突（`CONFLICT.USER`）
  - `500` 注册联动失败（含 profile 初始化失败）

### BFF 层契约骨架
```ts
// src/types/auth.ts
export interface RegisterProfileSnapshot {
  id: string
  userId: string
  nickname: string
  avatarUrl?: string
  bio?: string
}

export interface RegisterResponse {
  userId: string
  orgName: string
  profile: RegisterProfileSnapshot
}

// src/bff/auth/go-auth-client.ts
export interface GoRegisterResult {
  userId: string
  orgName: string
  profile: RegisterProfileSnapshot
}
```

## 3.2 GraphQL 接口（Profile 域）

- 查询：`myUserProfile`
- 变更：`updateMyProfile(input: UpdateMyProfileInput!)`
- 错误 union：`ProfileNotFound`、`InvalidProfileInput`（按 contract）

## 3.3 Mock 数据契约

### 目标
1. 成功态：返回完整 `user + profile`
2. 缺失态：返回 `ProfileNotFound`
3. 更新失败态：返回 `InvalidProfileInput`

### Mock 工厂骨架
```ts
// src/mocks/data/org/profile-factory.ts
export interface MockMyUserProfileScenario {
  type: 'success' | 'profileNotFound' | 'invalidInput'
}

export function createMockMyUserProfilePayload(scenario: MockMyUserProfileScenario) {
  // TODO: worker 实现
}

export function createMockUpdateMyProfilePayload(scenario: MockMyUserProfileScenario) {
  // TODO: worker 实现
}
```

---

## 4. TypeScript 类型骨架

## 4.1 全局共享类型（`src/types/profile.ts`）

```ts
export type ProfileLoadStatus = 'idle' | 'loading' | 'success' | 'error'

export interface UserProfileView {
  userId: string
  phone: string
  userName: string
  status: 'REGISTERED' | 'ACTIVE' | 'SUSPENDED'
  profileId: string
  nickname: string
  avatarUrl?: string
  bio?: string
  createdAt: string
  updatedAt: string
}

export interface UpdateMyProfileFormValues {
  nickname?: string
  avatarUrl?: string
  bio?: string
}

export interface ProfileDomainError {
  type: 'UserNotFound' | 'ProfileNotFound' | 'InvalidProfileInput' | 'Unknown'
  message: string
  suggestion?: string
}
```

## 4.2 页面私有类型（`app/org/[orgName]/profile/_hooks/types.ts`）

```ts
import type { UserProfileView, UpdateMyProfileFormValues, ProfileDomainError } from '@/types'

export interface ProfilePageState {
  mode: 'view' | 'edit'
  saving: boolean
}

export interface ProfilePageData {
  profile: UserProfileView | null
  loading: boolean
  error: ProfileDomainError | null
}

export interface UseProfileEditFormReturn {
  initialValues: UpdateMyProfileFormValues
  submit: (values: UpdateMyProfileFormValues) => Promise<void>
  reset: () => void
}
```

## 4.3 全局 Hook 签名（`web/hooks/user/*`）

```ts
import type { ProfileDomainError, UpdateMyProfileFormValues, UserProfileView } from '@/types'

export interface UseMyUserProfileReturn {
  data: UserProfileView | null
  loading: boolean
  error: ProfileDomainError | null
  refetch: () => Promise<void>
}

export function useMyUserProfile(): UseMyUserProfileReturn {
  // TODO: worker 实现
}

export interface UseUpdateMyProfileReturn {
  mutate: (input: UpdateMyProfileFormValues) => Promise<UserProfileView | null>
  loading: boolean
  error: ProfileDomainError | null
}

export function useUpdateMyProfile(): UseUpdateMyProfileReturn {
  // TODO: worker 实现
}
```

## 4.4 页面组件签名（`_components/*`）

```tsx
import type { UserProfileView, UpdateMyProfileFormValues } from '@/types'

export interface ProfileOverviewPanelProps {
  profile: UserProfileView
  onEdit: () => void
}

export function ProfileOverviewPanel({ profile, onEdit }: ProfileOverviewPanelProps): JSX.Element {
  // TODO: worker 实现
}

export interface ProfileEditFormProps {
  initialValues: UpdateMyProfileFormValues
  saving: boolean
  onSubmit: (values: UpdateMyProfileFormValues) => Promise<void>
  onCancel: () => void
}

export function ProfileEditForm({ initialValues, saving, onSubmit, onCancel }: ProfileEditFormProps): JSX.Element {
  // TODO: worker 实现
}
```

---

## 5. 页面/组件分层与状态流

## 5.1 页面分层
- `app/org/[orgName]/profile/page.tsx`：Profile 概览入口（组合组件）
- `app/org/[orgName]/profile/edit/page.tsx`：Profile 编辑入口（组合组件）
- 页面私有逻辑集中在 `_hooks`，避免把业务逻辑留在 page.tsx

## 5.2 状态流（单向）
1. 页面初始化 -> `useMyUserProfile` 查询 `myUserProfile`
2. Hook 将 GraphQL payload 映射为 `UserProfileView`
3. `ProfileOverviewPanel` 渲染数据
4. 进入编辑页 -> `ProfileEditForm` 使用 RHF + Zod（worker 实现）
5. 提交 -> `useUpdateMyProfile` 调 mutation
6. 成功后：
   - 刷新 `myUserProfile` 查询缓存
   - 回到概览并展示最新数据
7. 失败时按 union 错误类型转换为统一 `ProfileDomainError`

## 5.3 与现有布局的关系
- 继续复用 `AppLayout`
- 通过 `orgName` 上下文走 Org GraphQL endpoint，不新建跨层依赖
- `UserMenu` 可增加「个人资料」入口（仅路由跳转，避免业务耦合）

---

## 6. 与现有 contract 的对接点

## 6.1 合同输入源
- GraphQL：`contract/graph/org/schema/*.graphql`（profile 类型与 operation 依赖此处）
- OpenAPI：`contract/openapi/auth.yaml`（register 响应新增 `profile`）

## 6.2 前端落点
1. `web/graphql/queries/profile.ts` 定义 `MY_USER_PROFILE`
2. `web/graphql/mutations/profile.ts` 定义 `UPDATE_MY_PROFILE`
3. 执行 `npm run codegen`，生成：
   - `src/generated/graphql.ts`
   - `src/mocks/handlers/org/generated.ts`

## 6.3 当前仓库前置说明
- 若前端仓库尚未同步 `contract/` 目录，先完成 contract 同步后再执行 codegen（禁止手改 contract）。

---

## 7. 需要新增/修改的文件清单

## 新增
- `modelcraft-front/src/app/org/[orgName]/profile/page.tsx`
- `modelcraft-front/src/app/org/[orgName]/profile/edit/page.tsx`
- `modelcraft-front/src/app/org/[orgName]/profile/_components/ProfileOverviewPanel.tsx`
- `modelcraft-front/src/app/org/[orgName]/profile/_components/ProfileEditForm.tsx`
- `modelcraft-front/src/app/org/[orgName]/profile/_components/index.ts`
- `modelcraft-front/src/app/org/[orgName]/profile/_hooks/use-profile-page-state.ts`
- `modelcraft-front/src/app/org/[orgName]/profile/_hooks/use-profile-page-data.ts`
- `modelcraft-front/src/app/org/[orgName]/profile/_hooks/use-profile-edit-form.ts`
- `modelcraft-front/src/app/org/[orgName]/profile/_hooks/types.ts`
- `modelcraft-front/src/app/org/[orgName]/profile/_hooks/index.ts`
- `modelcraft-front/src/web/graphql/queries/profile.ts`
- `modelcraft-front/src/web/graphql/mutations/profile.ts`
- `modelcraft-front/src/web/hooks/user/use-my-user-profile.ts`
- `modelcraft-front/src/web/hooks/user/use-update-my-profile.ts`
- `modelcraft-front/src/types/profile.ts`
- `modelcraft-front/src/mocks/data/org/profile-factory.ts`

## 修改
- `modelcraft-front/src/web/graphql/queries/index.ts`
- `modelcraft-front/src/web/graphql/mutations/index.ts`
- `modelcraft-front/src/types/index.ts`
- `modelcraft-front/src/types/auth.ts`
- `modelcraft-front/src/bff/auth/go-auth-client.ts`
- `modelcraft-front/src/app/api/bff/auth/register/route.ts`
- `modelcraft-front/src/mocks/handlers/index.ts`
- `modelcraft-front/src/web/components/features/layout/UserMenu.tsx`（仅入口跳转）

---

## 8. front-worker 并行开发任务拆分建议

## Task A：Contract 对齐 + GraphQL 操作定义
- 产物：`web/graphql/queries/profile.ts`、`web/graphql/mutations/profile.ts`、index 导出
- 依赖：contract 已同步

## Task B：类型层与 BFF 注册契约升级
- 产物：`types/profile.ts`、`types/auth.ts`、`go-auth-client.ts`、`/api/bff/auth/register`
- 依赖：OpenAPI contract 同步

## Task C：Profile 全局 Hooks
- 产物：`use-my-user-profile.ts`、`use-update-my-profile.ts`
- 依赖：Task A + codegen 产物

## Task D：页面骨架与私有模块
- 产物：`app/org/[orgName]/profile/**`（page + `_components` + `_hooks`）
- 依赖：Task C

## Task E：Mock 场景补齐
- 产物：`mocks/data/org/profile-factory.ts`、handlers 汇总
- 依赖：Task A + codegen

## Task F：入口集成与联调回归
- 产物：UserMenu/Profile 路由入口、基本联调验证
- 依赖：Task B + C + D

---

## 9. Worker 实现注意事项（必须遵守）

1. **只读 contract**：禁止修改 `contract/`，仅通过同步机制更新。  
2. **GraphQL 类型来源**：禁止手写操作类型，统一从 `@/generated/graphql` 导入。  
3. **Hook 放置规则**：全局 hook 放 `web/hooks/user/`，页面私有放 `_hooks/`。  
4. **类型放置规则**：新增全局类型放 `src/types/profile.ts`，并在 `src/types/index.ts` re-export。  
5. **页面拆分规则**：`page.tsx` 只做组合，业务逻辑下沉 `_hooks`。  
6. **UI 约束**：基础控件使用 `@/components/ui`（shadcn/ui），不要自建原子组件。  
7. **mock 策略**：头像能力先走 mock，不阻塞主链路；需覆盖 `ProfileNotFound` 错误场景。  
8. **codegen 必做**：新增 query/mutation 或 contract 更新后执行 `npm run codegen`。  
9. **BFF 错误映射**：注册接口需保留现有错误语义并扩展 profile 快照字段，避免破坏登录注册主流程。