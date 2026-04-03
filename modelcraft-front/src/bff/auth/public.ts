// Public API for Web layer to access auth utilities
export {
  removeToken,
  getToken,
  getUserInfoFromToken,
  getOrgNameFromToken,
  isAuthenticated,
  isTokenNearExpiry,
  refreshAccessToken,
} from './casdoor'
export type { UserInfo } from './casdoor'
export { getOrgPath, getWelcomePath, getCurrentOrgName } from './token-utils'
