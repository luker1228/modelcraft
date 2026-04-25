// src/bff/end-user/end-user-go-client.ts
// 终端用户 Go Backend Client（对称 bff/auth/go-auth-client.ts）
// Mock 策略：所有调用在内部判断 NEXT_PUBLIC_API_MOCKING=enabled，走 mock 分支

import type { GoEndUserError } from '@/types/end-user-auth'

const GO_BACKEND_INTERNAL_URL =
  process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'
const INTERNAL_TOKEN =
  process.env.INTERNAL_TOKEN ?? process.env.INTERNAL_SERVICE_TOKEN ?? ''

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

export interface EndUserDatabaseLite {
  name: string
}

export interface EndUserDatabaseCatalogResult {
  databases: EndUserDatabaseLite[]
  totalCount: number
  page: number
  pageSize: number
}

export interface EndUserModelLite {
  id: string
  name: string
  title: string
  databaseName: string
}

export interface EndUserModelCatalogResult {
  models: EndUserModelLite[]
  totalCount: number
  page: number
  pageSize: number
}

export interface EndUserInitPrivateDBResult {
  success: boolean
  requestId?: string
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

/** Private DB 未初始化（需要用户确认后初始化） */
export class EndUserPrivateDBNotInitializedError extends Error {
  requestId?: string

  constructor(message = '私有库未初始化') {
    super(message)
    this.name = 'EndUserPrivateDBNotInitializedError'
  }
}

/** 参数校验失败 */
export class EndUserParamInvalidError extends Error {
  requestId?: string

  constructor(message = '参数校验失败') {
    super(message)
    this.name = 'EndUserParamInvalidError'
  }
}

/** 未授权（通常为 internal token 不匹配） */
export class EndUserUnauthorizedError extends Error {
  requestId?: string

  constructor(message = '未授权') {
    super(message)
    this.name = 'EndUserUnauthorizedError'
  }
}

/** 上游通用错误（保留 HTTP 状态） */
export class EndUserUpstreamError extends Error {
  status: number
  requestId?: string
  code?: string

  constructor(message: string, status: number) {
    super(message)
    this.name = 'EndUserUpstreamError'
    this.status = status
  }
}

// ============================================================================
// 内部工具函数
// ============================================================================

/**
 * 解析 Go 后端错误响应
 */
function attachRequestId<T extends Error>(err: T, requestId?: string): T {
  if (requestId && typeof err === 'object' && err) {
    ;(err as T & { requestId?: string }).requestId = requestId
  }
  return err
}

function createInternalHeaders(orgName?: string, projectSlug?: string): { headers: Record<string, string>; requestId: string } {
  const requestId = `bff-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  const scopedHeaders: Record<string, string> = {}
  if (orgName) scopedHeaders['X-Org-Name'] = orgName
  if (projectSlug) scopedHeaders['X-Project-Slug'] = projectSlug

  return {
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': INTERNAL_TOKEN,
      'X-Request-ID': requestId,
      ...scopedHeaders,
    },
    requestId,
  }
}

function normalizeGoErrorCode(code?: string): string {
  if (!code) return ''
  return code.split('.')[0] || code
}

async function parseGoError(res: Response, fallbackRequestId?: string): Promise<Error> {
  const headerRequestId = res.headers.get('x-request-id') || undefined
  const requestId = headerRequestId ?? fallbackRequestId

  try {
    const data = (await res.json()) as GoEndUserError
    const { code, message } = data.error
    const normalizedCode = normalizeGoErrorCode(code)

    switch (normalizedCode) {
      case 'INVALID_CREDENTIALS':
        return attachRequestId(new EndUserInvalidCredentialsError(message), requestId)
      case 'ACCOUNT_DISABLED':
        return attachRequestId(new EndUserAccountDisabledError(message), requestId)
      case 'CONFLICT':
        return attachRequestId(new EndUserConflictError(message), requestId)
      case 'INVALID_REFRESH_TOKEN':
        return attachRequestId(new EndUserTokenError(message), requestId)
      case 'CLUSTER_NOT_CONFIGURED':
        return attachRequestId(new EndUserClusterNotConfiguredError(message), requestId)
      case 'PRIVATE_DB_NOT_INITIALIZED':
        return attachRequestId(new EndUserPrivateDBNotInitializedError(message), requestId)
      case 'PARAM_INVALID':
        return attachRequestId(new EndUserParamInvalidError(message), requestId)
      case 'UNAUTHORIZED':
        return attachRequestId(new EndUserUnauthorizedError(message), requestId)
      default:
        const upstream = new EndUserUpstreamError(
          message || `Go backend error: ${res.status}`,
          res.status
        )
        upstream.requestId = requestId
        upstream.code = normalizedCode || code
        return upstream
    }
  } catch {
    if (res.status === 401) {
      return attachRequestId(new EndUserUnauthorizedError(`Go backend error: ${res.status}`), requestId)
    }
    const upstream = new EndUserUpstreamError(`Go backend error: ${res.status}`, res.status)
    upstream.requestId = requestId
    return upstream
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
 * 调用 Go Backend /internal/v1/end-user/auth/login
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
  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/v1/end-user/auth/login`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  return res.json() as Promise<EndUserLoginResult>
}

/**
 * 调用 Go Backend /internal/v1/end-user/auth/register
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
  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/v1/end-user/auth/register`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  return res.json() as Promise<EndUserRegisterResult>
}

/**
 * 调用 Go Backend /internal/v1/end-user/auth/refresh (token rotation)
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
  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/v1/end-user/auth/refresh`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  return res.json() as Promise<EndUserRefreshResult>
}

/**
 * 调用 Go Backend /internal/v1/end-user/auth/logout
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
  const { headers } = createInternalHeaders(params.orgName, params.projectSlug)
  await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/v1/end-user/auth/logout`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  }).catch(() => {
    // Ignore errors - Cookie will be cleared regardless
  })
}

/**
 * 调用 Go Backend /internal/v1/end-user/auth/me
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
  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/v1/end-user/auth/me`, {
    method: 'GET',
    headers: {
      ...headers,
      'X-End-User-Id': params.userId,
      'X-Org-Name': params.orgName,
      'X-Project-Slug': params.projectSlug,
    },
  })

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  return res.json() as Promise<EndUserMeResult>
}

/**
 * 调用 Go Backend /internal/end-user/data/database-catalog
 */
export async function callGoEndUserDatabaseCatalog(params: {
  orgName: string
  projectSlug: string
  userId: string
  search?: string
  page?: number
  pageSize?: number
}): Promise<EndUserDatabaseCatalogResult> {
  if (USE_MOCK) {
    const allNames = ['modelcraft_app', 'modelcraft_strapi', 'modelcraft_analytics']
    const keyword = (params.search ?? '').trim().toLowerCase()
    const filtered = keyword ? allNames.filter((name) => name.toLowerCase().includes(keyword)) : allNames
    const page = Math.max(1, params.page ?? 1)
    const pageSize = Math.max(1, params.pageSize ?? 20)
    const offset = (page - 1) * pageSize
    const pageItems = filtered.slice(offset, offset + pageSize)

    return {
      databases: pageItems.map((name) => ({ name })),
      totalCount: filtered.length,
      page,
      pageSize,
    }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const searchParams = new URLSearchParams()
  if (params.search) searchParams.set('search', params.search)
  searchParams.set('page', String(params.page ?? 1))
  searchParams.set('pageSize', String(params.pageSize ?? 50))

  const res = await fetch(
    `${GO_BACKEND_INTERNAL_URL}/internal/end-user/data/database-catalog?${searchParams.toString()}`,
    {
      method: 'GET',
      headers: {
        ...headers,
        'X-End-User-Id': params.userId,
      },
    }
  )

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  type GoCatalogResponse = {
    databases?: Array<{ name?: string }>
    totalCount?: number
    page?: number
    pageSize?: number
  }
  const data = (await res.json()) as GoCatalogResponse

  return {
    databases: (data.databases ?? [])
      .map((item) => item?.name ?? '')
      .filter((name): name is string => name.length > 0)
      .map((name) => ({ name })),
    totalCount: data.totalCount ?? 0,
    page: data.page ?? 1,
    pageSize: data.pageSize ?? 50,
  }
}

/**
 * 调用 Go Backend /internal/end-user/data/model-catalog
 */

/**
 * 调用 Go Backend /internal/end-user/data/init-private-db
 */
export async function callGoEndUserInitPrivateDB(params: {
  orgName: string
  projectSlug: string
  userId?: string
}): Promise<EndUserInitPrivateDBResult> {
  if (USE_MOCK) {
    return { success: true }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const requestHeaders: Record<string, string> = {
    ...headers,
  }
  if (params.userId) {
    requestHeaders['X-End-User-Id'] = params.userId
  }

  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/internal/end-user/data/init-private-db`, {
    method: 'POST',
    headers: requestHeaders,
  })

  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  type GoInitResp = { success?: boolean; requestId?: string }
  const data = (await res.json()) as GoInitResp
  return {
    success: Boolean(data.success),
    requestId: data.requestId,
  }
}

export async function callGoEndUserModelCatalog(params: {
  orgName: string
  projectSlug: string
  userId: string
  databaseName: string
  search?: string
  page?: number
  pageSize?: number
}): Promise<EndUserModelCatalogResult> {
  if (USE_MOCK) {
    const allModels = [
      { id: 'mock-model-users', name: 'users', title: '用户', databaseName: params.databaseName },
      { id: 'mock-model-orders', name: 'orders', title: '订单', databaseName: params.databaseName },
    ]
    const keyword = (params.search ?? '').trim().toLowerCase()
    const filtered = keyword
      ? allModels.filter((m) => m.name.toLowerCase().includes(keyword) || m.title.toLowerCase().includes(keyword))
      : allModels

    return {
      models: filtered,
      totalCount: filtered.length,
      page: Math.max(1, params.page ?? 1),
      pageSize: Math.max(1, params.pageSize ?? 50),
    }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const searchParams = new URLSearchParams()
  searchParams.set('databaseName', params.databaseName)
  if (params.search) searchParams.set('search', params.search)
  searchParams.set('page', String(params.page ?? 1))
  searchParams.set('pageSize', String(params.pageSize ?? 200))

  const res = await fetch(
    `${GO_BACKEND_INTERNAL_URL}/internal/end-user/data/model-catalog?${searchParams.toString()}`,
    {
      method: 'GET',
      headers: {
        ...headers,
        'X-End-User-Id': params.userId,
      },
    }
  )
  if (!res.ok) {
    throw await parseGoError(res, requestId)
  }

  type GoModelCatalogResponse = {
    models?: Array<{
      id?: string
      name?: string
      title?: string
      databaseName?: string
    }>
    totalCount?: number
    page?: number
    pageSize?: number
  }
  const data = (await res.json()) as GoModelCatalogResponse
  return {
    models: (data.models ?? [])
      .filter((item) => item?.id && item?.name && item?.databaseName)
      .map((item) => ({
        id: item.id as string,
        name: item.name as string,
        title: item.title ?? '',
        databaseName: item.databaseName as string,
      })),
    totalCount: data.totalCount ?? 0,
    page: data.page ?? 1,
    pageSize: data.pageSize ?? 200,
  }
}

// Re-export org/project-scoped end-user management APIs so callers can use a single client module path.
export * from './end-user-go-client-v2'
