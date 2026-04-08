import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export interface TokenResponse {
  accessToken: string
  tokenType: string
}

export interface RegisterResponse {
  requestId: string
  userId: string
}

export interface LoginResponse {
  requestId: string
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface RefreshResponse {
  requestId: string
  userId: string
  refreshToken: string
  expiresAt: string
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

  async register(phone: string, password: string): Promise<RestResult<RegisterResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone, password }),
    })
    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as RegisterResponse }
    }
    return { status: res.status, error: body as RestErrorResponse }
  }

  async login(phone: string, password: string): Promise<RestResult<LoginResponse>> {
    const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone, password }),
    })
    const body = await res.json()
    if (res.ok) {
      return { status: res.status, data: body as LoginResponse }
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
}
