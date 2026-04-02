# Dynamic Form Implementation Plan

**Branch**: `feature/dynamic-form`  
**Worktree**: `/home/luke/modelcraft_project/.worktrees/dynamic-form`  
**Spec**: `modelcraft-front/docs/superpowers/specs/2026-04-01-dynamic-form-design.md`

---

## Overview

4 tasks, executed sequentially via subagent-driven-development.

---

## Task 1: Install RJSF dependencies

**Files**: `modelcraft-front/package.json`

```bash
cd /home/luke/modelcraft_project/.worktrees/dynamic-form/modelcraft-front
npm install @rjsf/core @rjsf/utils @rjsf/validator-ajv8 @rjsf/shadcn-ui
```

Commit after install.

---

## Task 2: Column real-time sync (refetchQueries)

**Files to modify**:
- `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`
- `src/web/components/model-editor/InsertFieldSheet.tsx`

**Exact changes**:

`page.tsx` — `REMOVE_FIELD` useMutation (~line 195):
```typescript
const [removeFieldMutation] = useMutation(REMOVE_FIELD, {
  context: projectScopedContext,
  refetchQueries: ['GetModel', 'GetModelJsonSchema'],  // ADD
})
```

`page.tsx` — `UPDATE_FIELD` useMutation (if present): add same `refetchQueries`.

`InsertFieldSheet.tsx` — `ADD_FIELDS` useMutation: change existing `refetchQueries: ['GetModel']` to:
```typescript
refetchQueries: ['GetModel', 'GetModelJsonSchema'],
```

---

## Task 3: ModelRecordForm component

**New files** (all under `modelcraft-front/src/web/components/model-editor/ModelRecordForm/`):
- `index.tsx` — RJSF Form wrapper
- `buildUiSchema.ts` — fields[] → uiSchema mapping
- `widgets/EnumSelect.tsx`
- `widgets/RelationPicker.tsx`
- `widgets/index.ts`

### `index.tsx` interface:
```typescript
interface ModelRecordFormProps {
  fields: Field[]                  // from src/types/index.ts Field type
  jsonSchema: RJSFSchema           // from @rjsf/utils
  initialData?: Record<string, unknown>
  onSubmit: (data: Record<string, unknown>) => Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string
}
```

Implementation:
- Query `GET_LOGICAL_FOREIGN_KEYS` with `{ modelId }` using `useProjectScopedClient()` from `src/bff/apollo/clients.ts`
- While FK loading: render skeleton (use `src/web/components/ui/skeleton.tsx` if exists, else simple div)
- Pass `formContext: { orgName, projectSlug, clusterName, databaseName, modelId, logicalForeignKeys }` to RJSF `<Form>`
- Use `@rjsf/shadcn-ui` validator
- Hide primary key fields via `uiSchema[fieldName]['ui:widget'] = 'hidden'`
- Submit/cancel buttons rendered by this component; `isSubmitting` controls loading state
- Server errors shown via toast (from `src/components/ui/` or existing toast)

### `buildUiSchema.ts` mapping:
```typescript
function buildUiSchema(
  fields: Field[],
  context: { orgName: string; projectSlug: string; clusterName: string; databaseName: string }
): UiSchema
```

| format | storageHint | output |
|--------|-------------|--------|
| ENUM | - | `ui:widget: 'EnumSelect'`, `ui:options: { enumValues: field.enum.options, multiple: false }` |
| ENUM_ARRAY | - | `ui:widget: 'EnumSelect'`, `ui:options: { enumValues: field.enum.options, multiple: true }` |
| RELATION | - | `ui:widget: 'RelationPicker'`, `ui:options: { relateFkId: field.relateFkId, ...context }` |
| DATE | - | `ui:widget: 'date'` |
| DATETIME | - | `ui:widget: 'datetime-local'` |
| TIME | - | `ui:widget: 'time'` |
| any | TEXT | `ui:widget: 'textarea'` (lower priority than format) |
| isPrimary=true | - | `ui:widget: 'hidden'` |

### `widgets/RelationPicker.tsx`:
- RJSF widget props: `value: string`, `onChange: (v: string) => void`, `formContext`, `uiSchema`
- From `formContext.logicalForeignKeys` find FK where `id === uiOptions.relateFkId` → get `refModelName`
- Create Apollo client: `createModelRuntimeClient(orgName, projectSlug, databaseName, refModelName)` from `src/bff/apollo/clients.ts`
- Query: `buildFindManyQuery(refModelName, ['id', 'name'])` from `src/bff/cms/runtime-query-builder.ts`
- UI: shadcn `Select` with client-side search filter input; first 50 records
- States: loading spinner, "暂无数据", "加载失败" + retry button
- Selected value stored as target record id

### `widgets/EnumSelect.tsx`:
- Single (`multiple: false`): shadcn `Select` from `src/web/components/ui/select.tsx`
- Multi (`multiple: true`): Checkbox array in `Popover` from `src/web/components/ui/`

---

## Task 4: Wire ModelRecordForm into DynamicModelTable

**File**: `src/web/components/model-editor/DynamicModelTable.tsx`

- Import `ModelRecordForm` from `./ModelRecordForm`
- Replace create Sheet inner form JSX with `<ModelRecordForm>` (keep Sheet container + open/close state)
- Replace edit Sheet inner form JSX with `<ModelRecordForm initialData={editFormData}>`
- Props to pass: `fields` (already in scope), `jsonSchema` (from existing `GetModelJsonSchema` query result), `orgName`, `projectSlug`, `clusterName`, `databaseName`, `modelId`
- `onSubmit` callback calls existing `createContent` / `updateContent` mutations (no mutation change needed here)
- `onCancel` closes the Sheet
- Service errors shown via existing toast pattern in the file

---

## Acceptance Criteria

1. Delete field → column disappears immediately (no refresh)
2. Add field → column appears immediately
3. ENUM → Select (single); ENUM_ARRAY → Checkbox Popover (multi)
4. DATE/DATETIME/TIME → date/datetime-local/time pickers
5. RELATION → searchable record selector
6. storageHint TEXT → textarea
7. Required fields: validation error displayed below field on submit
8. Number min/max exceeded: validation error displayed below field
9. RelationPicker load failure → shows error, other fields still work
10. Server errors → toast notification
