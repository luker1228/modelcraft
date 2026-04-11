/** 登录标识符类型 */
export type IdentifierType = 'PHONE' | 'USERNAME'

// ============================================================================
// BFF 请求/响应类型（前端 ↔ BFF）
// ============================================================================

/** BFF 登录请求 */
export interface LoginRequest {
  identifier: string
  identifierType: IdentifierType
  password: string
}

/** BFF 登录响应 */
export interface LoginResponse {
  accessToken: string
  expiresIn: number
  userId: string
  userName: string
  orgName: string
}

/** BFF 注册请求 */
export interface RegisterRequest {
  phone: string
  userName: string
  password: string
}

/** 注册后返回的个人资料快照 */
export interface RegisterProfileSnapshot {
  id: string
  userId: string
  nickname: string
  avatarUrl?: string
  bio?: string
}

/** BFF 注册响应 */
export interface RegisterResponse {
  userId: string
  orgName: string
  profile: RegisterProfileSnapshot
}

// ============================================================================
// Go 后端请求/响应类型（BFF ↔ Go Backend）
// ============================================================================

/** Go 后端登录请求 */
export interface GoLoginRequest {
  identifier: string
  identifierType: IdentifierType
  password: string
}

/** Go 后端登录响应 */
export interface GoLoginResponse {
  requestId: string
  userId: string
  userName: string
  orgName: string
  refreshToken: string
  expiresAt: string
}

/** Go 后端注册请求 */
export interface GoRegisterRequest {
  phone: string
  userName: string
  password: string
}

/** Go 后端注册响应 */
export interface GoRegisterResponse {
  requestId: string
  userId: string
  orgName: string
  profile?: RegisterProfileSnapshot
}

// ============================================================================
// 错误类型
// ============================================================================

/** Go 后端错误结构 */
export interface GoAuthError {
  requestId: string
  error: {
    code: string
    message: string
  }
}

/** Auth 错误码 */
export type AuthErrorCode =
  | 'PARAM_INVALID.AUTH'
  | 'CONFLICT.USER'
  | 'AUTHENTICATION_FAILED'

// ============================================================================
// 其他
// ============================================================================

/** JWT 解析后的用户信息 */
export interface AuthUser {
  id: string
  phone: string
  name: string
  userName?: string
}
