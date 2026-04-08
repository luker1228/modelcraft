'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { refreshAccessToken } from '@bff/auth/public'
import { useAuthStore } from '@shared/stores/auth-store'

const PUBLIC_ROUTES = ['/login', '/register', '/auth/callback']

/**
 * AuthProvider — periodic silent token refresh only.
 * Route-level auth guarding is handled by middleware.ts.
 */
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()

  useEffect(() => {
    if (PUBLIC_ROUTES.includes(pathname)) return

    const checkAndRefresh = async () => {
      const { accessToken, isTokenExpired } = useAuthStore.getState()
      if (!accessToken) return
      if (isTokenExpired()) {
        const newToken = await refreshAccessToken()
        if (!newToken) {
          useAuthStore.getState().clearAccessToken()
          router.push('/login')
        }
      }
    }

    const timer = setInterval(checkAndRefresh, 60_000)
    return () => clearInterval(timer)
  }, [pathname, router])

  return <>{children}</>
}
