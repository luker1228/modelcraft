// src/app/api/bff/org/[orgName]/end-user/users/[userId]/accessible-projects/route.ts
// BFF: 获取指定终端用户可访问的 Project 列表

import { NextRequest, NextResponse } from 'next/server'
import { callGoGetUserAccessibleProjects } from '@/bff/end-user/end-user-go-client'

interface RouteParams {
  params: Promise<{ orgName: string; userId: string }>
}

export async function GET(_req: NextRequest, { params }: RouteParams) {
  const { orgName, userId } = await params
  try {
    const projects = await callGoGetUserAccessibleProjects({ orgName, userId })
    return NextResponse.json({ projects })
  } catch {
    return NextResponse.json({ error: { message: '获取关联项目失败' } }, { status: 500 })
  }
}
