# 动态表单设计文档

**日期**: 2026-04-01
**状态**: 待实现
**范围**: modelcraft-front

---

## 背景

当前 `DynamicModelTable` 存在两个问题：

1. **列不实时同步**：字段增删后，表格列不更新，需手动刷新页面。根本原因是字段 mutation 缺少 `refetchQueries`，Apollo 缓存不失效。
2. **表单体验粗糙**：新增/编辑记录的 Sheet 表单使用手写 if-else 渲染字段，缺少日期选择器、枚举下拉、关联字段选择器等类型适配控件，且无实时校验反馈。

---

## 目标

1. 字段增删后，表格列**立即自动更新**，无需手动刷新。
2. 新增/编辑记录的 Sheet 表单使用 **RJSF（react-jsonschema-form）** 驱动，根据字段类型自动渲染合适的输入控件，并提供实时校验反馈。
3. 支持 **RELATION 关联字段**的记录选择器（优先支持的特殊字段类型）。

---

## 方案选型

### 列实时同步

**方案**：在字段变更 mutation 上添加 `refetchQueries`，Apollo Client 在 mutation 完成后自动 refetch 相关 query。

**涉及文件与 mutation**：
- `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`：`REMOVE_FIELD`
- `src/web/components/model-editor/InsertFieldSheet.tsx`：`ADD_FIELDS`、`UPDATE_FIELD`（UPDATE_FIELD 在 InsertFieldSheet 中，用于修改字段属性）

**实际 GraphQL operation names**（已验证，来自 `src/web/graphql/queries/model.ts`）：
```typescript
refetchQueries: ['GetModel', 'GetModelJsonSchema']
```

### 动态表单库

**选型**：`react-jsonschema-form`（RJSF）+ `@rjsf/shadcn-ui` 主题

**依赖安装**：
```bash
npm install @rjsf/core @rjsf/utils @rjsf/validator-ajv8 @rjsf/shadcn-ui
```

---

## 数据流设计

```
GET_MODEL_JSON_SCHEMA  →  schema prop  →  RJSF <Form>
                                               ↑
GET_MODEL fields[]     →  buildUiSchema()  →  uiSchema prop
                                               ↑
                          custom widgets  →  RelationPicker, EnumSelect
```

- **JSON Schema**（来自 `GetModelJsonSchema`）：驱动字段结构、类型、校验规则（minLength/maxLength/pattern/minimum/maximum）。
- **fields[]**（来自 `GetModel`）：补充 JSON Schema 缺失的 UI 信息——枚举选项、RELATION 目标查询上下文——通过 `buildUiSchema()` 注入到 `uiSchema` 的 `ui:options` 中。
- **custom widgets**：处理 ENUM、ENUM_ARRAY、RELATION 等无法用标准 HTML 控件表达的字段类型。

### 记录提交数据流

**新增记录**：
```typescript
// DynamicModelTable 现有 mutation（不变）
await createContent({ variables: { data } })
// data 为 RJSF onSubmit 返回的 formData，已排除 isPrimary 字段
```

**编辑记录**：
```typescript
// DynamicModelTable 现有 mutation（不变）
await updateContent({ variables: { where: { id: editItemId }, data } })
```

`ModelRecordForm` 的 `onSubmit` 只负责将 RJSF `formData` 向上传递，具体 mutation 调用由 `DynamicModelTable` 持有（保持现有模式不变）。

---

## 文件结构

```
src/web/components/model-editor/
├── ModelRecordForm/                      ← 新建
│   ├── index.tsx                         ← RJSF Form 封装，对外暴露 <ModelRecordForm>
│   ├── buildUiSchema.ts                  ← fields[] → uiSchema 映射逻辑
│   ├── widgets/
│   │   ├── RelationPicker.tsx            ← RELATION 字段：可搜索关联记录选择器
│   │   ├── EnumSelect.tsx                ← ENUM（单选）/ ENUM_ARRAY（多选）
│   │   └── index.ts                      ← 导出所有 custom widgets
│
├── DynamicModelTable.tsx                 ← 修改：替换手写 Sheet 表单为 <ModelRecordForm>，REMOVE_FIELD 加 refetchQueries
└── InsertFieldSheet.tsx                  ← 修改：ADD_FIELDS、UPDATE_FIELD 加 refetchQueries
```

> **注意**：`InsertFieldSheet` 是字段（列）属性管理的 Sheet，不是数据记录表单。`ModelRecordForm` 用于新增/编辑数据行，由 `DynamicModelTable` 内部的 Sheet 容器调用。Sheet 容器由 `DynamicModelTable` 管理，`ModelRecordForm` 只渲染表单内容。

---

## 组件设计

### `ModelRecordForm` (`index.tsx`)

**对外接口**：

```tsx
interface ModelRecordFormProps {
  fields: FieldDefinition[]                              // 来自 GET_MODEL，用于 buildUiSchema
  jsonSchema: RJSFSchema                                 // 来自 GET_MODEL_JSON_SCHEMA，传给 RJSF（import { RJSFSchema } from '@rjsf/utils'）
  initialData?: Record<string, unknown>                  // 编辑时传入现有数据，新增时为 undefined
  onSubmit: (data: Record<string, unknown>) => Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
  // RelationPicker 查询关联记录所需上下文，通过 RJSF formContext 传递给 widget
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string                                        // 用于查询 logicalForeignKeys
}
```

**实现要点**：
- 在组件内部使用 `modelId` 查询 `GET_LOGICAL_FOREIGN_KEYS`（来自 `src/web/graphql/queries/model.ts`），query 变量为 `{ modelId }`，结果存入本地 state。
- FK 查询状态处理：
  - **加载中**：渲染加载骨架屏（skeleton），不展示表单字段，避免 RelationPicker 在 FK 数据就绪前渲染
  - **查询失败**：仍然展示表单，RELATION 字段的 RelationPicker widget 从 `formContext.logicalForeignKeys`（空数组）中找不到 FK，自行展示错误状态，不阻塞其他字段
- 调用 `buildUiSchema(fields, { orgName, projectSlug, clusterName, databaseName })` 生成 `uiSchema`
- 通过 RJSF `formContext` 将上下文传递给 widgets：`<Form formContext={{ orgName, projectSlug, clusterName, databaseName, modelId, logicalForeignKeys }}>`
- 过滤主键字段：通过 `uiSchema` 将主键字段设置为 `ui:widget: 'hidden'`（`isPrimary` 仅存在于 `fields[]` 设计时元数据中，不在 JSON Schema properties 里，无法从 jsonSchema 侧过滤，只能通过 uiSchema 控制）
- 服务端错误通过 `toast` 显示，不注入 RJSF 内部
- 提交/取消按钮由本组件渲染，`isSubmitting` 控制 loading 状态

### `buildUiSchema` (`buildUiSchema.ts`)

**签名**：
```typescript
function buildUiSchema(
  fields: FieldDefinition[],
  context: { orgName: string; projectSlug: string; clusterName: string; databaseName: string }
): UiSchema
```

**字段映射规则**：

| 字段 format | storageHint | uiSchema 输出 |
|------------|-------------|--------------|
| `ENUM` | - | `ui:widget: 'EnumSelect'`，`ui:options: { enumValues: field.enum.options, multiple: false }` |
| `ENUM_ARRAY` | - | `ui:widget: 'EnumSelect'`，`ui:options: { enumValues: field.enum.options, multiple: true }` |
| `RELATION` | - | `ui:widget: 'RelationPicker'`，`ui:options: { relateFkId: field.relateFkId, ...context }` |
| `DATE` | - | `ui:widget: 'date'`（RJSF 内置） |
| `DATETIME` | - | `ui:widget: 'datetime-local'`（RJSF 内置） |
| `TIME` | - | `ui:widget: 'time'`（RJSF 内置） |
| 任意 | `TEXT` | `ui:widget: 'textarea'`（优先级低于 format） |
| 其他 | - | 默认（RJSF 根据 JSON Schema type 自动选择） |

**枚举数据来源**：`FieldDefinition.enum.options`（来自 `GET_MODEL` query，已包含 `{ code, label }` 列表）。`field.enum.isMultiSelect` 可辅助区分 ENUM 和 ENUM_ARRAY，但以 `field.format` 为主判断依据。

### `RelationPicker` (`widgets/RelationPicker.tsx`)

**接收参数**：
- RJSF widget 标准 props：`value`（string）、`onChange`（string → void）
- `formContext`（来自 RJSF formContext）：`{ orgName, projectSlug, clusterName, databaseName, modelId, logicalForeignKeys: LogicalForeignKey[] }`
- `uiSchema['ui:options']`：`{ relateFkId: string }`

> `logicalForeignKeys` 由父组件 `ModelRecordForm` 预先查询后注入到 `formContext`，避免每个 RelationPicker widget 重复请求。

**查询逻辑**：

**Step 1**：从 `formContext.logicalForeignKeys` 中找到 `id === relateFkId` 的 FK 记录，直接读取 `refModelName`（`GET_LOGICAL_FOREIGN_KEYS` 返回的 FK 对象上已包含 `refModelName`，无需额外查询）。

**Step 2**：用 `createModelRuntimeClient(orgName, projectSlug, databaseName, refModelName)` 创建运行时 Apollo Client（来自 `src/bff/apollo/clients.ts`），再用 `buildFindManyQuery(refModelName, ['id', 'name'])` 查询目标模型记录（取前 50 条）。

> `buildFindManyQuery` 签名：`(modelName: string, fields: string[] | FieldDefinition[]): DocumentNode`，来自 `src/bff/cms/runtime-query-builder.ts`

**显示标签**：下拉选项显示 `{record.name || record.id}`（优先 name 字段，无则降级到 id）。下拉标签格式：`"{refModelName} ({sourceFields.join(', ')})"` 作为 Select 标题。

**UI 行为**：
- 渲染为可搜索下拉框（shadcn `Select` + 搜索过滤，无需额外多选组件）
- Loading：显示 spinner，禁用交互
- 空结果：显示"暂无数据"
- 查询失败：显示"加载失败"并提供重试按钮
- 不支持分页（初版），取前 50 条
- 选中后存储目标记录的 id 字段值（主键）

### `EnumSelect` (`widgets/EnumSelect.tsx`)

**接收参数**：
- RJSF widget 标准 props：`value`（单选 string / 多选 string[]）、`onChange`
- `uiSchema['ui:options']`：`{ enumValues: Array<{code: string, label: string}>, multiple: boolean }`

**UI 实现**：
- 单选（`multiple: false`）：shadcn `Select`（已有组件 `src/components/ui/select.tsx`）
- 多选（`multiple: true`）：用 shadcn `Checkbox` 数组实现，渲染在 `Popover` 内（避免引入新组件）

---

## 列实时同步实现

```typescript
// src/app/.../model-editor/page.tsx — REMOVE_FIELD
const [removeFieldMutation] = useMutation(REMOVE_FIELD, {
  context: projectScopedContext,
  refetchQueries: ['GetModel', 'GetModelJsonSchema'],  // ← 新增
})

// src/web/components/model-editor/InsertFieldSheet.tsx — ADD_FIELDS
const [addFields] = useMutation(ADD_FIELDS, {
  context: projectScopedContext,
  refetchQueries: ['GetModel', 'GetModelJsonSchema'],  // ← 补充 GetModelJsonSchema
})

// src/web/components/model-editor/InsertFieldSheet.tsx — UPDATE_FIELD
const [updateField] = useMutation(UPDATE_FIELD, {
  context: projectScopedContext,
  refetchQueries: ['GetModel', 'GetModelJsonSchema'],  // ← 新增（字段类型变更影响表单渲染）
})
```

---

## 不在范围内（YAGNI）

- 字段联动/条件显示
- 内联编辑
- 自定义表单布局
- 文件上传字段
- 富文本编辑器
- RelationPicker 分页（初版取前 50 条）

---

## 验收标准

1. 删除字段后，表格列立即消失，无需手动刷新。
2. 新增字段后，表格列立即出现。
3. 新增/编辑记录的 Sheet 表单中：
   - ENUM 字段显示为下拉选择框（单选）
   - ENUM_ARRAY 字段显示为多选（Checkbox Popover）
   - DATE/DATETIME/TIME 显示日期/时间选择器
   - RELATION 字段显示关联记录搜索选择器
   - `storageHint === 'TEXT'` 的字段显示 textarea
4. 必填字段未填写时，提交后字段下方显示校验错误。
5. 数字字段超出 min/max 时，字段下方显示校验错误。
6. RelationPicker 加载失败时显示错误提示，不阻塞表单其他字段。
7. 服务端返回错误（如唯一键冲突）通过 toast 通知用户。
