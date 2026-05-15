'use client'

import { useState, useCallback } from 'react'
import { Sparkles } from 'lucide-react'
import { cn } from '@/shared/utils'
import type { FieldDefinition } from '@api-client/cms/public'
import type { FilterRow } from './filter-utils'
import { filterRowsToWhereJson } from './filter-utils'
import { StructuredFilterTab } from './StructuredFilterTab'
import { useCopilotKitAvailable } from './FilterCopilotActions'
import { AiQueryTab } from './AiQueryTab'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'

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
 * AI query lives in a Popover triggered by the ✨ button (only when CopilotKit available).
 */
export function FilterBar({ fields, onApply, onClear, children }: FilterBarProps) {
  const hasCopilot = useCopilotKitAvailable()
  const [rows, setRows] = useState<FilterRow[]>([])
  const [aiOpen, setAiOpen] = useState(false)

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

  return (
    <div className="flex min-h-[44px] flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2">
      {/* Structured filter chip bar — always visible */}
      <StructuredFilterTab
        fields={fields}
        rows={rows}
        onRowsChange={setRows}
        onApply={handleApply}
        onClear={handleClear}
        inline
      />

      {/* AI query popover button */}
      {hasCopilot && (
        <Popover open={aiOpen} onOpenChange={setAiOpen}>
          <PopoverTrigger asChild>
            <button
              type="button"
              className={cn(
                'flex h-[26px] shrink-0 items-center gap-1 rounded-sm border px-2 text-xs transition-colors',
                aiOpen
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border text-muted-foreground hover:border-primary/50 hover:text-foreground'
              )}
              title="AI 自然语言查询"
            >
              <Sparkles size={11} strokeWidth={1.5} />
              <span>AI 查询</span>
            </button>
          </PopoverTrigger>
          <PopoverContent
            side="bottom"
            align="start"
            className="w-[420px] p-0"
            onOpenAutoFocus={(e) => e.preventDefault()}
          >
            <AiQueryTab fields={fields} />
          </PopoverContent>
        </Popover>
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
