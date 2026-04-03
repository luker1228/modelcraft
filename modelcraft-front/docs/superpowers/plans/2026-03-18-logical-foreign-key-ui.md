# Logical Foreign Key UI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "逻辑外键" (Logical Foreign Key) section below the field definition table in the Edit Model Sheet, supporting listing, creating, and deleting logical foreign key pairs.

**Architecture:** Add GraphQL query + mutations for logical foreign keys to their respective files, add TypeScript-level error result types to `src/types/index.ts`, then extend the Edit Model Sheet in `page.tsx` with new state and UI. All changes are self-contained in three files.

**Tech Stack:** Next.js App Router, Apollo Client (useLazyQuery / useMutation), Tailwind CSS, shadcn/ui (Button, Select, Badge, toast)

---

## Chunk 1: GraphQL Operations & Types

### Task 1: Add GraphQL query for logical foreign keys

**Files:**
- Modify: `src/graphql/queries/model.ts`

- [ ] Append `GET_LOGICAL_FOREIGN_KEYS` to the bottom of `src/graphql/queries/model.ts`:

```ts
// 查询逻辑外键列表
export const GET_LOGICAL_FOREIGN_KEYS = gql`
  query GetLogicalForeignKeys($projectSlug: String!, $modelId: ID!) {
    logicalForeignKeys(projectSlug: $projectSlug, modelId: $modelId) {
      id
      pairId
      direction
      modelId
      refModelId
      sourceFields
      targetFields
    }
  }
`
```

---

### Task 2: Add GraphQL mutations for logical foreign keys

**Files:**
- Modify: `src/graphql/mutations/model.ts`

- [ ] Append `CREATE_LOGICAL_FOREIGN_KEY` and `DELETE_LOGICAL_FOREIGN_KEY` to the bottom of `src/graphql/mutations/model.ts`:

```ts
// 创建逻辑外键
export const CREATE_LOGICAL_FOREIGN_KEY = gql`
  mutation CreateLogicalForeignKey($projectSlug: String!, $input: CreateLogicalForeignKeyInput!) {
    createLogicalForeignKey(projectSlug: $projectSlug, input: $input) {
      result {
        __typename
        ... on LogicalForeignKey {
          id
          pairId
          direction
          modelId
          refModelId
          sourceFields
          targetFields
        }
        ... on FKColumnsNotFoundError {
          message
        }
        ... on FKFieldCountMismatchError {
          message
        }
      }
    }
  }
`

// 删除逻辑外键
export const DELETE_LOGICAL_FOREIGN_KEY = gql`
  mutation DeleteLogicalForeignKey($projectSlug: String!, $pairId: String!) {
    deleteLogicalForeignKey(projectSlug: $projectSlug, pairId: $pairId) {
      result {
        __typename
        ... on DeleteLogicalForeignKeySuccess {
          pairId
        }
        ... on FKNotFoundError {
          message
        }
        ... on FKPairHasRelateFieldsError {
          message
        }
      }
    }
  }
`
```

---

### Task 3: Extend TypeScript types for FK error results

**Files:**
- Modify: `src/types/index.ts`

- [ ] After the existing `CreateLogicalForeignKeyInput` interface (around line 448), add error/result types:

```ts
export interface FKColumnsNotFoundError {
  message: string
}

export interface FKFieldCountMismatchError {
  message: string
}

export interface FKNotFoundError {
  message: string
}

export interface FKPairHasRelateFieldsError {
  message: string
}

export type CreateLogicalForeignKeyResult =
  | (LogicalForeignKey & { __typename: 'LogicalForeignKey' })
  | (FKColumnsNotFoundError & { __typename: 'FKColumnsNotFoundError' })
  | (FKFieldCountMismatchError & { __typename: 'FKFieldCountMismatchError' })

export type DeleteLogicalForeignKeyResult =
  | ({ pairId: string } & { __typename: 'DeleteLogicalForeignKeySuccess' })
  | (FKNotFoundError & { __typename: 'FKNotFoundError' })
  | (FKPairHasRelateFieldsError & { __typename: 'FKPairHasRelateFieldsError' })
```

---

## Chunk 2: Edit Model Sheet - State & Hooks

### Task 4: Add imports and state for FK section

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] Add new lucide icons and GraphQL imports at the top of `page.tsx`.

In the lucide import block, add `Link2`, `Trash2` (already imported), `ChevronDown`, `GitMerge`, `Plus` (may already exist):
```ts
import {
  // ...existing...
  Link2,
  GitMerge,
} from 'lucide-react'
```

- [ ] Add new GraphQL imports (with the existing imports):
```ts
import { GET_LOGICAL_FOREIGN_KEYS } from '@/graphql/queries/model'
import {
  CREATE_LOGICAL_FOREIGN_KEY,
  DELETE_LOGICAL_FOREIGN_KEY,
} from '@/graphql/mutations/model'
```

- [ ] Add shadcn/ui `Select` component imports (after existing UI imports):
```ts
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
```

- [ ] Add `useToast` import if not already present:
```ts
import { useToast } from '@/components/ui/use-toast'
```

- [ ] After the existing FK-related state block (after `editFieldDescription` state, around line 153), add:

```ts
// 逻辑外键状态
const [fkList, setFkList] = useState<LogicalForeignKey[]>([])
const [fkLoading, setFkLoading] = useState(false)
const [fkFormOpen, setFkFormOpen] = useState(false)
const [fkRefModelId, setFkRefModelId] = useState('')
const [fkMappings, setFkMappings] = useState<{ sourceField: string; targetField: string }[]>([
  { sourceField: '', targetField: '' },
])
const [fkSubmitting, setFkSubmitting] = useState(false)
const [fkDeleteConfirm, setFkDeleteConfirm] = useState<string | null>(null) // pairId being confirmed
```

- [ ] Import `LogicalForeignKey` type at the top with other type imports (or use inline since it's in `src/types/index.ts`):
```ts
import type { LogicalForeignKey } from '@/types'
```

---

### Task 5: Add Apollo hooks and handler functions

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] After the existing `fetchModelDetail` lazy query (around line 164), add:

```ts
// 懒加载逻辑外键
const [fetchForeignKeys] = useLazyQuery(GET_LOGICAL_FOREIGN_KEYS, {
  fetchPolicy: 'network-only',
  context: orgScopedContext,
})

const [createFKMutation] = useMutation(CREATE_LOGICAL_FOREIGN_KEY, {
  context: orgScopedContext,
})

const [deleteFKMutation] = useMutation(DELETE_LOGICAL_FOREIGN_KEY, {
  context: orgScopedContext,
})
```

- [ ] Add `useToast` hook usage near other hooks:
```ts
const { toast } = useToast()
```

- [ ] Add helper to load FKs. Place after `handleCloseEditModel` function:

```ts
const loadForeignKeys = async (modelId: string) => {
  setFkLoading(true)
  try {
    const result = await fetchForeignKeys({
      variables: { projectSlug, modelId },
    })
    setFkList(result.data?.logicalForeignKeys ?? [])
  } catch {
    setFkList([])
  } finally {
    setFkLoading(false)
  }
}
```

- [ ] In the existing `handleOpenEditModel` (or wherever `fetchModelDetail` is called to open the sheet), also call `loadForeignKeys(modelId)` right after setting the model data. Find the section where `editModelId` is set and `fetchModelDetail` is called, and add:

```ts
// After setting editModelData:
setFkList([])
setFkFormOpen(false)
setFkMappings([{ sourceField: '', targetField: '' }])
setFkRefModelId('')
loadForeignKeys(modelId)
```

- [ ] Add the create FK handler:

```ts
const handleCreateFK = async () => {
  if (!editModelId || !fkRefModelId) return
  const validMappings = fkMappings.filter(m => m.sourceField && m.targetField)
  if (validMappings.length === 0) return

  setFkSubmitting(true)
  try {
    const result = await createFKMutation({
      variables: {
        projectSlug,
        input: {
          modelId: editModelId,
          refModelId: fkRefModelId,
          sourceFields: validMappings.map(m => m.sourceField),
          targetFields: validMappings.map(m => m.targetField),
        },
      },
    })
    const r = result.data?.createLogicalForeignKey?.result
    if (r?.__typename === 'LogicalForeignKey') {
      toast({ description: '外键创建成功' })
      setFkFormOpen(false)
      setFkRefModelId('')
      setFkMappings([{ sourceField: '', targetField: '' }])
      await loadForeignKeys(editModelId)
    } else if (r?.__typename === 'FKColumnsNotFoundError') {
      toast({ variant: 'destructive', description: `字段不存在：${r.message}` })
    } else if (r?.__typename === 'FKFieldCountMismatchError') {
      toast({ variant: 'destructive', description: '源字段与目标字段数量不匹配' })
    }
  } catch {
    toast({ variant: 'destructive', description: '创建外键失败，请重试' })
  } finally {
    setFkSubmitting(false)
  }
}
```

- [ ] Add the delete FK handler:

```ts
const handleDeleteFK = async (pairId: string) => {
  try {
    const result = await deleteFKMutation({
      variables: { projectSlug, pairId },
    })
    const r = result.data?.deleteLogicalForeignKey?.result
    if (r?.__typename === 'DeleteLogicalForeignKeySuccess') {
      toast({ description: '外键已删除' })
      if (editModelId) await loadForeignKeys(editModelId)
    } else if (r?.__typename === 'FKPairHasRelateFieldsError') {
      toast({ variant: 'destructive', description: '该外键关联了关系字段，请先删除相关字段' })
    } else if (r?.__typename === 'FKNotFoundError') {
      toast({ variant: 'destructive', description: '外键不存在' })
    }
  } catch {
    toast({ variant: 'destructive', description: '删除外键失败，请重试' })
  } finally {
    setFkDeleteConfirm(null)
  }
}
```

---

## Chunk 3: Edit Model Sheet - UI

### Task 6: Add the FK section UI to the Edit Model Sheet

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] In the Edit Model Sheet body, after the closing `</div>` of the "字段列表" section (around line 660, the closing `</div>` of `<div className="space-y-3">`), and before the final `</div>` (line 661, closing `<div className="space-y-4 py-4">`), insert the following JSX block:

```tsx
{/* 逻辑外键 */}
<div className="space-y-3">
  <div className="flex items-center justify-between">
    <h3 className="flex items-center gap-2 text-sm font-semibold text-foreground">
      <Link2 className="size-3.5" />
      逻辑外键
      <span className="text-xs font-normal text-muted-foreground">
        ({fkList.length} 个关系)
      </span>
    </h3>
    {!fkFormOpen && (
      <Button
        variant="outline"
        size="sm"
        className="h-7 text-xs"
        onClick={() => {
          setFkFormOpen(true)
          setFkRefModelId('')
          setFkMappings([{ sourceField: '', targetField: '' }])
        }}
      >
        <Plus className="mr-1 size-3" />
        添加关系
      </Button>
    )}
  </div>

  {/* FK 列表表格 */}
  {fkLoading ? (
    <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
      <Loader2 className="size-4 animate-spin" />
      加载中...
    </div>
  ) : fkList.length > 0 ? (
    <div className="overflow-hidden rounded-lg border border-border bg-card">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/30">
              <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                引用模型
              </th>
              <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                源字段
              </th>
              <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                目标字段
              </th>
              <th className="w-[70px] px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                方向
              </th>
              <th className="w-[50px] px-3 py-2 text-center text-xs font-medium text-muted-foreground" />
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {fkList.map((fk) => {
              const refModel = models.find(m => m.id === fk.refModelId)
              return (
                <tr key={fk.id} className="transition-colors hover:bg-muted/20">
                  <td className="px-3 py-2">
                    <span className="font-mono text-xs text-foreground">
                      {refModel?.name ?? fk.refModelId}
                    </span>
                    {refModel?.title && refModel.title !== refModel.name && (
                      <span className="ml-1 text-xs text-muted-foreground">
                        ({refModel.title})
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-2">
                    <span className="font-mono text-xs text-blue-700 dark:text-blue-400">
                      {fk.sourceFields.join(', ')}
                    </span>
                  </td>
                  <td className="px-3 py-2">
                    <span className="font-mono text-xs text-emerald-700 dark:text-emerald-400">
                      {fk.targetFields.join(', ')}
                    </span>
                  </td>
                  <td className="px-3 py-2">
                    <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${
                      fk.direction === 'NORMAL'
                        ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                        : 'bg-purple-50 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
                    }`}>
                      {fk.direction === 'NORMAL' ? '正向' : '反向'}
                    </span>
                  </td>
                  <td className="p-2 text-center">
                    {fkDeleteConfirm === fk.pairId ? (
                      <div className="flex items-center gap-1">
                        <button
                          className="rounded px-1.5 py-0.5 text-xs text-destructive hover:bg-destructive/10"
                          onClick={() => handleDeleteFK(fk.pairId)}
                        >
                          确认
                        </button>
                        <button
                          className="rounded px-1.5 py-0.5 text-xs text-muted-foreground hover:bg-muted"
                          onClick={() => setFkDeleteConfirm(null)}
                        >
                          取消
                        </button>
                      </div>
                    ) : (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="size-6 p-0 hover:bg-muted hover:text-destructive"
                        onClick={() => setFkDeleteConfirm(fk.pairId)}
                      >
                        <Trash2 className="size-3.5 text-muted-foreground" />
                      </Button>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  ) : !fkFormOpen ? (
    <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border py-6 text-muted-foreground">
      <Link2 className="mb-2 size-6 opacity-30" />
      <p className="text-sm">暂无逻辑外键</p>
      <p className="mt-1 text-xs">点击"添加关系"创建逻辑外键</p>
    </div>
  ) : null}

  {/* 内联创建表单 */}
  {fkFormOpen && (
    <div className="rounded-lg border border-border bg-muted/10 p-4 space-y-3">
      <h4 className="text-xs font-semibold text-foreground">新建逻辑外键</h4>

      {/* 引用模型选择 */}
      <div className="flex items-center gap-2">
        <label className="w-20 shrink-0 text-xs text-muted-foreground">引用模型</label>
        <Select value={fkRefModelId} onValueChange={(v) => {
          setFkRefModelId(v)
          setFkMappings([{ sourceField: '', targetField: '' }])
        }}>
          <SelectTrigger className="h-7 text-xs">
            <SelectValue placeholder="选择引用模型" />
          </SelectTrigger>
          <SelectContent>
            {models
              .filter(m => m.id !== editModelId)
              .map(m => (
                <SelectItem key={m.id} value={m.id} className="text-xs font-mono">
                  {m.name}{m.title && m.title !== m.name ? ` (${m.title})` : ''}
                </SelectItem>
              ))}
          </SelectContent>
        </Select>
      </div>

      {/* 字段映射 */}
      <div className="space-y-2">
        <label className="text-xs text-muted-foreground">字段映射（源字段 → 目标字段）</label>
        {fkMappings.map((mapping, idx) => {
          const refModelFields = fkRefModelId
            ? models.find(m => m.id === fkRefModelId) as (Model & { fields?: ModelField[] }) | undefined
            : undefined
          return (
            <div key={idx} className="flex items-center gap-2">
              {/* 源字段（当前模型） */}
              <Select
                value={mapping.sourceField}
                onValueChange={(v) => {
                  const next = [...fkMappings]
                  next[idx] = { ...next[idx], sourceField: v }
                  setFkMappings(next)
                }}
              >
                <SelectTrigger className="h-7 text-xs">
                  <SelectValue placeholder="源字段" />
                </SelectTrigger>
                <SelectContent>
                  {editModelData?.fields?.map(f => (
                    <SelectItem key={f.name} value={f.name} className="text-xs font-mono">
                      {f.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <span className="shrink-0 text-xs text-muted-foreground">→</span>

              {/* 目标字段（引用模型） */}
              <Select
                value={mapping.targetField}
                disabled={!fkRefModelId}
                onValueChange={(v) => {
                  const next = [...fkMappings]
                  next[idx] = { ...next[idx], targetField: v }
                  setFkMappings(next)
                }}
              >
                <SelectTrigger className="h-7 text-xs">
                  <SelectValue placeholder="目标字段" />
                </SelectTrigger>
                <SelectContent>
                  {(refModelFields as any)?.fields?.map((f: ModelField) => (
                    <SelectItem key={f.name} value={f.name} className="text-xs font-mono">
                      {f.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {/* 删除此行映射 */}
              {fkMappings.length > 1 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="size-6 shrink-0 p-0 hover:text-destructive"
                  onClick={() => setFkMappings(fkMappings.filter((_, i) => i !== idx))}
                >
                  <Trash2 className="size-3" />
                </Button>
              )}
            </div>
          )
        })}

        <Button
          variant="ghost"
          size="sm"
          className="h-6 text-xs text-muted-foreground hover:text-foreground"
          onClick={() => setFkMappings([...fkMappings, { sourceField: '', targetField: '' }])}
        >
          <Plus className="mr-1 size-3" />
          添加映射
        </Button>
      </div>

      {/* 操作按钮 */}
      <div className="flex justify-end gap-2 pt-1">
        <Button
          variant="outline"
          size="sm"
          className="h-7 text-xs"
          onClick={() => {
            setFkFormOpen(false)
            setFkRefModelId('')
            setFkMappings([{ sourceField: '', targetField: '' }])
          }}
        >
          取消
        </Button>
        <Button
          size="sm"
          className="h-7 text-xs"
          disabled={
            fkSubmitting ||
            !fkRefModelId ||
            fkMappings.every(m => !m.sourceField || !m.targetField)
          }
          onClick={handleCreateFK}
        >
          {fkSubmitting && <Loader2 className="mr-1 size-3 animate-spin" />}
          创建外键
        </Button>
      </div>
    </div>
  )}
</div>
```

---

### Task 7: Handle the case where refModel fields are not loaded

**Note:** The `models` list from `GET_MODELS` only includes basic model info (no fields). To populate target field options when creating a FK, we need to fetch the ref model's fields on demand.

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] Add state to cache ref model detail:
```ts
const [fkRefModelDetail, setFkRefModelDetail] = useState<ModelDetail | null>(null)
const [fkRefModelLoading, setFkRefModelLoading] = useState(false)
```

- [ ] Add a `useEffect` that fetches ref model detail when `fkRefModelId` changes (place near other effects):
```ts
useEffect(() => {
  if (!fkRefModelId || !projectSlug) {
    setFkRefModelDetail(null)
    return
  }
  let cancelled = false
  setFkRefModelLoading(true)
  fetchModelDetail({ variables: { projectSlug, id: fkRefModelId } })
    .then(result => {
      if (cancelled) return
      const m = result.data?.model?.model
      if (m) setFkRefModelDetail(m as ModelDetail)
    })
    .finally(() => {
      if (!cancelled) setFkRefModelLoading(false)
    })
  return () => { cancelled = true }
}, [fkRefModelId, projectSlug]) // eslint-disable-line react-hooks/exhaustive-deps
```

- [ ] In the inline form JSX, replace the `refModelFields` lookup in the target field `<SelectContent>` with `fkRefModelDetail?.fields`:

Change:
```tsx
{(refModelFields as any)?.fields?.map((f: ModelField) => (
```
To:
```tsx
{fkRefModelDetail?.fields?.map((f: ModelField) => (
```

And update the `disabled` prop on the target field Select:
```tsx
disabled={!fkRefModelId || fkRefModelLoading}
```

Also remove the now-unused `refModelFields` variable from inside the `.map()`.

---

## Chunk 4: Reset & Polish

### Task 8: Reset FK state when closing the Edit Model Sheet

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] In `handleCloseEditModel` (find the existing function), add reset for FK state:
```ts
setFkList([])
setFkFormOpen(false)
setFkRefModelId('')
setFkMappings([{ sourceField: '', targetField: '' }])
setFkDeleteConfirm(null)
setFkRefModelDetail(null)
```

### Task 9: Verify imports — ensure `Link2` and `Plus` are in the lucide import

**Files:**
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`

- [ ] Check that the lucide import block includes `Link2` and `Plus`. `Plus` may already be present. Add any missing ones:
```ts
import {
  Table2, Search, Plus, MoreVertical, ChevronsUpDown, X,
  Filter, Loader2, AlertTriangle, ExternalLink, Edit, Key,
  Settings, Trash2, Archive,
  Link2,  // ← add if missing
} from 'lucide-react'
```

### Task 10: Verify shadcn/ui Select component exists

**Files:**
- Check: `src/components/ui/select.tsx`

- [ ] Run:
```bash
ls src/components/ui/select.tsx
```
Expected: file exists. If it does not exist, install it:
```bash
npx shadcn@latest add select
```

### Task 11: Build check

- [ ] Run the TypeScript build to catch any type errors:
```bash
npm run build 2>&1 | head -60
```
Fix any reported errors before proceeding.

### Task 12: Commit

- [ ] Commit all changes:
```bash
git add \
  src/graphql/queries/model.ts \
  src/graphql/mutations/model.ts \
  src/types/index.ts \
  src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx
git commit -m "feat: add logical foreign key section to edit model sheet"
```
