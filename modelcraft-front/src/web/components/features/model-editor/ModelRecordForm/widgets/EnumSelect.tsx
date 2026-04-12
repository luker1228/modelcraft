'use client'

import React, { useState, useMemo } from 'react'
import type { RJSFSchema, WidgetProps } from '@rjsf/utils'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@web/components/ui/popover'
import { Checkbox } from '@web/components/ui/checkbox'
import { Button } from '@web/components/ui/button'
import { ChevronDown, X } from 'lucide-react'

interface EnumOption {
  code: string
  label: string
}

/**
 * EnumSelect widget for RJSF.
 *
 * Single-select: renders a shadcn <Select> dropdown.
 * Multi-select:  renders a <Popover> with a <Checkbox> list, value stored as string[].
 *
 * ui:options expected:
 *   enumValues: { code: string; label: string }[]
 *   multiple:   boolean
 */
export function EnumSelect(props: WidgetProps<unknown, RJSFSchema, Record<string, unknown>>) {
  const value = props.value as unknown
  const onChange = props.onChange as (nextValue: unknown) => void
  const disabled = props.disabled as boolean | undefined
  const label = props.label as string | undefined
  const uiSchema = props.uiSchema as Record<string, unknown> | undefined
  const uiOptions = uiSchema?.['ui:options'] as Record<string, unknown> | undefined
  const enumValues: EnumOption[] = useMemo(
    () => (uiOptions?.enumValues as EnumOption[] | undefined) ?? [],
    [uiOptions?.enumValues]
  )
  const multiple: boolean = (uiOptions?.multiple as boolean) ?? false

  const [open, setOpen] = useState(false)

  // Memoize current values array (for multi-select)
  // IMPORTANT: This hook must be called unconditionally to follow React's rules of hooks
  const currentValues: string[] = useMemo(
    () => (Array.isArray(value) ? (value as string[]) : []),
    [value]
  )

  const selectedLabels = useMemo(
    () =>
      currentValues
        .map((c) => enumValues.find((o) => o.code === c)?.label ?? c)
        .join(', '),
    [currentValues, enumValues]
  )

  const toggleValue = (code: string) => {
    const next = currentValues.includes(code)
      ? currentValues.filter((c) => c !== code)
      : [...currentValues, code]
    onChange(next)
  }

  // ---------- single select ----------
  if (!multiple) {
    const currentValue: string = typeof value === 'string' ? value : ''

    return (
      <Select
        value={currentValue}
        onValueChange={(val) => onChange(val)}
        disabled={disabled}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={`选择${label ?? ''}…`} />
        </SelectTrigger>
        <SelectContent>
          {enumValues.map((opt) => (
            <SelectItem key={opt.code} value={opt.code}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )
  }

  // ---------- multi select ----------

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          disabled={disabled}
          className="h-10 w-full justify-between font-normal"
          type="button"
        >
          <span className="flex-1 truncate text-left">
            {currentValues.length > 0 ? selectedLabels : `选择${label ?? ''}…`}
          </span>
          <ChevronDown className="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-64 p-2">
        <div className="space-y-1">
          {enumValues.map((opt) => (
            <label
              key={opt.code}
              className="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-muted"
            >
              <Checkbox
                checked={currentValues.includes(opt.code)}
                onCheckedChange={() => toggleValue(opt.code)}
                disabled={disabled}
              />
              <span className="text-foreground">{opt.label}</span>
            </label>
          ))}
          {enumValues.length === 0 && (
            <p className="px-2 py-1.5 text-sm text-muted-foreground">暂无选项</p>
          )}
        </div>
        {currentValues.length > 0 && (
          <div className="mt-2 border-t pt-2">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-full text-xs text-muted-foreground"
              type="button"
              onClick={() => onChange([])}
            >
              <X className="mr-1 size-3" />
              清除选择
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
