// tests-bdd/support/end-user-rest-client.ts
// End-User Auth REST API Client for BDD Tests

import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

// Types
export interface EndUserAuthResult {
  userId: string
  username: string
  accessToken: string
  refreshToken?: string
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
        org_name: this.orgName,
        project_slug: this.projectSlug,
        username: request.username,
        password: request.password,
      }),
    })

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body.userId,
          username: body.username,
          accessToken: body.accessToken || '',
          expiresIn: body.expiresIn || 3600,
        },
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
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
    params.append('org_name', this.orgName)
    params.append('project_slug', this.projectSlug)
    if (request.search) params.append('search', request.search)
    if (request.first) params.append('first', request.first.toString())
    if (request.after) params.append('after', request.after)

    const res = await fetch(`${API_BASE_URL}/internal/end-users?${params}`, {
      method: 'GET',
      headers: {
        'X-Internal-Token': internalToken,
      },
    })

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: {
          users: body.users || [],
          totalCount: body.totalCount || 0,
          hasNextPage: body.hasNextPage || false,
          endCursor: body.endCursor,
        },
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
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
        },
        body: JSON.stringify({
          org_name: this.orgName,
          project_slug: this.projectSlug,
          is_forbidden: isForbidden,
        }),
      }
    )

    const body = await res.json()

    if (res.ok) {
      return { status: res.status }
    }

    return {
      status: res.status,
      error: body.error as RestError,
    }
  }

  /**
   * 删除终端用户 (DELETE /internal/end-users/{userId})
   */
  async deleteEndUser(
    userId: string,
    internalToken: string
  ): Promise<RestResult<void>> {
    const params = new URLSearchParams()
    params.append('org_name', this.orgName)
    params.append('project_slug', this.projectSlug)

    const res = await fetch(
      `${API_BASE_URL}/internal/end-users/${userId}?${params}`,
      {
        method: 'DELETE',
        headers: {
          'X-Internal-Token': internalToken,
        },
      }
    )

    if (res.status === 204) {
      return { status: res.status }
    }

    const body = await res.json()
    return {
      status: res.status,
      error: body.error as RestError,
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
        },
        body: JSON.stringify({
          org_name: this.orgName,
          project_slug: this.projectSlug,
          username: request.username,
          password: request.password,
        }),
      }
    )

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body.userId,
          username: body.username,
          accessToken: body.accessToken,
          refreshToken: body.refreshToken,
          expiresIn: body.expiresIn || 3600,
        },
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
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
      },
      body: JSON.stringify({
        org_name: this.orgName,
        project_slug: this.projectSlug,
        username: request.username,
        password: request.password,
      }),
    })

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body.userId,
          username: body.username,
          accessToken: body.accessToken,
          refreshToken: body.refreshToken,
          expiresIn: body.expiresIn || 3600,
        },
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
    }
  }

  /**
   * 终端用户登出 (POST /internal/end-user/auth/logout)
   */
  async logoutEndUser(accessToken: string): Promise<RestResult<void>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/logout`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        org_name: this.orgName,
        project_slug: this.projectSlug,
      }),
    })

    if (res.status === 204) {
      return { status: res.status }
    }

    const body = await res.json()
    return {
      status: res.status,
      error: body.error as RestError,
    }
  }

  /**
   * 刷新 Token (POST /internal/end-user/auth/refresh)
   */
  async refreshEndUserToken(): Promise<RestResult<EndUserAuthResult>> {
    const res = await fetch(`${API_BASE_URL}/internal/end-user/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        org_name: this.orgName,
        project_slug: this.projectSlug,
      }),
      // 注意：refresh token 在 HttpOnly Cookie 中自动携带
      credentials: 'include',
    })

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: {
          userId: body.userId,
          username: body.username,
          accessToken: body.accessToken,
          refreshToken: body.refreshToken,
          expiresIn: body.expiresIn || 3600,
        },
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
    }
  }

  /**
   * 获取当前用户信息 (GET /internal/end-user/auth/me)
   */
  async getEndUserMe(accessToken: string): Promise<RestResult<EndUserInfo>> {
    const params = new URLSearchParams()
    params.append('org_name', this.orgName)
    params.append('project_slug', this.projectSlug)

    const res = await fetch(
      `${API_BASE_URL}/internal/end-user/auth/me?${params}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
        },
      }
    )

    const body = await res.json()

    if (res.ok) {
      return {
        status: res.status,
        data: body as EndUserInfo,
      }
    }

    return {
      status: res.status,
      error: body.error as RestError,
    }
  }
}
