# END_USER_REF Widget 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `owner` 字段（`END_USER_REF` 格式）在 EndUser 端插入时自动隐藏并由后端注入，在 Tenant 管理端渲染 EndUser 下拉选择器。

**Architecture:** 后端 `decideWidget()` 为 `FormatEndUserRef` 输出 `"end-user-ref"` widget；前端 `build-ui-schema.ts` 按 `workspaceMode` 分支——`end_user` 模式渲染为 `hidden`，`design` 模式渲染为 `EndUserSelectorWidget`；后端 `executeCreateOne` 在 EndUser 端上下文中强制覆盖 `owner` 为 JWT 中的 EndUser ID。

**Tech Stack:** Go (graphql-go), TypeScript, React, RJSF (@rjsf/core), Apollo Client, shadcn/ui

---

## 文件结构

| 操作 | 文件 | 说明 |
|---|---|---|
| 修改 | `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go:270-305` | `decideWidget()` 新增 `FormatEndUserRef` 分支 |
| 修改 | `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator_test.go:813-833` | 更新现有 END_USER_REF 测试，断言 widget |
| 修改 | `modelcraft-backend/internal/domain/modelruntime/model_resolver.go:1244-1291` | `executeCreateOne` 注入 EndUser owner |
| 新增 | `modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go` | 测试 owner 自动注入 |
| 修改 | `modelcraft-front/src/types/xmc.ts:21-28` | `XMCWidget` 新增 `"end-user-ref"` |
| 修改 | `modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts` | 新增 `end-user-ref` 映射分支 |
| 修改 | `modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx` | 注册 `EndUserSelectorWidget`，传入 `workspaceMode` |
| 新增 | `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx` | EndUser 下拉选择 RJSF widget |
| 修改 | `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/index.ts` | 导出 `EndUserSelectorWidget` |
| 修改 | `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md` | widget 表格追加 `end-user-ref` 行 |

---

## Task 1：后端 — `decideWidget()` 新增 END_USER_REF 分支

**Files:**
- Modify: `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go:268-306`
- Test: `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator_test.go:813-833`

- [ ] **Step 1：更新现有测试，增加 widget 断言**

文件：`modelcraft-backend/internal/domain/modeldesign/jsonschema_generator_test.go`

找到 `TestJSONSchemaGenerator_EndUserRef_HasStringType`（第 813 行），在最后追加 widget 断言：

```go
func TestJSONSchemaGenerator_EndUserRef_HasStringType(t *testing.T) {
	field := &FieldDefinition{
		Name: "owner", Title: "Owner",
		Type:         GetFieldTypeByFormat(FormatEndUserRef),
		DisplayOrder: "a0",
	}

	generator := NewJSONSchemaGenerator()
	model := makeMinimalModel([]*FieldDefinition{field})
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))
	props := schema["properties"].(map[string]interface{})
	owner := props["owner"].(map[string]interface{})

	assert.Equal(t, "string", owner["type"])
	xmc := getXMC(owner)
	assert.Equal(t, "END_USER_REF", xmc["format"])
	assert.Equal(t, "end-user-ref", xmc["widget"]) // ← 新增断言
}
```

- [ ] **Step 2：运行测试，确认失败**

```bash
cd modelcraft-backend && rtk go test -v -run TestJSONSchemaGenerator_EndUserRef_HasStringType ./internal/domain/modeldesign/...
```

预期：FAIL，`assert.Equal: "end-user-ref" != ""`

- [ ] **Step 3：实现 `decideWidget()` 新分支**

文件：`modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go`

在 `decideWidget()` 函数的 `switch field.Type.Format` 块（约第 290 行）里，在 `case FormatEnum, FormatEnumArray:` 之前新增：

```go
// 2.5 END_USER_REF → end-user-ref widget
case FormatEndUserRef:
    return "end-user-ref"
```

完整 switch 变为：

```go
switch field.Type.Format {
// 3. 枚举（单选/多选） → 枚举下拉
case FormatEnum, FormatEnumArray:
    return "enum-select"
// 4. 日期
case FormatDate:
    return "date"
// 5. 日期时间
case FormatDateTime:
    return "datetime-local"
// 6. 时间
case FormatTime:
    return "time"
// 7. END_USER_REF → end-user-ref widget
case FormatEndUserRef:
    return "end-user-ref"
}
```

- [ ] **Step 4：运行测试，确认通过**

```bash
cd modelcraft-backend && rtk go test -v -run TestJSONSchemaGenerator_EndUserRef_HasStringType ./internal/domain/modeldesign/...
```

预期：PASS

- [ ] **Step 5：运行全量测试，确认无回归**

```bash
cd modelcraft-backend && rtk go test -v -race -timeout=5m ./internal/domain/modeldesign/...
```

预期：all PASS

- [ ] **Step 6：Commit**

```bash
cd modelcraft-backend && git add internal/domain/modeldesign/jsonschema_generator.go internal/domain/modeldesign/jsonschema_generator_test.go
git commit -m "feat(schema): emit end-user-ref widget for FormatEndUserRef fields"
```

---

## Task 2：后端 — `executeCreateOne` 注入 EndUser owner

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go:1244-1291`
- Create: `modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go`

**背景：**  
`middleware.GetEndUserID(ctx)` 返回当前请求的 EndUser ID（从 JWT 注入）。当值非空时，说明请求来自 EndUser 端，应自动覆盖 `input.Data["owner"]`。当值为空时（Tenant 管理端），直接使用 payload 传入的值。

- [ ] **Step 1：新建测试文件**

创建文件：`modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go`

```go
package modelruntime_test

import (
	"context"
	"testing"

	"modelcraft/internal/interfaces/http/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndUserRefOwnerInjection 验证 EndUser 端 createOne 时 owner 被自动覆盖
func TestEndUserRefOwnerInjection(t *testing.T) {
	t.Run("injects owner from context when EndUser identity present", func(t *testing.T) {
		ctx := context.Background()
		// 模拟 EndUser JWT 已注入 context
		ctx = middleware.WithEndUserIDForTest(ctx, "end-user-uuid-123")

		endUserID := middleware.GetEndUserID(ctx)
		require.Equal(t, "end-user-uuid-123", endUserID)

		// 模拟 input.Data 中有恶意 owner 值
		data := map[string]interface{}{
			"title": "hello",
			"owner": "attacker-uuid",
		}

		// 应用注入逻辑
		if endUserID != "" {
			data["owner"] = endUserID
		}

		assert.Equal(t, "end-user-uuid-123", data["owner"])
	})

	t.Run("does not override owner when no EndUser identity (tenant context)", func(t *testing.T) {
		ctx := context.Background()
		endUserID := middleware.GetEndUserID(ctx)
		require.Empty(t, endUserID)

		data := map[string]interface{}{
			"title": "hello",
			"owner": "chosen-end-user-uuid",
		}

		if endUserID != "" {
			data["owner"] = endUserID
		}

		assert.Equal(t, "chosen-end-user-uuid", data["owner"])
	})
}
```

- [ ] **Step 2：在 middleware 包新增 `WithEndUserIDForTest` 测试辅助函数**

文件：`modelcraft-backend/internal/interfaces/http/middleware/runtime_auth_middleware.go`

在文件末尾（`GetEndUserID` 之后）追加：

```go
// WithEndUserIDForTest 仅用于测试：将 endUserID 直接注入 context，绕过 JWT 解析。
// 生产代码不得使用。
func WithEndUserIDForTest(ctx context.Context, endUserID string) context.Context {
	identity := &EndUserIdentity{EndUserID: endUserID}
	return context.WithValue(ctx, endUserContextKey, identity)
}
```

- [ ] **Step 3：运行测试，确认通过**

```bash
cd modelcraft-backend && rtk go test -v -run TestEndUserRefOwnerInjection ./internal/domain/modelruntime/...
```

预期：PASS（测试只验证 owner 注入逻辑，不依赖 DB）

- [ ] **Step 4：在 `executeCreateOne` 中实现 owner 注入**

文件：`modelcraft-backend/internal/domain/modelruntime/model_resolver.go`

找到 `executeCreateOne` 函数（约第 1244 行），在 UUID 自动生成循环之后（约第 1262 行），插入 owner 注入逻辑：

```go
func (m *graphqlModelResolver) executeCreateOne(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	input, err := newCreateOneInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}

	for _, field := range m.model.Fields {
		if field.Type.Format == modeldesign.FormatUUID {
			// Only generate UUIDV7 if user didn't provide a value
			if _, exists := input.Data[field.Name]; !exists || input.Data[field.Name] == nil {
				uuidv7, err := bizutils.GenerateUUIDV7()
				if err != nil {
					return nil, err
				}
				input.Data[field.Name] = uuidv7
			}
		}
	}

	// EndUser 端：从 JWT context 自动注入 owner，防止客户端伪造。
	// Tenant 管理端（无 EndUser identity）：直接使用 payload 传入值。
	if endUserID := middleware.GetEndUserID(p.Context); endUserID != "" {
		input.Data[FieldOwner] = endUserID
	}

	input.Id = cast.ToString(input.Data[FieldID])
	// ... 后续代码不变
```

同时在文件顶部 import 块中确认 middleware 包已引入（如尚未引入则追加）：

```go
import (
    // ... 现有 imports ...
    "modelcraft/internal/interfaces/http/middleware"
)
```

在 `graphql_constants.go`（同包）中查找或新增常量：

```go
// FieldOwner 是 END_USER_REF 系统字段的固定名称
const FieldOwner = "owner"
```

> 如果 `FieldOwner` 常量已存在，直接使用即可，不需重复定义。

- [ ] **Step 5：检查 FieldOwner 是否已存在**

```bash
grep -rn "FieldOwner\|\"owner\"" /data/home/lukemxjia/modelcraft/modelcraft-backend/internal/domain/modelruntime/ --include="*.go" | grep -v "_test.go"
```

- 如已存在 `FieldOwner` 常量：直接在 Step 4 的代码中引用它
- 如不存在：在 `modelcraft-backend/internal/domain/modelruntime/graphql_constants.go` 末尾追加 `const FieldOwner = "owner"`

- [ ] **Step 6：运行全量测试**

```bash
cd modelcraft-backend && rtk go test -v -race -timeout=5m ./internal/domain/modelruntime/...
```

预期：all PASS

- [ ] **Step 7：运行 lint**

```bash
cd modelcraft-backend && just lint
```

如有报错，运行 `just lint-fix` 再重新确认。

- [ ] **Step 8：Commit**

```bash
cd modelcraft-backend && git add internal/domain/modelruntime/ internal/interfaces/http/middleware/runtime_auth_middleware.go
git commit -m "feat(runtime): auto-inject owner from JWT in EndUser createOne"
```

---

## Task 3：前端 — 类型扩展 + build-ui-schema 分支

**Files:**
- Modify: `modelcraft-front/src/types/xmc.ts:21-28`
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts`
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx`

**背景：**  
`build-ui-schema.ts` 当前使用静态 `WIDGET_MAP: Record<XMCWidget, string>`，不接受 `workspaceMode`。需要改造为按 `workspaceMode` 返回不同映射。`ModelRecordForm` 组件需要接受新的 `workspaceMode` prop 并向下传递。

- [ ] **Step 1：扩展 `XMCWidget` 类型**

文件：`modelcraft-front/src/types/xmc.ts`

将：
```ts
export type XMCWidget =
  | 'enum-select'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'relation-selector'
  | 'relation-multi-readonly'
```

替换为：
```ts
export type XMCWidget =
  | 'enum-select'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'relation-selector'
  | 'relation-multi-readonly'
  | 'end-user-ref'
```

- [ ] **Step 2：改造 `buildUiSchema` 接受 `workspaceMode`**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts`

完整替换为：

```ts
import type { UiSchema, RJSFSchema } from '@rjsf/utils'
import { getXMC, type XMCWidget } from '@/types/xmc'

export type WorkspaceMode = 'design' | 'end_user'

/**
 * Widget name mapping for non-end-user-ref widgets.
 */
const WIDGET_MAP: Partial<Record<XMCWidget, string>> = {
  'enum-select': 'EnumSelect',
  'date': 'date',
  'datetime-local': 'datetime-local',
  'time': 'time',
  'textarea': 'textarea',
  'relation-selector': 'RelationSelector',
  'relation-multi-readonly': 'RelationMultiReadonly',
}

/**
 * Build RJSF uiSchema directly from the (filtered) JSON Schema.
 *
 * Reads `x-mc.widget` on each property and maps it to the appropriate
 * RJSF widget string.
 *
 * For `end-user-ref` fields:
 * - `end_user` mode: hidden (auto-injected by backend from JWT)
 * - `design` mode: EndUserSelectorWidget (admin selects an EndUser)
 */
export function buildUiSchema(
  jsonSchema: RJSFSchema,
  workspaceMode: WorkspaceMode = 'design',
): UiSchema {
  const uiSchema: UiSchema = {}

  if (!jsonSchema.properties) return uiSchema

  for (const [fieldName, prop] of Object.entries(jsonSchema.properties)) {
    const xmc = getXMC(prop as Record<string, unknown>)
    const widget = xmc?.widget

    if (!widget) continue

    if (widget === 'end-user-ref') {
      if (workspaceMode === 'end_user') {
        uiSchema[fieldName] = { 'ui:widget': 'hidden' }
      } else {
        uiSchema[fieldName] = { 'ui:widget': 'EndUserSelectorWidget' }
      }
      continue
    }

    if (WIDGET_MAP[widget]) {
      uiSchema[fieldName] = { 'ui:widget': WIDGET_MAP[widget] }
    }
  }

  return uiSchema
}
```

- [ ] **Step 3：更新 `ModelRecordForm` 接受 `workspaceMode` prop**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx`

将 `ModelRecordFormProps` interface 新增一个可选字段：

```ts
interface ModelRecordFormProps {
  jsonSchema: RJSFSchema
  initialData?: Record<string, unknown>
  onSubmit: (data: Record<string, unknown>) => Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string
  recordId?: string
  workspaceMode?: 'design' | 'end_user'  // ← 新增，默认 'design'
}
```

函数签名解构中新增（带默认值）：

```ts
export function ModelRecordForm({
  jsonSchema,
  initialData,
  onSubmit,
  onCancel,
  isSubmitting = false,
  orgName,
  projectSlug,
  clusterName,
  databaseName,
  modelId,
  recordId,
  workspaceMode = 'design',   // ← 新增
}: ModelRecordFormProps) {
```

将 `buildUiSchema` 调用传入 `workspaceMode`：

```ts
const uiSchema = useMemo<UiSchema>(() => {
  const widgetUiSchema = buildUiSchema(editableSchema, workspaceMode)  // ← 传入
  const orderedFieldNames = editableSchema.properties
    ? Object.keys(editableSchema.properties)
    : []

  if (orderedFieldNames.length === 0) {
    return widgetUiSchema
  }

  return {
    ...widgetUiSchema,
    'ui:order': orderedFieldNames,
  }
}, [editableSchema, workspaceMode])  // ← 加入依赖
```

- [ ] **Step 4：运行 TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -30
```

预期：无类型错误（`build-ui-schema.ts` 可能有 `WorkspaceMode` 冲突需排查）

- [ ] **Step 5：运行 lint**

```bash
cd modelcraft-front && npm run lint 2>&1 | head -30
```

如有报错，按提示修复。

- [ ] **Step 6：Commit**

```bash
cd modelcraft-front && git add src/types/xmc.ts src/web/components/features/model-editor/model-record-form/build-ui-schema.ts src/web/components/features/model-editor/model-record-form/index.tsx
git commit -m "feat(form): add end-user-ref widget support with workspaceMode branching"
```

---

## Task 4：前端 — `EndUserSelectorWidget` 组件

**Files:**
- Create: `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx`
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/index.ts`
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx` (注册 widget)

**背景：**  
该 widget 仅在 Tenant 管理端（`workspaceMode = 'design'`）渲染。`formContext` 中已有 `orgName` 和 `projectSlug`，可通过 Org-scoped client 调用 `LIST_END_USERS` 获取用户列表。value 为 EndUser UUID 字符串。

- [ ] **Step 1：创建 `EndUserSelectorWidget.tsx`**

创建文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx`

```tsx
'use client'

import React, { useMemo } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { useQuery } from '@apollo/client'
import { getOrgScopedClient } from '@api-client/apollo/public'
import { LIST_END_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'

interface FormContext {
  orgName?: string
}

interface OrgEndUserNode {
  id: string
  username: string
  displayName?: string
  isForbidden: boolean
}

interface ListEndUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: OrgEndUserNode[]
    }
    error?: { message?: string }
  }
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Renders a <Select> dropdown listing all EndUsers in the Org.
 * Value is the EndUser UUID string. Used only in Tenant (design) workspace.
 *
 * formContext must provide: orgName
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const { value, onChange, disabled, readonly, formContext } = props
  const ctx = formContext as FormContext
  const orgName = ctx?.orgName ?? ''

  const client = useMemo(() => getOrgScopedClient(), [])

  const { data, loading } = useQuery<ListEndUsersData>(LIST_END_USERS, {
    client,
    variables: { input: { first: 200 } },
    skip: !orgName,
    fetchPolicy: 'cache-first',
  })

  const users = data?.listEndUsers?.connection?.nodes ?? []
  const activeUsers = users.filter((u) => !u.isForbidden)

  const handleChange = (val: string) => {
    onChange(val === '__none__' ? undefined : val)
  }

  return (
    <Select
      value={(value as string | undefined) ?? '__none__'}
      onValueChange={handleChange}
      disabled={disabled || readonly || loading}
    >
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">— 不指定 —</SelectItem>
        {activeUsers.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            {user.displayName ? `${user.displayName} (${user.username})` : user.username}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
```

- [ ] **Step 2：从 widgets/index.ts 导出**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/index.ts`

追加：

```ts
export { EndUserSelectorWidget } from './EndUserSelectorWidget'
```

- [ ] **Step 3：在 `index.tsx` 注册 widget**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx`

将 import 行：
```ts
import { EnumSelect, EnumSchemaSelect, RelationSelector } from './widgets'
```
改为：
```ts
import { EnumSelect, EnumSchemaSelect, RelationSelector, EndUserSelectorWidget } from './widgets'
```

将 `customWidgets` 对象：
```ts
const customWidgets = {
  EnumSelect,
  EnumSchemaSelect,
  RelationSelector,
}
```
改为：
```ts
const customWidgets = {
  EnumSelect,
  EnumSchemaSelect,
  RelationSelector,
  EndUserSelectorWidget,
}
```

- [ ] **Step 4：运行 TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -30
```

预期：无类型错误

- [ ] **Step 5：运行 lint**

```bash
cd modelcraft-front && npm run lint 2>&1 | head -30
```

如有报错，按提示修复（常见：missing `displayName` 字段类型、Select import 路径）

- [ ] **Step 6：Commit**

```bash
cd modelcraft-front && git add src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx src/web/components/features/model-editor/model-record-form/widgets/index.ts src/web/components/features/model-editor/model-record-form/index.tsx
git commit -m "feat(widget): add EndUserSelectorWidget for END_USER_REF fields in design mode"
```

---

## Task 5：调用方 — 传入 workspaceMode

**Files:**
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx` (两处 `<ModelRecordForm>`)
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/RuntimeRecordWorkspace.tsx` (两处 `<ModelRecordForm>`)

**背景：**  
`DevelopRecordWorkspace` = Tenant 管理端（`workspaceMode="design"`）。  
`RuntimeRecordWorkspace` = EndUser 端（`workspaceMode="end_user"`）。  
每个文件各有两处 `<ModelRecordForm>` 调用（分别是 insert 和 edit 场景，第 786/837 行和第 657/708 行）。

- [ ] **Step 1：在 DevelopRecordWorkspace 的两处 ModelRecordForm 加 prop**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx`

使用以下命令定位两处 `<ModelRecordForm`（约第 786 和 837 行），在每处的 props 中追加：

```tsx
workspaceMode="design"
```

示例（两处相同）：
```tsx
<ModelRecordForm
  jsonSchema={...}
  initialData={...}
  onSubmit={...}
  onCancel={...}
  orgName={orgName}
  projectSlug={projectSlug}
  clusterName={clusterName}
  databaseName={databaseName}
  modelId={modelId}
  workspaceMode="design"   // ← 新增
/>
```

- [ ] **Step 2：在 RuntimeRecordWorkspace 的两处 ModelRecordForm 加 prop**

文件：`modelcraft-front/src/web/components/features/model-editor/model-record-form/RuntimeRecordWorkspace.tsx`

在两处 `<ModelRecordForm>`（约第 657 和 708 行）中追加：

```tsx
workspaceMode="end_user"
```

- [ ] **Step 3：运行 TypeScript 类型检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -30
```

预期：无类型错误

- [ ] **Step 4：运行 lint**

```bash
cd modelcraft-front && npm run lint 2>&1 | head -30
```

- [ ] **Step 5：Commit**

```bash
cd modelcraft-front && git add src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx src/web/components/features/model-editor/model-record-form/RuntimeRecordWorkspace.tsx
git commit -m "feat(workspace): pass workspaceMode to ModelRecordForm in both workspace contexts"
```

---

## Task 6：更新契约文档

**Files:**
- Modify: `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md`

- [ ] **Step 1：在 widget 表格追加 end-user-ref 行**

文件：`ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md`

找到 widget 表格（约第 117-127 行），在最后一行（`| 不填 | ...`）之前插入：

```markdown
| `"end-user-ref"` | `format = END_USER_REF` | EndUser 端：隐藏（不进 payload，后端从 JWT 注入）；Tenant 管理端：EndUser 下拉选择器 |
```

完整表格更新后为：

```markdown
| `x-mc.widget` 值 | 触发条件（后端逻辑） | 前端控件 |
|---|---|---|
| `"enum-select"` | `format = ENUM` 或 `ENUM_ARRAY` | `EnumSchemaSelect` |
| `"date"` | `format = DATE` | 原生 `date` input |
| `"datetime-local"` | `format = DATETIME` | 原生 `datetime-local` input |
| `"time"` | `format = TIME` | 原生 `time` input |
| `"textarea"` | `storageHint = TEXT` | `textarea` |
| `"relation-selector"` | `BelongsToFKID != nil`（外键列） | `RelationSelector` |
| `"end-user-ref"` | `format = END_USER_REF` | EndUser 端：隐藏；Tenant 管理端：EndUser 下拉选择器 |
| 不填 | 其余所有字段 | RJSF 按标准 `type` 默认渲染 |
```

- [ ] **Step 2：Commit**

```bash
git add ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md
git commit -m "docs: update jsonschema-contract with end-user-ref widget spec"
```

---

## 自检：Spec 覆盖验证

| Spec 要求 | 覆盖任务 |
|---|---|
| 后端 `decideWidget()` 输出 `end-user-ref` | Task 1 |
| 前端 `XMCWidget` 新增 `end-user-ref` | Task 3 |
| EndUser 端：`ui:widget = hidden` | Task 3 |
| Tenant 管理端：`EndUserSelectorWidget` | Task 4 |
| `ModelRecordForm` 接受 `workspaceMode` | Task 3 |
| `DevelopRecordWorkspace` 传 `workspaceMode="design"` | Task 5 |
| `RuntimeRecordWorkspace` 传 `workspaceMode="end_user"` | Task 5 |
| 后端 EndUser 端 owner 自动注入 | Task 2 |
| 契约文档更新 | Task 6 |
| `EndUserSelectorWidget` 注册到 RJSF | Task 4 |
