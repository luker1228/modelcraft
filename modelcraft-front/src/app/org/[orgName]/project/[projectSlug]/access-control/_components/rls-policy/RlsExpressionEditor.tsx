'use client'

import * as React from 'react'
import Link from 'next/link'
import { AlertCircle, CheckCircle2, Loader2, Play } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import type { RlsExprType } from '@/generated/graphql'
import {
  buildRlsCompletionItems,
  extractRlsCompletionContext,
  getRlsAvailableContexts,
  type RlsAuthVariable,
  type RlsExpressionField,
  validateRlsExpressionSyntax,
} from './rls-expression-utils'

interface DryRunResult {
  success: boolean
  message?: string
}

interface RlsExpressionEditorProps {
  label: string
  placeholder: string
  example?: string
  availableFields?: string[]
  modelFields?: RlsExpressionField[]
  authVariables?: RlsAuthVariable[]
  rootLabel?: 'row' | 'input'
  docsHref?: string
  value: string
  onChange: (value: string) => void
  exprType: RlsExprType
  onDryRun?: (input: { expression: string; exprType: RlsExprType; sampleInput?: string }) => Promise<DryRunResult>
  onValidationChange?: (valid: boolean | null) => void
}

export function RlsExpressionEditor({
  label,
  placeholder,
  example,
  availableFields = [],
  modelFields = [],
  authVariables = [],
  rootLabel = 'row',
  docsHref,
  value,
  onChange,
  exprType,
  onDryRun,
  onValidationChange,
}: RlsExpressionEditorProps) {
  const syntax = React.useMemo(() => validateRlsExpressionSyntax(value), [value])
  const [dryRunResult, setDryRunResult] = React.useState<DryRunResult | null>(null)
  const [dryRunning, setDryRunning] = React.useState(false)
  const [cursorPosition, setCursorPosition] = React.useState<number | null>(value.length)
  const [activeSuggestionIndex, setActiveSuggestionIndex] = React.useState(0)
  const textareaRef = React.useRef<HTMLTextAreaElement | null>(null)
  const availableContexts = React.useMemo(() => getRlsAvailableContexts(rootLabel), [rootLabel])
  const completionContext = React.useMemo(
    () => (cursorPosition === null ? null : extractRlsCompletionContext(value, cursorPosition)),
    [value, cursorPosition],
  )
  const completionItems = React.useMemo(
    () =>
      completionContext
        ? buildRlsCompletionItems({
            context: completionContext,
            rootLabel,
            fields: modelFields,
            authVariables,
          })
        : [],
    [authVariables, completionContext, modelFields, rootLabel],
  )
  const completionOpen = completionContext !== null

  React.useEffect(() => {
    setCursorPosition(value.length)
  }, [value])

  const onValidationChangeRef = React.useRef(onValidationChange)
  onValidationChangeRef.current = onValidationChange

  React.useEffect(() => {
    setDryRunResult(null)
    onValidationChangeRef.current?.(null)
  }, [value, exprType])

  React.useEffect(() => {
    setActiveSuggestionIndex(0)
  }, [completionContext?.root, completionContext?.query])

  const handleDryRun = async () => {
    if (!onDryRun || !syntax.valid || syntax.empty) return

    setDryRunning(true)
    try {
      const result = await onDryRun({ expression: value.trim(), exprType })
      setDryRunResult(result)
      onValidationChange?.(result.success)
    } finally {
      setDryRunning(false)
    }
  }

  const applyInsertion = (insertedValue: string, replaceStart?: number, replaceEnd?: number) => {
    const textarea = textareaRef.current
    if (!textarea) {
      onChange(`${value}${value && !/\s$/.test(value) ? ' ' : ''}${insertedValue}`)
      return
    }

    const start = replaceStart ?? textarea.selectionStart ?? value.length
    const end = replaceEnd ?? textarea.selectionEnd ?? value.length
    const before = value.slice(0, start)
    const after = value.slice(end)
    const prefix = before && !/\s$/.test(before) ? ' ' : ''
    const suffix = after && !/^\s/.test(after) ? ' ' : ''
    const insertion = `${prefix}${insertedValue}${suffix}`
    const nextValue = `${before}${insertion}${after}`
    onChange(nextValue)

    const nextCursor = start + insertion.length
    setCursorPosition(nextCursor)
    queueMicrotask(() => {
      textarea.focus()
      textarea.setSelectionRange(nextCursor, nextCursor)
    })
  }

  const handleInsertField = (field: string) => {
    applyInsertion(field)
  }

  const handleApplyCompletion = (completionValue: string) => {
    if (!completionContext) return
    applyInsertion(completionValue, completionContext.replaceStart, completionContext.replaceEnd)
  }

  const handleTextareaKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (!completionOpen || completionItems.length === 0) return

    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveSuggestionIndex((prev) => (prev + 1) % completionItems.length)
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveSuggestionIndex((prev) => (prev - 1 + completionItems.length) % completionItems.length)
      return
    }

    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault()
      handleApplyCompletion(completionItems[activeSuggestionIndex]?.value ?? '')
      return
    }

    if (event.key === 'Escape') {
      event.preventDefault()
      setCursorPosition(null)
    }
  }

  const hasValue = value.trim().length > 0
  const status = (() => {
    if (!hasValue) return null
    if (!syntax.valid) {
      return {
        tone: 'error' as const,
        icon: <AlertCircle className="size-3.5" strokeWidth={1.5} />,
        text: syntax.message,
      }
    }
    if (dryRunResult) {
      return {
        tone: dryRunResult.success ? 'success' as const : 'error' as const,
        icon: dryRunResult.success ? (
          <CheckCircle2 className="size-3.5" strokeWidth={1.5} />
        ) : (
          <AlertCircle className="size-3.5" strokeWidth={1.5} />
        ),
        text: dryRunResult.message ?? (dryRunResult.success ? 'Dry run 通过' : 'Dry run 失败'),
      }
    }
    return {
      tone: 'success' as const,
      icon: <CheckCircle2 className="size-3.5" strokeWidth={1.5} />,
      text: 'CEL 语法初步通过',
    }
  })()

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between gap-3">
        <Label>{label}</Label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="h-7 gap-1.5 px-2 text-xs"
          onClick={handleDryRun}
          disabled={!onDryRun || !syntax.valid || syntax.empty || dryRunning}
        >
          {dryRunning ? (
            <Loader2 className="size-3.5 animate-spin" />
          ) : (
            <Play className="size-3.5" strokeWidth={1.5} />
          )}
          Dry run
        </Button>
      </div>
      <Textarea
        ref={textareaRef}
        value={value}
        onChange={(event) => {
          setCursorPosition(event.target.selectionStart ?? event.target.value.length)
          onChange(event.target.value)
        }}
        onClick={(event) => setCursorPosition(event.currentTarget.selectionStart ?? value.length)}
        onKeyUp={(event) => setCursorPosition(event.currentTarget.selectionStart ?? value.length)}
        onSelect={(event) => setCursorPosition(event.currentTarget.selectionStart ?? value.length)}
        onKeyDown={handleTextareaKeyDown}
        placeholder={placeholder}
        rows={6}
        className="min-h-[150px] font-mono text-xs leading-5"
      />
      {completionOpen && (
        <div className="overflow-hidden rounded-md border border-border bg-popover shadow-sm">
          <div className="border-b border-border/70 px-3 py-2 text-[11px] text-muted-foreground">
            {completionContext?.root === 'auth'
              ? '认证变量补全'
              : `${completionContext?.root} 字段补全`}
          </div>
          {completionItems.length > 0 ? (
            <div className="max-h-56 overflow-y-auto p-1">
              {completionItems.map((item, index) => (
                <button
                  key={item.key}
                  type="button"
                  aria-label={`${item.label} 候选`}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => handleApplyCompletion(item.value)}
                  className={
                    index === activeSuggestionIndex
                      ? 'flex w-full items-start justify-between gap-3 rounded-sm bg-accent p-2 text-left'
                      : 'flex w-full items-start justify-between gap-3 rounded-sm p-2 text-left hover:bg-accent/60'
                  }
                >
                  <span className="font-mono text-xs text-foreground">{item.label}</span>
                  {item.description ? (
                    <span className="shrink-0 text-[11px] text-muted-foreground">{item.description}</span>
                  ) : null}
                </button>
              ))}
            </div>
          ) : (
            <div className="p-3 text-xs text-muted-foreground">没有匹配的可用项</div>
          )}
        </div>
      )}
      <div className="space-y-2 rounded-md border border-dashed border-border/70 bg-muted/20 p-3">
        <div className="flex items-center justify-between gap-3">
          <p className="text-xs leading-5 text-muted-foreground">
            支持 CEL 表达式，可引用 <code>{rootLabel}.字段名</code> 和 <code>auth.xxx</code>
            {example ? <>，示例：<code>{example}</code></> : null}
          </p>
          {docsHref ? (
            <Link
              href={docsHref}
              target="_blank"
              className="shrink-0 text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
            >
              查看示例
            </Link>
          ) : null}
        </div>
        <div className="flex flex-wrap gap-1.5">
          {availableContexts.map((context) => (
            <button
              key={context.key}
              type="button"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => handleInsertField(context.value)}
              className="rounded-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
              <Badge variant="outline" className="cursor-pointer font-mono text-[11px] hover:border-foreground/30 hover:text-foreground">
                {context.value}
              </Badge>
            </button>
          ))}
        </div>
        {availableFields.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {availableFields.map((field) => (
              <button
                key={field}
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => handleInsertField(field)}
                className="rounded-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <Badge variant="secondary" className="cursor-pointer font-mono text-[11px] hover:border-foreground/30 hover:text-foreground">
                  {field}
                </Badge>
              </button>
            ))}
          </div>
        )}
      </div>
      {status && (
        <div
          className={
            status.tone === 'success'
              ? 'text-success flex items-start gap-1.5 text-xs'
              : 'flex items-start gap-1.5 text-xs text-destructive'
          }
        >
          <span className="mt-0.5 shrink-0">{status.icon}</span>
          <span className="leading-5">{status.text}</span>
        </div>
      )}
    </div>
  )
}
