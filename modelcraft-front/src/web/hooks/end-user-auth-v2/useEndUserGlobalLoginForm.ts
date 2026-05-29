'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'
import { getEndUserInfoFromToken } from '@api-client/end-user/end-user-auth-client'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

interface EndUserGlobalLoginResponse {
  error?: { code?: string; message?: string }
  accessToken?: string
  expiresAt?: string
  orgName?: string
  projects?: EndUserAccessibleProject[]
}

const globalLoginSchema = z.object({
  username: z.string().min(1, '请输入用户名').max(64, '用户名最多 64 个字符'),
  password: z.string().min(1, '请输入密码').max(128, '密码最多 128 个字符'),
})

export type EndUserGlobalLoginFormValues = z.infer<typeof globalLoginSchema>

export interface UseEndUserGlobalLoginFormReturn {
  form: UseFormReturn<EndUserGlobalLoginFormValues>
  onSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  isLoading: boolean
  error: string | null
}

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

export function useEndUserGlobalLoginForm(
  redirectTo?: string | null
): UseEndUserGlobalLoginFormReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)
  const setUserInfo = useEndUserAuthStore((s) => s.setUserInfo)

  const form = useForm<EndUserGlobalLoginFormValues>({
    resolver: zodResolver(globalLoginSchema),
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = useCallback(
    (e?: React.BaseSyntheticEvent) =>
      form.handleSubmit(async (values) => {
        setIsLoading(true)
        setError(null)

        try {
          const res = await fetch('/api/bff/end-user/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({
              username: values.username,
              password: values.password,
            }),
          })

          const data = (await res.json()) as EndUserGlobalLoginResponse

          if (!res.ok) {
            setError(getErrorMessage(data.error?.code, data.error?.message))
            return
          }

          const tokenInfo = data.accessToken ? getEndUserInfoFromToken(data.accessToken) : null
          const orgName = data.orgName ?? tokenInfo?.orgName
          if (!orgName) {
            setError('登录成功，但缺少组织信息，请稍后重试')
            return
          }

          if (!data.accessToken) {
            router.push(`/end-user/${orgName}/no-project-access`)
            return
          }

          let expiresIn = 3600
          if (data.expiresAt) {
            const ms = new Date(data.expiresAt).getTime() - Date.now()
            if (ms > 0) expiresIn = Math.floor(ms / 1000)
          }

          setAccessToken(data.accessToken, expiresIn)
          setUserInfo({
            id: tokenInfo?.id ?? '',
            username: values.username,
            orgName,
            projectSlug: tokenInfo?.projectSlug ?? '',
          })

          sessionStorage.setItem(`eu_token_${orgName}`, data.accessToken)
          sessionStorage.setItem(
            `eu_token_expires_at_${orgName}`,
            String(Date.now() + expiresIn * 1000)
          )
          sessionStorage.setItem(
            `eu_accessible_projects_${orgName}`,
            JSON.stringify(data.projects ?? [])
          )

          const safeRedirect =
            redirectTo && redirectTo.startsWith(`/end-user/${orgName}/`)
              ? redirectTo
              : `/end-user/${orgName}/workspace`

          router.push(safeRedirect)
        } catch {
          setError('网络错误，请检查连接后重试')
        } finally {
          setIsLoading(false)
        }
      })(e),
    [form, redirectTo, router, setAccessToken, setUserInfo]
  )

  return { form, onSubmit, isLoading, error }
}
