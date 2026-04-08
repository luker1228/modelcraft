import { useEffect, useState } from 'react'
import { getToken, getUserInfoFromToken, removeToken } from '@bff/auth/public'
import { refreshAccessToken } from '@bff/auth/public'
import type { AuthUser } from '@/types/auth'
import { useAuthStore } from '@shared/stores/auth-store'

/**
 * Ensures a valid access token is present in memory.
 *
 * Middleware already guarantees the refresh_token cookie exists before this
 * hook runs. So the only case where in-memory token is missing is after a
 * page refresh — we silently exchange the cookie for a new access token.
 *
 * No redirect logic here: middleware owns that responsibility.
 */
export function useRequireAuth() {
  const [isLoading, setIsLoading] = useState(true)
  const [user, setUser] = useState<AuthUser | null>(null)

  useEffect(() => {
    async function restoreSession() {
      let token = getToken()
      console.log('[useRequireAuth] In-memory token present:', !!token)

      if (!token) {
        console.log('[useRequireAuth] Attempting silent refresh...')
        token = await refreshAccessToken()
        console.log('[useRequireAuth] Silent refresh:', token ? 'success' : 'failed')
      }

      if (token) {
        setUser(getUserInfoFromToken(token))
      }

      setIsLoading(false)
    }

    restoreSession()
  }, [])

  return { isLoading, user }
}

/**
 * Get current user info without any redirect side effects.
 */
export function useUser() {
  const [user, setUser] = useState<AuthUser | null>(null)

  useEffect(() => {
    const token = getToken()
    if (token) setUser(getUserInfoFromToken(token))
  }, [])

  const logout = () => {
    removeToken()
    useAuthStore.getState().clearAccessToken()
    setUser(null)
  }

  return { user, isAuthenticated: !!user, logout }
}
