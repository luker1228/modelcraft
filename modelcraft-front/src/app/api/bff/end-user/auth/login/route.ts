// src/app/api/bff/end-user/auth/login/route.ts
// 终端用户登录 BFF Route Handler（对称 api/bff/auth/login/route.ts）
// 不写内联 mock，调用 Slice A 的 end-user-go-client（内部处理 mock）

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoEndUserLogin,
  EndUserInvalidCredentialsError,
  EndUserAccountDisabledError,
  EndUserClusterNotConfiguredError,
  EndUserParamInvalidError,
} from '@/bff/end-user/end-user-go-client'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import type { EndUserLoginRequest, EndUserAuthResponse, EndUserBffError } from '@/types/end-user-auth'

interface LoginRequestBody {
  orgName?: unknown
  projectSlug?: unknown
  username?: unknown
  password?: unknown
}

export async function POST(req: NextRequest) {
  let body: LoginRequestBody
  try {
    body = (await req.json()) as LoginRequestBody
  } catch {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'Invalid JSON' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }

  const { orgName, projectSlug, username, password } = body

  // 参数校验
  if (!orgName || typeof orgName !== 'string') {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'orgName is required' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }
  if (!projectSlug || typeof projectSlug !== 'string') {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'projectSlug is required' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }
  if (!username || typeof username !== 'string') {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: '请输入用户名' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: '请输入密码' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }

  const loginRequest: EndUserLoginRequest = {
    orgName,
    projectSlug,
    username,
    password,
  }

  try {
    // 1. 调用 Go Backend 登录（内部处理 mock）
    const { userId, refreshToken } = await callGoEndUserLogin(loginRequest)

    // 2. BFF 签发 end-user access token（issuer = modelcraft-end-user）
    const accessToken = await signEndUserAccessToken({
      userId,
      orgName,
      projectSlug,
    })

    // 3. 写入 httpOnly Cookie（path 绑定到 Project）
    setEndUserRefreshTokenCookie(refreshToken, orgName, projectSlug)

    const response: EndUserAuthResponse = {
      accessToken,
      expiresIn: 3600,
    }

    return NextResponse.json(response)
  } catch (err) {
    console.error('[BFF] end-user login error:', err)

    // 错误分类处理
    if (err instanceof EndUserInvalidCredentialsError) {
      const errorRes: EndUserBffError = {
        error: { code: 'INVALID_CREDENTIALS', message: '用户名或密码错误' },
      }
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserAccountDisabledError) {
      const errorRes: EndUserBffError = {
        error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用' },
      }
      return NextResponse.json(errorRes, { status: 403 })
    }

    if (err instanceof EndUserClusterNotConfiguredError) {
      const errorRes: EndUserBffError = {
        error: { code: 'CLUSTER_NOT_CONFIGURED', message: '服务暂时不可用' },
      }
      return NextResponse.json(errorRes, { status: 503 })
    }

    if (err instanceof EndUserParamInvalidError) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: err.message },
      }
      return NextResponse.json(errorRes, { status: 400 })
    }

    // 未知错误
    const errorRes: EndUserBffError = {
      error: {
        code: 'INVALID_CREDENTIALS',
        message: err instanceof Error ? err.message : '登录失败，请稍后重试',
      },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }
}
