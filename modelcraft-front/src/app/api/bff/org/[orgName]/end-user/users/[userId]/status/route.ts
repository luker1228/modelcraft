// src/app/api/bff/org/[orgName]/end-user/users/[userId]/status/route.ts
// Org 级终端用户状态更新 BFF（启用 / 禁用）

import { NextRequest, NextResponse } from 'next/server'
import { callGoUpdateOrgEndUserStatus } from '@/bff/end-user/end-user-go-client'

interface RouteParams {
  params: Promise<{ orgName: string; userId: string }>
}

export async function PATCH(req: NextRequest, { params }: RouteParams) {
  const { orgName, userId } = await params
  try {
    const body: unknown = await req.json()
    const { status } = body as { status?: unknown }
    if (status !== 'ACTIVE' && status !== 'DISABLED') {
      return NextResponse.json(
        { error: { code: 'PARAM_INVALID', message: 'status 必须为 ACTIVE 或 DISABLED' } },
        { status: 400 }
      )
    }
    // Map frontend status → Go isForbidden
    await callGoUpdateOrgEndUserStatus({ orgName, userId, isForbidden: status === 'DISABLED' })
    return NextResponse.json({ ok: true })
  } catch {
    return NextResponse.json({ error: { message: '更新用户状态失败' } }, { status: 500 })
  }
}
