'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@shared/stores/auth-store'
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'
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
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({
          identifier: values.identifier,
          identifierType: values.identifierType,
          password: values.password,
        }),
      })

      const data = (await res.json()) as { accessToken?: string; error?: string; message?: string; orgName?: string; expiresIn?: number; userName?: string }

      if (!res.ok) {
        setError(data.message ?? data.error ?? '登录失败，请稍后重试')
        return
      }

      const accessToken = data.accessToken ?? ''

      // 存储 access token
      useAuthStore.getState().setAccessToken(accessToken, data.expiresIn ?? 3600)

      // 从 JWT payload 或响应字段中提取 userName / orgName
      const payload = parseJwtPayload(accessToken)
      const userName = data.userName ?? (payload.username as string) ?? ''
      const orgName = data.orgName ?? (payload.orgName as string) ?? userName

      // 记录默认组织和昵称并跳转到 workspace
      localStorage.setItem('defaultOrgName', orgName)
      localStorage.setItem('defaultUserName', userName)
      router.push(`/org/${orgName}/workspace`)
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { login, isLoading, error, identifierType, setIdentifierType }
}

/** 解析 JWT payload（不验签，仅读取字段） */
function parseJwtPayload(token: string): Record<string, unknown> {
  try {
    const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64)) as Record<string, unknown>
  } catch {
    return {}
  }
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
      const registerRes = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({
          phone: values.phone,
          userName: values.userName,
          password: values.password,
        }),
      })

      const registerData = (await registerRes.json()) as { accessToken?: string; error?: string; message?: string }

      if (!registerRes.ok) {
        setError(registerData.message ?? registerData.error ?? '注册失败，请稍后重试')
        return
      }

      // Step 2: 登录以获取 refresh_token cookie
      const loginRes = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({
          identifier: values.phone,
          identifierType: 'PHONE',
          password: values.password,
        }),
      })

      const loginData = (await loginRes.json()) as { accessToken?: string; error?: string; message?: string; orgName?: string; expiresIn?: number; userName?: string }

      if (!loginRes.ok) {
        setError('注册成功，请重新登录')
        router.push(TENANT_LOGIN_PATH)
        return
      }

      const accessToken = loginData.accessToken ?? registerData.accessToken ?? ''

      // Step 3: 存储 access token
      useAuthStore.getState().setAccessToken(accessToken, loginData.expiresIn ?? 3600)

      // Step 4: 从 JWT payload 或响应字段中提取 userName / orgName
      const payload = parseJwtPayload(accessToken)
      const userName = loginData.userName ?? (payload.username as string) ?? values.userName
      const orgName = loginData.orgName ?? (payload.orgName as string) ?? userName

      localStorage.setItem('defaultOrgName', orgName)
      localStorage.setItem('defaultUserName', userName)
      router.push(`/org/${orgName}/workspace`)
    } catch {
      setError('网络错误，请检查网络连接')
    } finally {
      setIsLoading(false)
    }
  }

  return { register, isLoading, error }
}
