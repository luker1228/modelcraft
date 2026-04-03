# Field Protocol & readOnly Design

**Date**: 2026-04-02
**Status**: Draft
**Scope**: Backend JSON Schema generator + Frontend field protocol layer

---

## 背景

`ModelRecordForm` 和 `DynamicModelTable` 当前存在两个 bug：

1. **RELATION 字段出现在创建/编辑表单中** — RELATION 字段是派生型字段，不应出现在 create/edit 表单里，但目前没有过滤。
2. **RELATION 字段在表格中显示为 `[object Object]`** — 表格单元格用 `String(item[field])` 渲染，RELATION 字段的值是对象（`{ id, name }`），转换结果是无意义字符串。

根本原因：没有一个集中的地方定义"遇到某类字段时应该怎么处理"。

---

## 目标

1. 用 JSON Schema 标准属性 `readOnly: true` 作为"此字段不可由用户编辑"的协议信号。
2. 建立前端 `fieldProtocol` 层，集中定义各字段类型在表单和表格中的行为。
3. 修复上述两个 bug，同时保持扩展性（未来新增字段类型只改协议层）。

---

## 方案

### 1. 后端：在 JSON Schema 中标记 `readOnly`

**文件**: `internal/domain/modeldesign/jsonschema_generator.go`

在 `applyCustomProperties()` 中，对以下两类字段追加 `"readOnly": true`：

- `field.IsPrimary == true` — 主键字段，由数据库生成，用户不可填写
- `field.Type.Format == FormatRelation` — 关联字段，只读展示，不可在表单中直接编辑

```go
// applyCustomProperties 中追加：
if field.IsPrimary || (field.Type != nil && field.Type.Format == FormatRelation) {
    schema["readOnly"] = true
}
```

生成的 JSON Schema 示例：

```json
{
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "x-isPrimary": true,
      "readOnly": true          // ← 新增
    },
    "user": {
      "type": "object",
      "x-relateFkId": "fk_abc",
      "readOnly": true          // ← 新增
    },
    "name": {
      "type": "string"
      // readOnly 不设置 → 可编辑
    }
  }
}
```

**设计说明**：
- `readOnly` 是 JSON Schema Draft 7 标准属性，符合规范
- 已有的 `x-isPrimary`、`x-relateFkId` 等扩展字段保留，供表格渲染使用
- RJSF 本身也识别 `readOnly`，即使前端过滤失效也有兜底

---

### 2. 前端：`filterJsonSchemaForForm()` 工具函数

**新文件**: `src/web/components/model-editor/ModelRecordForm/filterJsonSchemaForForm.ts`

```ts
import type { RJSFSchema } from '@rjsf/utils'

/**
 * 从 JSON Schema 中过滤 readOnly 字段，返回只包含可编辑字段的 schema。
 * 同时从 required 数组中移除被过滤的字段名。
 *
 * 协议约定：
 *   readOnly: true  →  不出现在 create/edit 表单中
 *   （主键字段、RELATION 字段由后端标记为 readOnly）
 */
export function filterJsonSchemaForForm(schema: RJSFSchema): RJSFSchema {
  if (!schema.properties) return schema

  const editableEntries = Object.entries(schema.properties).filter(
    ([, prop]) => !(prop as Record<string, unknown>).readOnly
  )

  const editableKeys = new Set(editableEntries.map(([k]) => k))

  return {
    ...schema,
    properties: Object.fromEntries(editableEntries),
    required: Array.isArray(schema.required)
      ? schema.required.filter((k) => editableKeys.has(k as string))
      : schema.required,
  }
}
```

---

### 3. 前端：`fieldProtocol.ts` — 集中定义字段渲染规则

**新文件**: `src/web/components/model-editor/fieldProtocol.ts`

字段协议是 JSON Schema property 对象到渲染行为的映射，驱动表单和表格两个场景。

```ts
import type { RJSFSchema } from '@rjsf/utils'

type SchemaProperty = Record<string, unknown>

/**
 * 判断字段是否应该出现在 create/edit 表单中。
 * 协议：readOnly: true 的字段不显示。
 */
export function shouldShowInForm(prop: SchemaProperty): boolean {
  return prop.readOnly !== true
}

/**
 * 渲染表格单元格值。
 * 协议：
 *   - RELATION 字段 (type=object, x-relateFkId 或 x-belongsToFkId 存在) → 显示 name 属性或 id
 *   - 普通值 → toString，截断至 100 字符
 *
 * 注：后端中 FormatRelation 同时用于 RelateFKID 和 BelongsToFKID 两类关联字段，
 * 两者在 JSON Schema 中均为 type=object，值均为 { id, name, ... } 对象。
 */
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''

  // RELATION 字段：值是 { id, name, ... } 对象
  // 同时检查 x-relateFkId 和 x-belongsToFkId，两者都使用 FormatRelation 格式
  if (prop.type === 'object' && (prop['x-relateFkId'] || prop['x-belongsToFkId'])) {
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      return String(rel.name ?? rel.id ?? '')
    }
    return ''
  }

  // 默认：转字符串并截断
  return String(value).slice(0, 100)
}

/**
 * 从完整 JSON Schema 中提取字段协议属性列表（按 x-displayOrder 排序）。
 */
export function getFieldProtocols(
  schema: RJSFSchema
): Array<{ name: string; prop: SchemaProperty }> {
  if (!schema.properties) return []

  return Object.entries(schema.properties)
    .map(([name, prop]) => ({ name, prop: prop as SchemaProperty }))
    .sort((a, b) => {
      const oa = String(a.prop['x-displayOrder'] ?? '')
      const ob = String(b.prop['x-displayOrder'] ?? '')
      return oa.localeCompare(ob)
    })
}
```

---

### 4. 前端：`ModelRecordForm` 使用过滤后的 schema

**文件**: `src/web/components/model-editor/ModelRecordForm/index.tsx`

在传给 RJSF 的 `schema` prop 上应用过滤：

```tsx
import { filterJsonSchemaForForm } from './filterJsonSchemaForForm'

// 在组件内：
const editableSchema = useMemo(
  () => filterJsonSchemaForForm(jsonSchema),
  [jsonSchema]
)

// 渲染时：
<Form
  schema={editableSchema}   // ← 使用过滤后的 schema，不再是原始 jsonSchema
  uiSchema={uiSchema}
  ...
/>
```

同时，`buildUiSchema.ts` 中做以下两处清理：

**1. 删除 RELATION → RelationPicker 的映射**（RELATION 字段已被过滤，映射无意义）：

```ts
// 删除以下片段：
if (field.format === 'RELATION') {
  uiSchema[field.name] = {
    'ui:widget': 'RelationPicker',
    ...
  }
  continue
}
```

**2. 删除 isPrimary → hidden 的映射**（主键字段已通过 `filterJsonSchemaForForm` 过滤，`ui:widget: hidden` 不再触达 RJSF，属于死代码）：

```ts
// 删除以下片段：
if (field.isPrimary) {
  uiSchema[field.name] = { 'ui:widget': 'hidden' }
  continue
}
```

> **注**：`RelationPicker` widget 本身可以保留，后续若有"关联字段内联展示"需求还可复用。

---

### 5. 前端：`DynamicModelTable` 使用 `renderCellValue`

**文件**: `src/web/components/model-editor/DynamicModelTable.tsx`

当前问题代码（约第 719 行）：
```tsx
{String(item[field]).slice(0, 100)}
```

修改为：
```tsx
import { renderCellValue, getFieldProtocols } from './fieldProtocol'

// 在组件顶部，从 jsonSchema 提取字段协议：
const fieldProtocols = useMemo(
  () => getFieldProtocols(jsonSchema),
  [jsonSchema]
)
const propByName = useMemo(
  () => Object.fromEntries(fieldProtocols.map(({ name, prop }) => [name, prop])),
  [fieldProtocols]
)

// 单元格渲染（field 是列迭代循环中的字符串列名，如 "id"、"user"、"name"）：
{renderCellValue(item[field], propByName[field] ?? {})}
```

---

## 数据流

```
后端 modelJsonSchema
  └─ id:    { type: "string", readOnly: true, x-isPrimary: true }
  └─ user:  { type: "object", readOnly: true, x-relateFkId: "fk_x" }
  └─ name:  { type: "string" }
         │
         ├─ filterJsonSchemaForForm()
         │      └─ 过滤 readOnly → { name: { type: "string" } }
         │             └─ 传给 RJSF → 表单只显示 name
         │
         └─ getFieldProtocols() + renderCellValue()
                └─ 表格列：id, user, name 都显示
                └─ user 单元格 → renderCellValue → "Alice"（不是 [object Object]）
```

---

## 修改文件清单

### 后端（`modelcraft-backend`）

| 文件 | 改动 |
|------|------|
| `internal/domain/modeldesign/jsonschema_generator.go` | `applyCustomProperties()` 追加 `readOnly: true` 逻辑（4 行） |
| `internal/domain/modeldesign/jsonschema_generator_test.go` | 补充断言：isPrimary 字段和 RELATION 字段有 `readOnly: true` |

### 前端（`modelcraft-front`）

| 文件 | 改动类型 | 说明 |
|------|----------|------|
| `src/web/components/model-editor/fieldProtocol.ts` | 新建 | 集中定义字段行为协议 |
| `src/web/components/model-editor/ModelRecordForm/filterJsonSchemaForForm.ts` | 新建 | JSON Schema 表单过滤工具 |
| `src/web/components/model-editor/ModelRecordForm/index.tsx` | 修改 | 应用 `filterJsonSchemaForForm` |
| `src/web/components/model-editor/ModelRecordForm/buildUiSchema.ts` | 修改 | 删除 RELATION → RelationPicker 映射 |
| `src/web/components/model-editor/DynamicModelTable.tsx` | 修改 | 用 `renderCellValue` 替换 `String(item[field])` |

---

## 不在本次范围内

- `RelationPicker` widget 的进一步功能改造
- 表格中 RELATION 字段的点击跳转
- 其他 `readOnly` 字段的特殊表格渲染（如主键高亮）

---

## 测试要点

1. 创建表单：RELATION 字段和主键字段不出现
2. 编辑表单：同上，且 `required` 不包含被过滤字段
3. 表格：RELATION 字段单元格显示 `name` 属性而非 `[object Object]`
4. 表格：普通字段渲染不受影响
5. 后端：`jsonschema_generator_test.go` 断言 `readOnly` 正确设置
6. 表格列排序：`x-displayOrder` 为字符串字典序键（如 `"a0"`、`"a1"`、`"a2"`），`getFieldProtocols` 按 `localeCompare` 排序，结果与后端 `DisplayOrder` 字段定义一致
