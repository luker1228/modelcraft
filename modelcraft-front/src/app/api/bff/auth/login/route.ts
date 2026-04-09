// src/app/api/bff/auth/login/route.ts
import { NextRequest, NextResponse } from 'next/server'
import {
  callGoLogin,
  AuthenticationError,
  AuthParamInvalidError,
} from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import { setRefreshTokenCookie } from '@/bff/auth/cookie-utils'
import type { IdentifierType, LoginResponse } from '@/types/auth'

interface LoginRequestBody {
  identifier?: unknown
  identifierType?: unknown
  password?: unknown
}

export async function POST(req: NextRequest) {
  let body: LoginRequestBody
  try {
    body = (await req.json()) as LoginRequestBody
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const { identifier, identifierType, password } = body

  // 参数校验
  if (!identifier || typeof identifier !== 'string') {
    return NextResponse.json(
      { error: '请输入手机号或用户名' },
      { status: 400 }
    )
  }
  if (
    !identifierType ||
    (identifierType !== 'PHONE' && identifierType !== 'USERNAME')
  ) {
    return NextResponse.json({ error: '登录类型无效' }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: '请输入密码' }, { status: 400 })
  }

  try {
    // 1. Call Go Backend login with identifier + identifierType + password
    const { userId, userName, orgName, refreshToken } = await callGoLogin({
      identifier,
      identifierType: identifierType as IdentifierType,
      password,
    })

    // 2. BFF signs Access Token
    const accessToken = await signAccessToken(userId)

    // 3. Set httpOnly Cookie with Refresh Token
    setRefreshTokenCookie(refreshToken)

    const response: LoginResponse = {
      accessToken,
      expiresIn: 3600,
      userId,
      userName,
      orgName,
    }

    return NextResponse.json(response)
  } catch (err) {
    console.error('BFF login error:', err)

    // 区分错误类型返回不同的错误消息和状态码
    if (err instanceof AuthenticationError) {
      // 根据 message 判断是用户不存在还是密码错误
      const message = err.message
      let userFriendlyMessage = '登录失败'

      if (
        message.includes('not found') ||
        message.includes('用户不存在') ||
        message.includes('手机号不存在') ||
        message.includes('用户名不存在')
      ) {
        userFriendlyMessage = '账号不存在'
      } else if (
        message.includes('incorrect password') ||
        message.includes('密码错误')
      ) {
        userFriendlyMessage = '密码错误'
      }

      return NextResponse.json({ error: userFriendlyMessage }, { status: 401 })
    }

    if (err instanceof AuthParamInvalidError) {
      return NextResponse.json({ error: err.message }, { status: 400 })
    }

    return NextResponse.json(
      { error: err instanceof Error ? err.message : '登录失败，请稍后重试' },
      { status: 401 }
    )
  }
}
