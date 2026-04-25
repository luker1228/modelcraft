// src/app/api/bff/org/[orgName]/end-user/auth/login/route.ts
// v1 Org 级 EndUser 登录 BFF
// 登录成功后返回可访问 Project 列表，由前端决定是否需要选择 Project

import { NextRequest, NextResponse } from 'next/server'
import { callGoEndUserLoginOrg } from '@/bff/end-user/end-user-go-client'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import {
  signPendingSessionToken,
  setPendingSessionCookie,
} from '@/bff/end-user/end-user-pending-session'
import {
  EndUserInvalidCredentialsError,
  EndUserAccountDisabledError,
  EndUserUnauthorizedError,
  EndUserUpstreamError,
} from '@/bff/end-user/end-user-go-client'
import type { EndUserLoginResponse } from '@/types/end-user-auth'

interface LoginBody {
  username?: unknown
  password?: unknown
}

function isInvalidCredentialMessage(message: string): boolean {
  return (
    message.includes('用户名或密码错误') ||
    message.toLowerCase().includes('invalid credentials')
  )
}

function isNoProjectAccessMessage(message: string): boolean {
  return (
    message.includes('暂无可访问项目') ||
    message.includes('暂无项目访问权限') ||
    message.toLowerCase().includes('no project access')
  )
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

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params

  let body: LoginBody
  try {
    body = (await req.json()) as LoginBody
  } catch {
    return NextResponse.json({ error: { code: 'PARAM_INVALID', message: 'Invalid JSON' } }, { status: 400 })
  }

  const { username, password } = body
  if (!username || typeof username !== 'string') {
    return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '请输入用户名' } }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '请输入密码' } }, { status: 400 })
  }

  try {
    const { userId, accessibleProjects } = await callGoEndUserLoginOrg({ orgName, username, password })

    // 无任何可访问项目 → 登录成功但进入待授权页（不视为错误）
    if (accessibleProjects.length === 0) {
      const res: EndUserLoginResponse = {
        singleProject: false,
        projects: [],
        noProjectAccess: true,
        message: '该账号暂无可访问项目，请联系管理员申请项目访问权限',
      }
      return NextResponse.json(res)
    }

    // 只有一个项目 → 直接签发正式 JWT
    if (accessibleProjects.length === 1) {
      const projectSlug = accessibleProjects[0].slug
      const accessToken = await signEndUserAccessToken({ userId, orgName, projectSlug })
      // refresh token cookie（path 绑定项目）
      setEndUserRefreshTokenCookie(`pending-refresh-${userId}`, orgName, projectSlug)

      const res: EndUserLoginResponse = { singleProject: true, projectSlug, accessToken }
      return NextResponse.json(res)
    }

    // 多个项目 → 签发 pending session，前端跳 select-project 页
    const pendingToken = await signPendingSessionToken(userId, orgName, accessibleProjects)
    setPendingSessionCookie(pendingToken)

    const res: EndUserLoginResponse = { singleProject: false, projects: accessibleProjects }
    return NextResponse.json(res)
  } catch (err) {
    const requestId = extractRequestId(err)

    if (err instanceof EndUserInvalidCredentialsError) {
      return NextResponse.json(
        {
          error: { code: 'INVALID_CREDENTIALS', message: '用户名或密码错误' },
          ...(requestId ? { requestId } : {}),
        },
        { status: 401 }
      )
    }
    if (err instanceof EndUserAccountDisabledError) {
      return NextResponse.json(
        {
          error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用，请联系管理员' },
          ...(requestId ? { requestId } : {}),
        },
        { status: 403 }
      )
    }
    if (err instanceof EndUserUnauthorizedError) {
      return NextResponse.json(
        {
          error: { code: 'UNAUTHORIZED', message: '未授权访问内部服务' },
          ...(requestId ? { requestId } : {}),
        },
        { status: 401 }
      )
    }
    if (err instanceof EndUserUpstreamError) {
      const status = err.status || 500
      const rawCode = err.code || (status >= 500 ? 'UPSTREAM_ERROR' : 'PARAM_INVALID')
      let normalizedCode = rawCode
      let normalizedMessage = err.message || '上游服务异常'

      if (rawCode === 'NO_PROJECT_ACCESS' || isNoProjectAccessMessage(err.message || '')) {
        const res: EndUserLoginResponse = {
          singleProject: false,
          projects: [],
          noProjectAccess: true,
          message: '该账号暂无可访问项目，请联系管理员申请项目访问权限',
        }
        return NextResponse.json(res)
      }

      if (rawCode === 'PARAM_INVALID' && isInvalidCredentialMessage(err.message)) {
        normalizedCode = 'INVALID_CREDENTIALS'
      }
      return NextResponse.json(
        {
          error: { code: normalizedCode, message: normalizedMessage },
          ...(requestId ? { requestId } : {}),
        },
        { status }
      )
    }

    console.error('[BFF v1] end-user login error:', err)
    return NextResponse.json(
      {
        error: { code: 'PARAM_INVALID', message: '登录失败，请稍后重试' },
        ...(requestId ? { requestId } : {}),
      },
      { status: 500 }
    )
  }
}
