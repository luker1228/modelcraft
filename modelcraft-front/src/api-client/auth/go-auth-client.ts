// src/api-client/auth/go-auth-client.ts
// 调用 modelcraft-gateway 的 /auth/* 端点
// Gateway 负责：认证、JWT 签发、httpOnly Cookie 管理
import type {
  GoAuthError,
} from '@/types/auth'

const GATEWAY_URL = ''

// ============================================================================
// 结果类型
// ============================================================================

export interface GoLoginResult {
  accessToken: string
}

export interface GoRegisterResult {
  accessToken: string
}

export interface GoRefreshResult {
  accessToken: string
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
// 错误解析
// ============================================================================

async function parseGatewayError(res: Response): Promise<Error> {
  try {
    const data = (await res.json()) as { code?: string; message?: string } | GoAuthError
    const code = 'error' in data ? data.error.code : (data as { code?: string }).code ?? ''
    const message =
      'error' in data ? data.error.message : (data as { message?: string }).message ?? `Gateway error: ${res.status}`

    if (code === 'PARAM_INVALID' || code === 'PARAM_INVALID.AUTH') {
      return new AuthParamInvalidError(message)
    }
    if (code === 'CONFLICT.USER') {
      return new UserConflictError(message)
    }
    if (code === 'AUTHENTICATION_FAILED' || code === 'AUTH_FAILED') {
      return new AuthenticationError(message, code)
    }
    return new Error(message)
  } catch {
    return new Error(`Gateway error: ${res.status}`)
  }
}

// ============================================================================
// API 调用函数
// ============================================================================

/**
 * 调用 Gateway POST /auth/login
 * Gateway 处理：验证凭证（→ Backend）、签发 access JWT、设 httpOnly Cookie
 */
export async function callGoLogin(params: {
  phone: string
  password: string
}): Promise<GoLoginResult> {
  const res = await fetch(`${GATEWAY_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGatewayError(res)
  }

  return res.json() as Promise<GoLoginResult>
}

/**
 * 调用 Gateway POST /auth/register
 */
export async function callGoRegister(params: {
  phone: string
  userName: string
  orgDisplayName: string
  organizationName?: string
  password: string
}): Promise<GoRegisterResult> {
  const res = await fetch(`${GATEWAY_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGatewayError(res)
  }

  return res.json() as Promise<GoRegisterResult>
}

/**
 * 调用 Gateway POST /auth/refresh
 * refresh token 通过 httpOnly Cookie 自动携带，无需手动传入
 */
export async function callGoRefresh(): Promise<GoRefreshResult> {
  const res = await fetch(`${GATEWAY_URL}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  })

  if (res.status === 401) {
    throw new TokenReuseError('Refresh token revoked or reused')
  }
  if (!res.ok) {
    throw await parseGatewayError(res)
  }

  return res.json() as Promise<GoRefreshResult>
}

/**
 * 调用 Gateway POST /auth/logout
 * Gateway 负责告知后端撤销 token 并清除 Cookie
 */
export async function callGoLogout(): Promise<void> {
  await fetch(`${GATEWAY_URL}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  }).catch(() => {
    // Ignore errors — UI logout proceeds regardless
  })
}
