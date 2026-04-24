'use client'

import * as React from 'react'

import { cn } from '@/shared/utils'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/web/components/ui/tooltip'
import type { EndUserRowScope } from '@/types'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface RowScopeSelectorProps {
  value: EndUserRowScope | null
  onChange: (scope: EndUserRowScope) => void
  /** Model 是否有 owner（EndUserRef 类型）字段 */
  hasOwnerField: boolean
  /** Model 是否有 dept_id 字段 */
  hasDeptIdField: boolean
  disabled?: boolean
}

// ---------------------------------------------------------------------------
// Option definition
// ---------------------------------------------------------------------------

interface ScopeOption {
  value: EndUserRowScope
  label: string
  description: string
  /** 返回 undefined 表示不 disabled；返回字符串作为 Tooltip 文案 */
  disabledReason?: (ctx: {
    hasOwnerField: boolean
    hasDeptIdField: boolean
  }) => string | undefined
}

const SCOPE_OPTIONS: ScopeOption[] = [
  {
    value: 'ALL',
    label: '全部行',
    description: '全部行，不过滤',
  },
  {
    value: 'SELF',
    label: '仅自己的行',
    description: '仅当前用户自己的行（需要 owner 字段）',
    disabledReason: ({ hasOwnerField }) =>
      hasOwnerField ? undefined : '该 Model 缺少 owner 字段（EndUserRef 类型）',
  },
  {
    value: 'DEPT',
    label: '所在部门的行',
    description: '当前用户所在部门的行（需要 dept_id 字段）',
    disabledReason: ({ hasDeptIdField }) =>
      hasDeptIdField ? undefined : '该 Model 缺少 dept_id 字段',
  },
  {
    value: 'DEPT_AND_CHILDREN',
    label: '部门及下级的行',
    description: '当前部门及所有下级部门的行（需要 dept_id 字段）',
    disabledReason: ({ hasDeptIdField }) =>
      hasDeptIdField ? undefined : '该 Model 缺少 dept_id 字段',
  },
]

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function RowScopeSelector({
  value,
  onChange,
  hasOwnerField,
  hasDeptIdField,
  disabled = false,
}: RowScopeSelectorProps) {
  return (
    <TooltipProvider>
      <div role="radiogroup" aria-label="行访问范围" className="flex flex-col gap-2">
        {SCOPE_OPTIONS.map((option) => {
          const disabledReason = option.disabledReason?.({ hasOwnerField, hasDeptIdField })
          const isDisabled = disabled || !!disabledReason
          const isChecked = value === option.value

          const radioEl = (
            <label
              key={option.value}
              className={cn(
                'flex items-start gap-3 rounded-md border border-border px-4 py-3 transition-colors',
                isChecked && !isDisabled && 'border-primary bg-primary/5',
                isDisabled
                  ? 'opacity-50 cursor-not-allowed'
                  : 'cursor-pointer hover:bg-muted/40',
              )}
            >
              {/* Native radio input */}
              <input
                type="radio"
                name="row-scope"
                value={option.value}
                checked={isChecked}
                disabled={isDisabled}
                onChange={() => {
                  if (!isDisabled) onChange(option.value)
                }}
                className={cn(
                  'mt-0.5 size-4 accent-primary',
                  isDisabled ? 'cursor-not-allowed' : 'cursor-pointer',
                )}
              />

              {/* Text block */}
              <div className="flex flex-col gap-0.5">
                <span className="text-sm font-medium text-foreground leading-none">
                  {option.label}
                </span>
                <span className="text-sm text-muted-foreground">{option.description}</span>
              </div>
            </label>
          )

          // Disabled options get a Tooltip explaining why
          if (disabledReason) {
            return (
              <Tooltip key={option.value}>
                {/* span wrapper: disabled elements don't fire pointer events */}
                <TooltipTrigger asChild>
                  <span className="block">{radioEl}</span>
                </TooltipTrigger>
                <TooltipContent side="right">{disabledReason}</TooltipContent>
              </Tooltip>
            )
          }

          return radioEl
        })}
      </div>
    </TooltipProvider>
  )
}
