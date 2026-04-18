// src/app/api/bff/end-user/auth/register/route.ts
// 终端用户注册 BFF Route Handler（对称 api/bff/auth/register/route.ts）
// 注册成功后自动登录，返回 access token 并写入 refresh token Cookie

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoEndUserRegister,
  EndUserConflictError,
  EndUserParamInvalidError,
  EndUserClusterNotConfiguredError,
  EndUserUnauthorizedError,
  EndUserUpstreamError,
} from '@/bff/end-user/end-user-go-client'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import type { EndUserRegisterRequest, EndUserAuthResponse, EndUserBffError } from '@/types/end-user-auth'

interface RegisterRequestBody {
  orgName?: unknown
  projectSlug?: unknown
  username?: unknown
  password?: unknown
}

function extractRequestId(err: unknown): string | undefined {
  if (
    err &&
    typeof err === 'object' &&
    'requestId' in err &&
    typeof (err as { requestId?: unknown }).requestId === 'string'
  ) {
    return (err as { requestId: string }).requestId
  }
  return undefined
}

export async function POST(req: NextRequest) {
  let body: RegisterRequestBody
  try {
    body = (await req.json()) as RegisterRequestBody
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

  const registerRequest: EndUserRegisterRequest = {
    orgName,
    projectSlug,
    username,
    password,
  }

  try {
    // 1. 调用 Go Backend 注册（内部处理 mock）
    const { userId, refreshToken } = await callGoEndUserRegister(registerRequest)

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
    console.error('[BFF] end-user register error:', err)

    // 错误分类处理
    if (err instanceof EndUserConflictError) {
      const errorRes: EndUserBffError = {
        error: { code: 'CONFLICT', message: '该用户名已被使用' },
      }
      const requestId = extractRequestId(err)
      if (requestId) errorRes.requestId = requestId
      return NextResponse.json(errorRes, { status: 409 })
    }

    if (err instanceof EndUserParamInvalidError) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: err.message },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 400 })
    }

    if (err instanceof EndUserClusterNotConfiguredError) {
      const errorRes: EndUserBffError = {
        error: { code: 'CLUSTER_NOT_CONFIGURED', message: '服务暂时不可用' },
      }
      const requestId = extractRequestId(err)
      if (requestId) errorRes.requestId = requestId
      return NextResponse.json(errorRes, { status: 503 })
    }

    if (err instanceof EndUserUnauthorizedError) {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: '未授权访问内部注册服务' },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserUpstreamError) {
      const errorRes: EndUserBffError = {
        error: {
          code: err.code === 'UNAUTHORIZED' ? 'UNAUTHORIZED' : 'PARAM_INVALID',
          message: err.message || '上游服务错误',
        },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: err.status || 500 })
    }

    // 未知错误
    const errorRes: EndUserBffError = {
      error: {
        code: 'PARAM_INVALID',
        message: err instanceof Error ? err.message : '注册失败，请稍后重试',
      },
    }
    const requestId = extractRequestId(err)
    if (requestId) errorRes.requestId = requestId
    return NextResponse.json(errorRes, { status: 400 })
  }
}
