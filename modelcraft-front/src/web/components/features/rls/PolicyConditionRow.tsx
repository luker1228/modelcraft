'use client'

import * as React from 'react'
import { X } from 'lucide-react'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/web/components/ui/select'
import { Input } from '@/web/components/ui/input'
import { Button } from '@/web/components/ui/button'
import type { RLSScalarOperator } from '@/types/rls'
import type { PolicyConditionRowProps } from './types'

/**
 * 可用的标量操作符
 */
const OPERATORS: Array<{ value: RLSScalarOperator; label: string; needsValue: boolean }> = [
  { value: '_eq', label: '等于', needsValue: true },
  { value: '_neq', label: '不等于', needsValue: true },
  { value: '_gt', label: '大于', needsValue: true },
  { value: '_gte', label: '大于等于', needsValue: true },
  { value: '_lt', label: '小于', needsValue: true },
  { value: '_lte', label: '小于等于', needsValue: true },
  { value: '_in', label: '在列表中', needsValue: true },
  { value: '_nin', label: '不在列表中', needsValue: true },
  { value: '_is_null', label: '是否为空', needsValue: false },
]

/**
 * 单行条件组件
 *
 * 包含字段选择、操作符选择、值输入和删除按钮
 */
export function PolicyConditionRow({
  condition,
  fields,
  authVariables,
  disabled,
  onChange,
  onRemove,
}: PolicyConditionRowProps) {
  const selectedOperator = OPERATORS.find((op) => op.value === condition.operator)
  const needsValue = selectedOperator?.needsValue ?? true

  const handleFieldChange = React.useCallback(
    (field: string) => {
      onChange({
        ...condition,
        field,
      })
    },
    [condition, onChange]
  )

  const handleOperatorChange = React.useCallback(
    (operator: RLSScalarOperator) => {
      onChange({
        ...condition,
        operator,
        // 切换操作符时，如果不需要值，清空值
        value: OPERATORS.find((op) => op.value === operator)?.needsValue ? condition.value : null,
      })
    },
    [condition, onChange]
  )

  const handleValueChange = React.useCallback(
    (value: string) => {
      let parsedValue: string | number | boolean | string[] = value

      // 尝试解析为数字
      if (!Number.isNaN(Number(value)) && value !== '') {
        parsedValue = Number(value)
      }
      // 处理数组输入（以逗号分隔）
      else if (condition.operator === '_in' || condition.operator === '_nin') {
        parsedValue = value.split(',').map((v) => v.trim())
      }

      onChange({
        ...condition,
        value: parsedValue,
      })
    },
    [condition, onChange]
  )

  const handleAuthVariableChange = React.useCallback(
    (authVar: string) => {
      onChange({
        ...condition,
        value: { _auth: authVar },
      })
    },
    [condition, onChange]
  )

  // 判断当前值是否为 auth 变量引用
  const isAuthValue = React.useMemo(() => {
    return typeof condition.value === 'object' && condition.value !== null && '_auth' in condition.value
  }, [condition.value])

  // 获取当前 auth 变量名
  const currentAuthVar = React.useMemo(() => {
    if (isAuthValue && typeof condition.value === 'object' && condition.value !== null) {
      return (condition.value as { _auth: string })._auth
    }
    return ''
  }, [isAuthValue, condition.value])

  // 获取显示的字符串值
  const displayValue = React.useMemo(() => {
    if (condition.value === null || condition.value === undefined) return ''
    if (isAuthValue) return ''
    if (Array.isArray(condition.value)) return condition.value.join(', ')
    return String(condition.value)
  }, [condition.value, isAuthValue])

  return (
    <div className="flex items-start gap-2 rounded-md border bg-background p-3">
      {/* 字段选择 */}
      <Select value={condition.field} onValueChange={handleFieldChange} disabled={disabled}>
        <SelectTrigger className="w-[160px]">
          <SelectValue placeholder="选择字段" />
        </SelectTrigger>
        <SelectContent>
          {fields.map((field) => (
            <SelectItem key={field.name} value={field.name}>
              {field.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* 操作符选择 */}
      <Select
        value={condition.operator}
        onValueChange={(value) => handleOperatorChange(value as RLSScalarOperator)}
        disabled={disabled}
      >
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="操作符" />
        </SelectTrigger>
        <SelectContent>
          {OPERATORS.map((op) => (
            <SelectItem key={op.value} value={op.value}>
              {op.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* 值类型选择（使用 Auth 变量或直接值） */}
      {needsValue && (
        <div className="flex flex-1 gap-2">
          <Select
            value={isAuthValue ? 'auth' : 'literal'}
            onValueChange={(value) => {
              if (value === 'auth') {
                handleAuthVariableChange(authVariables[0] || '')
              } else {
                onChange({
                  ...condition,
                  value: '',
                })
              }
            }}
            disabled={disabled}
          >
            <SelectTrigger className="w-[100px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="literal">直接值</SelectItem>
              <SelectItem value="auth">认证变量</SelectItem>
            </SelectContent>
          </Select>

          {isAuthValue ? (
            <Select
              value={currentAuthVar}
              onValueChange={handleAuthVariableChange}
              disabled={disabled}
            >
              <SelectTrigger className="flex-1">
                <SelectValue placeholder="选择变量" />
              </SelectTrigger>
              <SelectContent>
                {authVariables.map((variable) => (
                  <SelectItem key={variable} value={variable}>
                    {variable}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : (
            <Input
              className="flex-1"
              placeholder={condition.operator === '_in' || condition.operator === '_nin' ? '值1, 值2, ...' : '输入值'}
              value={displayValue}
              onChange={(e) => handleValueChange(e.target.value)}
              disabled={disabled}
            />
          )}
        </div>
      )}

      {/* 删除按钮 */}
      <Button
        variant="ghost"
        size="icon"
        className="shrink-0 text-muted-foreground hover:text-destructive"
        onClick={onRemove}
        disabled={disabled}
      >
        <X className="size-4" />
      </Button>
    </div>
  )
}
