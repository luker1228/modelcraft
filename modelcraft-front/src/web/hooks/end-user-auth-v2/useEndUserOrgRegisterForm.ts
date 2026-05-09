'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

// src/web/hooks/end-user-auth-v2/useEndUserOrgRegisterForm.ts
// Org 级终端用户自注册 hook（EndUser v2）
// 注册成功后自动登录，并按可访问项目数量分流

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface EndUserOrgRegisterBffResponse {
  error?: { code?: string; message?: string }
  accessToken?: string
  expiresAt?: string
  projects?: EndUserAccessibleProject[]
  refreshToken?: string
}

const orgRegisterSchema = z.object({
  username: z
    .string()
    .min(3, '用户名至少 3 位')
    .max(64, '用户名最多 64 位')
    .regex(/^[a-zA-Z0-9_-]{3,64}$/, '用户名仅支持字母、数字、下划线和中划线'),
  password: z
    .string()
    .min(8, '密码至少 8 位')
    .max(128, '密码最多 128 位')
    .regex(/[a-zA-Z]/, '密码必须包含至少 1 个字母')
    .regex(/[0-9]/, '密码必须包含至少 1 个数字'),
  confirmPassword: z.string().min(1, '请再次输入密码'),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
})

export type EndUserOrgRegisterFormValues = z.infer<typeof orgRegisterSchema>

export interface UseEndUserOrgRegisterFormReturn {
  form: UseFormReturn<EndUserOrgRegisterFormValues>
  onSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  isLoading: boolean
  error: string | null
}

const ERROR_MESSAGES: Record<string, string> = {
  CONFLICT: '该用户名已被使用',
  PARAM_INVALID: '注册参数无效，请检查后重试',
  ACCOUNT_DISABLED: '账号已被禁用，请联系管理员',
  NO_PROJECT_ACCESS: '当前账号暂无任何可访问的项目权限，请联系管理员',
}

function getErrorMessage(code?: string, fallback?: string): string {
  if (code && ERROR_MESSAGES[code]) return ERROR_MESSAGES[code]
  return fallback ?? '注册失败，请稍后重试'
}

export function useEndUserOrgRegisterForm(orgName: string): UseEndUserOrgRegisterFormReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<EndUserOrgRegisterFormValues>({
    resolver: zodResolver(orgRegisterSchema),
    defaultValues: { username: '', password: '', confirmPassword: '' },
  })

  const onSubmit = useCallback(
    (e?: React.BaseSyntheticEvent) => form.handleSubmit(async (values) => {
      setIsLoading(true)
      setError(null)

      try {
        const registerRes = await fetch(`/api/bff/org/${orgName}/end-user/auth/register`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({
            orgName,
            username: values.username,
            password: values.password,
          }),
        })

        const data = (await registerRes.json()) as EndUserOrgRegisterBffResponse
        if (!registerRes.ok) {
          setError(getErrorMessage(data.error?.code, data.error?.message))
          form.resetField('password')
          form.resetField('confirmPassword')
          form.setFocus('password')
          return
        }

        const projects: EndUserAccessibleProject[] = data.projects ?? []

        // 无论 0/1/N 个 Project，均跳转 workspace
        sessionStorage.setItem(`eu_accessible_projects_${orgName}`, JSON.stringify(projects))
        router.push(`/end-user/${orgName}/workspace`)
      } catch {
        setError('网络错误，请检查连接后重试')
      } finally {
        setIsLoading(false)
      }
    })(e),
    [form, orgName, router]
  )

  return { form, onSubmit, isLoading, error }
}
