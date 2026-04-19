'use client'

import * as React from 'react'
import { Plus, Loader2 } from 'lucide-react'

import { Button } from '@/web/components/ui/button'
import { ScrollArea } from '@/web/components/ui/scroll-area'
import type { AuthVariable, AuthVariableInput } from '@/types/rls'
import type { AuthVariableEditorProps } from './types'
import { AuthVariableRow } from './AuthVariableRow'

/**
 * 生成唯一 ID
 */
function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

/**
 * 认证变量编辑器组件
 *
 * 渲染多行 AuthVariableRow，支持添加和删除变量
 */
export function AuthVariableEditor({
  variables,
  onChange,
  disabled,
}: AuthVariableEditorProps) {
  // 本地状态，用于追踪每行的唯一 ID
  const [rows, setRows] = React.useState<
    Array<{ id: string; variable: AuthVariable }>
  >(() =>
    variables.map((v) => ({
      id: generateId(),
      variable: v,
    }))
  )

  // 同步外部变量变化
  React.useEffect(() => {
    // 只在长度变化时重新生成 IDs，保留现有行的 ID
    if (variables.length !== rows.length) {
      setRows(
        variables.map((v, index) => ({
          id: rows[index]?.id || generateId(),
          variable: v,
        }))
      )
    } else {
      // 更新现有行的变量数据
      setRows((prev) =>
        prev.map((row, index) => ({
          ...row,
          variable: variables[index] || row.variable,
        }))
      )
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [variables])

  const handleAddVariable = React.useCallback(() => {
    const newVariable: AuthVariable = {
      name: '',
      source: 'jwt.',
      type: 'STRING',
      isBuiltin: false,
    }
    onChange([...variables, newVariable])
  }, [variables, onChange])

  const handleUpdateVariable = React.useCallback(
    (index: number, updatedVariable: AuthVariable) => {
      const newVariables = [...variables]
      newVariables[index] = updatedVariable
      onChange(newVariables)
    },
    [variables, onChange]
  )

  const handleRemoveVariable = React.useCallback(
    (index: number) => {
      // 跳过内置变量
      if (variables[index]?.isBuiltin) return
      const newVariables = variables.filter((_, i) => i !== index)
      onChange(newVariables)
    },
    [variables, onChange]
  )

  const customVariables = rows.filter((row) => !row.variable.isBuiltin)
  const builtinVariables = rows.filter((row) => row.variable.isBuiltin)

  return (
    <div className="flex flex-col gap-4">
      <ScrollArea className="h-auto max-h-[400px]">
        <div className="flex flex-col gap-3 pr-4">
          {/* 内置变量（只读） */}
          {builtinVariables.map((row) => (
            <AuthVariableRow
              key={row.id}
              variable={row.variable}
              disabled={true}
            />
          ))}

          {/* 自定义变量（可编辑） */}
          {customVariables.map((row, customIndex) => {
            // 计算在原始数组中的索引
            const originalIndex = variables.findIndex(
              (v) => v.name === row.variable.name && v.source === row.variable.source
            )
            const index = originalIndex >= 0 ? originalIndex : customIndex

            return (
              <AuthVariableRow
                key={row.id}
                variable={row.variable}
                onChange={(updated) => handleUpdateVariable(index, updated)}
                onRemove={() => handleRemoveVariable(index)}
                disabled={disabled}
              />
            )
          })}

          {/* 空状态 */}
          {customVariables.length === 0 && (
            <div className="rounded-md border border-dashed p-6 text-center">
              <p className="text-sm text-muted-foreground">
                暂无自定义认证变量，点击下方按钮添加
              </p>
            </div>
          )}
        </div>
      </ScrollArea>

      {/* 添加变量按钮 */}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleAddVariable}
        disabled={disabled}
        className="w-fit"
      >
        {disabled ? (
          <Loader2 className="mr-2 size-4 animate-spin" />
        ) : (
          <Plus className="mr-2 size-4" />
        )}
        添加变量
      </Button>
    </div>
  )
}
