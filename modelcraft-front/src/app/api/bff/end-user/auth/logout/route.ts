// src/app/api/bff/end-user/auth/logout/route.ts
// 终端用户登出 BFF Route Handler（对称 api/bff/auth/logout/route.ts）
// Best-effort 调用 Go Backend revoke，无论成功失败都清除 Cookie

import { NextRequest, NextResponse } from 'next/server'
import { callGoEndUserLogout } from '@/bff/end-user/end-user-go-client'
import {
  getEndUserRefreshTokenFromCookie,
  clearEndUserRefreshTokenCookie,
} from '@/bff/end-user/end-user-cookie-utils'

interface LogoutRequestBody {
  orgName?: unknown
  projectSlug?: unknown
}

export async function POST(req: NextRequest) {
  // 从请求体获取 orgName 和 projectSlug，用于精确清除 Cookie
  let orgName = ''
  let projectSlug = ''

  try {
    const body = (await req.json()) as LogoutRequestBody
    orgName = typeof body.orgName === 'string' ? body.orgName : ''
    projectSlug = typeof body.projectSlug === 'string' ? body.projectSlug : ''
  } catch {
    // 如果解析失败，尝试从 URL 参数获取
    const url = new URL(req.url)
    orgName = url.searchParams.get('orgName') ?? ''
    projectSlug = url.searchParams.get('projectSlug') ?? ''
  }

  // 读取 refresh token（best-effort）
  const refreshToken = getEndUserRefreshTokenFromCookie()

  if (refreshToken && orgName && projectSlug) {
    // Best-effort 调用 Go Backend revoke（失败不影响 Cookie 清除）
    await callGoEndUserLogout({
      orgName,
      projectSlug,
      refreshToken,
    }).catch(() => {
      // 静默失败，Cookie 仍会被清除
    })
  }

  // 清除 Cookie（需要 orgName 和 projectSlug 以匹配 path）
  if (orgName && projectSlug) {
    clearEndUserRefreshTokenCookie(orgName, projectSlug)
  }

  return new NextResponse(null, { status: 204 })
}
