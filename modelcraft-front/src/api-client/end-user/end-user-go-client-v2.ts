// src/bff/end-user/end-user-go-client-v2.ts
// EndUser v1 — Org 级账号管理 + Project 访问控制 Go Client
// 默认走真实后端；仅在 END_USER_V2_MOCKING=enabled 时启用 mock

import type { EndUserAccessibleProject } from '@/types/end-user-auth'
import {
  EndUserInvalidCredentialsError,
  EndUserAccountDisabledError,
  EndUserConflictError,
  EndUserParamInvalidError,
  EndUserUpstreamError,
} from './end-user-go-client'

const GATEWAY_URL = ''

function shouldUseEndUserV2Mock(): boolean {
  return process.env.NEXT_PUBLIC_END_USER_V2_MOCKING === 'enabled'
}

// ============================================================================
// 结果类型
// ============================================================================

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
// 内部工具
// ============================================================================

function createInternalHeaders(orgName?: string, projectSlug?: string) {
  const requestId = `gw-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  const scopedHeaders: Record<string, string> = {}
  if (orgName) scopedHeaders['X-Org-Name'] = orgName
  if (projectSlug) scopedHeaders['X-Project-Slug'] = projectSlug

  return {
    headers: {
      'Content-Type': 'application/json',
      
      'X-Request-ID': requestId,
      ...scopedHeaders,
    } as Record<string, string>,
    requestId,
  }
}

async function parseGoError(res: Response, fallbackRequestId?: string): Promise<Error> {
  const headerRequestId = res.headers.get('x-request-id') ?? undefined
  try {
    const data = (await res.json()) as {
      requestId?: string
      error?: { code?: string; message?: string }
    }
    const requestId = data.requestId ?? headerRequestId ?? fallbackRequestId
    const code = data.error?.code ?? ''
    const message = data.error?.message ?? ''
    switch (code) {
      case 'INVALID_CREDENTIALS': return Object.assign(new EndUserInvalidCredentialsError(message), { requestId })
      case 'ACCOUNT_DISABLED':    return Object.assign(new EndUserAccountDisabledError(message), { requestId })
      case 'CONFLICT':            return Object.assign(new EndUserConflictError(message), { requestId })
      case 'PARAM_INVALID':       return Object.assign(new EndUserParamInvalidError(message), { requestId })
      default: {
        const err = new EndUserUpstreamError(message || `Go backend error: ${res.status}`, res.status)
        err.requestId = requestId ?? undefined
        return err
      }
    }
  } catch {
    const requestId = headerRequestId ?? fallbackRequestId
    const err = new EndUserUpstreamError(`Go backend error: ${res.status}`, res.status)
    err.requestId = requestId ?? undefined
    return err
  }
}

// ============================================================================
// Mock 数据
// ============================================================================

export interface MockOrgUser {
  id: string
  username: string
  password: string
  isForbidden: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

// Org 级用户池（v1：不绑定 project）
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
  disabled: {
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

// EndUser ↔ Project 访问控制（projectSlug → userId[]）
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
// v1 登录（返回 accessible projects，不签发 projectSlug JWT）
// ============================================================================

/**
 * 调用 Go Backend /internal/v1/end-user/auth/login
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

    // 查找该用户有权访问的 project
    const accessible: EndUserAccessibleProject[] = MOCK_PROJECTS.filter((p) => {
      const projectAccesses = MOCK_PROJECT_ACCESSES[p.slug]
      return projectAccesses && user.id in projectAccesses
    })

    // noaccess 用户无项目权限
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

// ── 后端响应类型（/internal/end-users 使用 items 字段，非 nodes）──────────────
interface InternalListEndUsersResponse {
  items: EndUserOrgUser[]
  totalCount: number
  pageInfo: { hasNextPage: boolean; endCursor?: string }
}

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

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const qs = new URLSearchParams()
  if (params.search) qs.set('search', params.search)
  if (params.first)  qs.set('first', String(params.first))
  if (params.after)  qs.set('after', params.after)
  const res = await fetch(
    `${GATEWAY_URL}/internal/end-users?${qs}`,
    { headers, credentials: 'include' }
  )
  if (!res.ok) throw await parseGoError(res, requestId)
  const data = await res.json() as InternalListEndUsersResponse
  // 后端返回 items，转换为前端期望的 nodes 结构
  return { nodes: data.items ?? [], totalCount: data.totalCount ?? 0 }
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

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const res = await fetch(`${GATEWAY_URL}/internal/end-users`, {
    method: 'POST', headers,
    body: JSON.stringify({ username: params.username, password: params.password }),
  })
  if (!res.ok) throw await parseGoError(res, requestId)
  const data = await res.json() as { endUser: EndUserOrgUser }
  return data.endUser
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

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const res = await fetch(`${GATEWAY_URL}/internal/end-users/${params.userId}/status`, {
    method: 'PATCH', headers,
    body: JSON.stringify({ isForbidden: params.isForbidden }),
  })
  if (!res.ok) throw await parseGoError(res, requestId)
  const data = await res.json() as { endUser: EndUserOrgUser }
  return data.endUser
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

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const res = await fetch(`${GATEWAY_URL}/internal/end-users/${params.userId}`, {
    method: 'DELETE', headers,
  })
  if (!res.ok) throw await parseGoError(res, requestId)
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

  const { headers, requestId } = createInternalHeaders(params.orgName)
  const res = await fetch(
    `${GATEWAY_URL}/internal/end-users/${params.userId}/accessible-projects`,
    { headers, credentials: 'include' }
  )
  if (!res.ok) throw await parseGoError(res, requestId)
  const data = await res.json() as { projects: EndUserAccessibleProject[] }
  return data.projects ?? []
}
