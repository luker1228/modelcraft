import { faker } from '@faker-js/faker'
import type {
  ModelRLSPolicy,
  AuthVariable,
  ProjectAuthSchema,
  ValidationResult,
  RLSPreset,
} from '@/types/rls'

/**
 * RLS Mock 数据工厂
 *
 * 基于 API Contract: plans/rls/api-contract.md
 */

// ============================================
// Preset JSON 表达式常量
// ============================================

/**
 * 预设策略对应的 JSON 表达式
 */
export const PRESET_JSON_EXPR: Record<
  RLSPreset,
  {
    selectPredicate: string
    insertCheck: string
    updatePredicate: string
    updateCheck: string
    deletePredicate: string
  }
> = {
  READ_WRITE_OWNER: {
    selectPredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    insertCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updatePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updateCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    deletePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
  },
  READ_ALL_WRITE_OWNER: {
    selectPredicate: 'true',
    insertCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updatePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    updateCheck: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
    deletePredicate: JSON.stringify({ owner: { _eq: { _auth: 'uid' } } }),
  },
  READ_ALL: {
    selectPredicate: 'true',
    insertCheck: 'false',
    updatePredicate: 'false',
    updateCheck: 'false',
    deletePredicate: 'false',
  },
  READ_WRITE_ALL: {
    selectPredicate: 'true',
    insertCheck: 'true',
    updatePredicate: 'true',
    updateCheck: 'true',
    deletePredicate: 'true',
  },
  NO_ACCESS: {
    selectPredicate: 'false',
    insertCheck: 'false',
    updatePredicate: 'false',
    updateCheck: 'false',
    deletePredicate: 'false',
  },
}

// ============================================
// Factory Functions
// ============================================

/**
 * 创建 Mock RLS Policy
 *
 * @param modelId - 模型 ID
 * @param preset - 预设策略类型，默认为 READ_WRITE_OWNER
 * @returns ModelRLSPolicy
 */
export function createMockRLSPolicy(
  modelId: string,
  preset: RLSPreset = 'READ_WRITE_OWNER',
): ModelRLSPolicy {
  const expr = PRESET_JSON_EXPR[preset]
  return {
    modelId,
    ...expr,
    preset,
    createdAt: faker.date.past().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
  }
}

/**
 * 创建 Mock Auth Variable
 *
 * @param override - 可选的覆盖属性
 * @returns AuthVariable
 */
export function createMockAuthVariable(
  override?: Partial<AuthVariable>,
): AuthVariable {
  const name = faker.helpers.slugify(faker.word.noun()).toLowerCase().replace(/-/g, '_')
  const source = `jwt.${faker.helpers.slugify(faker.word.noun()).toLowerCase().replace(/-/g, '_')}`
  const type = faker.helpers.arrayElement(['UUID', 'STRING', 'INTEGER']) as AuthVariable['type']

  return {
    name,
    source,
    type,
    ...override,
  }
}

/**
 * 创建内置 uid 变量（只读）
 *
 * @returns AuthVariable
 */
export function createBuiltinUidVariable(): AuthVariable {
  return {
    name: 'uid',
    source: 'jwt.user_id',
    type: 'UUID',
    isBuiltin: true,
  }
}

/**
 * 创建 Mock Project Auth Schema
 *
 * 包含默认的扩展变量（tenant_id, role）
 *
 * @returns ProjectAuthSchema
 */
export function createMockAuthSchema(): ProjectAuthSchema {
  return {
    variables: [
      { name: 'tenant_id', source: 'jwt.tenant_id', type: 'UUID' },
      { name: 'role', source: 'jwt.role', type: 'STRING' },
    ],
  }
}

/**
 * 创建空的 Project Auth Schema
 *
 * @returns ProjectAuthSchema
 */
export function createEmptyAuthSchema(): ProjectAuthSchema {
  return {
    variables: [],
  }
}

/**
 * 创建 Mock Validation Result
 *
 * @param valid - 是否校验通过，默认为 true
 * @returns ValidationResult
 */
export function createMockValidationResult(valid = true): ValidationResult {
  return {
    valid,
    errors: valid
      ? undefined
      : [
          {
            path: 'selectPredicate.owner._eq',
            message: 'Invalid field reference: owner',
            code: 'INVALID_FIELD',
          },
        ],
  }
}

/**
 * 创建包含多个错误的 Validation Result
 *
 * @param errors - 错误列表
 * @returns ValidationResult
 */
export function createValidationResultWithErrors(
  errors: { path: string; message: string; code: string }[],
): ValidationResult {
  return {
    valid: false,
    errors,
  }
}

// ============================================
// Preset Data Helpers
// ============================================

/**
 * 获取指定 Preset 的 JSON 表达式
 *
 * @param preset - 预设策略类型
 * @returns 五件套 JSON 表达式
 */
export function getPresetExpressions(preset: RLSPreset) {
  return PRESET_JSON_EXPR[preset]
}

/**
 * Preset 配置列表（用于 UI 展示）
 */
export interface PresetConfig {
  value: RLSPreset
  label: string
  description: string
  scenario: string
  isDangerous?: boolean
}

export const PRESET_CONFIGS: PresetConfig[] = [
  {
    value: 'READ_WRITE_OWNER',
    label: '读写自己',
    description: '终端用户只能读写归属于自己的数据',
    scenario: '适用于：任务管理、个人笔记、订单系统',
  },
  {
    value: 'READ_ALL_WRITE_OWNER',
    label: '读全部，写自己',
    description: '终端用户可以读取全部数据，但只能修改自己的数据',
    scenario: '适用于：评论、帖子、公开内容',
  },
  {
    value: 'READ_ALL',
    label: '只读全部',
    description: '终端用户只能读取全部数据，不能写入',
    scenario: '适用于：公告、商品目录、配置表',
  },
  {
    value: 'READ_WRITE_ALL',
    label: '读写全部',
    description: '终端用户可以读写任意数据（⚠️ 高危策略）',
    scenario: '适用于：完全开放的场景',
    isDangerous: true,
  },
  {
    value: 'NO_ACCESS',
    label: '无访问权限',
    description: '终端用户无法访问该模型的任何数据',
    scenario: '适用于：系统内部表、敏感数据',
  },
]
