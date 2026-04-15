'use client'

import React from 'react'
import type { WidgetProps } from '@rjsf/utils'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'

/**
 * EnumSchemaSelect widget for RJSF.
 *
 * Reads enum values directly from the JSON Schema `schema.enum` array
 * (populated by the backend's JSONSchemaGenerator). No ui:options needed.
 *
 * Replaces the native <select> with shadcn/ui <Select>.
 */
export function EnumSchemaSelect(props: WidgetProps) {
  const { onChange, disabled, label, schema } = props
  const value = props.value as string | undefined
  const enumValues = (schema.enum ?? []) as string[]
  const currentValue = typeof value === 'string' ? value : ''

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
        {enumValues.map((code) => (
          <SelectItem key={code} value={code}>
            {code}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
