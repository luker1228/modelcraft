# RLS 前端架构设计

> 基于 RLS PRD 的前端架构方案
> - 主 PRD: `ai-metadata/prd/rls/prd.md`
> - API 合约: `plans/rls/api-contract.md`
> - Policy 配置子页: `ai-metadata/prd/rls/05-policy-configuration.md`

---

## 1. 页面结构

### 1.1 模块划分

| 模块 | 职责 | 所属层 | 依赖 |
|------|------|--------|------|
| `RLSPolicyPanel` | Model 详情页"访问控制"Tab 内容 | Web Layer | `useRLSPolicy`, `RLSPresetSelector`, `PolicyConditionBuilder` |
| `AuthSchemaSection` | Project 设置页"认证变量"配置区 | Web Layer | `useAuthSchema`, `AuthVariableEditor` |
| `RLSPresetSelector` | 5 种策略卡片选择器 | Web Layer (features/rls) | 无 |
| `PolicyConditionBuilder` | 可视化条件构建器（Notion filter 风格） | Web Layer (features/rls) | `useRLSExprValidation` |
| `PolicyJSONPreview` | 实时 JSON 预览组件 | Web Layer (features/rls) | 无 |
| `AuthVariableEditor` | 认证变量增删改表单 | Web Layer (features/rls) | 无 |
| `useRLSPolicy` | RLS Policy 状态管理与 API 调用 | Web Layer (hooks/rls) | `@bff/apollo/public` |
| `useAuthSchema` | Auth Schema 状态管理与 API 调用 | Web Layer (hooks/rls) | `@bff/apollo/public` |
| `useRLSExprValidation` | RLS 表达式实时校验 Hook | Web Layer (hooks/rls) | `@bff/apollo/public` |

### 1.2 目录规划

```
# 新增/修改的文件清单

src/
├── web/
│   ├── components/
│   │   ├── features/
│   │   │   └── rls/                          # 新增：RLS 功能组件（绑定业务域）
│   │   │       ├── RLSPresetSelector.tsx     # Preset 选择器（5种策略卡片）
│   │   │       ├── PolicyConditionBuilder.tsx # 条件构建器（类似 Notion filter）
│   │   │       ├── PolicyConditionRow.tsx    # 单行条件组件
│   │   │       ├── PolicyJSONPreview.tsx     # JSON 预览组件
│   │   │       ├── AuthVariableEditor.tsx    # 认证变量编辑器
│   │   │       ├── AuthVariableRow.tsx       # 单行认证变量组件
│   │   │       └── index.ts                  # 统一导出
│   │   └── common/                           # 新增通用组件
│   │       ├── DangerConfirmDialog.tsx       # 高危操作二次确认弹窗
│   │       └── index.ts
│   ├── hooks/
│   │   └── rls/                              # 新增：RLS 业务域 Hooks
│   │       ├── use-rls-policy.ts             # RLS Policy CRUD
│   │       ├── use-auth-schema.ts            # Auth Schema CRUD
│   │       ├── use-rls-expr-validation.ts    # 表达式实时校验
│   │       ├── types.ts                      # 页面级类型定义
│   │       └── index.ts                      # 统一导出
│   └── graphql/                              # 扩展 GraphQL 操作定义
│       ├── mutations/
│       │   ├── rls.ts                        # 新增：RLS mutations
│       │   └── project.ts                    # 扩展：添加 setProjectAuthSchema
│       └── queries/
│           ├── rls.ts                        # 新增：RLS queries
│           └── project.ts                    # 扩展：添加 authSchema 字段
├── app/
│   └── org/[orgName]/project/[projectSlug]/
│       ├── model-editor/
│       │   └── _components/
│       │       ├── ModelDetailPanel.tsx      # 修改：添加"访问控制"Tab
│       │       └── RLSPolicyPanel.tsx        # 新增：访问控制 Tab 内容
│       └── settings/
│           └── page.tsx                      # 修改：添加"认证变量"配置区
└── types/
    ├── rls.ts                                # 新增：RLS 领域类型
    └── project.ts                            # 扩展：添加 AuthSchema 类型
```

### 1.3 Model 详情页 Tab 改造

```tsx
// app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelDetailPanel.tsx
// 在现有结构中新增 Tabs

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@web/components/ui/tabs'
import { Shield, Table, Key } from 'lucide-react'
import { RLSPolicyPanel } from './RLSPolicyPanel'

// 在 DrawerContent 中添加 Tabs 结构
<div className="min-h-0 flex-1 overflow-y-auto">
  <Tabs defaultValue="fields" className="flex h-full flex-col">
    <TabsList className="mx-6 mt-4 grid w-auto grid-cols-3">
      <TabsTrigger value="fields" className="gap-1.5">
        <Table className="size-3.5" />
        字段
      </TabsTrigger>
      <TabsTrigger value="foreign-keys" className="gap-1.5">
        <Key className="size-3.5" />
        外键
      </TabsTrigger>
      {/* 条件渲染：仅当 model.rlsPolicy != null 时显示 */}
      {state.editModelData?.rlsPolicy && (
        <TabsTrigger value="access-control" className="gap-1.5">
          <Shield className="size-3.5" />
          访问控制
        </TabsTrigger>
      )}
    </TabsList>

    <TabsContent value="fields" className="flex-1 px-6 py-4">
      {/* 现有字段列表内容 */}
    </TabsContent>

    <TabsContent value="foreign-keys" className="flex-1 px-6 py-4">
      <ForeignKeyPanel state={state} fkOps={fkOps} models={models} />
    </TabsContent>

    <TabsContent value="access-control" className="flex-1 px-6 py-4">
      <RLSPolicyPanel
        modelId={state.editModelId!}
        modelName={state.editModelData?.name || ''}
        policy={state.editModelData?.rlsPolicy}
        fields={state.editModelData?.fields || []}
        onPolicyUpdated={crud.refreshModelDetail}
      />
    </TabsContent>
  </Tabs>
</div>
```

### 1.4 Project 设置页改造

```tsx
// app/org/[orgName]/project/[projectSlug]/settings/page.tsx
// 在现有表单下方新增"认证变量"配置区

import { AuthSchemaSection } from '@web/components/features/rls/AuthSchemaSection'

// 在页面底部添加
<div className="mt-8 border-t border-border pt-8">
  <h3 className="mb-4 font-heading text-base font-semibold">认证变量</h3>
  <AuthSchemaSection
    projectSlug={params.projectSlug}
    orgName={params.orgName}
  />
</div>
```

---

## 2. 组件设计

### 2.1 RLSPresetSelector（Preset 选择器）

**职责**：展示 5 种预设策略卡片，支持选择和二次确认

**位置**：`web/components/features/rls/RLSPresetSelector.tsx`

```typescript
// 接口定义
export interface RLSPresetSelectorProps {
  value: RLSPreset | null  // 当前选中的 Preset
  onChange: (preset: RLSPreset, confirmed: boolean) => void  // 选择回调（高危策略需确认）
  disabled?: boolean
}

// 5 种 Preset 配置常量
export const RLS_PRESETS: Array<{
  value: RLSPreset
  label: string
  description: string
  scenario: string
  isDangerous?: boolean
}> = [
  {
    value: 'READ_WRITE_OWNER',
    label: '读写自己',
    description: '终端用户只能读写归属于自己的数据',
    scenario: '适用于：任务管理、个人笔记、订单系统'
  },
  {
    value: 'READ_ALL_WRITE_OWNER',
    label: '读全部，写自己',
    description: '终端用户可以读取全部数据，但只能修改自己的数据',
    scenario: '适用于：评论、帖子、公开内容'
  },
  {
    value: 'READ_ALL',
    label: '只读全部',
    description: '终端用户只能读取全部数据，不能写入',
    scenario: '适用于：公告、商品目录、配置表'
  },
  {
    value: 'READ_WRITE_ALL',
    label: '读写全部',
    description: '终端用户可以读写任意数据（⚠️ 高危策略）',
    scenario: '适用于：完全开放的场景',
    isDangerous: true
  },
  {
    value: 'NO_ACCESS',
    label: '无访问权限',
    description: '终端用户无法访问该模型的任何数据',
    scenario: '适用于：系统内部表、敏感数据'
  }
]

// 组件骨架
export function RLSPresetSelector({ value, onChange, disabled }: RLSPresetSelectorProps): JSX.Element {
  const [pendingPreset, setPendingPreset] = useState<RLSPreset | null>(null)

  const handleSelect = (preset: RLSPreset) => {
    if (preset === 'READ_WRITE_ALL') {
      setPendingPreset(preset)
      return
    }
    onChange(preset, true)
  }

  const handleConfirmDangerous = () => {
    if (pendingPreset) {
      onChange(pendingPreset, true)
      setPendingPreset(null)
    }
  }

  return (
    <>
      <div className="grid grid-cols-1 gap-3">
        {RLS_PRESETS.map((preset) => (
          // 卡片布局：图标 + 标题 + 描述 + 场景
          // 选中状态高亮显示
          // 高危策略显示警告图标
        ))}
      </div>

      {/* 高危策略二次确认弹窗 */}
      <DangerConfirmDialog
        open={pendingPreset !== null}
        onOpenChange={() => setPendingPreset(null)}
        title="确认选择高危策略？"
        description="此策略允许所有终端用户读写任意数据，包括其他用户的数据，请确认你了解风险。"
        onConfirm={handleConfirmDangerous}
      />
    </>
  )
}
```

### 2.2 PolicyConditionBuilder（条件构建器）

**职责**：可视化构建 RLS 条件表达式（类似 Notion filter）

**位置**：`web/components/features/rls/PolicyConditionBuilder.tsx`

```typescript
// 接口定义
export interface PolicyConditionBuilderProps {
  modelId: string
  exprType: RLSExprType  // SELECT_PREDICATE / INSERT_CHECK / ...
  value: JsonExpr
  fields: Array<{ name: string; format: FormatType }>  // 模型字段列表
  authVariables: string[]  // 已声明的 auth 变量（含内置 uid）
  onChange: (value: JsonExpr) => void
  onValidationResult?: (result: ValidationResult) => void
}

// 条件行数据结构
export interface ConditionRow {
  id: string
  field: string        // 字段名或 _auth / _ref / _exists
  operator: RLSScalarOperator  // _eq / _neq / _gt / _in / _is_null 等
  value: JsonValue     // 值或 {_auth: "uid"} / {_ref: "..."}
}

// 组件骨架
export function PolicyConditionBuilder({
  modelId,
  exprType,
  value,
  fields,
  authVariables,
  onChange,
  onValidationResult
}: PolicyConditionBuilderProps): JSX.Element {
  const [conditions, setConditions] = useState<ConditionRow[]>([])
  const [logicalOp, setLogicalOp] = useState<'AND' | 'OR'>('AND')
  const { validate, validating } = useRLSExprValidation()

  // 将 JsonExpr 解析为 ConditionRow[]
  // 将 ConditionRow[] 编译为 JsonExpr
  // 实时校验：debounce 500ms 后调用 validateRLSExpr

  return (
    <div className="space-y-3">
      {/* 操作类型标题 */}
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{getExprTypeLabel(exprType)}</span>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">关系：</span>
          <ToggleGroup value={logicalOp} onValueChange={setLogicalOp}>
            <ToggleGroupItem value="AND">且</ToggleGroupItem>
            <ToggleGroupItem value="OR">或</ToggleGroupItem>
          </ToggleGroup>
        </div>
      </div>

      {/* 条件行列表 */}
      {conditions.map((condition) => (
        <PolicyConditionRow
          key={condition.id}
          condition={condition}
          fields={fields}
          authVariables={authVariables}
          disabled={validating}
          onChange={(updated) => updateCondition(condition.id, updated)}
          onRemove={() => removeCondition(condition.id)}
        />
      ))}

      {/* 添加按钮 */}
      <Button variant="outline" size="sm" onClick={addCondition}>
        <Plus className="mr-1 size-3.5" />
        添加条件
      </Button>

      {/* CHECK 限制提示：insertCheck/updateCheck 不允许 _exists/_ref */}
      {(exprType === 'INSERT_CHECK' || exprType === 'UPDATE_CHECK') && (
        <Alert variant="info" className="text-xs">
          <Info className="size-3.5" />
          <AlertDescription>
            INSERT/UPDATE 检查不支持 EXISTS 子查询和跨表引用
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
```

### 2.3 PolicyJSONPreview（JSON 预览）

**职责**：实时展示当前表达式对应的 JSON，支持语法高亮

**位置**：`web/components/features/rls/PolicyJSONPreview.tsx`

```typescript
export interface PolicyJSONPreviewProps {
  value: JsonExpr
  className?: string
}

export function PolicyJSONPreview({ value, className }: PolicyJSONPreviewProps): JSX.Element {
  const formatted = useMemo(() => JSON.stringify(value, null, 2), [value])

  return (
    <div className={cn("rounded-md border bg-muted/30 p-3", className)}>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">JSON 预览</span>
        <CopyButton text={formatted} />
      </div>
      <pre className="overflow-x-auto font-mono text-xs leading-relaxed text-foreground">
        <code>{formatted}</code>
      </pre>
    </div>
  )
}
```

### 2.4 RLSPolicyPanel（访问控制 Tab 内容）

**职责**：组合 Preset 选择器、条件构建器和 JSON 预览

**位置**：`app/org/[orgName]/project/[projectSlug]/model-editor/_components/RLSPolicyPanel.tsx`

```typescript
export interface RLSPolicyPanelProps {
  modelId: string
  modelName: string
  policy: ModelRLSPolicy | null | undefined
  fields: Array<{ name: string; format: FormatType }>
  onPolicyUpdated?: () => void
}

export function RLSPolicyPanel({
  modelId,
  modelName,
  policy,
  fields,
  onPolicyUpdated
}: RLSPolicyPanelProps): JSX.Element {
  const [activeTab, setActiveTab] = useState<'preset' | 'advanced'>('preset')
  const { updatePolicy, updating, error } = useRLSPolicy()

  // 处理 Preset 选择
  const handlePresetChange = async (preset: RLSPreset) => {
    const jsonExpr = presetToJsonExpr(preset)
    await updatePolicy({
      modelId,
      selectPredicate: jsonExpr.selectPredicate,
      insertCheck: jsonExpr.insertCheck,
      updatePredicate: jsonExpr.updatePredicate,
      updateCheck: jsonExpr.updateCheck,
      deletePredicate: jsonExpr.deletePredicate
    })
    onPolicyUpdated?.()
  }

  return (
    <div className="space-y-6">
      {/* 当前策略概览 */}
      <div className="rounded-lg border bg-muted/20 p-4">
        <h4 className="mb-2 text-sm font-semibold">当前策略</h4>
        <div className="flex items-center gap-2">
          <Badge variant={policy?.preset ? 'default' : 'secondary'}>
            {policy?.preset || '自定义'}
          </Badge>
          {policy?.preset === 'READ_WRITE_ALL' && (
            <Badge variant="destructive">高危</Badge>
          )}
        </div>
      </div>

      {/* 配置模式切换 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="preset">选择预设</TabsTrigger>
          <TabsTrigger value="advanced">自定义条件</TabsTrigger>
        </TabsList>

        <TabsContent value="preset" className="space-y-4 pt-4">
          <RLSPresetSelector
            value={policy?.preset || null}
            onChange={handlePresetChange}
            disabled={updating}
          />
        </TabsContent>

        <TabsContent value="advanced" className="space-y-4 pt-4">
          {/* 五件套分别配置 */}
          <Accordion type="multiple" defaultValue={['select']}>
            <AccordionItem value="select">
              <AccordionTrigger>SELECT 条件</AccordionTrigger>
              <AccordionContent className="space-y-3">
                <PolicyConditionBuilder
                  modelId={modelId}
                  exprType="SELECT_PREDICATE"
                  value={parseJsonExpr(policy?.selectPredicate)}
                  fields={fields}
                  authVariables={authVariables}
                  onChange={(expr) => updatePredicate('selectPredicate', expr)}
                />
              </AccordionContent>
            </AccordionItem>
            {/* INSERT、UPDATE、DELETE 类似 */}
          </Accordion>

          {/* 实时 JSON 预览 */}
          <PolicyJSONPreview value={compiledPolicy} />
        </TabsContent>
      </Tabs>

      {/* 保存反馈 */}
      {error && <Alert variant="destructive">{error.message}</Alert>}
    </div>
  )
}
```

### 2.5 AuthVariableEditor（认证变量编辑器）

**职责**：Project 设置页认证变量的增删改

**位置**：`web/components/features/rls/AuthVariableEditor.tsx`

```typescript
export interface AuthVariableEditorProps {
  variables: AuthVariable[]
  onChange: (variables: AuthVariable[]) => void
  disabled?: boolean
}

export function AuthVariableEditor({
  variables,
  onChange,
  disabled
}: AuthVariableEditorProps): JSX.Element {
  // uid 内置，不可编辑
  const builtinVariables: AuthVariable[] = [
    { name: 'uid', source: 'jwt.user_id', type: 'UUID', isBuiltin: true }
  ]

  const handleAdd = () => {
    onChange([...variables, { name: '', source: '', type: 'STRING' }])
  }

  const handleUpdate = (index: number, updated: AuthVariable) => {
    const next = [...variables]
    next[index] = updated
    onChange(next)
  }

  const handleRemove = (index: number) => {
    onChange(variables.filter((_, i) => i !== index))
  }

  return (
    <div className="space-y-3">
      {/* 内置变量（只读） */}
      {builtinVariables.map((v) => (
        <AuthVariableRow key={v.name} variable={v} disabled />
      ))}

      {/* 自定义变量 */}
      {variables.map((v, i) => (
        <AuthVariableRow
          key={i}
          variable={v}
          onChange={(updated) => handleUpdate(i, updated)}
          onRemove={() => handleRemove(i)}
          disabled={disabled}
        />
      ))}

      <Button variant="outline" size="sm" onClick={handleAdd} disabled={disabled}>
        <Plus className="mr-1 size-3.5" />
        添加变量
      </Button>
    </div>
  )
}
```

### 2.6 删除 Owner 字段二次确认

**修改位置**：`app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/use-field-operations.ts`

```typescript
// 在 handleRemoveField 中检测 owner 字段
const handleRemoveField = async (field: Field) => {
  // 检测是否为 owner 字段（EndUserRef 类型）
  if (field.format === 'END_USER_REF' && field.name === 'owner') {
    // 显示二次确认弹窗
    const confirmed = await showConfirmDialog({
      title: '删除 owner 字段？',
      description: '删除后该模型的数据隔离将关闭，所有终端用户可访问全量数据。此操作同时会删除关联的访问控制策略。',
      confirmText: '确认删除',
      cancelText: '取消',
      variant: 'destructive'
    })
    if (!confirmed) return
  }

  // 执行删除操作
  await removeFieldMutation({
    modelID: state.editModelId!,
    fieldName: field.name
  })
}
```

---

## 3. GraphQL Client

### 3.1 新增 GraphQL 文件

**位置**：`src/web/graphql/mutations/rls.ts`

```typescript
import { gql } from '@apollo/client'

// 设置 Model RLS 策略
export const SET_MODEL_RLS_POLICY = gql`
  mutation SetModelRLSPolicy($input: SetModelRLSPolicyInput!) {
    setModelRLSPolicy(input: $input) {
      policy {
        modelId
        selectPredicate
        insertCheck
        updatePredicate
        updateCheck
        deletePredicate
        preset
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on ModelHasNoOwnerField {
          message
          suggestion
        }
        ... on InvalidRLSExpression {
          message
          suggestion
          path
        }
        ... on InvalidAuthVariable {
          message
          suggestion
          variable
        }
      }
    }
  }
`

// 校验 RLS 表达式
export const VALIDATE_RLS_EXPR = gql`
  mutation ValidateRLSExpr($input: ValidateRLSExprInput!) {
    validateRLSExpr(input: $input) {
      result {
        valid
        errors {
          path
          message
          code
        }
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on InvalidRLSExpression {
          message
          suggestion
          path
        }
      }
    }
  }
`
```

**位置**：`src/web/graphql/queries/rls.ts`

```typescript
import { gql } from '@apollo/client'

// 获取 Model RLS 策略
export const GET_MODEL_RLS_POLICY = gql`
  query ModelRLSPolicy($modelId: ID!) {
    modelRLSPolicy(modelId: $modelId) {
      modelId
      selectPredicate
      insertCheck
      updatePredicate
      updateCheck
      deletePredicate
      preset
      createdAt
      updatedAt
    }
  }
`
```

**扩展**：`src/web/graphql/mutations/project.ts`

```typescript
// 在现有文件中添加
export const SET_PROJECT_AUTH_SCHEMA = gql`
  mutation SetProjectAuthSchema($input: SetProjectAuthSchemaInput!) {
    setProjectAuthSchema(input: $input) {
      authSchema {
        variables {
          name
          source
          type
        }
      }
      error {
        __typename
        ... on ProjectNotFound {
          message
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`
```

**扩展**：`src/web/graphql/queries/project.ts`

```typescript
// 在获取 Project 的 query 中扩展 authSchema 字段
// 或在单独的 query 中获取
export const GET_PROJECT_AUTH_SCHEMA = gql`
  query GetProjectAuthSchema($orgName: String!, $projectSlug: String!) {
    project(orgName: $orgName, slug: $projectSlug) {
      id
      authSchema {
        variables {
          name
          source
          type
        }
      }
    }
  }
`
```

### 3.2 扩展现有 Query

**修改**：`src/web/graphql/queries/model.ts`

```typescript
// 在 GET_MODEL / GET_MODELS 等 query 中添加 rlsPolicy 字段
export const GET_MODEL_DETAIL = gql`
  query GetModelDetail($id: ID!) {
    model(id: $id) {
      id
      # ... 现有字段
      rlsPolicy {
        modelId
        selectPredicate
        insertCheck
        updatePredicate
        updateCheck
        deletePredicate
        preset
      }
    }
  }
`
```

---

## 4. 状态管理

### 4.1 RLS 领域类型

**位置**：`src/types/rls.ts`

```typescript
// RLS 预设策略枚举
export type RLSPreset =
  | 'READ_WRITE_OWNER'
  | 'READ_ALL_WRITE_OWNER'
  | 'READ_ALL'
  | 'READ_WRITE_ALL'
  | 'NO_ACCESS'

// RLS 表达式类型
export type RLSExprType =
  | 'SELECT_PREDICATE'
  | 'INSERT_CHECK'
  | 'UPDATE_PREDICATE'
  | 'UPDATE_CHECK'
  | 'DELETE_PREDICATE'

// 标量操作符
export type RLSScalarOperator =
  | '_eq'
  | '_neq'
  | '_gt'
  | '_gte'
  | '_lt'
  | '_lte'
  | '_in'
  | '_nin'
  | '_is_null'

// 逻辑操作符
export type RLSLogicalOperator = '_and' | '_or' | '_not'

// RLS JSON 表达式（递归类型）
export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonExpr
  | JsonValue[]

export interface JsonExpr {
  [key: string]: JsonValue
}

// RLS 策略
export interface ModelRLSPolicy {
  modelId: string
  selectPredicate: string  // JSON string
  insertCheck: string
  updatePredicate: string
  updateCheck: string
  deletePredicate: string
  preset: RLSPreset | null
  createdAt: string
  updatedAt: string
}

// 认证变量
export type AuthVariableType = 'UUID' | 'STRING' | 'INTEGER'

export interface AuthVariable {
  name: string
  source: string
  type: AuthVariableType
  isBuiltin?: boolean
}

export interface ProjectAuthSchema {
  variables: AuthVariable[]
}

// 校验结果
export interface ValidationError {
  path: string
  message: string
  code: string
}

export interface ValidationResult {
  valid: boolean
  errors?: ValidationError[]
}
```

### 4.2 Hooks 接口

**位置**：`src/web/hooks/rls/use-rls-policy.ts`

```typescript
export interface UseRLSPolicyOptions {
  modelId: string
}

export interface UseRLSPolicyReturn {
  policy: ModelRLSPolicy | null
  loading: boolean
  error: Error | null
  updating: boolean
  updatePolicy: (input: SetModelRLSPolicyInput) => Promise<void>
  refresh: () => Promise<void>
}

export function useRLSPolicy(options: UseRLSPolicyOptions): UseRLSPolicyReturn {
  // TODO: worker 实现
}
```

**位置**：`src/web/hooks/rls/use-rls-expr-validation.ts`

```typescript
export interface UseRLSExprValidationReturn {
  validating: boolean
  lastResult: ValidationResult | null
  validate: (input: ValidateRLSExprInput) => Promise<ValidationResult>
}

export function useRLSExprValidation(): UseRLSExprValidationReturn {
  // TODO: worker 实现
  // 内部使用 debounce 实现实时校验
}
```

**位置**：`src/web/hooks/rls/use-auth-schema.ts`

```typescript
export interface UseAuthSchemaOptions {
  projectSlug: string
  orgName: string
}

export interface UseAuthSchemaReturn {
  schema: ProjectAuthSchema | null
  loading: boolean
  error: Error | null
  saving: boolean
  saveSchema: (variables: AuthVariable[]) => Promise<void>
  refresh: () => Promise<void>
}

export function useAuthSchema(options: UseAuthSchemaOptions): UseAuthSchemaReturn {
  // TODO: worker 实现
}
```

---

## 5. BFF Mock 契约

### 5.1 Mock 数据结构

**位置**：`src/mocks/data/project/rls-factory.ts`

```typescript
import { faker } from '@faker-js/faker'
import type {
  ModelRLSPolicy,
  AuthVariable,
  ProjectAuthSchema,
  ValidationResult
} from '@/types/rls'

// 预设策略对应的 JSON 表达式
export const PRESET_JSON_EXPR = {
  READ_WRITE_OWNER: {
    selectPredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    insertCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updatePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updateCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    deletePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } })
  },
  READ_ALL_WRITE_OWNER: {
    selectPredicate: 'true',
    insertCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updatePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updateCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    deletePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } })
  },
  READ_ALL: {
    selectPredicate: 'true',
    insertCheck: 'false',
    updatePredicate: 'false',
    updateCheck: 'false',
    deletePredicate: 'false'
  },
  READ_WRITE_ALL: {
    selectPredicate: 'true',
    insertCheck: 'true',
    updatePredicate: 'true',
    updateCheck: 'true',
    deletePredicate: 'true'
  },
  NO_ACCESS: {
    selectPredicate: 'false',
    insertCheck: 'false',
    updatePredicate: 'false',
    updateCheck: 'false',
    deletePredicate: 'false'
  }
}

// 创建 Mock RLS Policy
export function createMockRLSPolicy(
  modelId: string,
  preset: keyof typeof PRESET_JSON_EXPR = 'READ_WRITE_OWNER'
): ModelRLSPolicy {
  const expr = PRESET_JSON_EXPR[preset]
  return {
    modelId,
    ...expr,
    preset,
    createdAt: faker.date.past().toISOString(),
    updatedAt: faker.date.recent().toISOString()
  }
}

// 创建 Mock Auth Variable
export function createMockAuthVariable(override?: Partial<AuthVariable>): AuthVariable {
  return {
    name: faker.word.identifier(),
    source: `jwt.${faker.word.identifier()}`,
    type: faker.helpers.arrayElement(['UUID', 'STRING', 'INTEGER']),
    ...override
  }
}

// 创建 Mock Project Auth Schema
export function createMockAuthSchema(): ProjectAuthSchema {
  return {
    variables: [
      { name: 'tenant_id', source: 'jwt.tenant_id', type: 'UUID' },
      { name: 'role', source: 'jwt.role', type: 'STRING' }
    ]
  }
}

// 创建 Mock Validation Result
export function createMockValidationResult(valid = true): ValidationResult {
  return {
    valid,
    errors: valid ? undefined : [
      {
        path: 'selectPredicate.owner._eq',
        message: 'Invalid field reference: owner',
        code: 'INVALID_FIELD'
      }
    ]
  }
}
```

### 5.2 Mock Handler 配置

**位置**：`src/mocks/handlers/project/rls.ts`

```typescript
import { graphql } from 'msw'
import {
  createMockRLSPolicy,
  createMockAuthSchema,
  createMockValidationResult,
  PRESET_JSON_EXPR
} from '@/mocks/data/project/rls-factory'

export const rlsHandlers = [
  // Get Model RLS Policy
  graphql.query('ModelRLSPolicy', ({ variables }) => {
    const { modelId } = variables
    return Response.json({
      data: {
        modelRLSPolicy: createMockRLSPolicy(modelId, 'READ_WRITE_OWNER')
      }
    })
  }),

  // Set Model RLS Policy
  graphql.mutation('SetModelRLSPolicy', ({ variables }) => {
    const { input } = variables
    return Response.json({
      data: {
        setModelRLSPolicy: {
          policy: {
            modelId: input.modelId,
            selectPredicate: input.selectPredicate,
            insertCheck: input.insertCheck,
            updatePredicate: input.updatePredicate,
            updateCheck: input.updateCheck,
            deletePredicate: input.deletePredicate,
            preset: null,  // 自定义策略返回 null
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          },
          error: null
        }
      }
    })
  }),

  // Validate RLS Expression
  graphql.mutation('ValidateRLSExpr', ({ variables }) => {
    const { input } = variables
    // 模拟：包含非法字段时返回错误
    const hasInvalidField = input.expression.includes('invalid_field')
    return Response.json({
      data: {
        validateRLSExpr: {
          result: createMockValidationResult(!hasInvalidField),
          error: null
        }
      }
    })
  }),

  // Get Project Auth Schema
  graphql.query('GetProjectAuthSchema', () => {
    return Response.json({
      data: {
        project: {
          id: 'mock-project-id',
          authSchema: createMockAuthSchema()
        }
      }
    })
  }),

  // Set Project Auth Schema
  graphql.mutation('SetProjectAuthSchema', ({ variables }) => {
    const { input } = variables
    return Response.json({
      data: {
        setProjectAuthSchema: {
          authSchema: {
            variables: input.variables
          },
          error: null
        }
      }
    })
  })
]
```

### 5.3 Model Factory 扩展

**扩展**：`src/mocks/data/project/model-factory.ts`

```typescript
// 在 createMockModel 中添加 rlsPolicy 字段
export function createMockModel(override = {}) {
  const modelId = faker.string.uuid()
  const hasOwner = override.hasOwner !== false  // 默认有 owner 字段

  return {
    id: modelId,
    name: faker.word.noun(),
    title: faker.commerce.productName(),
    // ... 现有字段
    rlsPolicy: hasOwner ? createMockRLSPolicy(modelId, 'READ_WRITE_OWNER') : null,
    // owner 字段在 fields 数组中
    fields: [
      // ... 现有字段
      ...(hasOwner ? [createMockOwnerField()] : [])
    ],
    ...override
  }
}

// 创建 owner 字段（EndUserRef）
export function createMockOwnerField() {
  return {
    id: faker.string.uuid(),
    name: 'owner',
    title: '所有者',
    schemaType: 'STRING',
    format: 'END_USER_REF',  // 新增 FormatType
    // ... 其他字段属性
  }
}
```

---

## 6. 实现优先级

### P0 - 核心功能（阻塞其他开发）

1. **GraphQL Client 定义**
   - `src/web/graphql/mutations/rls.ts` - SET_MODEL_RLS_POLICY, VALIDATE_RLS_EXPR
   - `src/web/graphql/queries/rls.ts` - GET_MODEL_RLS_POLICY
   - 扩展 `src/web/graphql/mutations/project.ts` - SET_PROJECT_AUTH_SCHEMA
   - **前置依赖**: 后端 API Contract 完成并通过 `front-contract-pull` 同步

2. **类型定义**
   - `src/types/rls.ts` - RLS 领域类型

3. **Mock 数据工厂**
   - `src/mocks/data/project/rls-factory.ts`
   - `src/mocks/handlers/project/rls.ts`

### P1 - 主要组件

4. **RLSPresetSelector 组件**
   - `src/web/components/features/rls/RLSPresetSelector.tsx`
   - **交互细节**: READ_WRITE_ALL 高危策略二次确认

5. **RLSPolicyPanel 页面组件**
   - `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/RLSPolicyPanel.tsx`
   - **前置依赖**: ModelDetailPanel Tabs 改造完成

6. **ModelDetailPanel Tabs 改造**
   - 添加"访问控制"Tab（条件渲染）

### P2 - 高级功能

7. **PolicyConditionBuilder 组件**
   - `src/web/components/features/rls/PolicyConditionBuilder.tsx`
   - `src/web/components/features/rls/PolicyConditionRow.tsx`

8. **PolicyJSONPreview 组件**
   - `src/web/components/features/rls/PolicyJSONPreview.tsx`

9. **Hooks 实现**
   - `src/web/hooks/rls/use-rls-policy.ts`
   - `src/web/hooks/rls/use-rls-expr-validation.ts`

### P3 - Project 级别配置

10. **AuthVariableEditor 组件**
    - `src/web/components/features/rls/AuthVariableEditor.tsx`

11. **useAuthSchema Hook**
    - `src/web/hooks/rls/use-auth-schema.ts`

12. **Project 设置页扩展**
    - 添加"认证变量"配置区

### P4 - 交互增强

13. **删除 Owner 字段二次确认**
    - 修改 `use-field-operations.ts`

14. **实时校验集成**
    - 在 PolicyConditionBuilder 中集成 useRLSExprValidation

---

## 7. Worker 实现注意事项

### 7.1 开发阶段 Mock 启用

```bash
# 开发阶段启用 MSW
echo "NEXT_PUBLIC_API_MOCKING=enabled" >> .env.local

# 联调阶段关闭 Mock
# 注释掉或删除 .env.local 中的 NEXT_PUBLIC_API_MOCKING
```

### 7.2 Codegen 流程

```bash
# 1. 后端完成 API Contract 变更后，前端同步
front-contract-pull

# 2. 生成 TypeScript 类型和 MSW handlers
npm run codegen

# 3. 检查生成的类型是否符合预期
cat src/generated/graphql.ts | grep -A 10 "ModelRLSPolicy"
```

### 7.3 测试检查清单

- [ ] Model 详情页在有 owner 字段时显示"访问控制"Tab
- [ ] Model 详情页在无 owner 字段时不显示"访问控制"Tab
- [ ] 选择 READ_WRITE_ALL 预设时弹出二次确认弹窗
- [ ] 删除 owner 字段时弹出二次确认弹窗
- [ ] 实时校验在输入停止 500ms 后触发
- [ ] INSERT_CHECK / UPDATE_CHECK 添加 _exists 时报错提示
- [ ] Project 设置页可正常增删认证变量
- [ ] uid 内置变量不可编辑

### 7.4 性能注意

- `useRLSExprValidation` 必须使用 debounce（建议 500ms）
- `PolicyConditionBuilder` 的 JSON 编译使用 useMemo 缓存
- Model 列表查询暂时不展开 `rlsPolicy` 字段，仅在详情页加载
