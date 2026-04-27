// src/bff/end-user/end-user-auth-client.ts
// 终端用户客户端侧认证工具（对称 bff/auth/auth-client.ts）
// 提供 getEndUserToken、refreshEndUserAccessToken、fetchAndCacheEndUserInfo 等客户端工具

import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import type { EndUserInfo, EndUserJWTPayload, EndUserAuthResponse, EndUserMeResponse } from '@/types/end-user-auth'

export type { EndUserInfo }

interface EndUserRefreshParams {
  orgName: string
  projectSlug: string
}

// ============================================================================
// JWT 解析工具
// ============================================================================

/**
 * Decode JWT token (without verification - use for client-side display only)
 */
function decodeEndUserJWT(token: string): EndUserJWTPayload | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload) as EndUserJWTPayload
  } catch {
    return null
  }
}

// ============================================================================
// Token 操作
// ============================================================================

/**
 * 从 in-memory store 获取 end-user access token
 */
export function getEndUserToken(): string | null {
  return useEndUserAuthStore.getState().accessToken
}

/**
 * 清除 end-user session（token + userInfo）
 */
export function removeEndUserSession(): void {
  useEndUserAuthStore.getState().clearSession()
}

// 防并发刷新
let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

/**
 * 使用 Cookie 中的 end_user_refresh_token 做 silent refresh。
 * 并发调用共享同一个请求。
 * @returns 新的 access token 或 null（刷新失败）
 */
export async function refreshEndUserAccessToken(
  params?: EndUserRefreshParams
): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) {
    return _refreshPromise
  }

  _isRefreshing = true
  _refreshPromise = (async () => {
    try {
      const res = await fetch(`/end-user/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          orgName: params?.orgName,
          projectSlug: params?.projectSlug,
        }),
        credentials: 'include',
      })

      if (!res.ok) {
        useEndUserAuthStore.getState().clearSession()
        return null
      }

      const data = (await res.json()) as EndUserAuthResponse
      if (data.accessToken && data.expiresIn) {
        const store = useEndUserAuthStore.getState()
        store.setAccessToken(data.accessToken, data.expiresIn)

        // 从 JWT 解析 org/project 上下文并缓存到 userInfo（username 待 /me 填充）
        const decoded = decodeEndUserJWT(data.accessToken)
        if (decoded) {
          store.setUserInfo({
            id: decoded.sub,
            username: '', // 待 fetchAndCacheEndUserInfo 填充
            orgName: decoded.org_name,
            projectSlug: decoded.project_slug,
          })
        }

        return data.accessToken
      }

      return null
    } catch {
      return null
    } finally {
      _isRefreshing = false
      _refreshPromise = null
    }
  })()

  return _refreshPromise
}

/**
 * 获取终端用户信息并填充 store.userInfo。
 * 在 silent refresh 成功后调用，确保 username 可用于右上角展示。
 * @returns EndUserInfo 或 null（获取失败）
 */
export async function fetchAndCacheEndUserInfo(): Promise<EndUserInfo | null> {
  try {
    const token = getEndUserToken()
    if (!token) return null

    const res = await fetch(`/end-user/auth/me`, {
      credentials: 'include',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!res.ok) return null

    const data = (await res.json()) as EndUserMeResponse
    const store = useEndUserAuthStore.getState()

    // 从当前 store 中读取 org/project 上下文
    const currentInfo = store.userInfo

    const info: EndUserInfo = {
      id: data.id,
      username: data.username,
      orgName: currentInfo?.orgName ?? '',
      projectSlug: currentInfo?.projectSlug ?? '',
    }

    store.setUserInfo(info)
    return info
  } catch {
    return null
  }
}

/**
 * 检查是否已认证（token 存在且未过期）
 */
export function isEndUserAuthenticated(): boolean {
  const { accessToken, isTokenExpired } = useEndUserAuthStore.getState()
  return !!accessToken && !isTokenExpired()
}

/**
 * 从 token 解析用户信息（仅用于客户端显示，无验证）
 */
export function getEndUserInfoFromToken(token: string): Partial<EndUserInfo> | null {
  const decoded = decodeEndUserJWT(token)
  if (!decoded) return null

  return {
    id: decoded.sub,
    orgName: decoded.org_name,
    projectSlug: decoded.project_slug,
  }
}
