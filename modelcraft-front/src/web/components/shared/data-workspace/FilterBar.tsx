'use client'

import { useState, useCallback } from 'react'
import { Search, X } from 'lucide-react'
import type { FieldDefinition } from '@api-client/cms/public'
import { RecordQueryBar } from './RecordQueryBar'
import type { FilterRow } from './filter-utils'
import { filterRowsToWhereJson } from './filter-utils'
import { StructuredFilterTab } from './StructuredFilterTab'

export interface FilterBarProps {
  /** Field definitions from the current model's jsonSchema. */
  fields: FieldDefinition[]
  /**
   * Called when the user applies a structured filter.
   * Receives the generated where JSON string.
   */
  onApply: (whereJson: string) => void
  /** Called to clear the active filter. */
  onClear: () => void
  searchValue: string
  onSearchChange: (value: string) => void
  searchPlaceholder: string
  summaryText?: string
}

/**
 * FilterBar — inline chip-bar style filter, replacing the old collapsible panel.
 *
 * Structured filter chips are always visible in the bar.
 * A "查询" button explicitly triggers the filter, a "清除筛选" button resets it.
 */
export function FilterBar({
  fields,
  onApply,
  onClear,
  searchValue,
  onSearchChange,
  searchPlaceholder,
  summaryText,
}: FilterBarProps) {
  const [rows, setRows] = useState<FilterRow[]>([])

  const handleApply = useCallback(() => {
    const normalizedRows = rows.map((row) => {
      if (row.operator === 'equals_true') return { ...row, operator: 'equals', value: 'true' }
      if (row.operator === 'equals_false') return { ...row, operator: 'equals', value: 'false' }
      return row
    })
    const where = filterRowsToWhereJson(normalizedRows)
    if (!where) {
      onClear()
      return
    }
    onApply(JSON.stringify(where))
  }, [rows, onApply, onClear])

  const handleClear = useCallback(() => {
    setRows([])
    onClear()
  }, [onClear])

  const hasRows = rows.length > 0

  return (
    <RecordQueryBar
      leftContent={
        <>
          <StructuredFilterTab
            fields={fields}
            rows={rows}
            onRowsChange={setRows}
            onApply={handleApply}
            onClear={handleClear}
            inline
          />
          <button
            type="button"
            onClick={handleApply}
            className="flex h-[26px] shrink-0 items-center gap-1 rounded-sm border border-primary/60 bg-primary/5 px-2 text-xs text-primary transition-colors hover:border-primary hover:bg-primary/10"
          >
            <Search size={11} strokeWidth={1.5} />
            <span>查询</span>
          </button>
          {hasRows && (
            <button
              type="button"
              onClick={handleClear}
              className="flex h-[26px] shrink-0 items-center gap-1 rounded-sm border border-border/60 px-2 text-xs text-muted-foreground transition-colors hover:border-destructive/50 hover:text-destructive"
              title="清除所有筛选"
            >
              <X size={11} strokeWidth={1.5} />
              <span>清除筛选</span>
            </button>
          )}
        </>
      }
      searchValue={searchValue}
      onSearchChange={onSearchChange}
      searchPlaceholder={searchPlaceholder}
      summaryText={summaryText}
    />
  )
}

// Keep old name as alias so existing imports don't break during transition
export { FilterBar as FilterPanel }
export type { FilterBarProps as FilterPanelProps }
