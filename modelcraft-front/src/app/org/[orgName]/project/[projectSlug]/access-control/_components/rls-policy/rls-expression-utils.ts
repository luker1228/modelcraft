import type { RlsAction, RlsExprType } from '@/generated/graphql'

export type RlsExpressionKind = 'using' | 'check'

type SyntaxResult =
  | { valid: true; empty: boolean }
  | { valid: false; empty: false; message: string }

export function validateRlsExpressionSyntax(value: string): SyntaxResult {
  const trimmed = value.trim()
  if (!trimmed) return { valid: true, empty: true }

  try {
    const parsed: unknown = JSON.parse(trimmed)
    const isObject = typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)
    const isBoolean = typeof parsed === 'boolean'

    if (isObject || isBoolean) {
      return { valid: true, empty: false }
    }

    return {
      valid: false,
      empty: false,
      message: '表达式必须是 JSON 对象或布尔常量',
    }
  } catch (error) {
    return {
      valid: false,
      empty: false,
      message: error instanceof Error ? `JSON 语法错误：${error.message}` : 'JSON 语法错误',
    }
  }
}

export function getRlsExpressionType(action: RlsAction, kind: RlsExpressionKind): RlsExprType {
  if (kind === 'check') {
    return action === 'update' ? 'UPDATE_CHECK' : 'INSERT_CHECK'
  }

  if (action === 'update') return 'UPDATE_PREDICATE'
  if (action === 'delete') return 'DELETE_PREDICATE'
  return 'SELECT_PREDICATE'
}
