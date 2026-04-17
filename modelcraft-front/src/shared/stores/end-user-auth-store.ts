// src/shared/stores/end-user-auth-store.ts
// 终端用户认证状态 Store（对称 auth-store.ts）
// 关键差异：额外 userInfo 字段（username 不在 JWT 中，需从 /me 填充）；clearSession 同时清 token 和 userInfo

import { create } from 'zustand'
import type { EndUserInfo } from '@/types/end-user-auth'

interface EndUserAuthState {
  // Token 状态
  accessToken: string | null
  expiresAt: number | null // Unix timestamp（毫秒）

  // 用户信息（login 后从 /me 接口获取，可延迟填充）
  userInfo: EndUserInfo | null

  // Actions
  setAccessToken: (token: string, expiresIn: number) => void
  setUserInfo: (info: EndUserInfo) => void
  clearSession: () => void // 同时清 token 和 userInfo
  isTokenExpired: () => boolean
}

export const useEndUserAuthStore = create<EndUserAuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,
  userInfo: null,

  setAccessToken: (token: string, expiresIn: number) => {
    set({
      accessToken: token,
      expiresAt: Date.now() + expiresIn * 1000,
    })
  },

  setUserInfo: (info: EndUserInfo) => set({ userInfo: info }),

  clearSession: () =>
    set({
      accessToken: null,
      expiresAt: null,
      userInfo: null,
    }),

  isTokenExpired: () => {
    const { expiresAt } = get()
    if (!expiresAt) return true
    // 提前 5 分钟触发 silent refresh（与开发者策略一致）
    return Date.now() > expiresAt - 5 * 60 * 1000
  },
}))
