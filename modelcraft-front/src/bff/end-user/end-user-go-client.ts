// src/bff/end-user/end-user-go-client.ts
// 终端用户 Go Backend Client（对称 bff/auth/go-auth-client.ts）
// Mock 策略：所有调用在内部判断 NEXT_PUBLIC_API_MOCKING=enabled，走 mock 分支

import type { GoEndUserError } from '@/types/end-user-auth'

const GO_BACKEND_INTERNAL_URL =
  process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'

// 判断是否启用 mock（审计修正：mock 策略在此文件内部，不在 route handler 内联）
const USE_MOCK = process.env.NEXT_PUBLIC_API_MOCKING === 'enabled'

// ============================================================================
// 结果类型（BFF 内部使用）
// ============================================================================

export interface EndUserLoginResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserRegisterResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserRefreshResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface EndUserMeResult {
  id: string
  username: string
  createdAt: string
}

// ============================================================================
// 错误类型
// ============================================================================

/** 凭证错误（用户名/密码错误，或用户不存在） */
export class EndUserInvalidCredentialsError extends Error {
  constructor(message = '用户名或密码错误') {
    super(message)
    this.name = 'EndUserInvalidCredentialsError'
  }
}

/** 账号已被禁用 */
export class EndUserAccountDisabledError extends Error {
  constructor(message = '该账号已被禁用') {
    super(message)
    this.name = 'EndUserAccountDisabledError'
  }
}

/** 用户名冲突（注册/创建时） */
export class EndUserConflictError extends Error {
  constructor(message = '该用户名已被使用') {
    super(message)
    this.name = 'EndUserConflictError'
  }
}

/** Token 无效或已过期（含 reuse 攻击检测） */
export class EndUserTokenError extends Error {
  constructor(message = 'Token 无效或已过期') {
    super(message)
    this.name = 'EndUserTokenError'
  }
}

/** Project 未关联 DatabaseCluster */
export class EndUserClusterNotConfiguredError extends Error {
  constructor(message = '服务暂时不可用') {
    super(message)
    this.name = 'EndUserClusterNotConfiguredError'
  }
}

/** 参数校验失败 */
export class EndUserParamInvalidError extends Error {
  constructor(message = '参数校验失败') {
    super(message)
    this.name = 'EndUserParamInvalidError'
  }
}

// ============================================================================
// 内部工具函数
// ============================================================================

/**
 * 解析 Go 后端错误响应
 */
async function parseGoError(res: Response): Promise<Error> {
  try {
    const data = (await res.json()) as GoEndUserError
    const { code, message } = data.error

    switch (code) {
      case 'INVALID_CREDENTIALS':
        return new EndUserInvalidCredentialsError(message)
      case 'ACCOUNT_DISABLED':
        return new EndUserAccountDisabledError(message)
      case 'CONFLICT':
        return new EndUserConflictError(message)
      case 'INVALID_REFRESH_TOKEN':
        return new EndUserTokenError(message)
      case 'CLUSTER_NOT_CONFIGURED':
        return new EndUserClusterNotConfiguredError(message)
      case 'PARAM_INVALID':
        return new EndUserParamInvalidError(message)
      default:
        return new Error(message || `Go backend error: ${res.status}`)
    }
  } catch {
    return new Error(`Go backend error: ${res.status}`)
  }
}

// ============================================================================
// Mock 数据场景
// ============================================================================

// Mock 用户数据库
const MOCK_USERS: Record<string, { userId: string; password: string; username: string; isForbidden: boolean; createdAt: string }> = {
  alice: {
    userId: 'mock-user-001',
    password: 'password123',
    username: 'alice',
    isForbidden: false,
    createdAt: '2026-04-10T08:00:00Z',
  },
  disabled: {
    userId: 'mock-user-disabled',
    password: 'password123',
    username: 'disabled',
    isForbidden: true,
    createdAt: '2026-04-05T08:00:00Z',
  },
}

// Mock refresh token → userId 映射
const MOCK_REFRESH_TOKENS: Record<string, string> = {
  'mock-refresh-token-alice': 'mock-user-001',
  'mock-refresh-token-disabled': 'mock-user-disabled',
}

// ============================================================================
// API 调用函数（内部判断 mock/真实）
// ============================================================================

/**
 * 调用 Go Backend /internal/end-user/auth/login
 */
export async function callGoEndUserLogin(params: {
  orgName: string
  projectSlug: string
  username: string
  password: string
}): Promise<EndUserLoginResult> {
  if (USE_MOCK) {
    // Mock 逻辑
    const user = MOCK_USERS[params.username]

    if (!user || user.password !== params.password) {
      throw new EndUserInvalidCredentialsError()
    }

    if (user.isForbidden) {
      throw new EndUserAccountDisabledError()
    }

    return {
      userId: user.userId,
      refreshToken: `mock-refresh-token-${params.username}`,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    }
  }

  // 真实后端调用（暂返回 NOT_IMPLEMENTED）
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': process.env.INTERNAL_SERVICE_TOKEN ?? '',
    },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  return res.json() as Promise<EndUserLoginResult>
}

/**
 * 调用 Go Backend /internal/end-user/auth/register
 */
export async function callGoEndUserRegister(params: {
  orgName: string
  projectSlug: string
  username: string
  password: string
}): Promise<EndUserRegisterResult> {
  if (USE_MOCK) {
    // Mock 逻辑
    if (MOCK_USERS[params.username]) {
      throw new EndUserConflictError()
    }

    // 密码强度检查（mock）
    if (params.username === 'weak' || params.password.length < 6) {
      throw new EndUserParamInvalidError('密码强度不足')
    }

    // 创建新用户
    const newUserId = `mock-user-${Date.now()}`
    MOCK_USERS[params.username] = {
      userId: newUserId,
      password: params.password,
      username: params.username,
      isForbidden: false,
      createdAt: new Date().toISOString(),
    }
    MOCK_REFRESH_TOKENS[`mock-refresh-token-${params.username}`] = newUserId

    return {
      userId: newUserId,
      refreshToken: `mock-refresh-token-${params.username}`,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    }
  }

  // 真实后端调用
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': process.env.INTERNAL_SERVICE_TOKEN ?? '',
    },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  return res.json() as Promise<EndUserRegisterResult>
}

/**
 * 调用 Go Backend /internal/end-user/auth/refresh (token rotation)
 */
export async function callGoEndUserRefresh(params: {
  orgName: string
  projectSlug: string
  refreshToken: string
}): Promise<EndUserRefreshResult> {
  if (USE_MOCK) {
    // Mock 逻辑
    const userId = MOCK_REFRESH_TOKENS[params.refreshToken]

    if (!userId) {
      throw new EndUserTokenError('无效的 refresh token')
    }

    // 检查用户是否被禁用
    const user = Object.values(MOCK_USERS).find((u) => u.userId === userId)
    if (user?.isForbidden) {
      throw new EndUserAccountDisabledError()
    }

    // Token rotation：生成新 token
    const newRefreshToken = `mock-refresh-token-rotated-${Date.now()}`
    MOCK_REFRESH_TOKENS[newRefreshToken] = userId
    // 保留旧 token 一段时间（mock 中简化处理，不删除旧 token）

    return {
      userId,
      refreshToken: newRefreshToken,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    }
  }

  // 真实后端调用
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': process.env.INTERNAL_SERVICE_TOKEN ?? '',
    },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  return res.json() as Promise<EndUserRefreshResult>
}

/**
 * 调用 Go Backend /internal/end-user/auth/logout
 * Best-effort，catch 静默失败
 */
export async function callGoEndUserLogout(params: {
  orgName: string
  projectSlug: string
  refreshToken: string
}): Promise<void> {
  if (USE_MOCK) {
    // Mock 逻辑：删除 refresh token
    delete MOCK_REFRESH_TOKENS[params.refreshToken]
    return
  }

  // 真实后端调用（best-effort）
  await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/auth/logout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': process.env.INTERNAL_SERVICE_TOKEN ?? '',
    },
    body: JSON.stringify(params),
  }).catch(() => {
    // Ignore errors - Cookie will be cleared regardless
  })
}

/**
 * 调用 Go Backend /internal/end-user/auth/me
 * BFF 已验证 JWT，传 userId/orgName/projectSlug 给 Go（X-End-User-Id 等 Header）
 */
export async function callGoEndUserMe(params: {
  orgName: string
  projectSlug: string
  userId: string
}): Promise<EndUserMeResult> {
  if (USE_MOCK) {
    // Mock 逻辑：根据 userId 查找用户
    const user = Object.values(MOCK_USERS).find((u) => u.userId === params.userId)

    if (!user) {
      throw new EndUserInvalidCredentialsError('用户不存在')
    }

    if (user.isForbidden) {
      throw new EndUserAccountDisabledError()
    }

    return {
      id: user.userId,
      username: user.username,
      createdAt: user.createdAt,
    }
  }

  // 真实后端调用
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/auth/me`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': process.env.INTERNAL_SERVICE_TOKEN ?? '',
      'X-End-User-Id': params.userId,
      'X-Org-Name': params.orgName,
      'X-Project-Slug': params.projectSlug,
    },
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  return res.json() as Promise<EndUserMeResult>
}
