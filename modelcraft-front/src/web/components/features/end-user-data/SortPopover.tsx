'use client'

import { useState } from 'react'
import { ArrowUpDown, Check } from 'lucide-react'
import { cn } from '@/shared/utils'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import { Button } from '@web/components/ui/button'
import type { FieldDefinition } from '@api-client/cms/public'

export interface SortState {
  /** Field to sort by. Empty string means no user-selected sort. */
  field: string
  direction: 'asc' | 'desc'
  /** When true, always appends {id: 'desc'} as a tiebreaker. */
  stableSort: boolean
}

export function buildOrderBy(
  sort: SortState
): Record<string, string>[] | undefined {
  const result: Record<string, string>[] = []

  if (sort.field) {
    result.push({ [sort.field]: sort.direction })
  }

  // Stable sort: append id:desc as tiebreaker (skip if id is already the sort field)
  if (sort.stableSort && sort.field !== 'id') {
    result.push({ id: 'desc' })
  }

  return result.length > 0 ? result : undefined
}

interface SortPopoverProps {
  fields: FieldDefinition[]
  sort: SortState
  onSortChange: (sort: SortState) => void
}

export function SortPopover({ fields, sort, onSortChange }: SortPopoverProps) {
  const [open, setOpen] = useState(false)

  const displayFields = fields.filter((f) => !f.name.startsWith('_'))
  const isActive = Boolean(sort.field)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn(
            'h-[26px] border-transparent px-2.5 text-xs font-normal',
            isActive
              ? 'border border-primary text-primary'
              : 'text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <ArrowUpDown className="mr-1.5 size-3.5" />
          <span>排序</span>
          {isActive && (
            <span className="ml-1.5 font-mono text-[10px] text-primary opacity-80">
              {sort.field} {sort.direction}
            </span>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent side="bottom" align="start" className="w-64 p-3">
        <p className="mb-2.5 text-xs font-medium text-foreground">排序</p>

        {/* Field + direction row */}
        <div className="flex gap-1.5">
          {/* Field selector */}
          <div className="relative flex-1">
            <select
              value={sort.field}
              onChange={(e) =>
                onSortChange({ ...sort, field: e.target.value })
              }
              className="h-8 w-full cursor-pointer appearance-none rounded-md border border-input bg-background py-0 pl-2.5 pr-6 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
            >
              <option value="">— 不排序 —</option>
              {displayFields.map((f) => (
                <option key={f.name} value={f.name}>
                  {f.name}
                </option>
              ))}
            </select>
            <svg
              className="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground/60"
              width="10" height="10" viewBox="0 0 16 16" fill="currentColor"
            >
              <path d="M8 10.5L3 5.5h10z" />
            </svg>
          </div>

          {/* Direction toggle */}
          <div className="flex overflow-hidden rounded-md border border-input">
            {(['asc', 'desc'] as const).map((dir) => (
              <button
                key={dir}
                type="button"
                onClick={() => onSortChange({ ...sort, direction: dir })}
                className={cn(
                  'px-2.5 py-0 text-xs transition-colors',
                  sort.direction === dir
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-background text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                {dir}
              </button>
            ))}
          </div>
        </div>

        {/* Stable sort toggle */}
        <button
          type="button"
          onClick={() =>
            onSortChange({ ...sort, stableSort: !sort.stableSort })
          }
          className="mt-3 flex w-full items-center justify-between rounded-md p-1 text-xs transition-colors hover:bg-muted"
        >
          <span className="text-muted-foreground">
            稳定排序
            <span className="ml-1 text-[10px] text-muted-foreground/60">
              (追加 id: desc)
            </span>
          </span>
          <span
            className={cn(
              'flex size-4 items-center justify-center rounded-sm border transition-colors',
              sort.stableSort
                ? 'border-primary bg-primary text-primary-foreground'
                : 'border-input bg-background text-transparent'
            )}
          >
            <Check size={10} strokeWidth={2.5} />
          </span>
        </button>

        {/* Clear button */}
        {isActive && (
          <button
            type="button"
            onClick={() => onSortChange({ field: '', direction: 'asc', stableSort: sort.stableSort })}
            className="mt-2 w-full text-center text-[10px] text-muted-foreground hover:text-foreground"
          >
            清除排序
          </button>
        )}
      </PopoverContent>
    </Popover>
  )
}
