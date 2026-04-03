'use client'

import React from 'react'
import type { BaseInputTemplateProps, RJSFSchema } from '@rjsf/utils'
import { getInputProps } from '@rjsf/utils'

/**
 * Extract typed fields from WidgetProps to avoid `any` index signature pollution.
 * WidgetProps extends GenericObjectType which has `[name: string]: any`,
 * so direct destructuring triggers @typescript-eslint/no-unsafe-* rules.
 */
function extractBaseInputFields(props: BaseInputTemplateProps<string, RJSFSchema, Record<string, unknown>>) {
  return {
    id: props.id as string,
    value: props.value as unknown,
    required: props.required as boolean,
    disabled: props.disabled as boolean,
    readonly: props.readonly as boolean,
    autofocus: props.autofocus as boolean,
    onChange: props.onChange as (value: unknown) => void,
    onChangeOverride: props.onChangeOverride as ((event: React.ChangeEvent<HTMLInputElement>) => void) | undefined,
    onBlur: props.onBlur as (id: string, value: string) => void,
    onFocus: props.onFocus as (id: string, value: string) => void,
    options: props.options as Record<string, unknown>,
    schema: props.schema,
    uiSchema: props.uiSchema as Record<string, unknown> | undefined,
    rawErrors: props.rawErrors as string[] | undefined,
    type: props.type as string | undefined,
  }
}

/**
 * Custom RJSF BaseInputTemplate.
 *
 * Replaces the default Bootstrap <input> with the project design-system Input styling.
 * Handles: text, number, email, url, password, date, datetime-local, time, color.
 */
export function BaseInputTemplate(props: BaseInputTemplateProps<string, RJSFSchema, Record<string, unknown>>) {
  const {
    id,
    value,
    required,
    disabled,
    readonly,
    autofocus,
    onChange,
    onChangeOverride,
    onBlur,
    onFocus,
    options,
    schema,
    uiSchema,
    rawErrors,
    type,
  } = extractBaseInputFields(props)

  const inputProps = getInputProps(schema, type, options as never) as Record<string, unknown>
  const hasError = rawErrors && rawErrors.length > 0

  const handleChange = onChangeOverride
    ? onChangeOverride
    : ({ target: { value: v } }: React.ChangeEvent<HTMLInputElement>) => {
        const emptyValue = options.emptyValue
        onChange(v === '' ? (emptyValue ?? '') : v)
      }

  const handleBlur = ({ target: { value: v } }: React.FocusEvent<HTMLInputElement>) =>
    onBlur(id, v)

  const handleFocus = ({ target: { value: v } }: React.FocusEvent<HTMLInputElement>) =>
    onFocus(id, v)

  // Textarea for multi-line text (ui:widget textarea or schema maxLength hint)
  if ((inputProps.type as string) === 'textarea' || uiSchema?.['ui:widget'] === 'textarea') {
    return (
      <textarea
        id={id}
        value={String(value ?? '')}
        required={required}
        disabled={disabled || readonly}
        autoFocus={autofocus}
        rows={4}
        className={[
          'flex min-h-[80px] w-full rounded-md border bg-transparent px-3 py-2',
          'text-sm shadow-sm placeholder:text-muted-foreground',
          'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
          'disabled:cursor-not-allowed disabled:opacity-50',
          hasError ? 'border-destructive' : 'border-input',
        ].join(' ')}
        onChange={(e) => {
          const emptyValue = options.emptyValue
          onChange(e.target.value === '' ? (emptyValue ?? '') : e.target.value)
        }}
        onBlur={(e) => onBlur(id, e.target.value)}
        onFocus={(e) => onFocus(id, e.target.value)}
      />
    )
  }

  return (
    <input
      id={id}
      value={String(value ?? '')}
      required={required}
      disabled={disabled || readonly}
      autoFocus={autofocus}
      {...inputProps}
      className={[
        'flex h-9 w-full rounded-md border bg-transparent px-3 py-1',
        'text-sm shadow-sm placeholder:text-muted-foreground',
        'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
        'disabled:cursor-not-allowed disabled:opacity-50',
        hasError ? 'border-destructive' : 'border-input',
      ].join(' ')}
      onChange={handleChange}
      onBlur={handleBlur}
      onFocus={handleFocus}
    />
  )
}
