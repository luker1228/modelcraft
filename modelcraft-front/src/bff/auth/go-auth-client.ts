// src/bff/auth/go-auth-client.ts

const GO_BACKEND_INTERNAL_URL = process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'

export interface GoLoginResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface GoRefreshResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export class TokenReuseError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'TokenReuseError'
  }
}

// callGoLogin calls Go Backend /api/auth/login
// Pass user info extracted from Casdoor token
export async function callGoLogin(params: {
  externalId: string
  email: string
  name: string
}): Promise<GoLoginResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  if (!res.ok) throw new Error(`Go login failed: ${res.status}`)
  return res.json() as Promise<GoLoginResult>
}

// callGoRefresh calls Go Backend /api/auth/refresh (token rotation)
export async function callGoRefresh(refreshToken: string): Promise<GoRefreshResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  })
  if (res.status === 401) throw new TokenReuseError('Refresh token revoked or reused')
  if (!res.ok) throw new Error(`Go refresh failed: ${res.status}`)
  return res.json() as Promise<GoRefreshResult>
}

// callGoLogout calls Go Backend /api/auth/logout
export async function callGoLogout(refreshToken: string): Promise<void> {
  await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/logout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  }).catch(() => {
    // Ignore errors - Cookie will be cleared regardless
  })
}
