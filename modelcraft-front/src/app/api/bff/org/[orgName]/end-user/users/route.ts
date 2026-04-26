// src/app/api/bff/org/[orgName]/end-user/users/route.ts
// Org 级终端用户列表 + 创建 BFF

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoListOrgEndUsers,
  callGoCreateOrgEndUser,
} from '@/bff/end-user/end-user-go-client'
import { EndUserConflictError, EndUserParamInvalidError } from '@/bff/end-user/end-user-go-client'

interface RouteParams {
  params: Promise<{ orgName: string }>
}

export async function GET(_req: NextRequest, { params }: RouteParams) {
  const { orgName } = await params
  try {
    const result = await callGoListOrgEndUsers({ orgName, first: 500 })
    const users = result.nodes.map((u) => ({
      id: u.id,
      username: u.username,
      displayName: undefined,
      status: u.isForbidden ? 'DISABLED' : 'ACTIVE',
      createdAt: u.createdAt,
    }))
    return NextResponse.json({ users })
  } catch {
    return NextResponse.json({ error: { message: '获取终端用户列表失败' } }, { status: 500 })
  }
}

export async function POST(req: NextRequest, { params }: RouteParams) {
  const { orgName } = await params
  try {
    const body: unknown = await req.json()
    const { username, password } = body as {
      username?: unknown
      password?: unknown
    }
    if (typeof username !== 'string' || !username.trim()) {
      return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '用户名不能为空' } }, { status: 400 })
    }
    if (typeof password !== 'string' || !password.trim()) {
      return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '密码不能为空' } }, { status: 400 })
    }
    const raw = await callGoCreateOrgEndUser({ orgName, username: username.trim(), password: password.trim() })
    const user = {
      id: raw.id,
      username: raw.username,
      displayName: undefined,
      status: raw.isForbidden ? 'DISABLED' : 'ACTIVE',
      createdAt: raw.createdAt,
    }
    return NextResponse.json({ user }, { status: 201 })
  } catch (e) {
    if (e instanceof EndUserConflictError) {
      return NextResponse.json({ error: { code: 'CONFLICT', message: '用户名已存在' } }, { status: 409 })
    }
    if (e instanceof EndUserParamInvalidError) {
      return NextResponse.json({ error: { code: 'PARAM_INVALID', message: e.message } }, { status: 400 })
    }
    return NextResponse.json({ error: { message: '创建终端用户失败' } }, { status: 500 })
  }
}
