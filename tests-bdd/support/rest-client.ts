import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export interface TokenResponse {
  accessToken: string
  tokenType: string
}

export interface RegisterProfileSnapshot {
  id: string
  userId: string
  nickname: string
  avatarUrl?: string
  bio?: string
}

export interface RegisterResponse {
  requestId: string
  userId: string
  orgName: string
  profile?: RegisterProfileSnapshot
}

export interface LoginResponse {
  requestId: string
  userId: string
  userName: string
  orgName: string
  refreshToken: string
  expiresAt: string
}

export interface RefreshResponse {
  requestId: string
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface InitOrgResponse {
  requestId: string
  success: boolean
  orgName: string
  displayName: string
  alreadyExists: boolean
}

export interface MembershipInfo {
  orgId: string
  orgName: string
  displayName: string
  role: string
  joinedAt: number
}

export interface GetMembershipsResponse {
  requestId: string
  memberships: MembershipInfo[]
}

export interface WebhookResponse {
  message: string
  userID?: string
}

export interface RestErrorResponse {
  requestId: string
  error: {
    code: string
    message: string
  }
}

export interface RestResult<T> {
  status: number
  data?: T
  error?: RestErrorResponse
}

export class RestClient {
  private buildUserNameFromPhone(phone: string): string {
    const suffix = phone.slice(-8)
    return `user${suffix}`
  }

  async getTokenByCode(code: string): Promise<TokenResponse> {
    const res = await fetch(`${API_BASE_URL}/api/auth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
    if (!res.ok) {
      throw new Error(`Auth failed: ${res.status} ${await res.text()}`)
    }
    return res.json() as Promise<TokenResponse>
  }

  async register(
    phone: string,
    password: string,
    userName?: string
  ): Promise<RestResult<RegisterResponse>> {
    const normalizedUserName = userName ?? this.buildUserNameFromPhone(phone)
    const res = await fetch(`${API_BASE_URL}/api/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone, userName: normalizedUserName, password }),
    })
    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as RegisterResponse }
    }

    // 兼容旧后端：当调用方未显式传 userName 时，允许回退到仅 phone+password 的老协议
    if (!userName && res.status === 400) {
      const legacyRes = await fetch(`${API_BASE_URL}/api/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone, password }),
      })
      const legacyBody = await legacyRes.json()
      if (legacyRes.ok) {
        return { status: legacyRes.status, data: legacyBody as RegisterResponse }
      }
      return { status: legacyRes.status, error: legacyBody as RestErrorResponse }
    }

    return { status: res.status, error: body as RestErrorResponse }
  }

  async login(
    identifier: string,
    password: string,
    identifierType: 'PHONE' | 'USERNAME' = 'PHONE'
  ): Promise<RestResult<LoginResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ identifier, identifierType, password }),
    })
    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as LoginResponse }
    }

    // 兼容旧后端：手机号登录可回退到 { phone, password }
    if (identifierType === 'PHONE' && /^1[3-9]\d{9}$/.test(identifier) && res.status === 400) {
      const legacyRes = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: identifier, password }),
      })
      const legacyBody = await legacyRes.json()
      if (legacyRes.ok) {
        return { status: legacyRes.status, data: legacyBody as LoginResponse }
      }
      return { status: legacyRes.status, error: legacyBody as RestErrorResponse }
    }

    return { status: res.status, error: body as RestErrorResponse }
  }

  async refresh(refreshToken: string): Promise<RestResult<RefreshResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    })
    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as RefreshResponse }
    }
    return { status: res.status, error: body as RestErrorResponse }
  }

  async logout(refreshToken: string): Promise<RestResult<void>> {
    const res = await fetch(`${API_BASE_URL}/api/auth/logout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    })
    if (res.status === 204) {
      return { status: res.status }
    }
    const body = await res.json()
    return { status: res.status, error: body as RestErrorResponse }
  }

  async initOrganization(
    accessToken: string,
    displayName: string,
    organizationName?: string
  ): Promise<RestResult<InitOrgResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/org/init`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        displayName,
        ...(organizationName ? { organizationName } : {}),
      }),
    })

    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as InitOrgResponse }
    }
    return { status: res.status, error: body as RestErrorResponse }
  }

  async getUserMemberships(accessToken: string): Promise<RestResult<GetMembershipsResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/user/memberships`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })

    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as GetMembershipsResponse }
    }
    return { status: res.status, error: body as RestErrorResponse }
  }

  async handleAuthProviderWebhook(payload: Record<string, unknown>): Promise<RestResult<WebhookResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/webhook/auth_provider`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })

    const raw = await res.text()
    let body: unknown = null

    if (raw.length > 0) {
      try {
        body = JSON.parse(raw)
      } catch {
        return {
          status: res.status,
          error: {
            requestId: '',
            error: {
              code: 'NON_JSON_RESPONSE',
              message: `webhook returned non-JSON response: ${raw.slice(0, 200)}`,
            },
          },
        }
      }
    }

    if (res.ok) {
      return { status: res.status, data: (body ?? {}) as WebhookResponse }
    }

    return {
      status: res.status,
      error: (body as RestErrorResponse) ?? {
        requestId: '',
        error: {
          code: `HTTP_${res.status}`,
          message: 'webhook request failed',
        },
      },
    }
  }
}
