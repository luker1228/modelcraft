'use client'

// src/web/hooks/end-user-auth/useRequireEndUserAuth.ts
// 终端用户页面级守卫 hook（对称 hooks/auth/use-auth.ts 的 useRequireAuth）
// middleware 已保证 end_user_refresh_token cookie 存在才能进入 /data 路由，
// 但页面刷新后 in-memory token 会丢失，需要 silent refresh 恢复。

import { useEffect, useState, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import {
  refreshEndUserAccessToken,
  fetchAndCacheEndUserInfo,
  removeEndUserSession,
} from '@api-client/end-user/public'
import type { EndUserInfo } from '@/types/end-user-auth'

interface UseRequireEndUserAuthReturn {
  /** 是否正在加载（恢复 session 中） */
  isLoading: boolean
}

/**
 * 页面级守卫 hook。
 * middleware 已保证 end_user_refresh_token cookie 存在才能进入此 layout，
 * 但页面刷新后 in-memory token 会丢失，需要 silent refresh 恢复。
 */
export function useRequireEndUserAuth(): UseRequireEndUserAuthReturn {
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()
  const params = useParams<{ orgName: string; projectSlug: string }>()

  // 提取参数避免 useEffect 依赖问题
  const orgName = params.orgName
  const projectSlug = params.projectSlug

  const restoreSession = useCallback(async () => {
    const store = useEndUserAuthStore.getState()
    const { accessToken, isTokenExpired } = store

    // token 存在且未过期：直接放行
    if (accessToken && !isTokenExpired()) {
      // userInfo 可能因页面刷新丢失，异步补充（不阻塞渲染）
      if (!store.userInfo) {
        void fetchAndCacheEndUserInfo()
      }
      setIsLoading(false)
      return
    }

    // token 不存在或已过期：尝试 silent refresh
    const newToken = await refreshEndUserAccessToken({
      orgName,
      projectSlug,
    })
    if (!newToken) {
      // refresh 失败（cookie 过期/revoked），重定向到登录页
      const loginUrl = `/end-user/${orgName}/login`
      const currentPath = typeof window !== 'undefined' ? window.location.pathname : ''
      router.replace(`${loginUrl}?redirect=${encodeURIComponent(currentPath)}`)
      return
    }

    // refresh 成功：填充 userInfo 供右上角展示
    void fetchAndCacheEndUserInfo()
    setIsLoading(false)
  }, [orgName, projectSlug, router])

  useEffect(() => {
    void restoreSession()
  }, [restoreSession])

  return { isLoading }
}

interface UseEndUserReturn {
  /** 当前终端用户信息 */
  user: EndUserInfo | null
  /** 登出函数 */
  logout: () => Promise<void>
}

/**
 * 获取当前终端用户信息 + logout。
 * 不含守卫逻辑，适合右上角 UserMenu 等非守卫场景。
 */
export function useEndUser(): UseEndUserReturn {
  const userInfo = useEndUserAuthStore((s) => s.userInfo)
  const router = useRouter()
  const params = useParams<{ orgName: string }>()

  const orgName = params.orgName

  const logout = useCallback(async () => {
    // 调用 BFF logout（best-effort）
    await fetch(`/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    }).catch(() => {
      // 静默忽略错误
    })

    // 清除客户端 session
    removeEndUserSession()

    // 重定向到登录页
    router.replace(`/end-user/${orgName}/login`)
  }, [orgName, router])

  return { user: userInfo, logout }
}
