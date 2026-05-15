'use client'

import { useCallback, useRef } from 'react'
import { X, Plus } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { cn } from '@/shared/utils'
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
// FilterChip — single inline condition pill
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
  const currentOp = operators.find((op) => op.value === row.operator) ?? operators[0]

  // Auto-size the value input: measure with a hidden span
  const valueSpanRef = useRef<HTMLSpanElement>(null)

  return (
    <div
      className={cn(
        'group flex h-[26px] shrink-0 items-stretch rounded-sm border bg-muted',
        row.value.trim() || isBool ? 'border-border' : 'border-dashed border-border/60'
      )}
    >
      {/* Field selector — styled as plain text */}
      <div className="relative inline-flex shrink-0 items-center">
        <span className="pointer-events-none flex h-full select-none items-center pl-2 pr-1 text-xs text-muted-foreground">
          {row.field || '字段'}
        </span>
        <select
          value={row.field}
          onChange={(e) => onUpdate(row.id, { field: e.target.value })}
          className="absolute inset-0 size-full cursor-pointer appearance-none bg-transparent opacity-0"
          title="选择字段"
        >
          {displayFields.map((f) => (
            <option key={f.name} value={f.name}>
              {f.name}
            </option>
          ))}
        </select>
      </div>

      {/* Operator selector — shows label */}
      <div className="relative inline-block shrink-0">
        {/* Invisible sizer so the cell is exactly as wide as the label */}
        <span className="invisible block shrink-0 px-1 text-xs">
          {currentOp?.label ?? 'equals'}
        </span>
        <select
          value={row.operator}
          onChange={(e) => onUpdate(row.id, { operator: e.target.value })}
          className="absolute inset-0 size-full cursor-pointer appearance-none bg-transparent px-1 text-center text-xs text-foreground opacity-0"
          title="选择操作符"
        >
          {operators.map((op) => (
            <option key={op.value} value={op.value}>
              {op.label}
            </option>
          ))}
        </select>
        {/* Visible label */}
        <span className="pointer-events-none absolute inset-0 flex items-center justify-center font-mono text-xs text-foreground">
          {currentOp?.label ?? 'equals'}
        </span>
      </div>

      {/* Value input — inline, auto-width */}
      {isBool ? null : (
        <div className="relative inline-block max-w-[180px]">
          {/* Hidden span to measure value width */}
          <span
            ref={valueSpanRef}
            className="invisible block whitespace-pre px-1 text-xs"
          >
            {row.value || ' '}
          </span>
          <input
            type="text"
            value={row.value}
            onChange={(e) => onUpdate(row.id, { value: e.target.value })}
            onKeyDown={(e) => {
              if (e.key === 'Enter') onApply()
            }}
            placeholder="值"
            className="absolute inset-0 size-full bg-transparent px-1 text-xs text-foreground placeholder:text-muted-foreground/50 focus:outline-none"
          />
        </div>
      )}

      {/* Remove button — visible on hover */}
      <button
        type="button"
        onClick={() => onRemove(row.id)}
        aria-label={`删除 ${row.field} 筛选条件`}
        className="flex h-full items-center justify-center px-1 text-muted-foreground/40 transition-colors hover:text-foreground group-hover:text-muted-foreground"
      >
        <X size={12} strokeWidth={1.5} />
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
}

let _idCounter = 0
function nextId() {
  return `row-${++_idCounter}`
}

// ---------------------------------------------------------------------------
// StructuredFilterTab — chip bar layout
// ---------------------------------------------------------------------------

export function StructuredFilterTab({
  fields,
  rows,
  onRowsChange,
  onApply,
  onClear,
}: StructuredFilterTabProps) {
  const displayFields = fields.filter((f) => !f.name.startsWith('_'))

  const addRow = useCallback(() => {
    const firstField = displayFields[0]
    const operators = getOperatorsForField(firstField)
    onRowsChange([
      ...rows,
      {
        id: nextId(),
        field: firstField?.name ?? '',
        operator: getDefaultOperator(operators),
        value: '',
        fieldType: getFieldType(firstField),
      },
    ])
  }, [rows, onRowsChange, displayFields])

  const removeRow = useCallback(
    (id: string) => {
      onRowsChange(rows.filter((r) => r.id !== id))
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

  return (
    <div className="flex flex-col">
      {/* Chip bar */}
      <div className="flex flex-wrap items-center gap-1.5 px-3 py-2.5">
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

        {/* Add filter chip */}
        <button
          type="button"
          onClick={addRow}
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
      </div>

      {/* Footer */}
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
