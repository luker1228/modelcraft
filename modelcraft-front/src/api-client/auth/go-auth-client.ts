// src/bff/auth/go-auth-client.ts
import type {
  IdentifierType,
  GoLoginResponse,
  GoRegisterResponse,
  GoAuthError,
  RegisterProfileSnapshot,
} from '@/types/auth'

const GO_BACKEND_INTERNAL_URL =
  process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'

// ============================================================================
// 结果类型（BFF 内部使用）
// ============================================================================

export interface GoLoginResult {
  userId: string
  userName: string
  orgName: string
  refreshToken: string
  expiresAt: string
}

export interface GoRegisterResult {
  userId: string
  orgName: string
  profile?: RegisterProfileSnapshot
}

export interface GoRefreshResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

// ============================================================================
// 错误类型
// ============================================================================

/** 认证失败错误（用户不存在/密码错误） */
export class AuthenticationError extends Error {
  constructor(
    message: string,
    public readonly code: string = 'AUTHENTICATION_FAILED'
  ) {
    super(message)
    this.name = 'AuthenticationError'
  }
}

/** 参数校验错误 */
export class AuthParamInvalidError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'AuthParamInvalidError'
  }
}

/** 冲突错误（手机号/用户名已存在） */
export class UserConflictError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'UserConflictError'
  }
}

/** Token 重用错误 */
export class TokenReuseError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'TokenReuseError'
  }
}

// ============================================================================
// API 调用函数
// ============================================================================

/**
 * 解析 Go 后端错误响应
 */
async function parseGoError(res: Response): Promise<Error> {
  try {
    const data = (await res.json()) as GoAuthError
    const { code, message } = data.error

    if (code === 'PARAM_INVALID.AUTH') {
      return new AuthParamInvalidError(message)
    }
    if (code === 'CONFLICT.USER') {
      return new UserConflictError(message)
    }
    if (code === 'AUTHENTICATION_FAILED') {
      return new AuthenticationError(message, code)
    }
    return new Error(message || `Go backend error: ${res.status}`)
  } catch {
    return new Error(`Go backend error: ${res.status}`)
  }
}

/**
 * 调用 Go Backend /api/auth/login
 * @param params identifier + identifierType + password
 */
export async function callGoLogin(params: {
  identifier: string
  identifierType: IdentifierType
  password: string
}): Promise<GoLoginResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  const data = (await res.json()) as GoLoginResponse
  return {
    userId: data.userId,
    userName: data.userName,
    orgName: data.orgName,
    refreshToken: data.refreshToken,
    expiresAt: data.expiresAt,
  }
}

/**
 * 调用 Go Backend /api/auth/register
 * @param params phone + userName + password
 */
export async function callGoRegister(params: {
  phone: string
  userName: string
  password: string
}): Promise<GoRegisterResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res)
  }

  const data = (await res.json()) as GoRegisterResponse

  const profile = data.profile
    ? {
        id: data.profile.id,
        userId: data.profile.userId,
        nickname: data.profile.nickname,
        avatarUrl: data.profile.avatarUrl,
        bio: data.profile.bio,
      }
    : undefined

  return {
    userId: data.userId,
    orgName: data.orgName,
    profile,
  }
}

/**
 * 调用 Go Backend /api/auth/refresh (token rotation)
 */
export async function callGoRefresh(
  refreshToken: string
): Promise<GoRefreshResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  })

  if (res.status === 401) {
    throw new TokenReuseError('Refresh token revoked or reused')
  }
  if (!res.ok) {
    throw await parseGoError(res)
  }

  return res.json() as Promise<GoRefreshResult>
}

/**
 * 调用 Go Backend /api/auth/logout
 */
export async function callGoLogout(refreshToken: string): Promise<void> {
  await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/logout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  }).catch(() => {
    // Ignore errors - Cookie will be cleared regardless
  })
}
