// src/app/api/bff/org/[orgName]/end-user/auth/login/route.ts
// v2 Org 级 EndUser 登录 BFF
// 登录成功后返回可访问 Project 列表，由前端决定是否需要选择 Project

import { NextRequest, NextResponse } from 'next/server'
import { callGoEndUserLoginV2 } from '@/bff/end-user/end-user-go-client-v2'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import {
  signPendingSessionToken,
  setPendingSessionCookie,
} from '@/bff/end-user/end-user-pending-session'
import {
  EndUserInvalidCredentialsError,
  EndUserAccountDisabledError,
} from '@/bff/end-user/end-user-go-client'
import type { EndUserLoginResponseV2 } from '@/types/end-user-auth'

interface LoginBody {
  username?: unknown
  password?: unknown
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
    const { userId, accessibleProjects } = await callGoEndUserLoginV2({ orgName, username, password })

    // 无任何可访问项目 → 报错
    if (accessibleProjects.length === 0) {
      const res: EndUserLoginResponseV2 = {
        error: { code: 'NO_PROJECT_ACCESS', message: '您暂无项目访问权限，请联系管理员授权' },
      }
      return NextResponse.json(res, { status: 403 })
    }

    // 只有一个项目 → 直接签发正式 JWT
    if (accessibleProjects.length === 1) {
      const projectSlug = accessibleProjects[0].slug
      const accessToken = await signEndUserAccessToken({ userId, orgName, projectSlug })
      // refresh token cookie（path 绑定项目）
      setEndUserRefreshTokenCookie(`pending-refresh-${userId}`, orgName, projectSlug)

      const res: EndUserLoginResponseV2 = { singleProject: true, projectSlug, accessToken }
      return NextResponse.json(res)
    }

    // 多个项目 → 签发 pending session，前端跳 select-project 页
    const pendingToken = await signPendingSessionToken(userId, orgName, accessibleProjects)
    setPendingSessionCookie(pendingToken)

    const res: EndUserLoginResponseV2 = { singleProject: false, projects: accessibleProjects }
    return NextResponse.json(res)
  } catch (err) {
    if (err instanceof EndUserInvalidCredentialsError) {
      return NextResponse.json(
        { error: { code: 'INVALID_CREDENTIALS', message: '用户名或密码错误' } },
        { status: 401 }
      )
    }
    if (err instanceof EndUserAccountDisabledError) {
      return NextResponse.json(
        { error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用，请联系管理员' } },
        { status: 403 }
      )
    }
    console.error('[BFF v2] end-user login error:', err)
    return NextResponse.json(
      { error: { code: 'PARAM_INVALID', message: '登录失败，请稍后重试' } },
      { status: 500 }
    )
  }
}
