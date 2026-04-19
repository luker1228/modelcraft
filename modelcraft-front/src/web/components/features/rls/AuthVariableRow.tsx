'use client'

import * as React from 'react'
import { Trash2 } from 'lucide-react'

import { Input } from '@/web/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/web/components/ui/select'
import { Button } from '@/web/components/ui/button'
import type { AuthVariable, AuthVariableType } from '@/types/rls'
import type { AuthVariableRowProps } from './types'

/**
 * 认证变量类型选项
 */
const VARIABLE_TYPES: Array<{ value: AuthVariableType; label: string }> = [
  { value: 'UUID', label: 'UUID' },
  { value: 'STRING', label: '字符串' },
  { value: 'INTEGER', label: '整数' },
]

/**
 * 单行认证变量组件
 *
 * 包含变量名输入、source 输入、type 下拉选择和删除按钮
 * 内置 uid 变量只读显示
 */
export function AuthVariableRow({
  variable,
  onChange,
  onRemove,
  disabled,
}: AuthVariableRowProps) {
  const isBuiltin = variable.isBuiltin ?? false
  const isDisabled = disabled || isBuiltin

  const handleNameChange = React.useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (isBuiltin || !onChange) return
      onChange({
        ...variable,
        name: e.target.value,
      })
    },
    [isBuiltin, variable, onChange]
  )

  const handleSourceChange = React.useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (isBuiltin || !onChange) return
      onChange({
        ...variable,
        source: e.target.value,
      })
    },
    [isBuiltin, variable, onChange]
  )

  const handleTypeChange = React.useCallback(
    (value: AuthVariableType) => {
      if (isBuiltin || !onChange) return
      onChange({
        ...variable,
        type: value,
      })
    },
    [isBuiltin, variable, onChange]
  )

  const handleRemove = React.useCallback(() => {
    if (isBuiltin || !onRemove) return
    onRemove()
  }, [isBuiltin, onRemove])

  return (
    <div className="flex items-center gap-3 rounded-md border bg-background p-3">
      {/* 变量名输入 */}
      <div className="flex flex-1 flex-col gap-1">
        <label className="text-xs text-muted-foreground">变量名</label>
        <Input
          placeholder="变量名（如 tenant_id）"
          value={variable.name}
          onChange={handleNameChange}
          disabled={isDisabled}
          className="h-8"
        />
      </div>

      {/* Source 输入 */}
      <div className="flex flex-1 flex-col gap-1">
        <label className="text-xs text-muted-foreground">来源路径</label>
        <Input
          placeholder="jwt.xxx 格式"
          value={variable.source}
          onChange={handleSourceChange}
          disabled={isDisabled}
          className="h-8 font-mono text-sm"
        />
      </div>

      {/* Type 下拉选择 */}
      <div className="flex w-28 flex-col gap-1">
        <label className="text-xs text-muted-foreground">类型</label>
        <Select
          value={variable.type}
          onValueChange={(value) => handleTypeChange(value as AuthVariableType)}
          disabled={isDisabled}
        >
          <SelectTrigger className="h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {VARIABLE_TYPES.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                {type.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* 删除按钮（内置变量不显示） */}
      {!isBuiltin && onRemove && (
        <Button
          variant="ghost"
          size="icon"
          className="mt-5 shrink-0 text-muted-foreground hover:text-destructive"
          onClick={handleRemove}
          disabled={disabled}
        >
          <Trash2 className="size-4" />
        </Button>
      )}

      {/* 内置变量标识 */}
      {isBuiltin && (
        <div className="mt-5 shrink-0 px-3 text-xs text-muted-foreground">
          内置
        </div>
      )}
    </div>
  )
}
