// src/bff/end-user/end-user-auth-client.ts
// 终端用户客户端侧认证工具（对称 bff/auth/auth-client.ts）
// 提供 getEndUserToken、refreshEndUserAccessToken、fetchAndCacheEndUserInfo 等客户端工具

import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { decodeJWTPayload } from '@shared/utils/jwt'
import type { EndUserInfo, EndUserJWTPayload, EndUserAuthResponse, EndUserMeResponse } from '@/types/end-user-auth'

export type { EndUserInfo }

export interface EndUserRefreshParams {
  orgName: string
  projectSlug?: string
}

// ============================================================================
// JWT 解析工具
// ============================================================================

function decodeEndUserJWT(token: string): EndUserJWTPayload | null {
  return decodeJWTPayload<EndUserJWTPayload>(token)
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
 * 使用 mc_enduser_refresh_token HttpOnly cookie 做 silent refresh。
 * 并发调用共享同一个请求。BFF refresh route 负责从 cookie 读取 refreshToken 并轮换。
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
      const orgName = params?.orgName ?? ''

      // cookie 由浏览器自动携带（HttpOnly），BFF 负责从 cookie 注入到后端请求 body
      const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgName }),
        credentials: 'include',
      })

      if (!res.ok) {
        useEndUserAuthStore.getState().clearSession()
        return null
      }

      const data = (await res.json()) as EndUserAuthResponse

      if (data.accessToken) {
        // 优先用 expiresAt（Go 后端返回 ISO 8601），退回到 expiresIn，再退回到默认 1h
        const expiresIn = data.expiresAt
          ? Math.max(1, Math.floor((new Date(data.expiresAt).getTime() - Date.now()) / 1000))
          : (data.expiresIn ?? 3600)

        const store = useEndUserAuthStore.getState()
        store.setAccessToken(data.accessToken, expiresIn)

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
 * @param orgName - 组织名称（用于构造正确的 BFF 路径）
 * @returns EndUserInfo 或 null（获取失败）
 */
export async function fetchAndCacheEndUserInfo(orgName: string): Promise<EndUserInfo | null> {
  try {
    const token = getEndUserToken()
    if (!token) return null

    const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/me`, {
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
