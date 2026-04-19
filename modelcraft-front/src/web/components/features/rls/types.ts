import type {
  RLSPreset,
  RLSExprType,
  JsonExpr,
  JsonValue,
  RLSScalarOperator,
  ValidationResult,
  ConditionRow,
  ModelRLSPolicy,
  AuthVariable,
} from '@/types/rls'

/**
 * RLS 功能组件 Props 类型定义
 *
 * 基于架构方案: plans/rls/front-architecture.md
 */

// ============================================
// RLSPresetSelector
// ============================================

/**
 * RLS Preset 选择器 Props
 */
export interface RLSPresetSelectorProps {
  /** 当前选中的 Preset */
  value: RLSPreset | null

  /**
   * 选择回调
   * @param preset - 选中的 Preset
   * @param confirmed - 是否已确认（高危策略需二次确认）
   */
  onChange: (preset: RLSPreset, confirmed: boolean) => void

  /** 是否禁用 */
  disabled?: boolean
}

// ============================================
// PolicyConditionBuilder
// ============================================

/**
 * 字段信息（用于条件构建器）
 */
export interface PolicyFieldInfo {
  /** 字段名 */
  name: string

  /** 字段格式类型 */
  format: string
}

/**
 * Policy 条件构建器 Props
 */
export interface PolicyConditionBuilderProps {
  /** 模型 ID（用于字段白名单校验） */
  modelId: string

  /** 表达式类型 */
  exprType: RLSExprType

  /** 当前表达式值 */
  value: JsonExpr

  /** 模型字段列表 */
  fields: PolicyFieldInfo[]

  /** 已声明的 auth 变量列表（含内置 uid） */
  authVariables: string[]

  /** 值变更回调 */
  onChange: (value: JsonExpr) => void

  /** 校验结果回调（用于实时显示校验状态） */
  onValidationResult?: (result: ValidationResult) => void
}

// ============================================
// PolicyConditionRow
// ============================================

/**
 * 单行条件 Props
 */
export interface PolicyConditionRowProps {
  /** 条件数据 */
  condition: ConditionRow

  /** 可用字段列表 */
  fields: PolicyFieldInfo[]

  /** 可用 auth 变量 */
  authVariables: string[]

  /** 是否禁用 */
  disabled?: boolean

  /** 变更回调 */
  onChange: (condition: ConditionRow) => void

  /** 删除回调 */
  onRemove: () => void
}

// ============================================
// PolicyJSONPreview
// ============================================

/**
 * Policy JSON 预览 Props
 */
export interface PolicyJSONPreviewProps {
  /** 表达式值 */
  value: JsonExpr

  /** 自定义类名 */
  className?: string
}

// ============================================
// RLSPolicyPanel
// ============================================

/**
 * RLS Policy 面板 Props
 * （Model 详情页"访问控制"Tab 内容）
 */
export interface RLSPolicyPanelProps {
  /** 模型 ID */
  modelId: string

  /** 模型名称（用于显示） */
  modelName: string

  /** 当前 Policy（无 owner 字段时为 null） */
  policy: ModelRLSPolicy | null | undefined

  /** 模型字段列表 */
  fields: PolicyFieldInfo[]

  /** Policy 更新成功回调 */
  onPolicyUpdated?: () => void
}

// ============================================
// AuthVariableEditor
// ============================================

/**
 * 认证变量编辑器 Props
 */
export interface AuthVariableEditorProps {
  /** 变量列表 */
  variables: AuthVariable[]

  /** 变更回调 */
  onChange: (variables: AuthVariable[]) => void

  /** 是否禁用 */
  disabled?: boolean
}

// ============================================
// AuthVariableRow
// ============================================

/**
 * 单行认证变量 Props
 */
export interface AuthVariableRowProps {
  /** 变量数据 */
  variable: AuthVariable

  /** 变更回调（内置变量不触发） */
  onChange?: (variable: AuthVariable) => void

  /** 删除回调（内置变量不触发） */
  onRemove?: () => void

  /** 是否禁用（内置变量为 true） */
  disabled?: boolean
}

// ============================================
// AuthSchemaSection
// ============================================

/**
 * 认证变量配置区 Props
 * （Project 设置页使用）
 */
export interface AuthSchemaSectionProps {
  /** 组织名称 */
  orgName: string

  /** 项目 slug */
  projectSlug: string
}

// ============================================
// DangerConfirmDialog
// ============================================

/**
 * 高危操作二次确认弹窗 Props
 */
export interface DangerConfirmDialogProps {
  /** 是否打开 */
  open: boolean

  /** 打开状态变更回调 */
  onOpenChange: (open: boolean) => void

  /** 弹窗标题 */
  title: string

  /** 弹窗描述 */
  description: string

  /** 确认按钮文字 */
  confirmText?: string

  /** 取消按钮文字 */
  cancelText?: string

  /** 确认回调 */
  onConfirm: () => void
}

// ============================================
// Utility Types
// ============================================

/**
 * 逻辑操作符类型
 */
export type LogicalOperator = 'AND' | 'OR'

/**
 * 值类型选项
 */
export type ValueType = 'literal' | 'auth' | 'ref'

/**
 * 条件值结构
 */
export interface ConditionValue {
  /** 值类型 */
  type: ValueType

  /** 值内容（根据类型不同含义不同） */
  value: JsonValue
}
