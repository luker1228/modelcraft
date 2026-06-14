import { describe, expect, it } from 'vitest'
import {
  getRlsExpressionType,
  shouldShowRlsCheckExpression,
  shouldShowRlsUsingExpression,
  validateRlsExpressionSyntax,
} from './rls-expression-utils'

describe('validateRlsExpressionSyntax', () => {
  it('accepts empty optional expressions', () => {
    expect(validateRlsExpressionSyntax('')).toEqual({ valid: true, empty: true })
    expect(validateRlsExpressionSyntax('   ')).toEqual({ valid: true, empty: true })
  })

  it('accepts CEL expressions and boolean constants', () => {
    expect(validateRlsExpressionSyntax('row.owner_id == auth.user_id')).toEqual({
      valid: true,
      empty: false,
    })
    expect(validateRlsExpressionSyntax('input.status in ["draft", "pending"]')).toEqual({
      valid: true,
      empty: false,
    })
    expect(validateRlsExpressionSyntax('true')).toEqual({ valid: true, empty: false })
    expect(validateRlsExpressionSyntax('false')).toEqual({ valid: true, empty: false })
  })

  it('rejects JSON syntax and expressions without allowed roots', () => {
    expect(validateRlsExpressionSyntax('{"owner_id":{"equals":"{{user_id}}"}}')).toEqual({
      valid: false,
      empty: false,
      message: '请输入 CEL 表达式，例如 row.owner_id == auth.user_id',
    })
    expect(validateRlsExpressionSyntax('owner_id == user_id')).toEqual({
      valid: false,
      empty: false,
      message: '表达式需要引用 row、input 或 auth',
    })
  })
})

describe('getRlsExpressionType', () => {
  it('maps using expressions by action', () => {
    expect(getRlsExpressionType('read', 'using')).toBe('SELECT_PREDICATE')
    expect(getRlsExpressionType('update', 'using')).toBe('UPDATE_PREDICATE')
    expect(getRlsExpressionType('delete', 'using')).toBe('DELETE_PREDICATE')
  })

  it('maps check expressions by action', () => {
    expect(getRlsExpressionType('create', 'check')).toBe('INSERT_CHECK')
    expect(getRlsExpressionType('update', 'check')).toBe('UPDATE_CHECK')
  })
})

describe('action expression visibility', () => {
  it('shows using only for read and delete', () => {
    expect(shouldShowRlsUsingExpression('read')).toBe(true)
    expect(shouldShowRlsCheckExpression('read')).toBe(false)
    expect(shouldShowRlsUsingExpression('delete')).toBe(true)
    expect(shouldShowRlsCheckExpression('delete')).toBe(false)
  })

  it('shows check only for create', () => {
    expect(shouldShowRlsUsingExpression('create')).toBe(false)
    expect(shouldShowRlsCheckExpression('create')).toBe(true)
  })

  it('shows both expressions for update', () => {
    expect(shouldShowRlsUsingExpression('update')).toBe(true)
    expect(shouldShowRlsCheckExpression('update')).toBe(true)
  })
})
