'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, type UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { fetchAndCacheEndUserInfo, getEndUserInfoFromToken } from '@bff/end-user/public'
import { mapEndUserErrorCode, type EndUserAuthResponse, type EndUserBffError } from '@/types/end-user-auth'

const endUserRegisterSchema = z.object({
  username: z.string().min(1, '请输入用户名').max(64, '用户名最多 64 个字符'),
  password: z.string().min(6, '密码至少 6 位').max(128, '密码最多 128 个字符'),
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
        const res = await fetch('/api/bff/end-user/auth/register', {
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
          const errorMessage = mapEndUserErrorCode(data.error?.code, res.status)
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

        void fetchAndCacheEndUserInfo()
        router.replace(`/u/${orgName}/${projectSlug}/data`)
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
