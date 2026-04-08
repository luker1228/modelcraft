/** BFF 登录请求 */
export interface LoginRequest {
  phone: string
  password: string
}

/** BFF 登录响应 */
export interface LoginResponse {
  accessToken: string
  expiresIn: number
}

/** BFF 注册请求 */
export interface RegisterRequest {
  phone: string
  password: string
}

/** BFF 注册响应 */
export interface RegisterResponse {
  success: boolean
}

/** JWT 解析后的用户信息 */
export interface AuthUser {
  id: string
  phone: string
  name: string
}
