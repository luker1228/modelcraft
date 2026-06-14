import type { RlsAction, RlsExprType } from '@/generated/graphql'

export type RlsExpressionKind = 'using' | 'check'

type SyntaxResult =
  | { valid: true; empty: boolean }
  | { valid: false; empty: false; message: string }

export function validateRlsExpressionSyntax(value: string): SyntaxResult {
  const trimmed = value.trim()
  if (!trimmed) return { valid: true, empty: true }

  if (trimmed.startsWith('{')) {
    return {
      valid: false,
      empty: false,
      message: '请输入 CEL 表达式，例如 row.owner_id == auth.user_id',
    }
  }

  if (trimmed === 'true' || trimmed === 'false') {
    return { valid: true, empty: false }
  }

  const hasAllowedRoot = /\b(row|input|auth)\.[A-Za-z_][A-Za-z0-9_]*/.test(trimmed)
  if (!hasAllowedRoot) {
    return {
      valid: false,
      empty: false,
      message: '表达式需要引用 row、input 或 auth',
    }
  }

  return { valid: true, empty: false }
}

export function getRlsExpressionType(action: RlsAction, kind: RlsExpressionKind): RlsExprType {
  if (kind === 'check') {
    return action === 'update' ? 'UPDATE_CHECK' : 'INSERT_CHECK'
  }

  if (action === 'update') return 'UPDATE_PREDICATE'
  if (action === 'delete') return 'DELETE_PREDICATE'
  return 'SELECT_PREDICATE'
}

export function shouldShowRlsUsingExpression(action: RlsAction): boolean {
  return action === 'read' || action === 'update' || action === 'delete'
}

export function shouldShowRlsCheckExpression(action: RlsAction): boolean {
  return action === 'create' || action === 'update'
}
