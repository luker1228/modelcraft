// src/types/end-user-auth.ts
// 终端用户认证相关类型定义（对称 auth.ts，不污染开发者类型）

// ============================================================
// BFF 请求/响应类型（前端 ↔ BFF Route Handler）
// ============================================================

/** BFF 终端用户登录请求 */
export interface EndUserLoginRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

/** BFF 终端用户注册请求 */
export interface EndUserRegisterRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

/** BFF 登录/注册统一响应格式 */
export interface EndUserAuthResponse {
  accessToken: string // BFF 自签 end-user JWT（1h）
  expiresIn: number // 3600
}

/** BFF /me 接口响应 */
export interface EndUserMeResponse {
  id: string
  username: string
  createdAt: string // ISO 8601
}

/** BFF 错误响应（统一格式） */
export interface EndUserBffError {
  error: {
    code: EndUserErrorCode
    message: string
  }
  requestId?: string
}

// ============================================================
// Go Backend 内网请求/响应类型（BFF 内部使用）
// ============================================================

/** Go 后端登录响应 */
export interface GoEndUserLoginResponse {
  userId: string
  refreshToken: string
  expiresAt: string // ISO 8601
}

/** Go 后端 refresh 响应 */
export interface GoEndUserRefreshResponse {
  userId: string
  refreshToken: string
  expiresAt: string
}

/** Go 后端 /me 响应 */
export interface GoEndUserMeResponse {
  id: string
  username: string
  createdAt: string
}

// ============================================================
// Go Backend 错误结构（BFF 内部解析用）
// ============================================================

export interface GoEndUserError {
  error: {
    code: string
    message: string
  }
}

// ============================================================
// 错误码
// ============================================================

export type EndUserErrorCode =
  | 'INVALID_CREDENTIALS' // 用户名/密码错误（防枚举，统一返回）
  | 'ACCOUNT_DISABLED' // 账号已禁用
  | 'CONFLICT' // 用户名已存在
  | 'INVALID_REFRESH_TOKEN' // refresh token 无效/过期/已 revoke
  | 'CLUSTER_NOT_CONFIGURED' // Project 未关联 Cluster
  | 'PRIVATE_DB_NOT_INITIALIZED' // Private DB 缺失，需用户确认初始化
  | 'PARAM_INVALID' // 参数校验失败
  | 'UNAUTHORIZED' // JWT 缺失或无效
  | 'NOT_IMPLEMENTED' // 真实后端分支未实现

/**
 * 错误码 → 用户提示文案
 * @param code - Go 后端返回的错误码
 * @param httpStatus - HTTP 状态码
 * @param backendMessage - 后端原始错误文案（可选）
 * @returns 用户可读的错误提示
 */
export function mapEndUserErrorCode(
  code: string | undefined,
  httpStatus: number,
  backendMessage?: string
): string {
  const msg = backendMessage?.toLowerCase() ?? ''

  if (code === 'INVALID_CREDENTIALS') return '用户名或密码错误，请重试'
  if (code === 'ACCOUNT_DISABLED') return '该账号已被禁用，请联系管理员'
  if (code === 'CONFLICT') return '该用户名已被使用'
  if (code === 'CLUSTER_NOT_CONFIGURED') return '服务暂时不可用，请联系管理员'
  if (code === 'PRIVATE_DB_NOT_INITIALIZED') return '私有库尚未初始化，请先确认初始化'
  if (code === 'PARAM_INVALID') {
    if (msg.includes('password must be at least 8 characters')) {
      return '密码至少 8 位'
    }
    if (msg.includes('password must contain at least one letter')) {
      return '密码必须包含至少 1 个字母'
    }
    if (msg.includes('password must contain at least one digit')) {
      return '密码必须包含至少 1 个数字'
    }
    if (msg.includes('username must be 3-64 characters')) {
      return '用户名需为 3-64 位，且仅支持字母、数字、下划线和中划线'
    }
    return '输入参数无效，请检查后重试'
  }
  if (code === 'INVALID_REFRESH_TOKEN') return '登录已过期，请重新登录'
  if (code === 'UNAUTHORIZED') return '服务认证失败，请联系管理员检查 INTERNAL_TOKEN 配置'
  if (httpStatus >= 500) return '登录服务暂时不可用，请稍后重试'
  return '登录服务暂时不可用，请稍后重试'
}

// ============================================================
// JWT Payload 类型（客户端 decode 用，仅供展示）
// ============================================================

export interface EndUserJWTPayload {
  sub: string // userId
  org_name: string
  project_slug: string
  role: 'end_user'
  exp?: number
  iat?: number
}

// ============================================================
// Store 类型
// ============================================================

/** JWT payload 中解析出来的终端用户信息（部分字段来自 /me 接口） */
export interface EndUserInfo {
  id: string // userId（JWT sub claim）
  username: string // 从 /me 接口获取后填充（JWT 中无此字段）
  orgName: string // 来自 JWT org_name claim
  projectSlug: string // 来自 JWT project_slug claim
}

// ============================================================
// EndUser v2 — Org 级登录类型
// ============================================================

/** v2 可访问的 Project */
export interface EndUserAccessibleProject {
  slug: string
  title: string
}

/**
 * v2 BFF 登录响应（三种结果）
 * - singleProject: true  → 直接跳转，accessToken 已签发
 * - singleProject: false → 进入 select-project 页面
 * - error                → 登录失败（凭证错误 / 无项目权限 / 账号禁用）
 */
export type EndUserLoginResponseV2 =
  | {
      singleProject: true
      projectSlug: string
      accessToken: string
    }
  | {
      singleProject: false
      projects: EndUserAccessibleProject[]
    }
  | {
      error: {
        code: 'NO_PROJECT_ACCESS' | 'INVALID_CREDENTIALS' | 'ACCOUNT_DISABLED' | 'PARAM_INVALID'
        message: string
      }
    }

/** v2 BFF select-project 请求 */
export interface EndUserSelectProjectRequest {
  projectSlug: string
}

/** v2 BFF select-project 响应 */
export interface EndUserSelectProjectResponse {
  accessToken: string
  projectSlug: string
}

/** v2 BFF select-project 错误 */
export interface EndUserSelectProjectError {
  error: {
    code: 'PROJECT_ACCESS_DENIED' | 'PENDING_SESSION_INVALID' | 'PARAM_INVALID'
    message: string
  }
}

/** v2 pending session JWT payload（临时，不含 projectSlug） */
export interface EndUserPendingSessionPayload {
  userId: string
  orgName: string
  projects: EndUserAccessibleProject[]
}
