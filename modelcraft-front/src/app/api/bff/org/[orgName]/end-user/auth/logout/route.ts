// src/app/api/bff/org/[orgName]/end-user/auth/logout/route.ts
// v2 登出：清除 pending session cookie 和 refresh token cookie

import { NextRequest, NextResponse } from 'next/server'
import { clearEndUserRefreshTokenCookie } from '@/bff/end-user/end-user-cookie-utils'
import { clearPendingSessionCookie } from '@/bff/end-user/end-user-pending-session'

export async function POST(
  _req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params

  // 清除正式 refresh token（path 不含 projectSlug，覆盖性清除）
  clearEndUserRefreshTokenCookie(orgName, '')
  // 清除 pending session
  clearPendingSessionCookie()

  return NextResponse.json({ success: true })
}
