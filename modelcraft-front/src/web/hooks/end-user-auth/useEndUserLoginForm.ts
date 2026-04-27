'use client'

// src/web/hooks/end-user-auth/useEndUserLoginForm.ts
// 终端用户登录表单 hook（对称 hooks/auth/use-auth-form.ts）
// 封装 react-hook-form + zod 校验 + BFF 调用 + 错误映射

import { useState, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { fetchAndCacheEndUserInfo, getEndUserInfoFromToken } from '@api-client/end-user/public'
import { mapEndUserErrorCode, type EndUserAuthResponse, type EndUserBffError } from '@/types/end-user-auth'

// ============================================================================
// Zod Schema
// ============================================================================

const endUserLoginSchema = z.object({
  username: z.string().min(1, '请输入用户名').max(64, '用户名最多 64 个字符'),
  password: z.string().min(1, '请输入密码').max(128, '密码最多 128 个字符'),
})

export type EndUserLoginFormValues = z.infer<typeof endUserLoginSchema>

// ============================================================================
// Hook Interface
// ============================================================================

interface UseEndUserLoginFormReturn {
  /** react-hook-form 表单实例 */
  form: UseFormReturn<EndUserLoginFormValues>
  /** 表单提交处理函数（绑定到 form.handleSubmit） */
  onSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  /** 是否正在提交 */
  isLoading: boolean
  /** 错误提示信息 */
  error: string | null
}

// ============================================================================
// Hook Implementation
// ============================================================================

/**
 * 终端用户登录表单 hook。
 * @param orgName - 组织名称
 * @param projectSlug - 项目标识
 * @returns 表单实例、提交函数、加载状态、错误信息
 */
export function useEndUserLoginForm(
  orgName: string,
  projectSlug: string
): UseEndUserLoginFormReturn {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<EndUserLoginFormValues>({
    resolver: zodResolver(endUserLoginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  const handleSubmit = useCallback(
    async (values: EndUserLoginFormValues) => {
      setIsLoading(true)
      setError(null)

      try {
        const res = await fetch(`/end-user/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            orgName,
            projectSlug,
            username: values.username,
            password: values.password,
          }),
          credentials: 'same-origin',
        })

        if (!res.ok) {
          const data = (await res.json()) as EndUserBffError
          const errorMessage = mapEndUserErrorCode(data.error?.code, res.status, data.error?.message)
          setError(errorMessage)

          // PRD 要求：错误后清空密码字段并自动聚焦
          form.resetField('password')
          form.setFocus('password')
          return
        }

        const data = (await res.json()) as EndUserAuthResponse
        const store = useEndUserAuthStore.getState()

        // 存储 access token
        store.setAccessToken(data.accessToken, data.expiresIn)

        // 从 JWT 解析用户信息并缓存
        const tokenInfo = getEndUserInfoFromToken(data.accessToken)
        if (tokenInfo) {
          store.setUserInfo({
            id: tokenInfo.id ?? '',
            username: '', // 待 fetchAndCacheEndUserInfo 填充
            orgName: tokenInfo.orgName ?? orgName,
            projectSlug: tokenInfo.projectSlug ?? projectSlug,
          })
        }

        // 异步填充 userInfo（不阻塞跳转）：登录后立即从 /me 接口获取 username，
        // 供数据管理页右上角展示；失败不影响主流程。
        void fetchAndCacheEndUserInfo()

        // 跳转：优先使用 ?redirect= 参数，否则跳转数据管理落地页
        const redirect = searchParams.get('redirect')
        const target = redirect ?? `/u/${orgName}/${projectSlug}/data`
        router.replace(target)
      } catch {
        setError('登录服务暂时不可用，请稍后重试')
      } finally {
        setIsLoading(false)
      }
    },
    [orgName, projectSlug, router, searchParams, form]
  )

  const onSubmit = form.handleSubmit(handleSubmit)

  return {
    form,
    onSubmit,
    isLoading,
    error,
  }
}
