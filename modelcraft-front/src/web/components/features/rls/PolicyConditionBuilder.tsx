'use client'

import * as React from 'react'
import { Plus } from 'lucide-react'

import { Button } from '@/web/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/web/components/ui/select'
import { Alert, AlertDescription } from '@/web/components/ui/alert'
import type { ConditionRow, JsonExpr, ValidationResult } from '@/types/rls'
import type { PolicyConditionBuilderProps, LogicalOperator } from './types'
import { PolicyConditionRow } from './PolicyConditionRow'

/**
 * 生成唯一 ID
 */
function generateId(): string {
  return `cond_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`
}

/**
 * 将条件行转换为 JSON 表达式
 */
function buildJsonExpr(
  conditions: ConditionRow[],
  logicalOperator: LogicalOperator
): JsonExpr | null {
  if (conditions.length === 0) return null

  // 单个条件直接返回
  if (conditions.length === 1) {
    const cond = conditions[0]
    return {
      [cond.field]: {
        [cond.operator]: cond.value,
      },
    }
  }

  // 多个条件使用逻辑操作符
  const conditionExprs = conditions.map((cond) => ({
    [cond.field]: {
      [cond.operator]: cond.value,
    },
  }))

  return {
    [logicalOperator === 'AND' ? '_and' : '_or']: conditionExprs,
  }
}

/**
 * 从 JSON 表达式解析条件行
 */
function parseJsonExpr(expr: JsonExpr | null): { conditions: ConditionRow[]; operator: LogicalOperator } {
  if (!expr) {
    return { conditions: [], operator: 'AND' }
  }

  const keys = Object.keys(expr)

  // 检查是否是逻辑操作符
  if (keys.length === 1 && (keys[0] === '_and' || keys[0] === '_or')) {
    const operator = keys[0] === '_and' ? 'AND' : 'OR'
    const conditions: ConditionRow[] = []

    const exprs = expr[keys[0]] as JsonExpr[]
    if (Array.isArray(exprs)) {
      exprs.forEach((e) => {
        const fieldKeys = Object.keys(e)
        if (fieldKeys.length === 1) {
          const field = fieldKeys[0]
          const opObj = e[field] as JsonExpr
          const opKeys = Object.keys(opObj)
          if (opKeys.length === 1) {
            conditions.push({
              id: generateId(),
              field,
              operator: opKeys[0] as ConditionRow['operator'],
              value: opObj[opKeys[0]],
            })
          }
        }
      })
    }

    return { conditions, operator }
  }

  // 单个条件
  if (keys.length === 1) {
    const field = keys[0]
    const opObj = expr[field] as JsonExpr
    const opKeys = Object.keys(opObj)
    if (opKeys.length === 1) {
      return {
        conditions: [
          {
            id: generateId(),
            field,
            operator: opKeys[0] as ConditionRow['operator'],
            value: opObj[opKeys[0]],
          },
        ],
        operator: 'AND',
      }
    }
  }

  return { conditions: [], operator: 'AND' }
}

/**
 * 条件构建器容器
 *
 * 管理多行条件、逻辑关系切换和实时校验
 */
export function PolicyConditionBuilder({
  modelId,
  exprType,
  value,
  fields,
  authVariables,
  onChange,
  onValidationResult,
}: PolicyConditionBuilderProps) {
  const { conditions, operator } = React.useMemo(() => parseJsonExpr(value), [value])

  const [localConditions, setLocalConditions] = React.useState<ConditionRow[]>(conditions)
  const [logicalOperator, setLogicalOperator] = React.useState<LogicalOperator>(operator)
  const [validationErrors, setValidationErrors] = React.useState<string[]>([])

  // 同步外部 value 变化
  React.useEffect(() => {
    const parsed = parseJsonExpr(value)
    setLocalConditions(parsed.conditions)
    setLogicalOperator(parsed.operator)
  }, [value])

  // 构建 JSON 表达式并触发校验
  const buildAndValidate = React.useCallback(
    (conds: ConditionRow[], op: LogicalOperator) => {
      const expr = buildJsonExpr(conds, op)
      onChange(expr ?? {})

      // 实时校验
      const errors: string[] = []

      // 检查每个条件的完整性
      conds.forEach((cond, index) => {
        if (!cond.field) {
          errors.push(`条件 ${index + 1}: 请选择字段`)
        }
        if (!cond.operator) {
          errors.push(`条件 ${index + 1}: 请选择操作符`)
        }
        // 需要值的操作符检查值
        if (
          cond.operator !== '_is_null' &&
          (cond.value === null || cond.value === undefined || cond.value === '')
        ) {
          errors.push(`条件 ${index + 1}: 请输入值`)
        }
      })

      // 检查字段是否存在于白名单
      conds.forEach((cond, index) => {
        if (cond.field && !fields.some((f) => f.name === cond.field)) {
          errors.push(`条件 ${index + 1}: 字段 "${cond.field}" 不存在`)
        }
      })

      setValidationErrors(errors)

      const result: ValidationResult = {
        valid: errors.length === 0,
        errors: errors.map((msg, idx) => ({
          path: `${exprType}[${idx}]`,
          message: msg,
          code: 'VALIDATION_ERROR',
        })),
      }

      onValidationResult?.(result)

      return result.valid
    },
    [exprType, fields, onChange, onValidationResult]
  )

  const handleAddCondition = React.useCallback(() => {
    const newCondition: ConditionRow = {
      id: generateId(),
      field: fields[0]?.name || '',
      operator: '_eq',
      value: '',
    }
    const updated = [...localConditions, newCondition]
    setLocalConditions(updated)
    buildAndValidate(updated, logicalOperator)
  }, [fields, localConditions, logicalOperator, buildAndValidate])

  const handleRemoveCondition = React.useCallback(
    (index: number) => {
      const updated = localConditions.filter((_, i) => i !== index)
      setLocalConditions(updated)
      buildAndValidate(updated, logicalOperator)
    },
    [localConditions, logicalOperator, buildAndValidate]
  )

  const handleConditionChange = React.useCallback(
    (index: number, condition: ConditionRow) => {
      const updated = [...localConditions]
      updated[index] = condition
      setLocalConditions(updated)
      buildAndValidate(updated, logicalOperator)
    },
    [localConditions, logicalOperator, buildAndValidate]
  )

  const handleOperatorChange = React.useCallback(
    (op: LogicalOperator) => {
      setLogicalOperator(op)
      buildAndValidate(localConditions, op)
    },
    [localConditions, buildAndValidate]
  )

  return (
    <div className="space-y-4">
      {/* 逻辑操作符选择 */}
      {localConditions.length > 1 && (
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">条件关系：</span>
          <Select value={logicalOperator} onValueChange={(v) => handleOperatorChange(v as LogicalOperator)}>
            <SelectTrigger className="w-[100px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="AND">全部满足 (AND)</SelectItem>
              <SelectItem value="OR">任一满足 (OR)</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}

      {/* 条件列表 */}
      <div className="space-y-2">
        {localConditions.map((condition, index) => (
          <PolicyConditionRow
            key={condition.id}
            condition={condition}
            fields={fields}
            authVariables={authVariables}
            onChange={(cond) => handleConditionChange(index, cond)}
            onRemove={() => handleRemoveCondition(index)}
          />
        ))}
      </div>

      {/* 添加条件按钮 */}
      <Button
        variant="outline"
        size="sm"
        onClick={handleAddCondition}
        disabled={fields.length === 0}
      >
        <Plus className="mr-2 size-4" />
        添加条件
      </Button>

      {/* 校验错误提示 */}
      {validationErrors.length > 0 && (
        <Alert variant="destructive">
          <AlertDescription>
            <ul className="list-inside list-disc space-y-1">
              {validationErrors.map((error, index) => (
                <li key={index}>{error}</li>
              ))}
            </ul>
          </AlertDescription>
        </Alert>
      )}

      {/* 空状态提示 */}
      {localConditions.length === 0 && (
        <div className="rounded-md border border-dashed p-6 text-center">
          <p className="text-sm text-muted-foreground">
            暂无条件，点击上方按钮添加
          </p>
        </div>
      )}
    </div>
  )
}
