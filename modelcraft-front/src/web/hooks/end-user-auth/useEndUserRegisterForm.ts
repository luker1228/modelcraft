'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { fetchAndCacheEndUserInfo, getEndUserInfoFromToken } from '@api-client/end-user/public'
import { mapEndUserErrorCode, type EndUserAuthResponse, type EndUserBffError } from '@/types/end-user-auth'

const endUserRegisterSchema = z.object({
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

export type EndUserRegisterFormValues = z.infer<typeof endUserRegisterSchema>

interface UseEndUserRegisterFormReturn {
  form: UseFormReturn<EndUserRegisterFormValues>
  onSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  isLoading: boolean
  error: string | null
}

export function useEndUserRegisterForm(
  orgName: string,
  projectSlug: string
): UseEndUserRegisterFormReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<EndUserRegisterFormValues>({
    resolver: zodResolver(endUserRegisterSchema),
    defaultValues: {
      username: '',
      password: '',
      confirmPassword: '',
    },
  })

  const handleSubmit = useCallback(
    async (values: EndUserRegisterFormValues) => {
      setIsLoading(true)
      setError(null)

      try {
        const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/register`, {
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
          form.resetField('password')
          form.resetField('confirmPassword')
          form.setFocus('password')
          return
        }

        const data = (await res.json()) as EndUserAuthResponse
        const store = useEndUserAuthStore.getState()

        store.setAccessToken(data.accessToken, data.expiresIn)

        const tokenInfo = getEndUserInfoFromToken(data.accessToken)
        if (tokenInfo) {
          store.setUserInfo({
            id: tokenInfo.id ?? '',
            username: '',
            orgName: tokenInfo.orgName ?? orgName,
            projectSlug: tokenInfo.projectSlug ?? projectSlug,
          })
        }

        void fetchAndCacheEndUserInfo(orgName)
        router.replace(`/end-user/${orgName}/${projectSlug}/data`)
      } catch {
        setError('注册服务暂时不可用，请稍后重试')
      } finally {
        setIsLoading(false)
      }
    },
    [orgName, projectSlug, router, form]
  )

  return {
    form,
    onSubmit: form.handleSubmit(handleSubmit),
    isLoading,
    error,
  }
}
