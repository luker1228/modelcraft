'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
  error?: { code?: string }
}

/**
 * Ensures a valid end-user access token is available before the page renders.
 *
 * 1. Checks the in-memory zustand store first.
 * 2. Falls back to sessionStorage (saved from a previous page navigation).
 * 3. Falls back to a silent refresh via httpOnly cookie.
 * 4. Redirects to login if all sources fail.
 *
 * @returns true when a valid token is available (page should render).
 */
export function useEndUserTokenReady(orgName: string): boolean {
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)
  const router = useRouter()

  const [ready, setReady] = useState(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) return true

    if (typeof window !== 'undefined') {
      const savedToken = sessionStorage.getItem(`eu_token_${orgName}`)
      const savedExpiresAt = Number(sessionStorage.getItem(`eu_token_expires_at_${orgName}`) ?? '0')
      if (savedToken && Date.now() < savedExpiresAt - 5 * 60 * 1000) {
        const expiresIn = Math.floor((savedExpiresAt - Date.now()) / 1000)
        useEndUserAuthStore.getState().setAccessToken(savedToken, expiresIn)
        return true
      }
    }
    return false
  })

  useEffect(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) {
      setReady(true)
      return
    }

    void (async () => {
      try {
        const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/refresh`, {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ orgName }),
        })
        if (!res.ok) {
          router.replace(`/end-user/${orgName}/login`)
          return
        }
        const data = (await res.json()) as RefreshResponse
        if (data.accessToken) {
          let expiresIn = 3600
          if (data.expiresAt) {
            const ms = new Date(data.expiresAt).getTime() - Date.now()
            if (ms > 0) expiresIn = Math.floor(ms / 1000)
          }
          setAccessToken(data.accessToken, expiresIn)
          setReady(true)
        } else {
          router.replace(`/end-user/${orgName}/login`)
        }
      } catch {
        router.replace(`/end-user/${orgName}/login`)
      }
    })()
  }, [orgName, setAccessToken, router])

  return ready
}
