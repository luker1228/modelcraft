'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

import * as React from 'react'

import { cn } from '@/shared/utils'
import { Input } from '@/web/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/web/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/web/components/ui/table'
import type { ColumnAccessMode, ColumnPolicy } from '@/types'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface ModelField {
  name: string
  title: string
  format: string
}

interface ColumnPolicyEditorProps {
  fields: ModelField[]
  value: ColumnPolicy
  onChange: (policy: ColumnPolicy) => void
  /** 当前权限点的 action（影响可选 mode） */
  action: string
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** 所有可能的默认模式（不含 MASKED，MASKED 只作字段覆盖用） */
const DEFAULT_MODE_OPTIONS: ColumnAccessMode[] = ['VISIBLE', 'READONLY', 'HIDDEN']

const DEFAULT_MODE_LABELS: Record<ColumnAccessMode, string> = {
  VISIBLE: '可见',
  READONLY: '只读',
  MASKED: '脱敏',
  HIDDEN: '隐藏',
}

/** 字段覆盖 Select 占位值（表示「使用默认」） */
const USE_DEFAULT = '__default__'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * 根据 action 确定字段覆盖下拉中可出现的 mode
 */
function getAvailableModes(action: string): ColumnAccessMode[] {
  const modes: ColumnAccessMode[] = ['VISIBLE', 'HIDDEN']

  if (['SELECT', 'INSERT', 'UPDATE'].includes(action)) {
    modes.push('READONLY')
  }
  if (action === 'SELECT') {
    modes.push('MASKED')
  }

  // 按固定顺序排列
  const order: ColumnAccessMode[] = ['VISIBLE', 'READONLY', 'MASKED', 'HIDDEN']
  return order.filter((m) => modes.includes(m))
}

// ---------------------------------------------------------------------------
// Sub-component: default mode radio group
// ---------------------------------------------------------------------------

interface DefaultModePickerProps {
  value: ColumnAccessMode
  onChange: (mode: ColumnAccessMode) => void
}

function DefaultModePicker({ value, onChange }: DefaultModePickerProps) {
  return (
    <div className="flex items-center gap-1" role="radiogroup" aria-label="默认列访问模式">
      {DEFAULT_MODE_OPTIONS.map((mode) => (
        <label
          key={mode}
          className={cn(
            'flex items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm cursor-pointer transition-colors',
            value === mode
              ? 'border-primary bg-primary/5 text-foreground font-medium'
              : 'text-muted-foreground hover:bg-muted/40',
          )}
        >
          <input
            type="radio"
            name="column-default-mode"
            value={mode}
            checked={value === mode}
            onChange={() => onChange(mode)}
            className="size-3.5 accent-primary cursor-pointer"
          />
          {DEFAULT_MODE_LABELS[mode]}
        </label>
      ))}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function ColumnPolicyEditor({
  fields,
  value,
  onChange,
  action,
}: ColumnPolicyEditorProps) {
  const availableModes = getAvailableModes(action)

  // ---- helpers ----

  /** 获取某字段当前的覆盖规则 */
  const getRuleForField = (fieldName: string) =>
    value.rules.find((r) => r.fieldName === fieldName)

  /** 更新 defaultMode */
  const handleDefaultModeChange = (mode: ColumnAccessMode) => {
    onChange({ ...value, defaultMode: mode })
  }

  /** 更新某字段的覆盖 mode（选 USE_DEFAULT 则移除条目） */
  const handleFieldModeChange = (fieldName: string, mode: string) => {
    if (mode === USE_DEFAULT) {
      onChange({
        ...value,
        rules: value.rules.filter((r) => r.fieldName !== fieldName),
      })
      return
    }

    const accessMode = mode as ColumnAccessMode
    const existing = getRuleForField(fieldName)

    if (existing) {
      onChange({
        ...value,
        rules: value.rules.map((r) =>
          r.fieldName === fieldName
            ? { ...r, mode: accessMode, maskPattern: accessMode === 'MASKED' ? r.maskPattern : undefined }
            : r,
        ),
      })
    } else {
      onChange({
        ...value,
        rules: [...value.rules, { fieldName, mode: accessMode }],
      })
    }
  }

  /** 更新某字段的 maskPattern */
  const handleMaskPatternChange = (fieldName: string, maskPattern: string) => {
    onChange({
      ...value,
      rules: value.rules.map((r) =>
        r.fieldName === fieldName ? { ...r, maskPattern: maskPattern || undefined } : r,
      ),
    })
  }

  // ---- render ----

  return (
    <div className="flex flex-col gap-4">
      {/* Default mode picker */}
      <div className="flex items-center gap-3 rounded-md border border-border bg-muted/30 px-4 py-3">
        <span className="text-sm font-medium text-foreground shrink-0">默认模式：</span>
        <DefaultModePicker value={value.defaultMode} onChange={handleDefaultModeChange} />
      </div>

      {/* Per-field override table */}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[40%]">字段名</TableHead>
            <TableHead>覆盖策略（留空 = 使用默认模式）</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {fields.map((field) => {
            const rule = getRuleForField(field.name)
            const currentMode = rule?.mode ?? USE_DEFAULT
            const isMasked = currentMode === 'MASKED'

            return (
              <TableRow key={field.name}>
                {/* Field name column */}
                <TableCell>
                  <div className="flex flex-col gap-0.5">
                    <span className="text-sm font-medium text-foreground">{field.name}</span>
                    <span className="text-xs text-muted-foreground">{field.title}</span>
                  </div>
                </TableCell>

                {/* Override mode column */}
                <TableCell>
                  <div className="flex items-center gap-2">
                    <Select
                      value={currentMode}
                      onValueChange={(v) => handleFieldModeChange(field.name, v)}
                    >
                      <SelectTrigger className="h-8 w-36 text-sm">
                        <SelectValue placeholder="使用默认" />
                      </SelectTrigger>
                      <SelectContent>
                        {/* "Use default" option */}
                        <SelectItem value={USE_DEFAULT}>
                          <span className="text-muted-foreground">使用默认</span>
                        </SelectItem>

                        {/* Mode options filtered by action */}
                        {availableModes.map((mode) => (
                          <SelectItem key={mode} value={mode}>
                            {DEFAULT_MODE_LABELS[mode]}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>

                    {/* maskPattern input — only shown when MASKED is selected */}
                    {isMasked && (
                      <Input
                        className="h-8 w-44 text-sm"
                        placeholder="如 138****{last4}"
                        value={rule?.maskPattern ?? ''}
                        onChange={(e) => handleMaskPatternChange(field.name, e.target.value)}
                      />
                    )}
                  </div>
                </TableCell>
              </TableRow>
            )
          })}

          {/* Empty state */}
          {fields.length === 0 && (
            <TableRow>
              <TableCell colSpan={2} className="text-center text-muted-foreground py-6">
                暂无字段
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
