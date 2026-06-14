import { describe, expect, it } from 'vitest'
import {
  buildRlsCompletionItems,
  buildRlsExpressionHelp,
  extractRlsCompletionContext,
  getRlsExpressionType,
  getRlsAvailableContexts,
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
    expect(validateRlsExpressionSyntax('row.owner_id == auth.userid')).toEqual({
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
      message: '请输入 CEL 表达式，例如 row.owner_id == auth.userid',
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

describe('buildRlsExpressionHelp', () => {
  it('prefers auth-related fields for examples', () => {
    expect(
      buildRlsExpressionHelp('using', [
        { name: 'id', title: 'ID' },
        { name: 'owner_id', title: 'Owner' },
        { name: 'status', title: 'Status' },
      ]),
    ).toEqual({
      availableFields: ['row.id', 'row.owner_id', 'row.status'],
      example: 'row.owner_id == auth.userid',
      placeholder: '例如：row.owner_id == auth.userid',
      rootLabel: 'row',
    })
  })

  it('switches root for input-check expressions', () => {
    expect(buildRlsExpressionHelp('check', [{ name: 'owner_id', title: 'Owner' }])).toEqual({
      availableFields: ['input.owner_id'],
      example: 'input.owner_id == auth.userid',
      placeholder: '例如：input.owner_id == auth.userid',
      rootLabel: 'input',
    })
  })

  it('falls back to a stable default example when model has no ownership field', () => {
    expect(
      buildRlsExpressionHelp('using', [
        { name: 'id', title: 'ID' },
        { name: 'status', title: 'Status' },
      ]),
    ).toEqual({
      availableFields: ['row.id', 'row.status'],
      example: 'true',
      placeholder: '例如：true',
      rootLabel: 'row',
    })
  })
})

describe('extractRlsCompletionContext', () => {
  it('detects root token before cursor', () => {
    expect(extractRlsCompletionContext('row.own', 'row.own'.length)).toEqual({
      root: 'row',
      query: 'own',
      replaceStart: 0,
      replaceEnd: 'row.own'.length,
    })
  })

  it('returns null when cursor is not after a completion token', () => {
    expect(extractRlsCompletionContext('row.owner_id == auth.userid', 3)).toBeNull()
    expect(extractRlsCompletionContext('status == 1', 'status == 1'.length)).toBeNull()
  })
})

describe('buildRlsCompletionItems', () => {
  it('builds field candidates for active root', () => {
    expect(
      buildRlsCompletionItems({
        context: { root: 'row', query: 'own', replaceStart: 0, replaceEnd: 7 },
        rootLabel: 'row',
        fields: [
          { name: 'owner_id', title: 'Owner' },
          { name: 'status', title: 'Status' },
        ],
        authVariables: [],
      }),
    ).toEqual([
      {
        key: 'row.owner_id',
        value: 'row.owner_id',
        label: 'row.owner_id',
        description: 'Owner',
      },
    ])
  })

  it('returns auth candidates from auth schema', () => {
    expect(
      buildRlsCompletionItems({
        context: { root: 'auth', query: 'user', replaceStart: 0, replaceEnd: 9 },
        rootLabel: 'row',
        fields: [],
        authVariables: [
          { name: 'userid', type: 'string', source: 'X-MC-Auth-Userid' },
          { name: 'username', type: 'string', source: 'X-MC-Auth-Username' },
        ],
      }),
    ).toEqual([
      {
        key: 'auth.userid',
        value: 'auth.userid',
        label: 'auth.userid',
        description: 'string · X-MC-Auth-Userid',
      },
      {
        key: 'auth.username',
        value: 'auth.username',
        label: 'auth.username',
        description: 'string · X-MC-Auth-Username',
      },
    ])
  })
})

describe('getRlsAvailableContexts', () => {
  it('lists active roots for using expressions', () => {
    expect(getRlsAvailableContexts('row')).toEqual([
      {
        key: 'row',
        value: 'row.',
        description: '当前记录行上下文（Using Filter）',
      },
      {
        key: 'auth',
        value: 'auth.',
        description: '当前认证身份上下文',
      },
    ])
  })
})
