'use client'

import { useCallback } from 'react'
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
  { label: '包含', value: 'contains' },
  { label: '等于', value: 'equals' },
  { label: '不等于', value: 'not' },
  { label: '开头是', value: 'startsWith' },
  { label: '结尾是', value: 'endsWith' },
]

const NUMBER_OPERATORS: OperatorOption[] = [
  { label: '等于', value: 'equals' },
  { label: '不等于', value: 'not' },
  { label: '大于', value: 'gt' },
  { label: '大于等于', value: 'gte' },
  { label: '小于', value: 'lt' },
  { label: '小于等于', value: 'lte' },
]

const BOOLEAN_OPERATORS: OperatorOption[] = [
  { label: '是', value: 'equals_true' },
  { label: '否', value: 'equals_false' },
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

/**
 * Derive fieldType string for filterRowsToWhereJson from a FieldDefinition.
 * We pass it on the FilterRow so the util can coerce values correctly.
 */
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
// Props
// ---------------------------------------------------------------------------

export interface StructuredFilterTabProps {
  fields: FieldDefinition[]
  rows: FilterRow[]
  onRowsChange: (rows: FilterRow[]) => void
  onApply: () => void
  onClear: () => void
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

let _idCounter = 0
function nextId() {
  return `row-${++_idCounter}`
}

export function StructuredFilterTab({
  fields,
  rows,
  onRowsChange,
  onApply,
  onClear,
}: StructuredFilterTabProps) {
  // Display fields: exclude internal _-prefixed fields
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
          // When field changes, reset operator to the first valid one for new field type
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

  const hasAnyValue = rows.some((r) => r.field && r.value.trim())

  return (
    <div className="flex flex-col">
      <div className="flex flex-col gap-2 px-3 py-3">
        {rows.length === 0 ? (
          <p className="text-xs text-muted-foreground">暂无条件，点击"添加条件"开始筛选。</p>
        ) : (
          rows.map((row) => {
            const fieldDef = displayFields.find((f) => f.name === row.field)
            const operators = getOperatorsForField(fieldDef)
            const isBool = operators === BOOLEAN_OPERATORS

            return (
              <div key={row.id} className="flex items-center gap-1.5">
                {/* Field selector */}
                <select
                  value={row.field}
                  onChange={(e) => updateRow(row.id, { field: e.target.value })}
                  className="h-8 flex-1 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  {displayFields.map((f) => (
                    <option key={f.name} value={f.name}>
                      {f.name}
                    </option>
                  ))}
                </select>

                {/* Operator selector */}
                <select
                  value={row.operator}
                  onChange={(e) => updateRow(row.id, { operator: e.target.value })}
                  className="h-8 w-24 rounded-md border border-input bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  {operators.map((op) => (
                    <option key={op.value} value={op.value}>
                      {op.label}
                    </option>
                  ))}
                </select>

                {/* Value input — hidden for boolean (operator encodes value) */}
                {isBool ? (
                  <div className="h-8 flex-[1.2] rounded-md border border-input bg-muted/40 px-2 text-xs text-muted-foreground flex items-center">
                    {row.operator === 'equals_true' ? 'true' : 'false'}
                  </div>
                ) : (
                  <input
                    type="text"
                    value={row.value}
                    onChange={(e) => updateRow(row.id, { value: e.target.value })}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') onApply()
                    }}
                    className={cn(
                      'h-8 flex-[1.2] rounded-md border bg-background px-2 text-xs focus:outline-none focus:ring-1 focus:ring-ring',
                      row.value.trim() ? 'border-primary ring-1 ring-primary/30' : 'border-input'
                    )}
                  />
                )}

                {/* Remove button */}
                <button
                  type="button"
                  onClick={() => removeRow(row.id)}
                  className="flex size-6 shrink-0 items-center justify-center text-muted-foreground hover:text-foreground"
                  aria-label="删除条件"
                >
                  ×
                </button>
              </div>
            )
          })
        )}

        <button
          type="button"
          onClick={addRow}
          className="mt-0.5 self-start text-xs text-primary hover:underline"
        >
          + 添加条件
        </button>
      </div>

      <div className="flex items-center justify-between border-t border-border bg-muted/40 px-3 py-2">
        <button
          type="button"
          onClick={onClear}
          className="text-xs text-muted-foreground hover:text-foreground"
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
