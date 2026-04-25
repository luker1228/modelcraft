// src/app/api/bff/org/[orgName]/end-user/users/[userId]/route.ts
// Org 级终端用户删除 BFF

import { NextRequest, NextResponse } from 'next/server'
import { callGoDeleteOrgEndUser } from '@/bff/end-user/end-user-go-client-v2'

interface RouteParams {
  params: Promise<{ orgName: string; userId: string }>
}

export async function DELETE(_req: NextRequest, { params }: RouteParams) {
  const { orgName, userId } = await params
  try {
    await callGoDeleteOrgEndUser({ orgName, userId })
    return new NextResponse(null, { status: 204 })
  } catch {
    return NextResponse.json({ error: { message: '删除终端用户失败' } }, { status: 500 })
  }
}
