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
  EndUserUnauthorizedError,
  EndUserUpstreamError,
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

function parseEndUserRouteContext(referer: string | null): { orgName: string; projectSlug: string } | null {
  if (!referer) return null
  try {
    const url = new URL(referer)
    const match = url.pathname.match(/^\/u\/([^/]+)\/([^/]+)\//)
    if (!match) return null
    return {
      orgName: decodeURIComponent(match[1]),
      projectSlug: decodeURIComponent(match[2]),
    }
  } catch {
    return null
  }
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

  const routeContext = parseEndUserRouteContext(req.headers.get('referer'))

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

  if (
    routeContext &&
    (routeContext.orgName !== orgName || routeContext.projectSlug !== projectSlug)
  ) {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'orgName/projectSlug 与当前页面路由不一致' },
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
      const requestId = extractRequestId(err)
      if (requestId) errorRes.requestId = requestId
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserAccountDisabledError) {
      const errorRes: EndUserBffError = {
        error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用' },
      }
      const requestId = extractRequestId(err)
      if (requestId) errorRes.requestId = requestId
      return NextResponse.json(errorRes, { status: 403 })
    }

    if (err instanceof EndUserClusterNotConfiguredError) {
      const errorRes: EndUserBffError = {
        error: { code: 'CLUSTER_NOT_CONFIGURED', message: '服务暂时不可用' },
      }
      const requestId = extractRequestId(err)
      if (requestId) errorRes.requestId = requestId
      return NextResponse.json(errorRes, { status: 503 })
    }

    if (err instanceof EndUserParamInvalidError) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: err.message },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 400 })
    }

    if (err instanceof EndUserUnauthorizedError) {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: '未授权访问内部登录服务' },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserUpstreamError) {
      const errorRes: EndUserBffError = {
        error: {
          code: err.code === 'UNAUTHORIZED' ? 'UNAUTHORIZED' : 'INVALID_CREDENTIALS',
          message: err.message || '上游服务错误',
        },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: err.status || 500 })
    }

    // 未知错误
    const errorRes: EndUserBffError = {
      error: {
        code: 'INVALID_CREDENTIALS',
        message: err instanceof Error ? err.message : '登录失败，请稍后重试',
      },
    }
    const requestId = extractRequestId(err)
    if (requestId) errorRes.requestId = requestId
    return NextResponse.json(errorRes, { status: 401 })
  }
}
