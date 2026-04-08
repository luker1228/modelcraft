// Public API for Web layer to access auth utilities
export {
  removeToken,
  getToken,
  getUserInfoFromToken,
  isAuthenticated,
  isTokenNearExpiry,
  refreshAccessToken,
} from './auth-client'
export type { AuthUser } from '@/types/auth'
export { getOrgPath, getWelcomePath, getCurrentOrgName } from './token-utils'
