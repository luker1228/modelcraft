'use client'

import * as React from 'react'
import { AlertCircle, CheckCircle2, Loader2, Play } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import type { RlsExprType } from '@/generated/graphql'
import { validateRlsExpressionSyntax } from './rls-expression-utils'

interface DryRunResult {
  success: boolean
  message?: string
}

interface RlsExpressionEditorProps {
  label: string
  value: string
  onChange: (value: string) => void
  exprType: RlsExprType
  onDryRun?: (input: { expression: string; exprType: RlsExprType }) => Promise<DryRunResult>
}

export function RlsExpressionEditor({
  label,
  value,
  onChange,
  exprType,
  onDryRun,
}: RlsExpressionEditorProps) {
  const syntax = React.useMemo(() => validateRlsExpressionSyntax(value), [value])
  const [dryRunResult, setDryRunResult] = React.useState<DryRunResult | null>(null)
  const [dryRunning, setDryRunning] = React.useState(false)

  React.useEffect(() => {
    setDryRunResult(null)
  }, [value, exprType])

  const handleDryRun = async () => {
    if (!onDryRun || !syntax.valid || syntax.empty) return

    setDryRunning(true)
    try {
      setDryRunResult(await onDryRun({ expression: value.trim(), exprType }))
    } finally {
      setDryRunning(false)
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
      text: 'JSON 语法通过',
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
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder='例如：{"owner_id": {"equals": "{{user_id}}"}}'
        rows={6}
        className="min-h-[150px] font-mono text-xs leading-5"
      />
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
