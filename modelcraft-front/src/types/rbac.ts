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

// ============================================================================
// 核心实体类型
// ============================================================================

/**
 * 终端用户权限点：针对某个 Model 的单一操作授权
 */
export interface EndUserPermission {
  id: string
  modelId: string
  modelDisplayName: string
  action: EndUserPermissionAction
  rowScope: EndUserRowScope
  columnPolicy: ColumnPolicy
  displayName?: string
  description?: string
  createdAt: string
  updatedAt: string
}

/**
 * 终端用户权限包：一组权限点的集合，可授予角色或直接授予用户
 */
export interface EndUserPermissionBundle {
  id: string
  name: string
  description?: string
  permissions: EndUserPermission[]
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
  permissionBundles: EndUserPermissionBundle[]
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
  grantedAt: string
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
