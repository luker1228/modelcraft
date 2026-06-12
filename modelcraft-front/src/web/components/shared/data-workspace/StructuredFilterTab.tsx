'use client'

import { useCallback, useState } from 'react'
import { X, Plus } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { cn } from '@/shared/utils'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import type { FilterRow } from './filter-utils'
import type { FieldDefinition } from '@api-client/cms/public'

// ---------------------------------------------------------------------------
// Operator definitions per field storage type
// ---------------------------------------------------------------------------

interface OperatorOption {
  label: string
  value: string
}

const STRING_OPERATORS: OperatorOption[] = [
  { label: 'contains', value: 'contains' },
  { label: 'equals', value: 'equals' },
  { label: 'not', value: 'not' },
  { label: 'startsWith', value: 'startsWith' },
  { label: 'endsWith', value: 'endsWith' },
]

const NUMBER_OPERATORS: OperatorOption[] = [
  { label: 'equals', value: 'equals' },
  { label: 'not', value: 'not' },
  { label: 'gt', value: 'gt' },
  { label: 'gte', value: 'gte' },
  { label: 'lt', value: 'lt' },
  { label: 'lte', value: 'lte' },
]

const BOOLEAN_OPERATORS: OperatorOption[] = [
  { label: 'is true', value: 'equals_true' },
  { label: 'is false', value: 'equals_false' },
]

function getOperatorsForField(field: FieldDefinition | undefined): OperatorOption[] {
  if (!field) return STRING_OPERATORS
  const hint = field.storageHint?.toUpperCase() ?? ''
  const schema = field.schemaType?.toUpperCase() ?? ''
  if (['BOOL', 'BOOLEAN'].includes(hint) || schema === 'BOOLEAN') return BOOLEAN_OPERATORS
  if (
    ['INT', 'BIGINT', 'NUMBER', 'FLOAT', 'DECIMAL', 'INTEGER'].includes(hint) ||
    ['NUMBER', 'INTEGER'].includes(schema)
  )
    return NUMBER_OPERATORS
  return STRING_OPERATORS
}

function getDefaultOperator(operators: OperatorOption[]): string {
  return operators[0]?.value ?? 'equals'
}

function getFieldType(field: FieldDefinition | undefined): string {
  if (!field) return 'STRING'
  const hint = field.storageHint?.toUpperCase() ?? ''
  const schema = field.schemaType?.toUpperCase() ?? ''
  if (['BOOL', 'BOOLEAN'].includes(hint) || schema === 'BOOLEAN') return 'BOOLEAN'
  if (['INT', 'BIGINT', 'INTEGER'].includes(hint) || schema === 'INTEGER') return 'INT'
  if (['NUMBER', 'FLOAT', 'DECIMAL'].includes(hint) || schema === 'NUMBER') return 'NUMBER'
  return 'STRING'
}

// ---------------------------------------------------------------------------
// AddFilterPopover — pick field / operator / value, then confirm
// ---------------------------------------------------------------------------

interface AddFilterPopoverProps {
  displayFields: FieldDefinition[]
  onAdd: (row: Omit<FilterRow, 'id'>) => void
}

function AddFilterPopover({ displayFields, onAdd }: AddFilterPopoverProps) {
  const [open, setOpen] = useState(false)
  const firstField = displayFields[0]

  const [draftField, setDraftField] = useState(firstField?.name ?? '')
  const [draftOperator, setDraftOperator] = useState(() =>
    getDefaultOperator(getOperatorsForField(firstField))
  )
  const [draftValue, setDraftValue] = useState('')

  const fieldDef = displayFields.find((f) => f.name === draftField)
  const operators = getOperatorsForField(fieldDef)
  const isBool = operators === BOOLEAN_OPERATORS

  function handleFieldChange(name: string) {
    const def = displayFields.find((f) => f.name === name)
    const ops = getOperatorsForField(def)
    setDraftField(name)
    setDraftOperator(getDefaultOperator(ops))
    setDraftValue('')
  }

  function handleAdd() {
    onAdd({
      field: draftField,
      operator: draftOperator,
      value: isBool ? '' : draftValue,
      fieldType: getFieldType(fieldDef),
    })
    // Reset draft for next use
    setDraftValue('')
    setDraftOperator(getDefaultOperator(getOperatorsForField(fieldDef)))
    setOpen(false)
  }

  const canAdd = draftField && (isBool || draftValue.trim() !== '')

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          disabled={displayFields.length === 0}
          className={cn(
            'flex h-[26px] shrink-0 items-center gap-1 rounded-sm border border-dashed border-border/60 px-2 text-xs text-muted-foreground transition-colors',
            'hover:border-border hover:bg-muted hover:text-foreground',
            'disabled:cursor-not-allowed disabled:opacity-40'
          )}
        >
          <Plus size={11} strokeWidth={1.5} />
          添加筛选
        </button>
      </PopoverTrigger>

      <PopoverContent
        side="bottom"
        align="start"
        className="w-72 p-3"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        <p className="mb-3 text-xs font-medium text-foreground">添加筛选条件</p>

        <div className="flex flex-col gap-2.5">
          {/* Step 1: Field */}
          <div className="flex flex-col gap-1">
            <label className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
              字段
            </label>
            <div className="relative">
              <select
                value={draftField}
                onChange={(e) => handleFieldChange(e.target.value)}
                className="h-8 w-full cursor-pointer appearance-none rounded-md border border-input bg-background py-0 pl-2.5 pr-7 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              >
                {displayFields.map((f) => (
                  <option key={f.name} value={f.name}>{f.name}</option>
                ))}
              </select>
              <svg className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground/60" width="10" height="10" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 10.5L3 5.5h10z" />
              </svg>
            </div>
          </div>

          {/* Step 2: Operator */}
          <div className="flex flex-col gap-1">
            <label className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
              操作符
            </label>
            <div className="relative">
              <select
                value={draftOperator}
                onChange={(e) => setDraftOperator(e.target.value)}
                className="h-8 w-full cursor-pointer appearance-none rounded-md border border-input bg-background py-0 pl-2.5 pr-7 font-mono text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              >
                {operators.map((op) => (
                  <option key={op.value} value={op.value}>{op.label}</option>
                ))}
              </select>
              <svg className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground/60" width="10" height="10" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 10.5L3 5.5h10z" />
              </svg>
            </div>
          </div>

          {/* Step 3: Value (hidden for boolean) */}
          {!isBool && (
            <div className="flex flex-col gap-1">
              <label className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
                值
              </label>
              <input
                type="text"
                value={draftValue}
                onChange={(e) => setDraftValue(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter' && canAdd) handleAdd() }}
                placeholder="输入筛选值…"
                autoFocus
                className="h-8 w-full rounded-md border border-input bg-background px-2.5 text-xs text-foreground placeholder:text-muted-foreground/60 focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="mt-3.5 flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            取消
          </button>
          <Button
            size="sm"
            className="h-7 px-4 text-xs"
            onClick={handleAdd}
            disabled={!canAdd}
          >
            添加
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

// ---------------------------------------------------------------------------
// FilterChip — single committed condition pill
// ---------------------------------------------------------------------------

interface FilterChipProps {
  row: FilterRow
  displayFields: FieldDefinition[]
  onUpdate: (id: string, patch: Partial<FilterRow>) => void
  onRemove: (id: string) => void
  onApply: () => void
}

function FilterChip({ row, displayFields, onUpdate, onRemove, onApply }: FilterChipProps) {
  const fieldDef = displayFields.find((f) => f.name === row.field)
  const operators = getOperatorsForField(fieldDef)
  const isBool = operators === BOOLEAN_OPERATORS

  return (
    <div className="group flex h-[26px] shrink-0 items-stretch overflow-hidden rounded-sm border border-border bg-muted text-xs">
      {/* Part 1: Field */}
      <div className="relative flex shrink-0 items-center">
        <select
          value={row.field}
          onChange={(e) => onUpdate(row.id, { field: e.target.value })}
          className="h-full cursor-pointer appearance-none bg-transparent py-0 pl-2 pr-5 text-xs text-muted-foreground hover:bg-muted-foreground/10 hover:text-foreground focus:outline-none"
        >
          {displayFields.map((f) => (
            <option key={f.name} value={f.name}>{f.name}</option>
          ))}
        </select>
        <svg className="pointer-events-none absolute right-1 text-muted-foreground/50" width="8" height="8" viewBox="0 0 16 16" fill="currentColor">
          <path d="M8 10.5L3 5.5h10z" />
        </svg>
      </div>

      <div className="w-px self-stretch bg-border/60" />

      {/* Part 2: Operator */}
      <div className="relative flex shrink-0 items-center">
        <select
          value={row.operator}
          onChange={(e) => onUpdate(row.id, { operator: e.target.value })}
          className="h-full cursor-pointer appearance-none bg-transparent py-0 pl-2 pr-5 font-mono text-xs text-foreground hover:bg-muted-foreground/10 focus:outline-none"
        >
          {operators.map((op) => (
            <option key={op.value} value={op.value}>{op.label}</option>
          ))}
        </select>
        <svg className="pointer-events-none absolute right-1 text-muted-foreground/50" width="8" height="8" viewBox="0 0 16 16" fill="currentColor">
          <path d="M8 10.5L3 5.5h10z" />
        </svg>
      </div>

      {/* Part 3: Value */}
      {!isBool && (
        <>
          <div className="w-px self-stretch bg-border/60" />
          <input
            type="text"
            value={row.value}
            onChange={(e) => onUpdate(row.id, { value: e.target.value })}
            onKeyDown={(e) => { if (e.key === 'Enter') onApply() }}
            placeholder="值"
            className="w-24 bg-transparent px-2 py-0 text-xs text-foreground placeholder:text-muted-foreground/50 focus:outline-none"
          />
        </>
      )}

      {/* Remove */}
      <div className="w-px self-stretch bg-border/60 opacity-0 transition-opacity group-hover:opacity-100" />
      <button
        type="button"
        onClick={() => onRemove(row.id)}
        aria-label={`删除 ${row.field} 筛选条件`}
        className="flex w-0 shrink-0 items-center justify-center overflow-hidden text-muted-foreground transition-all hover:text-foreground group-hover:w-5"
      >
        <X size={11} strokeWidth={1.5} />
      </button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Props + ID counter
// ---------------------------------------------------------------------------

export interface StructuredFilterTabProps {
  fields: FieldDefinition[]
  rows: FilterRow[]
  onRowsChange: (rows: FilterRow[]) => void
  onApply: () => void
  onClear: () => void
  inline?: boolean
}

let _idCounter = 0
function nextId() {
  return `row-${++_idCounter}`
}

// ---------------------------------------------------------------------------
// StructuredFilterTab
// ---------------------------------------------------------------------------

export function StructuredFilterTab({
  fields,
  rows,
  onRowsChange,
  onApply,
  onClear,
  inline = false,
}: StructuredFilterTabProps) {
  const displayFields = fields.filter((f) => !f.name.startsWith('_'))

  const addRow = useCallback(
    (draft: Omit<FilterRow, 'id'>) => {
      const newRow: FilterRow = { ...draft, id: nextId() }
      const updated = [...rows, newRow]
      onRowsChange(updated)
      // Do not auto-apply in inline mode — user clicks 查询 to trigger
    },
    [rows, onRowsChange]
  )

  const removeRow = useCallback(
    (id: string) => {
      onRowsChange(rows.filter((r) => r.id !== id))
      // Do not auto-apply in inline mode
    },
    [rows, onRowsChange]
  )

  const updateRow = useCallback(
    (id: string, patch: Partial<FilterRow>) => {
      onRowsChange(
        rows.map((r) => {
          if (r.id !== id) return r
          const updated = { ...r, ...patch }
          if (patch.field !== undefined) {
            const newField = displayFields.find((f) => f.name === patch.field)
            const ops = getOperatorsForField(newField)
            updated.operator = getDefaultOperator(ops)
            updated.fieldType = getFieldType(newField)
            updated.value = ''
          }
          return updated
        })
      )
    },
    [rows, onRowsChange, displayFields]
  )

  const hasAnyValue = rows.some((r) => {
    if (!r.field) return false
    const fieldDef = displayFields.find((f) => f.name === r.field)
    const isBool = getOperatorsForField(fieldDef) === BOOLEAN_OPERATORS
    return isBool || r.value.trim() !== ''
  })

  const chipBar = (
    <div className={cn('flex flex-wrap items-center gap-1.5', !inline && 'px-3 py-2.5')}>
      {rows.map((row) => (
        <FilterChip
          key={row.id}
          row={row}
          displayFields={displayFields}
          onUpdate={updateRow}
          onRemove={removeRow}
          onApply={onApply}
        />
      ))}
      <AddFilterPopover displayFields={displayFields} onAdd={addRow} />
    </div>
  )

  if (inline) return chipBar

  return (
    <div className="flex flex-col">
      {chipBar}
      <div className="flex items-center justify-between border-t border-border bg-muted/30 px-3 py-2">
        <button
          type="button"
          onClick={onClear}
          disabled={rows.length === 0}
          className="text-xs text-muted-foreground transition-colors hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
        >
          清空
        </button>
        <Button
          size="sm"
          className="h-7 px-4 text-xs"
          onClick={onApply}
          disabled={!hasAnyValue}
        >
          应用
        </Button>
      </div>
    </div>
  )
}
