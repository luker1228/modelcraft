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
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

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
// Hook Implementation
// ============================================================================

/**
 * Org 级终端用户登录表单 hook。
 *
 * 登录分支：
 * - singleProject: true  → 直接写入 store，跳转数据页
 * - singleProject: false + noProjectAccess=true → 跳转待授权页
 * - singleProject: false → 跳转选择 Project 页（accessible projects 存入 sessionStorage）
 * - error                → 显示错误信息
 */
export function useEndUserOrgLoginForm(orgName: string): UseEndUserOrgLoginFormReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const setEndUserToken = useEndUserAuthStore((s) => s.setAccessToken)

  const form = useForm<EndUserOrgLoginFormValues>({
    resolver: zodResolver(orgLoginSchema),
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = useCallback(
    form.handleSubmit(async (values) => {
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

        const projects: EndUserAccessibleProject[] = data.projects ?? []

        if (projects.length === 0) {
          router.push(`/end-user/${orgName}/no-project-access`)
          return
        }

        if (data.refreshToken) {
          sessionStorage.setItem(`eu_refresh_token_${orgName}`, data.refreshToken)
        }

        if (projects.length === 1) {
          const projectSlug = projects[0].slug
          sessionStorage.setItem(`eu_selected_project_${orgName}`, projectSlug)
          const expiresIn = data.expiresAt
            ? Math.max(1, Math.floor((new Date(data.expiresAt).getTime() - Date.now()) / 1000))
            : 3600
          setEndUserToken(data.accessToken ?? '', expiresIn)
          router.push(`/end-user/${orgName}/${projectSlug}/data`)
          return
        }

        sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
        router.push(`/end-user/${orgName}/select-project`)
      } catch {
        setError('网络错误，请检查连接后重试')
      } finally {
        setIsLoading(false)
      }
    }),
    [form, orgName, router, setEndUserToken]
  )

  return { form, onSubmit, isLoading, error }
}
