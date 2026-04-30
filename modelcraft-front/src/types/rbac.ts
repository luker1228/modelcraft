// EndUser RBAC 类型定义
// 对齐 prd/rbac/ 设计文档与 api-contract

// ============================================================================
// 列访问模式与列策略
// ============================================================================

export type ColumnAccessMode = 'VISIBLE' | 'READONLY' | 'MASKED' | 'HIDDEN'

export interface ColumnRule {
  fieldName: string
  mode: ColumnAccessMode
  maskPattern?: string
}

export interface ColumnPolicy {
  defaultMode: ColumnAccessMode
  rules: ColumnRule[]
}

// ============================================================================
// 权限动作与行策略枚举
// ============================================================================

export type EndUserPermissionAction = 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE' | 'EXPORT'
export type EndUserRowScope = 'ALL' | 'SELF' | 'DEPT' | 'DEPT_AND_CHILDREN'

/** 预设策略枚举 */
export type EndUserPermissionPreset =
  | 'READ_WRITE_ALL'
  | 'READ_ALL'
  | 'READ_WRITE_OWNER'
  | 'READ_ALL_WRITE_OWNER'

/** Bundle item 来源类型 */
export type DataPermissionGrantType = 'PRESET' | 'CUSTOM'

// ============================================================================
// 核心实体类型
// ============================================================================

/**
 * 终端用户权限点：针对某个 Model 的单一操作授权
 */
export interface EndUserPermission {
  id: string
  modelId: string
  action: EndUserPermissionAction
  rowScope: EndUserRowScope
  columnPolicy: ColumnPolicy
  displayName?: string
  modelDisplayName?: string
  description?: string
  /** 若为预设策略实例化产生，则记录对应的预设类型；否则为 null/undefined */
  preset?: string | null
  createdAt: string
  updatedAt: string
}

/**
 * Bundle 内的数据权限配置项（item-centric）。
 * 同一 bundle 下同一 modelId 最多一个 item。
 */
export interface EndUserBundleDataPermissionItem {
  id: string
  bundleId: string
  modelId: string
  /** 来源类型：PRESET 或 CUSTOM */
  grantType: DataPermissionGrantType
  /** grantType=PRESET 时非空 */
  preset?: EndUserPermissionPreset | null
  /** grantType=CUSTOM 时非空 */
  customPermissionId?: string | null
  /** grantType=CUSTOM 时，引用的自定义权限点摘要 */
  customPermission?: EndUserPermission | null
  sortOrder: number
  createdAt: string
  updatedAt: string
}

/**
 * 快照中的 data permission item 条目（item-centric）
 */
export interface EndUserPermissionSnapshotItemEntry {
  modelId: string
  grantType: DataPermissionGrantType
  preset?: EndUserPermissionPreset | null
  customPermissionId?: string | null
  sortOrder: number
}

/**
 * 权限包历史快照中的权限点条目（兼容旧格式）
 */
export interface EndUserPermissionBundleSnapshotEntry {
  sortOrder: number
  permissionId: string
}

/**
 * 权限包历史快照
 */
export interface EndUserPermissionBundleSnapshot {
  version: number
  createdAt: string
  createdBy?: string | null
  restoredFrom?: number | null
  /** Item-centric 快照条目（新格式） */
  items: EndUserPermissionSnapshotItemEntry[]
  /** 兼容旧格式 */
  permissions: EndUserPermissionBundleSnapshotEntry[]
}

/**
 * 终端用户权限包：一组 data permission item 的集合，可授予角色或直接授予用户
 */
export interface EndUserPermissionBundle {
  id: string
  /** URL 友好的对外标识符，同项目内唯一，创建时由用户指定或从名称自动派生，之后不可修改 */
  slug: string
  name: string
  description?: string
  /** Item-centric 数据权限列表（新字段，每个模型最多一个 item） */
  dataPermissionItems: EndUserBundleDataPermissionItem[]
  /** 兼容旧字段（将逐步废弃） */
  permissions: Array<{ sortOrder: number; permission: EndUserPermission }>
  /** 当前版本号（每次权限列表变更后递增） */
  currentVersion: number
  /** 最近历史快照列表（最多 5 个，按 version DESC 排列） */
  snapshots: EndUserPermissionBundleSnapshot[]
  createdAt: string
  updatedAt: string
}

/**
 * 终端用户角色：包含若干权限包，可分配给用户
 */
export interface EndUserRole {
  id: string
  name: string
  description?: string
  /** 是否为系统内置隐式角色（如 SYSTEM_AUTHENTICATED_USER） */
  isImplicit: boolean
  permissionBundles: Array<{ bundle: EndUserPermissionBundle }>
  createdAt: string
  updatedAt: string
}

// ============================================================================
// 用户授权关联
// ============================================================================

/** 直接授予用户的权限包关联 */
export interface EndUserBundleAssignment {
  endUserId: string
  bundle: EndUserPermissionBundle
  assignedAt: string
}

/** 授予用户的角色关联 */
export interface EndUserRoleAssignment {
  endUserId: string
  role: EndUserRole
  assignedAt: string
}

// ============================================================================
// 有效权限（三通道合并结果）
// ============================================================================

/** 单个 Model 上的有效授权条目 */
export interface EffectiveGrant {
  action: EndUserPermissionAction
  columnPolicy: ColumnPolicy
  rowScope: EndUserRowScope
}

/** 某个用户在某个 Model 上的全部有效权限（隐式角色 + 角色包 + 直接包 合并） */
export interface EffectivePermissions {
  endUserId: string
  modelId: string
  grants: EffectiveGrant[]
}
