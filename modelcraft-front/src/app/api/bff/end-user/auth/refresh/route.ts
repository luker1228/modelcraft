// src/app/api/bff/end-user/auth/refresh/route.ts
// 终端用户 Token 刷新 BFF Route Handler（对称 api/bff/auth/refresh/route.ts）
// 读取 Cookie → 调用 Go token rotation → 重写 Cookie → 返回新 access token

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoEndUserRefresh,
  EndUserTokenError,
  EndUserAccountDisabledError,
} from '@/bff/end-user/end-user-go-client'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import {
  getEndUserRefreshTokenFromCookie,
  setEndUserRefreshTokenCookie,
  clearEndUserRefreshTokenCookie,
} from '@/bff/end-user/end-user-cookie-utils'
import type { EndUserAuthResponse, EndUserBffError } from '@/types/end-user-auth'

interface RefreshRequestBody {
  orgName?: unknown
  projectSlug?: unknown
}

export async function POST(req: NextRequest) {
  // 从请求体获取 orgName 和 projectSlug
  let orgName = ''
  let projectSlug = ''

  try {
    const body = (await req.json()) as RefreshRequestBody
    orgName = typeof body.orgName === 'string' ? body.orgName : ''
    projectSlug = typeof body.projectSlug === 'string' ? body.projectSlug : ''
  } catch {
    // 如果解析失败，尝试从 URL 参数获取
    const url = new URL(req.url)
    orgName = url.searchParams.get('orgName') ?? ''
    projectSlug = url.searchParams.get('projectSlug') ?? ''
  }

  // 读取 refresh token
  const refreshToken = getEndUserRefreshTokenFromCookie()

  if (!refreshToken) {
    const errorRes: EndUserBffError = {
      error: { code: 'INVALID_REFRESH_TOKEN', message: 'No refresh token' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }

  if (!orgName || !projectSlug) {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'orgName and projectSlug are required' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }

  try {
    // 1. 调用 Go Backend token rotation
    const { userId, refreshToken: newRefreshToken } = await callGoEndUserRefresh({
      orgName,
      projectSlug,
      refreshToken,
    })

    // 2. Rotate httpOnly Cookie
    setEndUserRefreshTokenCookie(newRefreshToken, orgName, projectSlug)

    // 3. 签发新的 access token
    const accessToken = await signEndUserAccessToken({
      userId,
      orgName,
      projectSlug,
    })

    const response: EndUserAuthResponse = {
      accessToken,
      expiresIn: 3600,
    }

    return NextResponse.json(response)
  } catch (err) {
    console.error('[BFF] end-user refresh error:', err)

    // Token 无效或已过期，清除 Cookie 强制重新登录
    if (err instanceof EndUserTokenError) {
      clearEndUserRefreshTokenCookie(orgName, projectSlug)
      const errorRes: EndUserBffError = {
        error: { code: 'INVALID_REFRESH_TOKEN', message: 'Token 无效或已过期' },
      }
      return NextResponse.json(errorRes, { status: 401 })
    }

    // 账号被禁用
    if (err instanceof EndUserAccountDisabledError) {
      clearEndUserRefreshTokenCookie(orgName, projectSlug)
      const errorRes: EndUserBffError = {
        error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用' },
      }
      return NextResponse.json(errorRes, { status: 403 })
    }

    // 其他错误，清除 Cookie
    clearEndUserRefreshTokenCookie(orgName, projectSlug)
    const errorRes: EndUserBffError = {
      error: { code: 'INVALID_REFRESH_TOKEN', message: 'Unauthorized' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }
}
