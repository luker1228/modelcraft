/**
 * RLS (Row Level Security) 领域类型定义
 *
 * 基于 API Contract: plans/rls/api-contract.md
 */

// ============================================
// Enums
// ============================================

/**
 * RLS 预设策略枚举
 */
export type RLSPreset =
  | 'READ_WRITE_OWNER'
  | 'READ_ALL_WRITE_OWNER'
  | 'READ_ALL'
  | 'READ_WRITE_ALL'
  | 'NO_ACCESS'

/**
 * RLS 表达式类型
 */
export type RLSExprType =
  | 'SELECT_PREDICATE'
  | 'INSERT_CHECK'
  | 'UPDATE_PREDICATE'
  | 'UPDATE_CHECK'
  | 'DELETE_PREDICATE'

/**
 * Auth Schema 变量类型
 */
export type AuthVariableType = 'UUID' | 'STRING' | 'INTEGER'

/**
 * 标量操作符
 */
export type RLSScalarOperator =
  | '_eq'
  | '_neq'
  | '_gt'
  | '_gte'
  | '_lt'
  | '_lte'
  | '_in'
  | '_nin'
  | '_is_null'

/**
 * 逻辑操作符
 */
export type RLSLogicalOperator = '_and' | '_or' | '_not'

// ============================================
// JSON 表达式类型
// ============================================

/**
 * RLS JSON 值类型（递归类型）
 */
export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonExpr
  | JsonValue[]

/**
 * RLS JSON 表达式
 */
export interface JsonExpr {
  [key: string]: JsonValue
}

// ============================================
// Domain Types
// ============================================

/**
 * Model RLS 策略（五件套）
 */
export interface ModelRLSPolicy {
  /** 所属模型 ID */
  modelId: string

  /** SELECT USING 谓词（JSON 字符串） */
  selectPredicate: string

  /** INSERT WITH CHECK 谓词（JSON 字符串） */
  insertCheck: string

  /** UPDATE USING 谓词（JSON 字符串） */
  updatePredicate: string

  /** UPDATE WITH CHECK 谓词（JSON 字符串） */
  updateCheck: string

  /** DELETE USING 谓词（JSON 字符串） */
  deletePredicate: string

  /** 当前策略匹配的 Preset，自定义组合时返回 null */
  preset: RLSPreset | null

  /** 创建时间 */
  createdAt: string

  /** 更新时间 */
  updatedAt: string
}

/**
 * 认证变量定义
 */
export interface AuthVariable {
  /** 变量名（如 "tenant_id"） */
  name: string

  /** JWT 来源路径（如 "jwt.tenant_id"） */
  source: string

  /** 变量类型 */
  type: AuthVariableType

  /** 是否为内置变量（如 uid） */
  isBuiltin?: boolean
}

/**
 * Project 认证变量配置
 */
export interface ProjectAuthSchema {
  /** 认证变量列表（不含内置 uid） */
  variables: AuthVariable[]
}

// ============================================
// Validation Types
// ============================================

/**
 * 校验错误详情
 */
export interface ValidationError {
  /** 错误位置（如 "selectPredicate"、"insertCheck._exists"） */
  path: string

  /** 错误描述 */
  message: string

  /** 错误码 */
  code: string
}

/**
 * RLS 表达式校验结果
 */
export interface ValidationResult {
  /** 是否校验通过 */
  valid: boolean

  /** 校验错误信息列表（valid=false 时返回） */
  errors?: ValidationError[]
}

// ============================================
// Input Types
// ============================================

/**
 * 设置 Model RLS 策略输入
 */
export interface SetModelRLSPolicyInput {
  /** 模型 ID */
  modelId: string

  /** SELECT USING 谓词（JSON 字符串） */
  selectPredicate: string

  /** INSERT WITH CHECK 谓词（JSON 字符串） */
  insertCheck: string

  /** UPDATE USING 谓词（JSON 字符串） */
  updatePredicate: string

  /** UPDATE WITH CHECK 谓词（JSON 字符串） */
  updateCheck: string

  /** DELETE USING 谓词（JSON 字符串） */
  deletePredicate: string
}

/**
 * 校验 RLS 表达式输入
 */
export interface ValidateRLSExprInput {
  /** 所属模型 ID（用于字段名白名单校验） */
  modelId: string

  /** 要校验的谓词类型 */
  exprType: RLSExprType

  /** 表达式 JSON 字符串 */
  expression: string
}

/**
 * Auth 变量输入
 */
export interface AuthVariableInput {
  /** 变量名 */
  name: string

  /** JWT 来源路径 */
  source: string

  /** 变量类型 */
  type: AuthVariableType
}

/**
 * 设置 Project 认证变量配置输入
 */
export interface SetProjectAuthSchemaInput {
  /** 项目 slug */
  projectSlug: string

  /** 认证变量列表（uid 内置，无需声明） */
  variables: AuthVariableInput[]
}

// ============================================
// Component State Types
// ============================================

/**
 * 条件行数据结构（用于 PolicyConditionBuilder）
 */
export interface ConditionRow {
  /** 唯一标识 */
  id: string

  /** 字段名或 _auth / _ref / _exists */
  field: string

  /** 操作符 */
  operator: RLSScalarOperator

  /** 值或 {_auth: "uid"} / {_ref: "..."} */
  value: JsonValue
}

/**
 * Preset 配置项
 */
export interface RLSPresetConfig {
  /** Preset 值 */
  value: RLSPreset

  /** 显示标签 */
  label: string

  /** 描述 */
  description: string

  /** 适用场景 */
  scenario: string

  /** 是否为高危策略 */
  isDangerous?: boolean
}
