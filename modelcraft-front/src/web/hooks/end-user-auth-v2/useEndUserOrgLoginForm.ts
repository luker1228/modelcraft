'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

// src/web/hooks/end-user-auth-v2/useEndUserOrgLoginForm.ts
// Org 级终端用户登录表单 hook（EndUser v2）
// 封装 react-hook-form + zod 校验 + v2 BFF 调用 + 登录分支处理

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

// ============================================================================
// BFF Response Types
// ============================================================================

interface EndUserOrgLoginBffResponse {
  error?: { code?: string; message?: string }
  accessToken?: string
  expiresAt?: string
  projects?: EndUserAccessibleProject[]
  refreshToken?: string
}

// ============================================================================
// Zod Schema
// ============================================================================

const orgLoginSchema = z.object({
  username: z.string().min(1, '请输入用户名').max(64, '用户名最多 64 个字符'),
  password: z.string().min(1, '请输入密码').max(128, '密码最多 128 个字符'),
})

export type EndUserOrgLoginFormValues = z.infer<typeof orgLoginSchema>

// ============================================================================
// Hook Interface
// ============================================================================

export interface UseEndUserOrgLoginFormReturn {
  form: UseFormReturn<EndUserOrgLoginFormValues>
  onSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  isLoading: boolean
  error: string | null
}

// ============================================================================
// Error Messages
// ============================================================================

const ERROR_MESSAGES: Record<string, string> = {
  INVALID_CREDENTIALS: '用户名或密码错误',
  ACCOUNT_DISABLED: '账号已被禁用，请联系管理员',
  NO_PROJECT_ACCESS: '当前账号暂无任何可访问的项目权限，请联系管理员',
  PARAM_INVALID: '请求参数无效',
}

function getErrorMessage(code?: string, fallback?: string): string {
  if (code && ERROR_MESSAGES[code]) return ERROR_MESSAGES[code]
  return fallback ?? '登录失败，请稍后重试'
}

// ============================================================================
// JWT Utilities
// ============================================================================

/** 解析 JWT payload（不验签，仅读取字段） */
function parseJwtPayload(token: string): Record<string, unknown> {
  try {
    const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64)) as Record<string, unknown>
  } catch {
    return {}
  }
}

// ============================================================================
// Hook Implementation
// ============================================================================

/**
 * Org 级终端用户登录表单 hook。
 *
 * 登录分支：
 * - 登录失败（4xx）  → 显示错误信息（error 字段）
 * - 登录成功 + 有项目权限 → 写入 store/sessionStorage，跳转 workspace
 * - 登录成功 + 无项目权限 → 后端不返回 accessToken，跳转 no-project-access 引导页
 * - 网络异常          → 显示网络错误提示
 */
export function useEndUserOrgLoginForm(orgName: string): UseEndUserOrgLoginFormReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)

  const form = useForm<EndUserOrgLoginFormValues>({
    resolver: zodResolver(orgLoginSchema),
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = useCallback(
    (e?: React.BaseSyntheticEvent) => form.handleSubmit(async (values) => {
      setIsLoading(true)
      setError(null)

      try {
        const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({
            orgName,
            username: values.username,
            password: values.password,
          }),
        })

        const data = (await res.json()) as EndUserOrgLoginBffResponse

        if (!res.ok) {
          setError(getErrorMessage(data.error?.code, data.error?.message))
          return
        }

        // 把 accessToken 写入 store 供后续 GraphQL 请求使用
        if (data.accessToken) {
          // expiresAt 是 ISO 8601，转换为 expiresIn 秒数
          let expiresIn = 3600
          if (data.expiresAt) {
            const ms = new Date(data.expiresAt).getTime() - Date.now()
            if (ms > 0) expiresIn = Math.floor(ms / 1000)
          }
          setAccessToken(data.accessToken, expiresIn)
          // 同时写入 sessionStorage，防止客户端导航后 Zustand store 内存被重置
          sessionStorage.setItem(`eu_token_${orgName}`, data.accessToken)
          sessionStorage.setItem(`eu_token_expires_at_${orgName}`, String(Date.now() + expiresIn * 1000))

          // 管理员从 end-user 入口登录 → 跳转管理后台
          const payload = parseJwtPayload(data.accessToken)
          if (payload.is_admin === true) {
            router.push(`/org/${orgName}/projects`)
            return
          }
        } else {
          // 无 accessToken = 登录成功但当前账号没有任何项目访问权限
          // 跳转到"暂无权限"引导页，而不是 workspace
          router.push(`/end-user/${orgName}/no-project-access`)
          return
        }

        const projects: EndUserAccessibleProject[] = data.projects ?? []

        // 写入可访问项目列表，跳转 dashboard
        sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
        router.push(`/end-user/${orgName}/dashboard`)
      } catch {
        setError('网络错误，请检查连接后重试')
      } finally {
        setIsLoading(false)
      }
    })(e),
    [form, orgName, router, setAccessToken]
  )

  return { form, onSubmit, isLoading, error }
}
