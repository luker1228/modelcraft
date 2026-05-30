// src/api-client/end-user/end-user-go-client.ts
// 终端用户 Go Backend Client（客户端侧，对称 bff/end-user/end-user-go-client.ts）
// 包含 Project 级认证（auth/login, refresh, logout, me, data）
// 以及 Org 级账号管理 + Project 访问控制（EndUser v1）

import type { GoEndUserError } from '@/types/end-user-auth'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

const GATEWAY_URL = ''

// 判断是否启用 mock（Project 级认证）
const USE_MOCK = process.env.NEXT_PUBLIC_API_MOCKING === 'enabled'

// 判断是否启用 Org/Project 访问控制 mock
function shouldUseEndUserV2Mock(): boolean {
  return process.env.NEXT_PUBLIC_END_USER_V2_MOCKING === 'enabled'
}

// ============================================================================
// 结果类型（客户端使用）
// ============================================================================

export interface EndUserLoginResult {
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

// ── Org 级账号管理 ────────────────────────────────────────────────────────────

export interface EndUserOrgLoginResult {
  userId: string
  /** 该用户可访问的 Project 列表（Org 级，不含 projectSlug 的 refresh token） */
  accessibleProjects: EndUserAccessibleProject[]
}

export interface EndUserOrgUser {
  id: string
  username: string
  isForbidden: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface EndUserOrgUserConnection {
  nodes: EndUserOrgUser[]
  totalCount: number
}

export interface EndUserProjectAccess {
  id: string
  endUser: Pick<EndUserOrgUser, 'id' | 'username' | 'isForbidden'>
  permissionBundleId: string
  permissionBundleName: string
  grantedBy: string
  grantedAt: string
}

export interface EndUserProjectAccessConnection {
  nodes: EndUserProjectAccess[]
  totalCount: number
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

/** 未授权 */
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

function attachRequestId<T extends Error>(err: T, requestId?: string): T {
  if (requestId && typeof err === 'object' && err) {
    ;(err as T & { requestId?: string }).requestId = requestId
  }
  return err
}

function createInternalHeaders(orgName?: string, projectSlug?: string): { headers: Record<string, string>; requestId: string } {
  const requestId = `gw-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  const scopedHeaders: Record<string, string> = {}
  if (orgName) scopedHeaders['X-Org-Name'] = orgName
  if (projectSlug) scopedHeaders['X-Project-Slug'] = projectSlug

  return {
    headers: {
      'Content-Type': 'application/json',
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
      default: {
        const upstream = new EndUserUpstreamError(
          message || `Go backend error: ${res.status}`,
          res.status
        )
        upstream.requestId = requestId
        upstream.code = normalizedCode || code
        return upstream
      }
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

// Mock 用户数据库（Project 级）
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

// Mock 数据（Org 级用户池，不绑定 project）
export interface MockOrgUser {
  id: string
  username: string
  password: string
  isForbidden: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

const MOCK_ORG_USERS: Record<string, MockOrgUser> = {
  alice: {
    id: 'mock-user-001', username: 'alice', password: 'password123',
    isForbidden: false, createdBy: 'admin',
    createdAt: '2025-03-01T00:00:00Z', updatedAt: '2025-03-01T00:00:00Z',
  },
  zhangsan: {
    id: 'mock-user-002', username: 'zhangsan', password: 'Pass1234',
    isForbidden: false, createdBy: 'admin',
    createdAt: '2025-03-05T00:00:00Z', updatedAt: '2025-03-05T00:00:00Z',
  },
  lisi: {
    id: 'mock-user-003', username: 'lisi', password: 'Pass1234',
    isForbidden: false, createdBy: 'admin',
    createdAt: '2025-03-10T00:00:00Z', updatedAt: '2025-03-10T00:00:00Z',
  },
  disabled_org: {
    id: 'mock-user-disabled', username: 'disabled', password: 'password123',
    isForbidden: true, createdBy: 'admin',
    createdAt: '2025-03-15T00:00:00Z', updatedAt: '2025-03-15T00:00:00Z',
  },
  noaccess: {
    id: 'mock-user-noaccess', username: 'noaccess', password: 'password123',
    isForbidden: false, createdBy: 'admin',
    createdAt: '2025-04-01T00:00:00Z', updatedAt: '2025-04-01T00:00:00Z',
  },
}

// EndUser ↔ Project 访问控制
const MOCK_PROJECT_ACCESSES: Record<string, Record<string, { accessId: string; permissionBundleId: string; permissionBundleName: string; grantedBy: string; grantedAt: string }>> = {
  'sales-system': {
    'mock-user-001': { accessId: 'access-001', permissionBundleId: 'bundle-sales', permissionBundleName: '销售员', grantedBy: 'admin', grantedAt: '2025-03-01T00:00:00Z' },
    'mock-user-002': { accessId: 'access-002', permissionBundleId: 'bundle-sales', permissionBundleName: '销售员', grantedBy: 'admin', grantedAt: '2025-03-05T00:00:00Z' },
  },
  'hr-system': {
    'mock-user-001': { accessId: 'access-003', permissionBundleId: 'bundle-hr', permissionBundleName: 'HR 专员', grantedBy: 'admin', grantedAt: '2025-03-10T00:00:00Z' },
  },
  'inventory': {
    'mock-user-003': { accessId: 'access-004', permissionBundleId: 'bundle-viewer', permissionBundleName: '查看者', grantedBy: 'admin', grantedAt: '2025-03-20T00:00:00Z' },
  },
}

const MOCK_PROJECTS: EndUserAccessibleProject[] = [
  { slug: 'sales-system', title: '销售系统' },
  { slug: 'hr-system',    title: 'HR 系统' },
  { slug: 'inventory',    title: '库存管理' },
]

// ============================================================================
// Project 级认证函数
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

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v1/end-user/auth/login`, {
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
 * 调用 Go Backend /internal/v1/end-user/auth/refresh (token rotation)
 */
export async function callGoEndUserRefresh(params: {
  orgName: string
  projectSlug: string
  refreshToken: string
}): Promise<EndUserRefreshResult> {
  if (USE_MOCK) {
    const userId = MOCK_REFRESH_TOKENS[params.refreshToken]

    if (!userId) {
      throw new EndUserTokenError('无效的 refresh token')
    }

    const user = Object.values(MOCK_USERS).find((u) => u.userId === userId)
    if (user?.isForbidden) {
      throw new EndUserAccountDisabledError()
    }

    const newRefreshToken = `mock-refresh-token-rotated-${Date.now()}`
    MOCK_REFRESH_TOKENS[newRefreshToken] = userId

    return {
      userId,
      refreshToken: newRefreshToken,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v1/end-user/auth/refresh`, {
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
    delete MOCK_REFRESH_TOKENS[params.refreshToken]
    return
  }

  const { headers } = createInternalHeaders(params.orgName, params.projectSlug)
  await fetch(`${GATEWAY_URL}/internal/v1/end-user/auth/logout`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  }).catch(() => {
    // Ignore errors - Cookie will be cleared regardless
  })
}

/**
 * 调用 Go Backend /internal/v1/end-user/auth/me
 */
export async function callGoEndUserMe(params: {
  orgName: string
  projectSlug: string
  userId: string
}): Promise<EndUserMeResult> {
  if (USE_MOCK) {
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

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v1/end-user/auth/me`, {
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
    `${GATEWAY_URL}/internal/end-user/data/database-catalog?${searchParams.toString()}`,
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

  const res = await fetch(`${GATEWAY_URL}/internal/end-user/data/init-private-db`, {
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

/**
 * 调用 Go Backend /internal/end-user/data/model-catalog
 */
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
    `${GATEWAY_URL}/internal/end-user/data/model-catalog?${searchParams.toString()}`,
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

// ============================================================================
// Org 级 EndUser 登录（返回 accessible projects）
// ============================================================================

/**
 * 调用 Go Backend /internal/v1/end-user/auth/login（Org 级）
 * 返回 userId + 该用户在当前 Org 下可访问的 Project 列表
 */
export async function callGoEndUserLoginOrg(params: {
  orgName: string
  username: string
  password: string
}): Promise<EndUserOrgLoginResult> {
  if (shouldUseEndUserV2Mock()) {
    const user = MOCK_ORG_USERS[params.username]
    if (!user || user.password !== params.password) throw new EndUserInvalidCredentialsError()
    if (user.isForbidden) throw new EndUserAccountDisabledError()

    const accessible: EndUserAccessibleProject[] = MOCK_PROJECTS.filter((p) => {
      const projectAccesses = MOCK_PROJECT_ACCESSES[p.slug]
      return projectAccesses && user.id in projectAccesses
    })

    if (user.username === 'noaccess') {
      return { userId: user.id, accessibleProjects: [] }
    }

    return { userId: user.id, accessibleProjects: accessible }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const res = await fetch(`${GATEWAY_URL}/internal/v1/end-user/auth/login`, {
    method: 'POST',
    headers,
    body: JSON.stringify(params),
  })
  if (!res.ok) throw await parseGoError(res, requestId)

  const data = await res.json() as {
    userId: string
    accessibleProjects?: EndUserAccessibleProject[]
    projects?: EndUserAccessibleProject[]
  }

  return {
    userId: data.userId,
    accessibleProjects: data.accessibleProjects ?? data.projects ?? [],
  }
}

// ============================================================================
// Org 级 EndUser CRUD
// ============================================================================

export async function callGoListOrgEndUsers(params: {
  orgName: string
  search?: string
  first?: number
  after?: string
}): Promise<EndUserOrgUserConnection> {
  if (shouldUseEndUserV2Mock()) {
    let nodes = Object.values(MOCK_ORG_USERS).filter((u) => u.username !== 'noaccess')
    if (params.search) {
      nodes = nodes.filter((u) => u.username.includes(params.search!))
    }
    return { nodes: nodes.map(({ password: _pw, ...u }) => u as EndUserOrgUser), totalCount: nodes.length }
  }

  throw new Error('Deprecated: /internal/end-users has been removed; migrate to Org GraphQL.')
}

export async function callGoCreateOrgEndUser(params: {
  orgName: string
  username: string
  password: string
}): Promise<EndUserOrgUser> {
  if (shouldUseEndUserV2Mock()) {
    if (MOCK_ORG_USERS[params.username]) throw new EndUserConflictError()
    if (params.password.length < 8) throw new EndUserParamInvalidError('密码至少 8 位，含字母与数字')
    const id = `mock-user-${Date.now()}`
    const now = new Date().toISOString()
    MOCK_ORG_USERS[params.username] = {
      id, username: params.username, password: params.password,
      isForbidden: false, createdBy: 'admin', createdAt: now, updatedAt: now,
    }
    return { id, username: params.username, isForbidden: false, createdBy: 'admin', createdAt: now, updatedAt: now }
  }

  throw new Error('Deprecated: /internal/end-users has been removed; migrate to Org GraphQL.')
}

export async function callGoUpdateOrgEndUserStatus(params: {
  orgName: string
  userId: string
  isForbidden: boolean
}): Promise<EndUserOrgUser> {
  if (shouldUseEndUserV2Mock()) {
    const user = Object.values(MOCK_ORG_USERS).find((u) => u.id === params.userId)
    if (!user) throw new EndUserParamInvalidError('用户不存在')
    user.isForbidden = params.isForbidden
    user.updatedAt = new Date().toISOString()
    return { ...user } as EndUserOrgUser
  }

  throw new Error('Deprecated: /internal/end-users has been removed; migrate to Org GraphQL.')
}

export async function callGoDeleteOrgEndUser(params: {
  orgName: string
  userId: string
}): Promise<void> {
  if (shouldUseEndUserV2Mock()) {
    const key = Object.keys(MOCK_ORG_USERS).find((k) => MOCK_ORG_USERS[k].id === params.userId)
    if (!key) throw new EndUserParamInvalidError('用户不存在')
    delete MOCK_ORG_USERS[key]
    return
  }

  throw new Error('Deprecated: /internal/end-users has been removed; migrate to Org GraphQL.')
}

// ============================================================================
// Project 级 EndUser 访问控制
// ============================================================================

export async function callGoListProjectEndUserAccesses(params: {
  orgName: string
  projectSlug: string
  search?: string
  first?: number
}): Promise<EndUserProjectAccessConnection> {
  if (shouldUseEndUserV2Mock()) {
    const projectAccesses = MOCK_PROJECT_ACCESSES[params.projectSlug] ?? {}
    let nodes: EndUserProjectAccess[] = Object.entries(projectAccesses).map(([userId, a]) => {
      const user = Object.values(MOCK_ORG_USERS).find((u) => u.id === userId)
      return {
        id: a.accessId,
        endUser: { id: userId, username: user?.username ?? userId, isForbidden: user?.isForbidden ?? false },
        permissionBundleId: a.permissionBundleId,
        permissionBundleName: a.permissionBundleName,
        grantedBy: a.grantedBy,
        grantedAt: a.grantedAt,
      }
    })
    if (params.search) {
      nodes = nodes.filter((n) => n.endUser.username.includes(params.search!))
    }
    return { nodes, totalCount: nodes.length }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const qs = new URLSearchParams()
  if (params.search) qs.set('search', params.search)
  if (params.first)  qs.set('first', String(params.first))
  const res = await fetch(
    `${GATEWAY_URL}/internal/v2/end-user/project-accesses?${qs}`,
    { headers, credentials: 'include' }
  )
  if (!res.ok) throw await parseGoError(res, requestId)
  return res.json() as Promise<EndUserProjectAccessConnection>
}

export async function callGoGrantEndUserProjectAccess(params: {
  orgName: string
  projectSlug: string
  endUserId: string
  permissionBundleId: string
  permissionBundleName: string
}): Promise<EndUserProjectAccess> {
  if (shouldUseEndUserV2Mock()) {
    const user = Object.values(MOCK_ORG_USERS).find((u) => u.id === params.endUserId)
    if (!user) throw new EndUserParamInvalidError('用户不存在')
    const projectAccesses = MOCK_PROJECT_ACCESSES[params.projectSlug] ?? {}
    if (projectAccesses[params.endUserId]) throw new EndUserConflictError('该用户已拥有此项目的访问权')
    const accessId = `access-${Date.now()}`
    const now = new Date().toISOString()
    if (!MOCK_PROJECT_ACCESSES[params.projectSlug]) MOCK_PROJECT_ACCESSES[params.projectSlug] = {}
    MOCK_PROJECT_ACCESSES[params.projectSlug][params.endUserId] = {
      accessId, permissionBundleId: params.permissionBundleId,
      permissionBundleName: params.permissionBundleName, grantedBy: 'admin', grantedAt: now,
    }
    return {
      id: accessId,
      endUser: { id: user.id, username: user.username, isForbidden: user.isForbidden },
      permissionBundleId: params.permissionBundleId,
      permissionBundleName: params.permissionBundleName,
      grantedBy: 'admin',
      grantedAt: now,
    }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v2/end-user/project-accesses`, {
    method: 'POST', headers,
    body: JSON.stringify({ endUserId: params.endUserId, permissionBundleId: params.permissionBundleId }),
  })
  if (!res.ok) throw await parseGoError(res, requestId)
  return res.json() as Promise<EndUserProjectAccess>
}

export async function callGoRevokeEndUserProjectAccess(params: {
  orgName: string
  projectSlug: string
  accessId: string
}): Promise<void> {
  if (shouldUseEndUserV2Mock()) {
    const projectAccesses = MOCK_PROJECT_ACCESSES[params.projectSlug] ?? {}
    const userId = Object.keys(projectAccesses).find(
      (uid) => projectAccesses[uid].accessId === params.accessId
    )
    if (!userId) throw new EndUserParamInvalidError('访问权不存在')
    delete MOCK_PROJECT_ACCESSES[params.projectSlug][userId]
    return
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v2/end-user/project-accesses/${params.accessId}`, {
    method: 'DELETE', headers,
  })
  if (!res.ok) throw await parseGoError(res, requestId)
}

export async function callGoUpdateEndUserProjectAccess(params: {
  orgName: string
  projectSlug: string
  accessId: string
  permissionBundleId: string
  permissionBundleName: string
}): Promise<EndUserProjectAccess> {
  if (shouldUseEndUserV2Mock()) {
    const projectAccesses = MOCK_PROJECT_ACCESSES[params.projectSlug] ?? {}
    const entry = Object.entries(projectAccesses).find(([, a]) => a.accessId === params.accessId)
    if (!entry) throw new EndUserParamInvalidError('访问权不存在')
    const [userId, access] = entry
    access.permissionBundleId = params.permissionBundleId
    access.permissionBundleName = params.permissionBundleName
    const user = Object.values(MOCK_ORG_USERS).find((u) => u.id === userId)
    return {
      id: params.accessId,
      endUser: { id: userId, username: user?.username ?? userId, isForbidden: user?.isForbidden ?? false },
      permissionBundleId: params.permissionBundleId,
      permissionBundleName: params.permissionBundleName,
      grantedBy: access.grantedBy,
      grantedAt: access.grantedAt,
    }
  }

  const { headers, requestId } = createInternalHeaders(params.orgName, params.projectSlug)
  const res = await fetch(`${GATEWAY_URL}/internal/v2/end-user/project-accesses/${params.accessId}`, {
    method: 'PATCH', headers,
    body: JSON.stringify({ permissionBundleId: params.permissionBundleId }),
  })
  if (!res.ok) throw await parseGoError(res, requestId)
  return res.json() as Promise<EndUserProjectAccess>
}

/** 获取单个用户在当前 Org 中有权访问的 Project 列表（用于 Org 管理页的 Drawer） */
export async function callGoGetUserAccessibleProjects(params: {
  orgName: string
  userId: string
}): Promise<EndUserAccessibleProject[]> {
  if (shouldUseEndUserV2Mock()) {
    return MOCK_PROJECTS.filter((p) => {
      const projectAccesses = MOCK_PROJECT_ACCESSES[p.slug]
      return projectAccesses && params.userId in projectAccesses
    })
  }

  throw new Error('Deprecated: /internal/end-users has been removed; migrate to Org GraphQL.')
}
