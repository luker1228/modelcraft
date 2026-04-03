import { create } from 'zustand'

interface AuthState {
  accessToken: string | null
  expiresAt: number | null // Unix timestamp (milliseconds)
  setAccessToken: (token: string, expiresIn: number) => void
  clearAccessToken: () => void
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,

  setAccessToken: (token: string, expiresIn: number) => {
    set({
      accessToken: token,
      expiresAt: Date.now() + expiresIn * 1000,
    })
  },

  clearAccessToken: () => set({ accessToken: null, expiresAt: null }),

  isTokenExpired: () => {
    const { expiresAt } = get()
    if (!expiresAt) return true
    // Treat as expired 5 minutes early to trigger refresh
    return Date.now() > expiresAt - 5 * 60 * 1000
  },
}))
