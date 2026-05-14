# Filter Panel 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 EndUser 数据视图的工具栏下方添加可展开的筛选面板，用户通过 JSON 编辑器直接编写 GraphQL `where` 条件，后端执行真实过滤。

**Architecture:** 新建 `FilterPanel` 组件（包含 `WhereJsonEditor` 和 `FieldSchemaPanel` 两个子组件），挂载在 `EndUserRecordWorkspace` 工具栏下方。筛选状态分草稿态（`whereJsonDraft`，随输入变化）和已提交态（`whereJsonCommitted`，点击"应用筛选"后更新），只有已提交态才传入 `useQuery` 的 `where` 变量触发后端查询。

**Tech Stack:** React + TypeScript，Tailwind CSS，shadcn/ui（Button、Badge），lucide-react，原生 `<textarea>`（无需引入 CodeMirror，项目中不存在此依赖）

---

## 文件结构

| 操作 | 路径 | 职责 |
|------|------|------|
| **新建** | `src/web/components/features/end-user-data/FilterPanel.tsx` | 面板容器：持有两列布局，组合 WhereJsonEditor 和 FieldSchemaPanel |
| **新建** | `src/web/components/features/end-user-data/WhereJsonEditor.tsx` | 文本编辑器：textarea + JSON 实时校验 + 格式化/清空/应用按钮 |
| **新建** | `src/web/components/features/end-user-data/FieldSchemaPanel.tsx` | 字段速查面板：展示当前模型字段名/类型 + 操作符参考，点击插入片段 |
| **新建** | `src/web/components/features/end-user-data/filter-utils.ts` | 纯函数工具：`getFilterCount`、`isValidJson`、`formatJson` |
| **新建** | `src/web/components/features/end-user-data/filter-utils.test.ts` | 工具函数单元测试 |
| **修改** | `src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx` | 接入 filterOpen/whereJsonDraft/whereJsonCommitted 状态，渲染 FilterPanel，给 useQuery 传 where |
| **修改** | `src/web/components/features/end-user-data/index.ts` | 导出新组件（若外部需要） |

---

## Task 1: 工具函数 + 测试

**Files:**
- Create: `src/web/components/features/end-user-data/filter-utils.ts`
- Create: `src/web/components/features/end-user-data/filter-utils.test.ts`

- [ ] **Step 1: 写失败测试**

新建 `src/web/components/features/end-user-data/filter-utils.test.ts`：

```typescript
import { describe, it, expect } from 'vitest'
import { isValidJson, formatJson, getFilterCount } from './filter-utils'

describe('isValidJson', () => {
  it('returns true for valid JSON object', () => {
    expect(isValidJson('{"name": {"contains": "张"}}')).toBe(true)
  })
  it('returns false for empty string', () => {
    expect(isValidJson('')).toBe(false)
  })
  it('returns false for invalid JSON', () => {
    expect(isValidJson('{bad json')).toBe(false)
  })
  it('returns false for JSON array (not an object)', () => {
    expect(isValidJson('[1,2,3]')).toBe(false)
  })
})

describe('formatJson', () => {
  it('pretty-prints valid JSON', () => {
    expect(formatJson('{"a":1}')).toBe('{\n  "a": 1\n}')
  })
  it('returns original string on invalid JSON', () => {
    expect(formatJson('{bad')).toBe('{bad')
  })
})

describe('getFilterCount', () => {
  it('returns null for null input', () => {
    expect(getFilterCount(null)).toBeNull()
  })
  it('returns AND array length', () => {
    expect(getFilterCount('{"AND":[{},{}]}')).toBe(2)
  })
  it('returns OR array length', () => {
    expect(getFilterCount('{"OR":[{}]}')).toBe(1)
  })
  it('returns bullet for single-field condition', () => {
    expect(getFilterCount('{"name":{"contains":"张"}}')).toBe('•')
  })
  it('returns null for invalid JSON', () => {
    expect(getFilterCount('{bad')).toBeNull()
  })
  it('returns null for empty object', () => {
    expect(getFilterCount('{}')).toBeNull()
  })
})
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-front && npx vitest run src/web/components/features/end-user-data/filter-utils.test.ts
```

期望：FAIL，报 `Cannot find module './filter-utils'`

- [ ] **Step 3: 实现工具函数**

新建 `src/web/components/features/end-user-data/filter-utils.ts`：

```typescript
/**
 * Check if a string is a valid JSON object (not array, not primitive).
 */
export function isValidJson(value: string): boolean {
  if (!value.trim()) return false
  try {
    const parsed = JSON.parse(value)
    return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)
  } catch {
    return false
  }
}

/**
 * Pretty-print JSON. Returns original string on parse failure.
 */
export function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

/**
 * Count top-level filter conditions for the filter button badge.
 *
 * Returns:
 * - number  — length of AND/OR array
 * - '•'     — non-AND/OR structure (single-field condition)
 * - null    — no active filter (empty, invalid JSON, or empty object)
 *
 * AI note: this reads `whereJsonCommitted`. AI agents should write a valid
 * where JSON string and this function will derive the badge automatically.
 */
export function getFilterCount(whereJson: string | null): number | '•' | null {
  if (!whereJson) return null
  try {
    const parsed = JSON.parse(whereJson)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null
    if (Array.isArray(parsed.AND)) return parsed.AND.length
    if (Array.isArray(parsed.OR)) return parsed.OR.length
    const keys = Object.keys(parsed).filter((k) => k !== 'NOT')
    return keys.length > 0 ? '•' : null
  } catch {
    return null
  }
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-front && npx vitest run src/web/components/features/end-user-data/filter-utils.test.ts
```

期望：全部 PASS（8 个测试）

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/filter-utils.ts \
        modelcraft-front/src/web/components/features/end-user-data/filter-utils.test.ts
git commit -m "feat(front): add filter panel utility functions with tests"
```

---

## Task 2: FieldSchemaPanel 组件

**Files:**
- Create: `src/web/components/features/end-user-data/FieldSchemaPanel.tsx`

> 这个组件纯展示，无副作用，不需要单独的单元测试（会在集成中验证）。

- [ ] **Step 1: 创建组件**

新建 `src/web/components/features/end-user-data/FieldSchemaPanel.tsx`：

```typescript
import React from 'react'
import type { FieldDefinition } from '@api-client/cms/public'

// Operator reference per field storage type.
// Shown as static reference — not exhaustive, covers common cases.
const OPERATOR_REFERENCE = [
  'equals / not',
  'contains / startsWith',
  'gt / gte / lt / lte',
  'in: [...]',
  'AND / OR / NOT',
] as const

/**
 * Derives a short human-readable type label from a FieldDefinition.
 * Used only for display in the schema panel sidebar.
 */
function getTypeLabel(field: FieldDefinition): string {
  const fmt = field.format?.toUpperCase()
  if (fmt === 'RELATION') return 'Relation'
  const hint = field.storageHint?.toUpperCase()
  if (hint === 'BOOL' || hint === 'BOOLEAN') return 'Bool'
  if (hint === 'INT' || hint === 'BIGINT') return 'Int'
  if (hint === 'FLOAT' || hint === 'DECIMAL') return 'Float'
  if (hint === 'DATETIME' || hint === 'DATE') return 'Date'
  if (fmt === 'ENUM') return 'Enum'
  return 'String'
}

export interface FieldSchemaPanelProps {
  fields: FieldDefinition[]
  /** Called with a JSON snippet string when a field is clicked. */
  onFieldClick: (snippet: string) => void
}

/**
 * Sidebar showing the current model's field names, types, and operator reference.
 *
 * AI note: the field list here is the schema context an AI agent needs to
 * construct a valid where JSON. Pass this list to your AI prompt as context.
 */
export function FieldSchemaPanel({ fields, onFieldClick }: FieldSchemaPanelProps) {
  // Filter out internal/display fields (e.g. _displayName suffixed fields)
  const displayFields = fields.filter((f) => !f.name.startsWith('_'))

  function handleFieldClick(field: FieldDefinition) {
    // Insert a starter snippet: "fieldName": {}
    onFieldClick(`"${field.name}": {}`)
  }

  return (
    <div className="flex w-44 shrink-0 flex-col gap-3 rounded-md border border-border bg-card p-3 text-xs">
      <div>
        <p className="mb-1.5 font-medium text-foreground">字段</p>
        <div className="flex flex-col gap-1">
          {displayFields.map((field) => (
            <button
              key={field.name}
              type="button"
              onClick={() => handleFieldClick(field)}
              className="flex items-center justify-between rounded px-1.5 py-1 text-left hover:bg-muted"
              title={`点击插入 "${field.name}": {}`}
            >
              <span className="font-mono text-primary">{field.name}</span>
              <span className="rounded bg-muted px-1 text-[10px] text-muted-foreground">
                {getTypeLabel(field)}
              </span>
            </button>
          ))}
        </div>
      </div>

      <div>
        <p className="mb-1.5 font-medium text-foreground">操作符</p>
        <div className="rounded bg-muted px-2 py-1.5 font-mono text-[10px] leading-relaxed text-muted-foreground">
          {OPERATOR_REFERENCE.map((op) => (
            <div key={op}>{op}</div>
          ))}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/FieldSchemaPanel.tsx
git commit -m "feat(front): add FieldSchemaPanel component for filter sidebar"
```

---

## Task 3: WhereJsonEditor 组件

**Files:**
- Create: `src/web/components/features/end-user-data/WhereJsonEditor.tsx`

- [ ] **Step 1: 创建组件**

新建 `src/web/components/features/end-user-data/WhereJsonEditor.tsx`：

```typescript
import React, { useRef } from 'react'
import { Button } from '@web/components/ui/button'
import { cn } from '@/shared/utils'

export interface WhereJsonEditorProps {
  value: string
  onChange: (value: string) => void
  onFormat: () => void
  onClear: () => void
  onApply: () => void
  isValid: boolean
}

/**
 * Plain textarea JSON editor with validation status, format, clear, and apply actions.
 *
 * Exposes a ref-based `insertSnippet` method so parent components (FilterPanel)
 * can programmatically insert text at the current cursor position — used by
 * FieldSchemaPanel when a field is clicked.
 *
 * AI note: to programmatically set the filter, call `onChange` with the full
 * where JSON string, then call `onApply`. No ref manipulation needed from AI.
 */
export interface WhereJsonEditorRef {
  insertAtCursor: (snippet: string) => void
}

export const WhereJsonEditor = React.forwardRef<WhereJsonEditorRef, WhereJsonEditorProps>(
  function WhereJsonEditor({ value, onChange, onFormat, onClear, onApply, isValid }, ref) {
    const textareaRef = useRef<HTMLTextAreaElement>(null)

    React.useImperativeHandle(ref, () => ({
      insertAtCursor(snippet: string) {
        const el = textareaRef.current
        if (!el) return
        const start = el.selectionStart
        const end = el.selectionEnd
        const newValue = value.slice(0, start) + snippet + value.slice(end)
        onChange(newValue)
        // Restore cursor after the inserted snippet
        requestAnimationFrame(() => {
          el.selectionStart = start + snippet.length
          el.selectionEnd = start + snippet.length
          el.focus()
        })
      },
    }))

    const isEmpty = !value.trim()
    const showError = !isEmpty && !isValid

    return (
      <div className="flex flex-1 flex-col gap-2">
        {/* Header row */}
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-foreground">Where JSON</span>
          <div className="flex gap-1.5">
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={onFormat}
              disabled={isEmpty || !isValid}
            >
              格式化
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={onClear}
              disabled={isEmpty}
            >
              清空
            </Button>
          </div>
        </div>

        {/* Textarea */}
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={'{\n  "AND": [\n    { "fieldName": { "contains": "value" } }\n  ]\n}'}
          spellCheck={false}
          className={cn(
            'min-h-[120px] w-full resize-y rounded-md border bg-[#1e1e2e] p-3 font-mono text-[11px] leading-relaxed text-[#cdd6f4] placeholder:text-[#6c7086] focus:outline-none focus:ring-1',
            showError ? 'border-destructive focus:ring-destructive' : 'border-border focus:ring-ring'
          )}
        />

        {/* Footer row */}
        <div className="flex items-center justify-between">
          <span
            className={cn(
              'text-[11px]',
              isEmpty
                ? 'text-muted-foreground'
                : isValid
                  ? 'text-green-600'
                  : 'text-destructive'
            )}
          >
            {isEmpty ? '输入 where 条件后点击应用' : isValid ? '✓ 有效 JSON' : '✗ JSON 格式错误'}
          </span>
          <Button
            size="sm"
            className="h-7 px-3 text-xs"
            onClick={onApply}
            disabled={!isValid && !isEmpty}
          >
            应用筛选
          </Button>
        </div>
      </div>
    )
  }
)
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/WhereJsonEditor.tsx
git commit -m "feat(front): add WhereJsonEditor component with cursor-insert support"
```

---

## Task 4: FilterPanel 组件（组合两个子组件）

**Files:**
- Create: `src/web/components/features/end-user-data/FilterPanel.tsx`

- [ ] **Step 1: 创建组件**

新建 `src/web/components/features/end-user-data/FilterPanel.tsx`：

```typescript
import React, { useRef } from 'react'
import type { FieldDefinition } from '@api-client/cms/public'
import { WhereJsonEditor, type WhereJsonEditorRef } from './WhereJsonEditor'
import { FieldSchemaPanel } from './FieldSchemaPanel'
import { isValidJson, formatJson } from './filter-utils'

export interface FilterPanelProps {
  /** Field definitions from the current model's jsonSchema (runtimeFields). */
  fields: FieldDefinition[]
  /** Draft JSON string — changes on every keystroke, does NOT trigger a query. */
  whereJsonDraft: string
  /** Called on every keystroke in the editor. */
  onWhereJsonDraftChange: (json: string) => void
  /**
   * Called when the user clicks "应用筛选".
   * The parent is responsible for committing whereJsonDraft → whereJsonCommitted.
   *
   * AI note: to programmatically apply a filter, set whereJsonDraft to a valid
   * where JSON and then call onApply. The parent will commit and trigger the query.
   */
  onApply: () => void
  /**
   * Called when the user clicks "清空". Bypasses the draft/apply flow entirely
   * so the parent can atomically clear both draft and committed state.
   * (Cannot use onApply here because React batches state updates — the draft
   * cleared by setWhereJsonDraft('') would not be visible to onApply yet.)
   */
  onClear: () => void
}

export function FilterPanel({
  fields,
  whereJsonDraft,
  onWhereJsonDraftChange,
  onApply,
  onClear,
}: FilterPanelProps) {
  const editorRef = useRef<WhereJsonEditorRef>(null)

  const valid = isValidJson(whereJsonDraft)

  function handleFormat() {
    onWhereJsonDraftChange(formatJson(whereJsonDraft))
  }

  function handleClear() {
    onClear() // Parent atomically clears draft + committed state
  }

  function handleFieldClick(snippet: string) {
    editorRef.current?.insertAtCursor(snippet)
  }

  return (
    <div className="flex gap-3 border-b border-border bg-muted/30 px-4 py-3">
      <WhereJsonEditor
        ref={editorRef}
        value={whereJsonDraft}
        onChange={onWhereJsonDraftChange}
        onFormat={handleFormat}
        onClear={handleClear}
        onApply={onApply}
        isValid={valid}
      />
      <FieldSchemaPanel fields={fields} onFieldClick={handleFieldClick} />
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/FilterPanel.tsx
git commit -m "feat(front): add FilterPanel composing editor and schema sidebar"
```

---

## Task 5: 接入 EndUserRecordWorkspace

**Files:**
- Modify: `src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx`

这是改动最集中的一步，分三个小步骤：状态、工具栏按钮、面板渲染 + useQuery。

- [ ] **Step 1: 添加筛选相关 import**

在文件顶部已有的 import 区域补充（在 `import { Filter, ...` 那一块附近）：

```typescript
// 新增 import（加在现有 lucide import 下方）
import { cn } from '@/shared/utils'
import { FilterPanel } from './FilterPanel'
import { getFilterCount } from './filter-utils'
```

- [ ] **Step 2: 添加筛选状态（在 searchKeyword state 附近，约第 177 行）**

在 `const [searchKeyword, setSearchKeyword] = useState('')` 下方添加：

```typescript
// --- Filter state ---
const [filterOpen, setFilterOpen] = useState(false)
// Draft: changes on every keystroke, does NOT trigger a query
const [whereJsonDraft, setWhereJsonDraft] = useState<string>('')
// Committed: only updated on "应用筛选", drives the actual GraphQL where clause
const [whereJsonCommitted, setWhereJsonCommitted] = useState<string | null>(null)

const whereInput = useMemo(() => {
  if (!whereJsonCommitted?.trim()) return undefined
  try {
    return JSON.parse(whereJsonCommitted) as Record<string, unknown>
  } catch {
    return undefined
  }
}, [whereJsonCommitted])

function handleApplyFilter() {
  const trimmed = whereJsonDraft.trim()
  setWhereJsonCommitted(trimmed || null)
}

function handleClearFilter() {
  // Atomically clear both draft and committed state so the filter is removed immediately.
  // We cannot call setWhereJsonDraft('') then handleApplyFilter() because React would
  // batch the updates and handleApplyFilter would still see the old draft value.
  setWhereJsonDraft('')
  setWhereJsonCommitted(null)
}

const filterCount = getFilterCount(whereJsonCommitted)
const hasActiveFilter = filterCount !== null
// --- End filter state ---
```

- [ ] **Step 3: 给 useQuery 传 where 变量**

找到约第 339 行的 `useQuery` call，`variables` 里加入 `where`：

```typescript
// 修改前：
variables: {
  take: 50,
  skip: 0,
},

// 修改后：
variables: {
  take: 50,
  skip: 0,
  where: whereInput,
},
```

- [ ] **Step 4: 替换工具栏里的"筛选"按钮**

找到约第 564 行的筛选 Button，替换为带三态样式和角标的版本：

```typescript
// 替换前：
<Button
  variant="ghost"
  size="sm"
  className="h-[26px] border-transparent px-2.5 text-xs font-normal text-muted-foreground hover:bg-muted hover:text-foreground"
>
  <Filter className="mr-1.5 size-3.5" />
  <span>筛选</span>
</Button>

// 替换后：
<Button
  variant="ghost"
  size="sm"
  onClick={() => setFilterOpen((open) => !open)}
  className={cn(
    'h-[26px] px-2.5 text-xs font-normal',
    filterOpen
      ? 'border border-primary text-primary ring-2 ring-primary/20'
      : hasActiveFilter
        ? 'border border-primary text-primary'
        : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'
  )}
>
  <Filter className="mr-1.5 size-3.5" />
  <span>筛选</span>
  {filterCount !== null && (
    <span className="ml-1.5 flex size-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
      {filterCount}
    </span>
  )}
</Button>
```

- [ ] **Step 5: 在工具栏后渲染 FilterPanel**

找到工具栏 `<div>` 的结束标签（`</div>` after the toolbar div），在其后、表格前添加：

```typescript
{/* 筛选面板（工具栏下方内联展开） */}
{filterOpen && (
  <FilterPanel
    fields={runtimeFields}
    whereJsonDraft={whereJsonDraft}
    onWhereJsonDraftChange={setWhereJsonDraft}
    onApply={handleApplyFilter}
    onClear={handleClearFilter}
  />
)}
```

- [ ] **Step 6: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx
git commit -m "feat(front): wire FilterPanel into EndUserRecordWorkspace"
```

---

## Task 6: 验收测试（手动 + lint）

**Files:**
- 无新文件，验证现有功能

- [ ] **Step 1: 运行 lint**

```bash
cd modelcraft-front && npm run lint
```

期望：0 errors

- [ ] **Step 2: 跑单元测试**

```bash
cd modelcraft-front && npx vitest run src/web/components/features/end-user-data/filter-utils.test.ts
```

期望：8 tests passed

- [ ] **Step 3: 手动验收 — 无筛选**

启动开发服务器，打开任意模型的 EndUser 数据视图：
- "筛选"按钮显示默认灰色样式，无角标
- 点击按钮：面板展开，右侧显示当前模型字段列表
- 再次点击：面板收起

- [ ] **Step 4: 手动验收 — 正常筛选**

在编辑器输入（替换 `name` 为模型中实际存在的字段名）：
```json
{"name": {"contains": "test"}}
```
- 底部显示 `✓ 有效 JSON`，"应用筛选"可点击
- 点击"应用筛选"：表格刷新，仅显示匹配的记录
- "筛选"按钮变蓝，角标显示 `•`（单字段条件）

用 AND 结构再试一次：
```json
{"AND": [{"name": {"contains": "test"}}]}
```
- 角标显示 `1`（AND 数组长度）

- [ ] **Step 5: 手动验收 — 非法 JSON**

在编辑器输入 `{bad json`：
- 底部显示 `✗ JSON 格式错误`
- "应用筛选"按钮禁用

- [ ] **Step 6: 手动验收 — 清空**

有激活筛选时点击"清空"：
- 编辑器清空
- 表格恢复全量数据
- "筛选"按钮角标消失

- [ ] **Step 7: 手动验收 — 字段点击插入**

在右侧字段列表点击任意字段名：
- 编辑器在光标位置插入 `"fieldName": {}`

- [ ] **Step 8: Final commit**

```bash
git add -p  # 确认无意外文件
git commit -m "feat(front): complete filter panel MVP with JSON editor and field schema sidebar"
```
