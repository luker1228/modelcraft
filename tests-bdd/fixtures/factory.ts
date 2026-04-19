import { randomUUID } from 'crypto'

/**
 * 生成并发安全的唯一名称，避免多次运行之间冲突。
 * 不使用连字符分隔，因为 model/enum 名称不允许包含连字符。
 * @example uniqueName('User') → 'Usera3f2b1c0'
 */
export const uniqueName = (prefix: string): string =>
  `${prefix}${randomUUID().replace(/-/g, '').slice(0, 8)}`

// ============================================================
// RLS (Row-Level Security) 测试数据生成函数
// ============================================================

/**
 * 生成 RLS 测试用的 JSON 表达式
 */
export const rlsExpr = {
  /**
   * owner 等于当前用户 ID
   * @example ownerEqualsUid() → {"owner":{"_eq":{"_auth":"uid"}}}
   */
  ownerEqualsUid: (): string =>
    JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),

  /**
   * 常量 true
   */
  alwaysTrue: (): string => 'true',

  /**
   * 常量 false
   */
  alwaysFalse: (): string => 'false',

  /**
   * 自定义字段比较表达式
   * @example fieldEquals('status', 'active') → {"status":{"_eq":"active"}}
   */
  fieldEquals: (field: string, value: string | { _auth: string }): string =>
    JSON.stringify({ [field]: { _eq: value } }),

  /**
   * AND 组合表达式
   */
  and: (...conditions: string[]): string =>
    JSON.stringify({ _and: conditions.map(c => JSON.parse(c)) }),

  /**
   * OR 组合表达式
   */
  or: (...conditions: string[]): string =>
    JSON.stringify({ _or: conditions.map(c => JSON.parse(c)) }),
}

/**
 * RLS Policy Preset 配置
 */
export const rlsPresets = {
  READ_WRITE_OWNER: {
    selectPredicate: rlsExpr.ownerEqualsUid(),
    insertCheck: rlsExpr.ownerEqualsUid(),
    updatePredicate: rlsExpr.ownerEqualsUid(),
    updateCheck: rlsExpr.ownerEqualsUid(),
    deletePredicate: rlsExpr.ownerEqualsUid(),
  },
  READ_ALL_WRITE_OWNER: {
    selectPredicate: rlsExpr.alwaysTrue(),
    insertCheck: rlsExpr.ownerEqualsUid(),
    updatePredicate: rlsExpr.ownerEqualsUid(),
    updateCheck: rlsExpr.ownerEqualsUid(),
    deletePredicate: rlsExpr.ownerEqualsUid(),
  },
  READ_ALL: {
    selectPredicate: rlsExpr.alwaysTrue(),
    insertCheck: rlsExpr.alwaysFalse(),
    updatePredicate: rlsExpr.alwaysFalse(),
    updateCheck: rlsExpr.alwaysFalse(),
    deletePredicate: rlsExpr.alwaysFalse(),
  },
  READ_WRITE_ALL: {
    selectPredicate: rlsExpr.alwaysTrue(),
    insertCheck: rlsExpr.alwaysTrue(),
    updatePredicate: rlsExpr.alwaysTrue(),
    updateCheck: rlsExpr.alwaysTrue(),
    deletePredicate: rlsExpr.alwaysTrue(),
  },
  NO_ACCESS: {
    selectPredicate: rlsExpr.alwaysFalse(),
    insertCheck: rlsExpr.alwaysFalse(),
    updatePredicate: rlsExpr.alwaysFalse(),
    updateCheck: rlsExpr.alwaysFalse(),
    deletePredicate: rlsExpr.alwaysFalse(),
  },
}

/**
 * Auth Schema 变量类型
 */
export type AuthVariableType = 'UUID' | 'STRING' | 'INTEGER'

/**
 * 生成 Auth Schema 变量配置
 */
export const createAuthVariable = (
  name: string,
  source: string,
  type: AuthVariableType = 'STRING'
): { name: string; source: string; type: AuthVariableType } => ({
  name,
  source,
  type,
})

/**
 * 生成测试用的 EndUser 用户名
 * @example endUserName('alice') → 'alice_a3f2b1c0'
 */
export const endUserName = (base: string): string =>
  `${base}_${randomUUID().replace(/-/g, '').slice(0, 8)}`

/**
 * 生成测试用的记录名称
 * @example recordName('Order') → 'Order_a3f2b1c0'
 */
export const recordName = (base: string): string =>
  `${base}_${randomUUID().replace(/-/g, '').slice(0, 8)}`
