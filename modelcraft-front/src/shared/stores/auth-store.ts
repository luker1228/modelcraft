import { create } from 'zustand'
import { persist } from 'zustand/middleware'

function parseIsAdmin(token: string): boolean | null {
  try {
    const payload: unknown = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    if (typeof payload === 'object' && payload !== null && 'is_admin' in payload) {
      return (payload as { is_admin: unknown }).is_admin === true
    }
    return null
  } catch {
    return null
  }
}

interface AuthState {
  accessToken: string | null
  expiresAt: number | null // Unix timestamp (milliseconds)
  isAdmin: boolean | null
  setAccessToken: (token: string, expiresIn: number) => void
  clearAccessToken: () => void
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      expiresAt: null,
      isAdmin: null,

      setAccessToken: (token: string, expiresIn: number) => {
        set({
          accessToken: token,
          expiresAt: Date.now() + expiresIn * 1000,
          isAdmin: parseIsAdmin(token),
        })
      },

      clearAccessToken: () => set({ accessToken: null, expiresAt: null, isAdmin: null }),

      isTokenExpired: () => {
        const { expiresAt } = get()
        if (!expiresAt) return true
        // Treat as expired 5 minutes early to trigger refresh
        return Date.now() > expiresAt - 5 * 60 * 1000
      },
    }),
    {
      name: 'modelcraft-auth-store',
      partialize: (state) => ({
        accessToken: state.accessToken,
        expiresAt: state.expiresAt,
        isAdmin: state.isAdmin,
      }),
    }
  )
)
