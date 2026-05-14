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
