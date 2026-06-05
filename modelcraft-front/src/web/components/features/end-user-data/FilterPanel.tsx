'use client'

import { useState, useCallback } from 'react'
import { Search, X } from 'lucide-react'
import type { FieldDefinition } from '@api-client/cms/public'
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
  /** Extra content (search input etc.) rendered to the right of the chips. */
  children?: React.ReactNode
}

/**
 * FilterBar — inline chip-bar style filter, replacing the old collapsible panel.
 *
 * Structured filter chips are always visible in the bar.
 * A "查询" button explicitly triggers the filter, a "清除筛选" button resets it.
 */
export function FilterBar({ fields, onApply, onClear, children }: FilterBarProps) {
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
    <div className="flex min-h-[44px] flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2">
      {/* Structured filter chip bar — always visible, no auto-apply */}
      <StructuredFilterTab
        fields={fields}
        rows={rows}
        onRowsChange={setRows}
        onApply={handleApply}
        onClear={handleClear}
        inline
      />

      {/* 查询 button — always visible, triggers filter */}
      <button
        type="button"
        onClick={handleApply}
        className="flex h-[26px] shrink-0 items-center gap-1 rounded-sm border border-primary/60 bg-primary/5 px-2 text-xs text-primary transition-colors hover:border-primary hover:bg-primary/10"
      >
        <Search size={11} strokeWidth={1.5} />
        <span>查询</span>
      </button>

      {/* Clear all filters button — only shown when filters are active */}
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

      {/* Slot for search input / record count etc. */}
      {children && (
        <div className="ml-auto flex items-center gap-2">
          {children}
        </div>
      )}
    </div>
  )
}

// Keep old name as alias so existing imports don't break during transition
export { FilterBar as FilterPanel }
export type { FilterBarProps as FilterPanelProps }
