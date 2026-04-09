// src/bff/auth/auth-client.ts
// Pure JWT utility functions extracted from casdoor.ts — no Casdoor SDK dependency.

import { useAuthStore } from '@shared/stores/auth-store'
import type { AuthUser } from '@/types/auth'

export type { AuthUser }

/** JWT payload structure for our self-signed access tokens */
interface JWTPayload {
  exp?: number
  user_id?: string
  sub?: string
  phone?: string
  name?: string
  user_name?: string
  userName?: string
}

/**
 * Decode JWT token (without verification - use for client-side display only)
 */
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

/**
 * Check if token is expired
 */
export function isTokenExpired(token: string): boolean {
  const decoded = decodeJWT(token)
  if (!decoded || !decoded.exp) {
    return true
  }

  const currentTime = Math.floor(Date.now() / 1000)
  return decoded.exp < currentTime
}

/**
 * Check if token expires within the given threshold (seconds).
 * Default threshold is 300 seconds (5 minutes).
 */
export function isTokenNearExpiry(token: string, thresholdSeconds = 300): boolean {
  const decoded = decodeJWT(token)
  if (!decoded || !decoded.exp) {
    return true
  }
  const currentTime = Math.floor(Date.now() / 1000)
  return decoded.exp - currentTime < thresholdSeconds
}

/**
 * Get user info from JWT token
 */
export function getUserInfoFromToken(token: string): AuthUser | null {
  const decoded = decodeJWT(token)
  if (!decoded) {
    return null
  }

  return {
    id: decoded.user_id || decoded.sub || '',
    phone: decoded.phone || '',
    name: decoded.name || '',
    userName: decoded.user_name || decoded.userName,
  }
}

/**
 * Get JWT token from in-memory store
 */
export function getToken(): string | null {
  return useAuthStore.getState().accessToken
}

/**
 * Remove JWT token (logout)
 */
export function removeToken(): void {
  useAuthStore.getState().clearAccessToken()
}

// Prevent concurrent refresh requests
let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

/**
 * Refresh the access token using the httpOnly Cookie refresh token.
 * Returns the new access token, or null if refresh failed.
 * Concurrent calls will share the same refresh request.
 */
export async function refreshAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) {
    return _refreshPromise
  }

  _isRefreshing = true
  _refreshPromise = (async () => {
    try {
      // No body needed - refresh token is sent automatically via httpOnly Cookie
      const response = await fetch('/api/bff/auth/refresh', {
        method: 'POST',
        credentials: 'same-origin',
      })

      if (!response.ok) {
        // Refresh token invalid or expired - clear access token
        useAuthStore.getState().clearAccessToken()
        return null
      }

      const data = (await response.json()) as { accessToken?: string; expiresIn?: number }
      const newAccessToken: string | undefined = data.accessToken
      const expiresIn: number | undefined = data.expiresIn

      if (newAccessToken && expiresIn) {
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

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  const { accessToken, isTokenExpired: checkExpired } = useAuthStore.getState()
  if (!accessToken) {
    return false
  }
  return !checkExpired()
}
