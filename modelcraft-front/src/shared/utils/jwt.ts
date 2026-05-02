/**
 * JWT 客户端解析工具
 *
 * 仅用于客户端展示（无签名验证）。
 * 开发者 token 和终端用户 token 共用同一套 decode 实现。
 */

/**
 * 解析 JWT payload，返回泛型对象。
 * 失败时返回 null（格式错误、base64 解码失败等）。
 */
export function decodeJWTPayload<T = Record<string, unknown>>(token: string): T | null {
  try {
    const base64Url = token.split('.')[1]
    if (!base64Url) return null
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload) as T
  } catch {
    return null
  }
}

/**
 * 检查 token 是否已过期（基于 exp claim，单位秒）。
 */
export function isJWTExpired(token: string): boolean {
  const payload = decodeJWTPayload<{ exp?: number }>(token)
  if (!payload?.exp) return true
  return payload.exp < Math.floor(Date.now() / 1000)
}

/**
 * 检查 token 是否即将过期（默认 5 分钟内）。
 */
export function isJWTNearExpiry(token: string, thresholdSeconds = 300): boolean {
  const payload = decodeJWTPayload<{ exp?: number }>(token)
  if (!payload?.exp) return true
  return payload.exp - Math.floor(Date.now() / 1000) < thresholdSeconds
}
