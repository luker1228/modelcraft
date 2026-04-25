// src/app/api/bff/org/[orgName]/end-user/auth/select-project/route.ts
// v2 选择 Project，验证 pending session 并签发带 projectSlug 的正式 JWT

import { NextRequest, NextResponse } from 'next/server'
import { signEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import { setEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import {
  verifyPendingSessionToken,
  getPendingSessionFromCookie,
  clearPendingSessionCookie,
} from '@/bff/end-user/end-user-pending-session'
import type { EndUserSelectProjectResponse, EndUserSelectProjectError } from '@/types/end-user-auth'

interface SelectProjectBody {
  projectSlug?: unknown
}

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params

  let body: SelectProjectBody
  try {
    body = (await req.json()) as SelectProjectBody
  } catch {
    const err: EndUserSelectProjectError = { error: { code: 'PARAM_INVALID', message: 'Invalid JSON' } }
    return NextResponse.json(err, { status: 400 })
  }

  const { projectSlug } = body
  if (!projectSlug || typeof projectSlug !== 'string') {
    const err: EndUserSelectProjectError = { error: { code: 'PARAM_INVALID', message: '请提供 projectSlug' } }
    return NextResponse.json(err, { status: 400 })
  }

  // 读取 pending session cookie
  const pendingToken = getPendingSessionFromCookie()
  if (!pendingToken) {
    const err: EndUserSelectProjectError = { error: { code: 'PENDING_SESSION_INVALID', message: '登录会话已过期，请重新登录' } }
    return NextResponse.json(err, { status: 401 })
  }

  let session
  try {
    session = await verifyPendingSessionToken(pendingToken)
  } catch {
    clearPendingSessionCookie()
    const err: EndUserSelectProjectError = { error: { code: 'PENDING_SESSION_INVALID', message: '登录会话无效，请重新登录' } }
    return NextResponse.json(err, { status: 401 })
  }

  // 验证 orgName 一致
  if (session.orgName !== orgName) {
    clearPendingSessionCookie()
    const err: EndUserSelectProjectError = { error: { code: 'PENDING_SESSION_INVALID', message: '请求与登录会话不一致' } }
    return NextResponse.json(err, { status: 401 })
  }

  // 验证 projectSlug 在可访问列表内
  const allowed = session.projects.some((p) => p.slug === projectSlug)
  if (!allowed) {
    const err: EndUserSelectProjectError = { error: { code: 'PROJECT_ACCESS_DENIED', message: '您没有该项目的访问权限' } }
    return NextResponse.json(err, { status: 403 })
  }

  // 签发正式 access token
  const accessToken = await signEndUserAccessToken({
    userId: session.userId,
    orgName,
    projectSlug,
  })

  // 写入 refresh token cookie，清除 pending session
  setEndUserRefreshTokenCookie(`refresh-${session.userId}-${projectSlug}`, orgName, projectSlug)
  clearPendingSessionCookie()

  const res: EndUserSelectProjectResponse = { accessToken, projectSlug }
  return NextResponse.json(res)
}
