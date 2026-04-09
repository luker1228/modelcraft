'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@shared/stores/auth-store'
import type { LoginFormValues, RegisterFormValues } from '@/shared/validation/auth'
import type { LoginResponse, RegisterResponse, IdentifierType } from '@/types/auth'

interface UseLoginReturn {
  login: (values: LoginFormValues) => Promise<void>
  isLoading: boolean
  error: string | null
  identifierType: IdentifierType
  setIdentifierType: (type: IdentifierType) => void
}

interface UseRegisterReturn {
  register: (values: RegisterFormValues) => Promise<void>
  isLoading: boolean
  error: string | null
}

export function useLogin(): UseLoginReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [identifierType, setIdentifierType] = useState<IdentifierType>('PHONE')

  const login = async (values: LoginFormValues) => {
    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/bff/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          identifier: values.identifier,
          identifierType: values.identifierType,
          password: values.password,
        }),
      })

      const data = (await res.json()) as LoginResponse & { error?: string }

      if (!res.ok) {
        setError(data.error ?? '登录失败，请稍后重试')
        return
      }

      // 存储 access token
      useAuthStore.getState().setAccessToken(data.accessToken, data.expiresIn)

      // 跳转到 workspace
      router.push(`/org/${data.orgName}/workspace`)
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { login, isLoading, error, identifierType, setIdentifierType }
}

export function useRegister(): UseRegisterReturn {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const register = async (values: RegisterFormValues) => {
    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/bff/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phone: values.phone,
          userName: values.userName,
          password: values.password,
        }),
      })

      const data = (await res.json()) as RegisterResponse & { error?: string }

      if (!res.ok) {
        setError(data.error ?? '注册失败，请稍后重试')
        return
      }

      // 注册成功后直接跳转到 workspace（PRD 要求：注册成功后直接进入工作区）
      router.push(`/org/${data.orgName}/workspace`)
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { register, isLoading, error }
}
