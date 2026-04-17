// src/bff/end-user/public.ts
// Public API for Web layer to access end-user auth utilities
// 对称 bff/auth/public.ts，仅导出客户端安全的函数

export {
  getEndUserToken,
  removeEndUserSession,
  refreshEndUserAccessToken,
  fetchAndCacheEndUserInfo,
  isEndUserAuthenticated,
  getEndUserInfoFromToken,
} from './end-user-auth-client'

export type { EndUserInfo } from '@/types/end-user-auth'
