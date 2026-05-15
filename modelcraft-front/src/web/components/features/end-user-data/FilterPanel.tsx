'use client'

import { useState, useCallback } from 'react'
import { cn } from '@/shared/utils'
import type { FieldDefinition } from '@api-client/cms/public'
import type { FilterRow } from './filter-utils'
import { filterRowsToWhereJson } from './filter-utils'
import { StructuredFilterTab } from './StructuredFilterTab'
import { useCopilotKitAvailable } from './FilterCopilotActions'
import { AiQueryTab } from './AiQueryTab'

type TabId = 'structured' | 'ai'

export interface FilterPanelProps {
  /** Field definitions from the current model's jsonSchema. */
  fields: FieldDefinition[]
  /**
   * Called when the user applies a structured filter.
   * Receives the generated where JSON string (ready to JSON.parse and pass to useQuery).
   */
  onApply: (whereJson: string) => void
  /**
   * Called to clear the active filter and reset the table to full data.
   * Invoked both by the "清空" button and on Tab switch.
   */
  onClear: () => void
}

/**
 * Filter panel with two independent tabs:
 * - "筛选": Structured row-based filter builder (field + operator + value)
 * - "AI 查询": Natural language query via modelcraft-agent (only rendered when CopilotKit available)
 *
 * Tab switching always calls onClear to reset the current query first.
 */
export function FilterPanel({ fields, onApply, onClear }: FilterPanelProps) {
  const hasCopilot = useCopilotKitAvailable()
  const [activeTab, setActiveTab] = useState<TabId>('structured')
  const [rows, setRows] = useState<FilterRow[]>([])

  const handleTabSwitch = useCallback(
    (tab: TabId) => {
      if (tab === activeTab) return
      // Tab switch = full reset per design decision
      setRows([])
      onClear()
      setActiveTab(tab)
    },
    [activeTab, onClear]
  )

  const handleApply = useCallback(() => {
    // Normalize boolean pseudo-operators before passing to filterRowsToWhereJson:
    // equals_true → operator: 'equals', value: 'true'
    // equals_false → operator: 'equals', value: 'false'
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
    <div className="border-b border-border bg-background">
      {/* Tab bar */}
      <div className="flex border-b border-border bg-muted/40">
        <button
          type="button"
          onClick={() => handleTabSwitch('structured')}
          className={cn(
            'px-4 py-2 text-xs font-medium transition-colors',
            activeTab === 'structured'
              ? 'border-b-2 border-primary bg-background text-foreground'
              : 'text-muted-foreground hover:text-foreground'
          )}
        >
          🔧 筛选
        </button>
        {hasCopilot && (
          <button
            type="button"
            onClick={() => handleTabSwitch('ai')}
            className={cn(
              'px-4 py-2 text-xs font-medium transition-colors',
              activeTab === 'ai'
                ? 'border-b-2 border-primary bg-background text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            ✨ AI 查询
          </button>
        )}
      </div>

      {/* Tab content */}
      {activeTab === 'structured' && (
        <StructuredFilterTab
          fields={fields}
          rows={rows}
          onRowsChange={setRows}
          onApply={handleApply}
          onClear={handleClear}
        />
      )}
      {activeTab === 'ai' && hasCopilot && (
        <AiQueryTab fields={fields} />
      )}
    </div>
  )
}
