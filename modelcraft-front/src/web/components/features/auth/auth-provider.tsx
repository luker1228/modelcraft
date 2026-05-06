'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { refreshAccessToken } from '@api-client/auth/public'
import { TENANT_LOGIN_PATH, TENANT_REGISTER_PATH } from '@shared/constants/routes'
import { useAuthStore } from '@shared/stores/auth-store'

const PUBLIC_ROUTES = [TENANT_LOGIN_PATH, TENANT_REGISTER_PATH]
const END_USER_ROUTE_RE = /^\/u\/[^/]+\/[^/]+(\/|$)/

/**
 * AuthProvider — periodic silent token refresh only.
 * Route-level auth guarding is handled by middleware.ts.
 */
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()

  useEffect(() => {
    if (END_USER_ROUTE_RE.test(pathname)) return
    if (PUBLIC_ROUTES.includes(pathname)) return

    const checkAndRefresh = async () => {
      const { accessToken, isTokenExpired } = useAuthStore.getState()
      if (!accessToken) return
      if (isTokenExpired()) {
        const newToken = await refreshAccessToken()
        if (!newToken) {
          useAuthStore.getState().clearAccessToken()
          router.push(TENANT_LOGIN_PATH)
        }
      }
    }

    const timer = setInterval(checkAndRefresh, 60_000)
    return () => clearInterval(timer)
  }, [pathname, router])

  return <>{children}</>
}
