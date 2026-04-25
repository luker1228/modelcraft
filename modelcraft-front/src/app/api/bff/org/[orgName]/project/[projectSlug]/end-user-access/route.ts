// src/app/api/bff/org/[orgName]/project/[projectSlug]/end-user-access/route.ts
// Project 级终端用户访问控制列表 + 授权 BFF（EndUser v1）

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoListProjectEndUserAccesses,
  callGoGrantEndUserProjectAccess,
} from '@/bff/end-user/end-user-go-client-v2'
import { EndUserConflictError, EndUserParamInvalidError } from '@/bff/end-user/end-user-go-client'

interface RouteParams {
  params: Promise<{ orgName: string; projectSlug: string }>
}

export async function GET(_req: NextRequest, { params }: RouteParams) {
  const { orgName, projectSlug } = await params
  try {
    const conn = await callGoListProjectEndUserAccesses({ orgName, projectSlug })
    // Map Go types → frontend EndUserProjectAccessEntry shape (expose accessId for revoke/update)
    const accesses = conn.nodes.map((a) => ({
      accessId: a.id,
      userId: a.endUser.id,
      username: a.endUser.username,
      displayName: undefined,
      permissionBundle: a.permissionBundleName || undefined,
      permissionBundleId: a.permissionBundleId,
      grantedAt: a.grantedAt,
    }))
    return NextResponse.json({ accesses })
  } catch {
    return NextResponse.json({ error: { message: '获取访问控制列表失败' } }, { status: 500 })
  }
}

export async function POST(req: NextRequest, { params }: RouteParams) {
  const { orgName, projectSlug } = await params
  try {
    const body = await req.json()
    const { userId, permissionBundle } = body as { userId?: unknown; permissionBundle?: unknown }
    if (typeof userId !== 'string' || !userId.trim()) {
      return NextResponse.json({ error: { code: 'PARAM_INVALID', message: '用户 ID 不能为空' } }, { status: 400 })
    }
    // Map frontend permissionBundle label → bundleId + bundleName
    const bundleName = typeof permissionBundle === 'string' ? permissionBundle : ''
    const access = await callGoGrantEndUserProjectAccess({
      orgName,
      projectSlug,
      endUserId: userId.trim(),
      permissionBundleId: bundleName || 'default',
      permissionBundleName: bundleName,
    })
    return NextResponse.json({
      access: {
        accessId: access.id,
        userId: access.endUser.id,
        username: access.endUser.username,
        permissionBundle: access.permissionBundleName || undefined,
        permissionBundleId: access.permissionBundleId,
        grantedAt: access.grantedAt,
      }
    }, { status: 201 })
  } catch (e) {
    if (e instanceof EndUserConflictError) {
      return NextResponse.json({ error: { code: 'CONFLICT', message: '该用户已有访问权限' } }, { status: 409 })
    }
    if (e instanceof EndUserParamInvalidError) {
      return NextResponse.json({ error: { code: 'PARAM_INVALID', message: e.message } }, { status: 400 })
    }
    return NextResponse.json({ error: { message: '授权失败' } }, { status: 500 })
  }
}
