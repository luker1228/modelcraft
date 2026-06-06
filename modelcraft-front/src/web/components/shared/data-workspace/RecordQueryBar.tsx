'use client'

import { Search } from 'lucide-react'

import { Input } from '@web/components/ui/input'

interface RecordQueryBarProps {
  leftContent?: React.ReactNode
  searchValue: string
  onSearchChange: (value: string) => void
  searchPlaceholder: string
  summaryText?: string
  searchInputClassName?: string
}

export function RecordQueryBar({
  leftContent,
  searchValue,
  onSearchChange,
  searchPlaceholder,
  summaryText,
  searchInputClassName = 'h-[26px] w-40 text-xs',
}: RecordQueryBarProps) {
  return (
    <div className="flex min-h-[44px] flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2">
      {leftContent}

      <div className="ml-auto flex items-center gap-2">
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1.5">
            <Search className="size-3.5 shrink-0 text-muted-foreground" />
            <Input
              value={searchValue}
              onChange={(event) => onSearchChange(event.target.value)}
              placeholder={searchPlaceholder}
              className={searchInputClassName}
            />
          </div>
          {summaryText ? (
            <span className="text-xs text-muted-foreground">{summaryText}</span>
          ) : null}
        </div>
      </div>
    </div>
  )
}
