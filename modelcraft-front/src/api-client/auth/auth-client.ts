// src/api-client/auth/auth-client.ts
// JWT utility functions for client-side authentication — token decode, expiry check, refresh.

import { useAuthStore } from '@shared/stores/auth-store'
import type { AuthUser } from '@/types/auth'

export type { AuthUser }

/** JWT payload structure for gateway-issued access tokens */
interface JWTPayload {
  exp?: number
  sub?: string      // userID (gateway JWT uses sub)
  user_id?: string  // legacy field
  phone?: string
  name?: string
  userName?: string
  username?: string // gateway field
}

export function decodeJWT(token: string): JWTPayload | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join(''),
    )
    return JSON.parse(jsonPayload) as JWTPayload
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const decoded = decodeJWT(token)
  if (!decoded || !decoded.exp) return true
  return decoded.exp < Math.floor(Date.now() / 1000)
}

export function isTokenNearExpiry(token: string, thresholdSeconds = 300): boolean {
  const decoded = decodeJWT(token)
  if (!decoded || !decoded.exp) return true
  return decoded.exp - Math.floor(Date.now() / 1000) < thresholdSeconds
}

export function getUserInfoFromToken(token: string): AuthUser | null {
  const decoded = decodeJWT(token)
  if (!decoded) return null
  return {
    id: decoded.sub || decoded.user_id || '',
    phone: decoded.phone || '',
    name: decoded.name || '',
    userName: decoded.username || decoded.userName || '',
  }
}

export function getToken(): string | null {
  return useAuthStore.getState().accessToken
}

export function removeToken(): void {
  useAuthStore.getState().clearAccessToken()
}

let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

/**
 * Refresh the access token by calling the gateway /auth/refresh endpoint.
 * The httpOnly Cookie refresh token is sent automatically by the browser.
 * Concurrent calls share the same in-flight request.
 */
export async function refreshAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) {
    return _refreshPromise
  }

  _isRefreshing = true
  _refreshPromise = (async () => {
    try {
      const response = await fetch(`/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
      })

      if (!response.ok) {
        useAuthStore.getState().clearAccessToken()
        return null
      }

      const data = (await response.json()) as { accessToken?: string }
      const newAccessToken = data.accessToken

      if (newAccessToken) {
        const decoded = decodeJWT(newAccessToken)
        const expiresIn = decoded?.exp
          ? decoded.exp - Math.floor(Date.now() / 1000)
          : 3600
        useAuthStore.getState().setAccessToken(newAccessToken, expiresIn)
        return newAccessToken
      }

      return null
    } catch {
      return null
    } finally {
      _isRefreshing = false
      _refreshPromise = null
    }
  })()

  return _refreshPromise
}

export function isAuthenticated(): boolean {
  const { accessToken, isTokenExpired: checkExpired } = useAuthStore.getState()
  if (!accessToken) return false
  return !checkExpired()
}
