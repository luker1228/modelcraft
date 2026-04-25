// src/app/api/bff/org/[orgName]/end-user/auth/register/route.ts
// v1 Org 级 EndUser 自注册 BFF
// 注册成功后自动登录，并按可访问 Project 数量决定跳转分支

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoCreateOrgEndUser,
  callGoEndUserLoginOrg,
} from '@/bff/end-user/end-user-go-client-v2'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import {
  signPendingSessionToken,
  setPendingSessionCookie,
} from '@/bff/end-user/end-user-pending-session'
import {
  EndUserAccountDisabledError,
  EndUserConflictError,
  EndUserInvalidCredentialsError,
  EndUserParamInvalidError,
  EndUserUnauthorizedError,
  EndUserUpstreamError,
} from '@/bff/end-user/end-user-go-client'
import type { EndUserLoginResponse } from '@/types/end-user-auth'

interface RegisterBody {
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

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params

  let body: RegisterBody
  try {
    body = (await req.json()) as RegisterBody
  } catch {
    return NextResponse.json(
      { error: { code: 'PARAM_INVALID', message: 'Invalid JSON' } },
      { status: 400 }
    )
  }

  const { username, password } = body
  if (!username || typeof username !== 'string') {
    return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '请输入用户名' } }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '请输入密码' } }, { status: 400 })
  }

  try {
    // 1) 先创建 Org 级终端用户账号
    await callGoCreateOrgEndUser({ orgName, username, password })

    // 2) 再执行登录流程，复用 v1 单项目/多项目分支
    const { userId, accessibleProjects } = await callGoEndUserLoginOrg({ orgName, username, password })

    if (accessibleProjects.length === 0) {
      const res: EndUserLoginResponse = {
        error: { code: 'NO_PROJECT_ACCESS', message: '您暂无项目访问权限，请联系管理员授权' },
      }
      return NextResponse.json(res, { status: 403 })
    }

    if (accessibleProjects.length === 1) {
      const projectSlug = accessibleProjects[0].slug
      const accessToken = await signEndUserAccessToken({ userId, orgName, projectSlug })
      setEndUserRefreshTokenCookie(`pending-refresh-${userId}`, orgName, projectSlug)

      const res: EndUserLoginResponse = { singleProject: true, projectSlug, accessToken }
      return NextResponse.json(res)
    }

    const pendingToken = await signPendingSessionToken(userId, orgName, accessibleProjects)
    setPendingSessionCookie(pendingToken)
    const res: EndUserLoginResponse = { singleProject: false, projects: accessibleProjects }
    return NextResponse.json(res)
  } catch (err) {
    const requestId = extractRequestId(err)

    if (err instanceof EndUserConflictError) {
      return NextResponse.json(
        {
          error: { code: 'CONFLICT', message: '该用户名已被使用' },
          ...(requestId ? { requestId } : {}),
        },
        { status: 409 }
      )
    }
    if (err instanceof EndUserParamInvalidError) {
      return NextResponse.json(
        {
          error: { code: 'PARAM_INVALID', message: err.message || '注册参数不合法' },
          ...(requestId ? { requestId } : {}),
        },
        { status: 400 }
      )
    }
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
      const code = err.code || (status >= 500 ? 'UPSTREAM_ERROR' : 'PARAM_INVALID')
      return NextResponse.json(
        {
          error: { code, message: err.message || '上游服务异常' },
          ...(requestId ? { requestId } : {}),
        },
        { status }
      )
    }

    console.error('[BFF v1] end-user register error:', err)
    return NextResponse.json(
      {
        error: { code: 'PARAM_INVALID', message: '注册失败，请稍后重试' },
        ...(requestId ? { requestId } : {}),
      },
      { status: 500 }
    )
  }
}
