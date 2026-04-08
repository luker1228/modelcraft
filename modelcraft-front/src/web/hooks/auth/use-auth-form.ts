'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@shared/stores/auth-store'
import type { LoginFormValues, RegisterFormValues } from '@/shared/validation/auth'

interface UseLoginReturn {
  login: (values: LoginFormValues) => Promise<void>
  isLoading: boolean
  error: string | null
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

  const login = async (values: LoginFormValues) => {
    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/bff/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: values.phone, password: values.password }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error ?? '登录失败，请稍后重试')
        return
      }

      useAuthStore.getState().setAccessToken(data.accessToken, data.expiresIn)
      router.push('/org-selector')
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { login, isLoading, error }
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
        body: JSON.stringify({ phone: values.phone, password: values.password }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error ?? '注册失败，请稍后重试')
        return
      }

      router.push('/login')
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { register, isLoading, error }
}
