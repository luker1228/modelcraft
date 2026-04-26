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
      const res = await fetch(`${process.env.NEXT_PUBLIC_GATEWAY_URL ?? ''}/auth/login`, {
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

      // 记录默认组织和昵称并跳转到 workspace
      localStorage.setItem('defaultOrgName', data.orgName)
      localStorage.setItem('defaultUserName', data.userName)
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
      // Step 1: 注册
      const registerRes = await fetch(`${process.env.NEXT_PUBLIC_GATEWAY_URL ?? ''}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phone: values.phone,
          userName: values.userName,
          password: values.password,
        }),
      })

      const registerData = (await registerRes.json()) as RegisterResponse & { error?: string }

      if (!registerRes.ok) {
        setError(registerData.error ?? '注册失败，请稍后重试')
        return
      }

      // Step 2: 自动登录获取 accessToken（使用刚注册的手机号 + 密码）
      const loginRes = await fetch(`${process.env.NEXT_PUBLIC_GATEWAY_URL ?? ''}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          identifier: values.phone,
          identifierType: 'PHONE',
          password: values.password,
        }),
      })

      const loginData = (await loginRes.json()) as LoginResponse & { error?: string }

      if (!loginRes.ok) {
        // 注册成功但登录失败，跳转到登录页让用户手动登录
        setError('注册成功，请重新登录')
        router.push('/auth/login')
        return
      }

      // Step 3: 存储 access token
      useAuthStore.getState().setAccessToken(loginData.accessToken, loginData.expiresIn)

      // Step 4: 记录默认组织和昵称并跳转到 workspace
      localStorage.setItem('defaultOrgName', loginData.orgName)
      localStorage.setItem('defaultUserName', registerData.profile.nickname || values.userName)
      router.push(`/org/${loginData.orgName}/workspace`)
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { register, isLoading, error }
}
