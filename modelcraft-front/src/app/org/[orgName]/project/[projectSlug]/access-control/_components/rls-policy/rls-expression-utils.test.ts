import { describe, expect, it } from 'vitest'
import {
  getRlsExpressionType,
  validateRlsExpressionSyntax,
} from './rls-expression-utils'

describe('validateRlsExpressionSyntax', () => {
  it('accepts empty optional expressions', () => {
    expect(validateRlsExpressionSyntax('')).toEqual({ valid: true, empty: true })
    expect(validateRlsExpressionSyntax('   ')).toEqual({ valid: true, empty: true })
  })

  it('accepts JSON objects and boolean constants', () => {
    expect(validateRlsExpressionSyntax('{"owner_id":{"equals":"{{user_id}}"}}')).toEqual({
      valid: true,
      empty: false,
    })
    expect(validateRlsExpressionSyntax('true')).toEqual({ valid: true, empty: false })
    expect(validateRlsExpressionSyntax('false')).toEqual({ valid: true, empty: false })
  })

  it('rejects malformed JSON and unsupported JSON values', () => {
    expect(validateRlsExpressionSyntax('{"owner_id":')).toMatchObject({
      valid: false,
      empty: false,
    })
    expect(validateRlsExpressionSyntax('"owner_id"')).toEqual({
      valid: false,
      empty: false,
      message: '表达式必须是 JSON 对象或布尔常量',
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
