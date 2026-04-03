'use client'

import React from 'react'
import type { FieldTemplateProps } from '@rjsf/utils'

/**
 * Custom RJSF FieldTemplate.
 *
 * Replaces the default Bootstrap-style wrapper with project design-system styling:
 * - Label: text-sm font-medium, semantic foreground color
 * - Required asterisk in destructive color
 * - Description in muted-foreground
 * - Inline error list in destructive color
 */
export function FieldTemplate({
  id,
  label,
  children,
  rawErrors,
  rawDescription,
  hidden,
  required,
  displayLabel,
  fieldPathId,
}: FieldTemplateProps) {
  if (hidden) {
    return <div className="hidden">{children}</div>
  }

  const path = fieldPathId?.path ?? []
  const rawFieldName = path[path.length - 1]
  const fieldName = typeof rawFieldName === 'string' ? rawFieldName : ''
  const labelWithName = label && fieldName ? `${label} (${fieldName})` : label

  return (
    <div className="space-y-1.5">
      {displayLabel && labelWithName && (
        <label
          htmlFor={id}
          className="block text-sm font-medium text-foreground"
        >
          {labelWithName}
          {required && (
            <span className="ml-0.5 text-destructive" aria-hidden="true">
              *
            </span>
          )}
        </label>
      )}
      {children}
      {rawDescription && (
        <p className="text-xs text-muted-foreground">{rawDescription}</p>
      )}
      {rawErrors && rawErrors.length > 0 && (
        <ul className="space-y-0.5">
          {rawErrors.map((error, i) => (
            <li key={i} className="text-xs text-destructive">
              {error}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
