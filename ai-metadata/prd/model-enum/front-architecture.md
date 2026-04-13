# Model Enum 前端架构方案（front-architecture）

> 目标：仅产出前端架构与可执行骨架，不写实现细节。  
> 依据：
> - `ai-metadata/prd/model-enum/00-model-enum.md`
> - `ai-metadata/prd/model-enum/01-field-create-enum-binding.md`
> - `ai-metadata/prd/model-enum/02-field-edit-format-immutable.md`
> - `ai-metadata/prd/model-enum/04-frontend-subpage-design.md`
> - `ai-metadata/prd/model-enum/api-contract.md`
> - `ai-metadata/front/development/{architecture,bff-design,code-conventions,react-best-practices,typescript-guide}.md`
> - `ai-metadata/front/style/quick-start.md`

---

## 1. 约束对齐（必须落地）

1. `ENUM` 创建必须使用 `relateEnumName`。  
2. `ENUM_LABEL` 创建必须使用 `enumRelationId`。  
3. 参数缺失/参数非法统一映射为 `InvalidInput`（`InvalidInput` / `InvalidInput`）。  
4. source 唯一冲突统一为 `FIELD_ENUM_SOURCE_CONFLICT`。  
5. 字段 `format` 编辑不可变（`FIELD_FORMAT_IMMUTABLE`）。

---

## 2. 模块划分

### 2.1 ModelEnumFieldPageOrchestrator
- 职责：承载「创建 ENUM / 创建 ENUM_LABEL / 编辑字段」三类子页编排，仅做页面组合
- 所属层：App
- 依赖：Web（页面私有 `_components` + `_hooks`）

### 2.2 ModelEnumFieldPages
- 职责：三类子页 UI 组件（创建 ENUM、创建 ENUM_LABEL、编辑只读 format）
- 所属层：Web（页面私有）
- 依赖：页面私有 hooks、`@/components/ui`

### 2.3 ModelEnumPageHooks
- 职责：页面级状态编排（表单态、加载态、保存态、错误态）
- 所属层：Web（页面私有）
- 依赖：Web 全局 hooks、`@/types`

### 2.4 ModelEnumDomainHooks
- 职责：封装 model-enum 领域动作（创建 ENUM、创建 ENUM_LABEL、编辑元信息、relation 查询/创建）
- 所属层：Web
- 依赖：BFF 门面 `@bff/model-enum/public`、`@/generated/graphql`

### 2.5 ModelEnumBffFacade
- 职责：对 Web 提供唯一 BFF 出口（门面），屏蔽 mock/real 数据源切换细节
- 所属层：BFF
- 依赖：BFF 内部 mock client、后续 real client（预留）

### 2.6 ModelEnumMockGateway
- 职责：前端阶段 mock 数据访问协议（不依赖真实后端）
- 所属层：BFF
- 依赖：`src/types/model-enum.ts`

### 2.7 ModelEnumTypes
- 职责：沉淀 model-enum 领域类型、错误模型、页面表单模型
- 所属层：Shared（`src/types`）
- 依赖：无

### 2.8 ModelEnumErrorMapper
- 职责：将 GraphQL/BFF 错误统一映射为前端可消费错误语义
- 所属层：Shared
- 依赖：`@/types/model-enum`、`@/generated/graphql`

---

## 3. 本次涉及目录结构（仅列变更路径）

```text
modelcraft-front/src/
├── app/org/[orgName]/projects/[projectSlug]/model-editor/
│   ├── _components/
│   │   ├── field-pages/
│   │   │   ├── CreateEnumFieldPage.tsx
│   │   │   ├── CreateEnumLabelFieldPage.tsx
│   │   │   ├── EditFieldImmutablePage.tsx
│   │   │   ├── EnumRelationSelector.tsx
│   │   │   └── index.ts
│   │   ├── FieldEditSheet.tsx                      # 修改（接入不可变编辑页）
│   │   └── ModelDetailPanel.tsx                    # 修改（挂载子页入口）
│   └── _hooks/
│       ├── use-create-enum-field-page.ts
│       ├── use-create-enum-label-field-page.ts
│       ├── use-edit-field-page.ts
│       ├── useFieldOperations.ts                   # 修改（切换到领域 hook）
│       ├── types.ts                                # 修改（新增页面私有类型）
│       └── index.ts                                # 修改（导出新增 hooks）
├── web/
│   ├── hooks/model/
│   │   └── use-model-enum-field.ts
│   └── graphql/
│       ├── queries/
│       │   ├── field-enum-relation.ts
│       │   └── index.ts                            # 修改导出
│       └── mutations/
│           ├── field-enum-relation.ts
│           ├── model.ts                            # 修改 AddFields/UpdateField 选择字段
│           └── index.ts                            # 修改导出
├── bff/model-enum/
│   ├── public.ts
│   ├── types.ts
│   └── mock-client.ts
├── shared/errors/
│   └── model-enum-error-mapper.ts
└── types/
    ├── model-enum.ts
    └── index.ts                                    # 修改 re-export
```

---

## 4. BFF Mock 接口定义（前端阶段唯一数据入口）

## 4.1 门面导出（`src/bff/model-enum/public.ts`）

```ts
import type {
  ModelEnumContextQuery,
  CreateEnumFieldCommand,
  CreateEnumLabelFieldCommand,
  UpdateFieldMetaCommand,
  CreateFieldEnumRelationCommand,
  ModelEnumActionResult,
  ModelEnumContextResult,
  FieldEnumRelationListResult,
} from './types'

export function queryModelEnumContext(query: ModelEnumContextQuery): Promise<ModelEnumContextResult> {
  // TODO: worker 实现
}

export function createEnumField(command: CreateEnumFieldCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function createEnumLabelField(command: CreateEnumLabelFieldCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function updateFieldMeta(command: UpdateFieldMetaCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function listFieldEnumRelations(query: ModelEnumContextQuery): Promise<FieldEnumRelationListResult> {
  // TODO: worker 实现
}

export function createFieldEnumRelation(command: CreateFieldEnumRelationCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}
```

## 4.2 BFF 契约类型（`src/bff/model-enum/types.ts`）

```ts
import type { ValidationConfigInput } from '@/generated/graphql'
import type {
  EnumSourceOption,
  EnumRelationOption,
  ModelEnumDomainError,
} from '@/types'

export interface ModelEnumContextQuery {
  orgName: string
  projectSlug: string
  modelId: string
}

export interface CreateEnumFieldCommand {
  orgName: string
  projectSlug: string
  modelId: string
  name: string
  title: string
  description?: string
  relateEnumName: string
}

export interface CreateEnumLabelFieldCommand {
  orgName: string
  projectSlug: string
  modelId: string
  name: string
  title: string
  description?: string
  sourceFieldName: string
  enumRelationId: string
}

export interface UpdateFieldMetaCommand {
  orgName: string
  projectSlug: string
  modelId: string
  fieldName: string
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}

export interface CreateFieldEnumRelationCommand {
  orgName: string
  projectSlug: string
  modelId: string
  sourceFieldName: string
  enumName: string
  labelFieldName: string
}

export interface ModelEnumContextResult {
  enumSources: EnumSourceOption[]
  relations: EnumRelationOption[]
  error: ModelEnumDomainError | null
}

export interface FieldEnumRelationListResult {
  relations: EnumRelationOption[]
  error: ModelEnumDomainError | null
}

export interface ModelEnumActionResult {
  success: boolean
  error: ModelEnumDomainError | null
}
```

## 4.3 Mock Client（`src/bff/model-enum/mock-client.ts`）

```ts
import type {
  ModelEnumContextQuery,
  CreateEnumFieldCommand,
  CreateEnumLabelFieldCommand,
  UpdateFieldMetaCommand,
  CreateFieldEnumRelationCommand,
  ModelEnumContextResult,
  FieldEnumRelationListResult,
  ModelEnumActionResult,
} from './types'

export function mockQueryModelEnumContext(query: ModelEnumContextQuery): Promise<ModelEnumContextResult> {
  // TODO: worker 实现
}

export function mockCreateEnumField(command: CreateEnumFieldCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function mockCreateEnumLabelField(command: CreateEnumLabelFieldCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function mockUpdateFieldMeta(command: UpdateFieldMetaCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}

export function mockListFieldEnumRelations(query: ModelEnumContextQuery): Promise<FieldEnumRelationListResult> {
  // TODO: worker 实现
}

export function mockCreateFieldEnumRelation(command: CreateFieldEnumRelationCommand): Promise<ModelEnumActionResult> {
  // TODO: worker 实现
}
```

---

## 5. TypeScript 类型骨架

## 5.1 全局类型（`src/types/model-enum.ts`）

```ts
import type { ValidationConfigInput } from '@/generated/graphql'

export type ModelEnumErrorType =
  | 'InvalidInput'
  | 'InvalidInput'
  | 'FieldEnumSourceConflict'
  | 'FieldFormatImmutable'
  | 'FieldReferenceInUse'
  | 'Unknown'

export type ModelEnumErrorCode =
  | 'FIELD_ENUM_SOURCE_CONFLICT'
  | 'FIELD_FORMAT_IMMUTABLE'
  | 'FIELD_REFERENCE_IN_USE'
  | 'UNKNOWN'

export interface ModelEnumDomainError {
  type: ModelEnumErrorType
  code?: ModelEnumErrorCode
  message: string
  suggestion?: string
}

export interface EnumSourceOption {
  fieldName: string
  title: string
  enumName: string
  occupied: boolean
}

export interface EnumRelationOption {
  id: string
  sourceFieldName: string
  enumName: string
  labelFieldName: string
}

export interface CreateEnumFieldFormValues {
  name: string
  title: string
  description?: string
  relateEnumName: string
}

export interface CreateEnumLabelFieldFormValues {
  name: string
  title: string
  description?: string
  sourceFieldName: string
  enumRelationId: string
}

export interface UpdateFieldMetaFormValues {
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}
```

## 5.2 类型导出（`src/types/index.ts`）

```ts
export * from './model-enum'
```

## 5.3 页面私有类型（`app/.../model-editor/_hooks/types.ts` 增量）

```ts
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  UpdateFieldMetaFormValues,
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
} from '@/types'

export interface EnumFieldPageState {
  saving: boolean
  error: ModelEnumDomainError | null
}

export interface CreateEnumFieldPageModel {
  defaults: CreateEnumFieldFormValues
  enumOptions: string[]
  state: EnumFieldPageState
}

export interface CreateEnumLabelFieldPageModel {
  defaults: CreateEnumLabelFieldFormValues
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  state: EnumFieldPageState
}

export interface EditFieldImmutablePageModel {
  fieldName: string
  format: string
  relateEnumName?: string
  enumRelationId?: string
  editable: UpdateFieldMetaFormValues
  state: EnumFieldPageState
}
```

---

## 6. 页面/组件/hooks 分层与职责

## 6.1 页面私有 hooks（`app/.../_hooks/*.ts`）

```ts
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  UpdateFieldMetaFormValues,
  ModelEnumDomainError,
} from '@/types'

export interface UseCreateEnumFieldPageReturn {
  submit: (values: CreateEnumFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumFieldPage(): UseCreateEnumFieldPageReturn {
  // TODO: worker 实现
}

export interface UseCreateEnumLabelFieldPageReturn {
  submit: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumLabelFieldPage(): UseCreateEnumLabelFieldPageReturn {
  // TODO: worker 实现
}

export interface UseEditFieldPageReturn {
  submit: (values: UpdateFieldMetaFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useEditFieldPage(): UseEditFieldPageReturn {
  // TODO: worker 实现
}
```

## 6.2 全局领域 hook（`src/web/hooks/model/use-model-enum-field.ts`）

```ts
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  UpdateFieldMetaFormValues,
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
} from '@/types'

export interface UseModelEnumContextParams {
  orgName: string
  projectSlug: string
  modelId: string
}

export interface UseModelEnumContextReturn {
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  loading: boolean
  error: ModelEnumDomainError | null
  refetch: () => Promise<void>
}

export function useModelEnumContext(params: UseModelEnumContextParams): UseModelEnumContextReturn {
  // TODO: worker 实现
}

export interface UseCreateEnumFieldReturn {
  mutate: (values: CreateEnumFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumField(params: UseModelEnumContextParams): UseCreateEnumFieldReturn {
  // TODO: worker 实现
}

export interface UseCreateEnumLabelFieldReturn {
  mutate: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useCreateEnumLabelField(params: UseModelEnumContextParams): UseCreateEnumLabelFieldReturn {
  // TODO: worker 实现
}

export interface UseUpdateFieldMetaReturn {
  mutate: (fieldName: string, values: UpdateFieldMetaFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
}

export function useUpdateFieldMeta(params: UseModelEnumContextParams): UseUpdateFieldMetaReturn {
  // TODO: worker 实现
}
```

## 6.3 页面私有组件（`app/.../_components/field-pages/*.tsx`）

```tsx
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  UpdateFieldMetaFormValues,
  EnumRelationOption,
  EnumSourceOption,
} from '@/types'

export interface CreateEnumFieldPageProps {
  enumOptions: string[]
  loading: boolean
  onSubmit: (values: CreateEnumFieldFormValues) => Promise<void>
  onCancel: () => void
}

export function CreateEnumFieldPage(props: CreateEnumFieldPageProps): JSX.Element {
  // TODO: worker 实现
}

export interface CreateEnumLabelFieldPageProps {
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  loading: boolean
  onCreateRelation: (sourceFieldName: string) => Promise<void>
  onSubmit: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  onCancel: () => void
}

export function CreateEnumLabelFieldPage(props: CreateEnumLabelFieldPageProps): JSX.Element {
  // TODO: worker 实现
}

export interface EditFieldImmutablePageProps {
  fieldName: string
  format: string
  relateEnumName?: string
  enumRelationId?: string
  loading: boolean
  onSubmit: (values: UpdateFieldMetaFormValues) => Promise<void>
  onCancel: () => void
}

export function EditFieldImmutablePage(props: EditFieldImmutablePageProps): JSX.Element {
  // TODO: worker 实现
}
```

---

## 7. 子页联动流程（创建 ENUM / 创建 ENUM_LABEL / 编辑）

## 7.1 创建 ENUM 字段
1. 页面进入「创建 ENUM」子页，加载 enum 下拉选项。  
2. 用户提交时仅允许 `format=ENUM + relateEnumName`。  
3. 调用 `createEnumField`（BFF 门面）。  
4. 成功：关闭子页并刷新字段列表。  
5. 失败：按错误映射展示（`InvalidInput` 等）。

## 7.2 创建 ENUM_LABEL 字段
1. 页面进入「创建 ENUM_LABEL」子页，加载 source ENUM 字段 + relation 列表。  
2. 用户选择 source 后，选择已有 relation 或触发新建 relation。  
3. 提交时仅允许 `format=ENUM_LABEL + enumRelationId`。  
4. 若 source 已占用，前端先阻断；后端兜底返回 `FIELD_ENUM_SOURCE_CONFLICT`。  
5. 成功后刷新字段与 relation 列表。

## 7.3 编辑字段（format 不可变）
1. 打开编辑子页时，`format` / `relateEnumName` / `enumRelationId` 仅只读展示。  
2. 仅允许提交 `title` / `description` / `validationConfig`。  
3. 若后端返回 `FIELD_FORMAT_IMMUTABLE`，提示并保持只读态。

---

## 8. 错误态映射策略（含 InvalidInput）

| 上游错误 | 识别字段 | 前端领域错误 | UI 处理策略 |
|---|---|---|---|
| `InvalidInput` | `__typename` | `type=InvalidInput` | 顶部错误提示 + 阻断提交 |
| `InvalidInput` | `__typename` | `type=InvalidInput` | 表单错误提示（字段级 + 表单级） |
| `FieldEnumSourceConflict` | `code=FIELD_ENUM_SOURCE_CONFLICT` | `type=FieldEnumSourceConflict` | source 标记占用并提示冲突 |
| `FieldFormatImmutable` | `code=FIELD_FORMAT_IMMUTABLE` | `type=FieldFormatImmutable` | 强制维持只读 + 提示不可变 |
| `FieldReferenceInUse` | `code=FIELD_REFERENCE_IN_USE` | `type=FieldReferenceInUse` | 阻断删除/更新并提示引用关系 |
| 未知错误 | fallback | `type=Unknown` | 通用错误提示 + 保留输入态 |

### 错误映射器骨架（`src/shared/errors/model-enum-error-mapper.ts`）

```ts
import type { ModelEnumDomainError } from '@/types'

export interface RawGraphQLErrorLike {
  __typename?: string
  code?: string
  message?: string
  suggestion?: string
}

export function mapModelEnumError(raw: RawGraphQLErrorLike | null | undefined): ModelEnumDomainError | null {
  // TODO: worker 实现
}
```

---

## 9. GraphQL 操作骨架（仅定义，不实现逻辑）

## 9.1 Query（`src/web/graphql/queries/field-enum-relation.ts`）

```ts
import { gql } from '@apollo/client'

export const GET_FIELD_ENUM_RELATIONS = gql`
  query GetFieldEnumRelations($modelID: ID!) {
    fieldEnumRelations(modelID: $modelID) {
      id
      modelId
      sourceFieldName
      labelFieldName
      enumName
      createdAt
      updatedAt
    }
  }
`
```

## 9.2 Mutation（`src/web/graphql/mutations/field-enum-relation.ts`）

```ts
import { gql } from '@apollo/client'

export const CREATE_FIELD_ENUM_RELATION = gql`
  mutation CreateFieldEnumRelation($input: CreateFieldEnumRelationInput!) {
    createFieldEnumRelation(input: $input) {
      relation {
        id
        sourceFieldName
        labelFieldName
        enumName
      }
      error {
        __typename
        ... on InvalidInput { message suggestion }
        ... on InvalidInput { message suggestion }
        ... on FieldEnumSourceConflict { message code suggestion }
      }
    }
  }
`
```

---

## 10. 需要新增/修改的文件清单

## 新增
- `modelcraft-front/src/types/model-enum.ts`
- `modelcraft-front/src/bff/model-enum/public.ts`
- `modelcraft-front/src/bff/model-enum/types.ts`
- `modelcraft-front/src/bff/model-enum/mock-client.ts`
- `modelcraft-front/src/web/hooks/model/use-model-enum-field.ts`
- `modelcraft-front/src/web/graphql/queries/field-enum-relation.ts`
- `modelcraft-front/src/web/graphql/mutations/field-enum-relation.ts`
- `modelcraft-front/src/shared/errors/model-enum-error-mapper.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/field-pages/CreateEnumFieldPage.tsx`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/field-pages/CreateEnumLabelFieldPage.tsx`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/field-pages/EditFieldImmutablePage.tsx`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/field-pages/EnumRelationSelector.tsx`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/field-pages/index.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/use-create-enum-field-page.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/use-create-enum-label-field-page.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/use-edit-field-page.ts`

## 修改
- `modelcraft-front/src/types/index.ts`
- `modelcraft-front/src/web/graphql/queries/index.ts`
- `modelcraft-front/src/web/graphql/mutations/index.ts`
- `modelcraft-front/src/web/graphql/mutations/model.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/types.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/index.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_hooks/useFieldOperations.ts`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/FieldEditSheet.tsx`
- `modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx`

---

## 11. Worker 实现注意事项（执行约束）

1. **分层边界**：Web 只能通过 `@bff/model-enum/public` 调用 BFF，禁止直连 BFF 内部实现。  
2. **Mock 优先**：前端阶段默认走 BFF mock client，不依赖真实后端；后续再补 real client。  
3. **Contract 只读**：禁止修改 `contract/`；若 schema 变化，先同步 contract 再 codegen。  
4. **Codegen 必做**：GraphQL 操作变更后必须执行 `npm run codegen`，类型统一来自 `@/generated/graphql`。  
5. **参数约束硬编码在提交层**：
   - ENUM 仅提交 `relateEnumName`
   - ENUM_LABEL 仅提交 `enumRelationId`
   - 编辑不提交 `format` 变更
6. **错误映射统一出口**：所有错误必须先走 `mapModelEnumError`，UI 不直接散落判断 `__typename`。  
7. **页面拆分**：`page.tsx` 保持组合层；业务逻辑全部下沉 `_hooks`，页面私有组件仅放 `_components/field-pages`。  
8. **UI 约束**：基础控件必须使用 `@/components/ui`（shadcn/ui），不新增自定义原子组件。  
9. **回归范围**：必须覆盖三条链路（创建 ENUM、创建 ENUM_LABEL、编辑 format 不可变）与 5 类错误（含 InvalidInput）。
