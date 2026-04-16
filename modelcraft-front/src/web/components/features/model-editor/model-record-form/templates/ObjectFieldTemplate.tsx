'use client'

import React from 'react'
import type { ObjectFieldTemplateProps } from '@rjsf/utils'

/**
 * Custom RJSF ObjectFieldTemplate.
 *
 * Removes the default <fieldset>/<legend> wrapper.
 * Renders a clean <div> with consistent vertical spacing between fields.
 * Nested objects get a subtle left border to indicate depth.
 */
export function ObjectFieldTemplate({
  title,
  description,
  properties,
  fieldPathId,
}: ObjectFieldTemplateProps) {
  const isRoot = fieldPathId.$id === 'root'

  return (
    <div className={isRoot ? 'space-y-4' : 'space-y-4 border-l-2 border-border pl-4'}>
      {!isRoot && title && (
        <p className="text-sm font-medium text-foreground">{title}</p>
      )}
      {description && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
      {properties.map((prop) => (
        <div key={prop.name}>{prop.content}</div>
      ))}
    </div>
  )
}
