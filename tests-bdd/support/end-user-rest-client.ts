// tests-bdd/support/end-user-rest-client.ts
// End-User Auth REST API Client for BDD Tests

import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'
const INTERNAL_TOKEN = process.env.INTERNAL_TOKEN ?? 'test-internal-token'

// Types
export interface AccessibleProject {
  slug: string
  title: string
}

export interface EndUserAuthResult {
  userId: string
  username?: string
  accessToken?: string
  refreshToken?: string
  projects?: AccessibleProject[]
  selectedProject?: string
  expiresIn: number
}

export interface EndUserInfo {
  id: string
  username: string
  isForbidden: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface EndUserListResult {
  users: EndUserInfo[]
  totalCount: number
  hasNextPage: boolean
  endCursor?: string
}

export interface RestError {
  code: string
  message: string
}

export interface RestResult<T> {
  status: number
  data?: T
  error?: RestError
}

// Request types
interface CreateEndUserRequest {
  username: string
  password: string
}

interface LoginEndUserRequest {
  username: string
  password: string
}

interface ListEndUsersRequest {
  search?: string
  first?: number
  after?: string
}

interface AnyBody {
  [key: string]: any
}

function parseExpiresIn(body?: AnyBody): number {
  if (typeof body?.expiresIn === 'number') {
    return body.expiresIn
  }
  return 3600
}

function toRestError(status: number, body: AnyBody | null): RestError {
  if (body?.error?.code && body?.error?.message) {
    return body.error as RestError
  }
  return {
    code: status === 401 ? 'UNAUTHORIZED' : 'SYSTEM_ERROR',
    message: body?.error?.message ?? body?.message ?? `HTTP ${status}`,
  }
}

/**
 * End-User REST Client
 * 调用后端内部接口 /internal/end-user/* 和 /internal/end-users/*
 */
export class EndUserRestClient {
  private orgName: string
  private projectSlug: string

  constructor(orgName: string, projectSlug: string) {
    this.orgName = orgName
    this.projectSlug = projectSlug
  }

  private async parseBody(res: Response): Promise<AnyBody | null> {
    const raw = await res.text()
    if (!raw) return null
    try {
      return JSON.parse(raw) as AnyBody
    } catch {
      return { message: raw }
    }
  }

  // ============================================================
  // 开发者管理接口 (需要 X-Internal-Token)
  // ============================================================

  /**
   * 创建终端用户 (POST /internal/end-users)
   */
  async createEndUser(
    request: CreateEndUserRequest,
    internalToken: string
  ): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Internal-Token': internalToken,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
      body: JSON.stringify({
        username: request.username,
        password: request.password,
      }),
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body?.endUser?.id,
          username: body?.endUser?.username,
          expiresIn: 3600,
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 查询终端用户列表 (GET /internal/end-users)
   */
  async listEndUsers(
    request: ListEndUsersRequest,
    internalToken: string
  ): Promise<RestResult<EndUserListResult>> {
    const params = new URLSearchParams()
    if (request.search) params.append('search', request.search)
    if (request.first) params.append('first', request.first.toString())
    if (request.after) params.append('after', request.after)

    const res = await fetch(`${API_BASE_URL}/internal/end-users?${params.toString()}`, {
      method: 'GET',
      headers: {
        'X-Internal-Token': internalToken,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          users: body?.items || [],
          totalCount: body?.totalCount || 0,
          hasNextPage: body?.pageInfo?.hasNextPage || false,
          endCursor: body?.pageInfo?.endCursor,
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 更新终端用户状态 (PATCH /internal/end-users/{userId}/status)
   */
  async updateEndUserStatus(
    userId: string,
    isForbidden: boolean,
    internalToken: string
  ): Promise<RestResult<void>> {
    const res = await fetch(
      `${API_BASE_URL}/internal/end-users/${userId}/status`,
      {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-Internal-Token': internalToken,
          'X-Org-Name': this.orgName,
          'X-Project-Slug': this.projectSlug,
        },
        body: JSON.stringify({
          isForbidden: isForbidden,
        }),
      }
    )

    const body = await this.parseBody(res)

    if (res.ok) {
      return { status: res.status }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 删除终端用户 (DELETE /internal/end-users/{userId})
   */
  async deleteEndUser(
    userId: string,
    internalToken: string
  ): Promise<RestResult<void>> {
    const res = await fetch(
      `${API_BASE_URL}/internal/end-users/${userId}`,
      {
        method: 'DELETE',
        headers: {
          'X-Internal-Token': internalToken,
          'X-Org-Name': this.orgName,
          'X-Project-Slug': this.projectSlug,
        },
      }
    )

    if (res.ok) {
      return { status: res.status }
    }

    const body = await this.parseBody(res)
    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  // ============================================================
  // 终端用户自助认证接口
  // ============================================================

  /**
   * 终端用户注册 (POST /internal/end-user/auth/register)
   */
  async registerEndUser(
    request: CreateEndUserRequest
  ): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(
      `${API_BASE_URL}/internal/end-user/auth/register`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Internal-Token': INTERNAL_TOKEN,
          'X-Org-Name': this.orgName,
          'X-Project-Slug': this.projectSlug,
        },
        body: JSON.stringify({
          username: request.username,
          password: request.password,
        }),
      }
    )

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body?.userId,
          refreshToken: body?.refreshToken,
          expiresIn: parseExpiresIn(body),
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 终端用户登录 (POST /internal/end-user/auth/login)
   */
  async loginEndUser(
    request: LoginEndUserRequest
  ): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Internal-Token': INTERNAL_TOKEN,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
      body: JSON.stringify({
        username: request.username,
        password: request.password,
      }),
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body?.userId,
          username: request.username,
          accessToken: body?.accessToken,
          refreshToken: body?.refreshToken,
          projects: body?.projects ?? [],
          expiresIn: parseExpiresIn(body),
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 选择项目上下文 (POST /internal/end-user/auth/select-project)
   */
  async selectProjectContext(
    refreshToken: string,
    projectSlug: string
  ): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/select-project`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Internal-Token': INTERNAL_TOKEN,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
      body: JSON.stringify({
        refreshToken,
        projectSlug,
      }),
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body?.userId,
          selectedProject: body?.selectedProject,
          expiresIn: 3600,
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 终端用户登出 (POST /internal/end-user/auth/logout)
   */
  async logoutEndUser(refreshToken: string): Promise<RestResult<void>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Internal-Token': INTERNAL_TOKEN,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
      body: JSON.stringify({
        refreshToken,
      }),
    })

    if (res.status === 204 || res.ok) {
      return { status: res.status }
    }

    const body = await this.parseBody(res)
    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 刷新 Token (POST /internal/end-user/auth/refresh)
   */
  async refreshEndUserToken(refreshToken: string): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Internal-Token': INTERNAL_TOKEN,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
      },
      body: JSON.stringify({
        refreshToken,
      }),
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body?.userId,
          accessToken: body?.accessToken,
          refreshToken: body?.refreshToken,
          projects: body?.projects ?? [],
          expiresIn: parseExpiresIn(body),
        },
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }

  /**
   * 获取当前用户信息 (GET /internal/end-user/auth/me)
   */
  async getEndUserMe(userId: string): Promise<RestResult<EndUserInfo>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/me`, {
      method: 'GET',
      headers: {
        'X-Internal-Token': INTERNAL_TOKEN,
        'X-Org-Name': this.orgName,
        'X-Project-Slug': this.projectSlug,
        'X-End-User-Id': userId,
      },
    })

    const body = await this.parseBody(res)

    if (res.ok) {
      return {
        status: res.status,
        data: body?.endUser as EndUserInfo,
      }
    }

    return {
      status: res.status,
      error: toRestError(res.status, body),
    }
  }
}
