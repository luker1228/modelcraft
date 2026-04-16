'use client'

import type { WidgetProps } from '@rjsf/utils'
import { Badge } from '@web/components/ui/badge'

interface RelationRecord {
  id?: string | null
  _displayName?: string | null
}

function toDisplayText(record: RelationRecord): string {
  const id = typeof record.id === 'string' ? record.id : ''
  const label = typeof record._displayName === 'string' ? record._displayName : ''
  if (label === '' && id !== '') return `空(${id})`
  if (label !== '' && id !== '') return `${label}(${id})`
  if (label !== '') return label
  return id
}

/**
 * Readonly widget for one-to-many RELATION fields.
 * Intended for schema-driven read-only rendering.
 */
export function RelationMultiReadonly(props: WidgetProps) {
  const rawValue = props.value as unknown
  const values: RelationRecord[] = Array.isArray(rawValue)
    ? rawValue.filter((item): item is RelationRecord => typeof item === 'object' && item !== null)
    : []

  if (values.length === 0) {
    return <span className="text-xs text-muted-foreground">暂无关联记录</span>
  }

  return (
    <div className="flex flex-wrap gap-1.5">
      {values.map((record, idx) => (
        <Badge key={`${record.id ?? 'unknown'}-${idx}`} variant="secondary" className="font-mono text-[11px]">
          {toDisplayText(record)}
        </Badge>
      ))}
    </div>
  )
}

