// src/app/api/bff/org/[orgName]/project/[projectSlug]/end-user-access/[accessId]/route.ts
// Project 级单个终端用户访问控制操作（撤销授权 / 更新权限包）BFF

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoRevokeEndUserProjectAccess,
  callGoUpdateEndUserProjectAccess,
} from '@/bff/end-user/end-user-go-client'

interface RouteParams {
  params: Promise<{ orgName: string; projectSlug: string; accessId: string }>
}

export async function DELETE(_req: NextRequest, { params }: RouteParams) {
  const { orgName, projectSlug, accessId } = await params
  try {
    await callGoRevokeEndUserProjectAccess({ orgName, projectSlug, accessId })
    return new NextResponse(null, { status: 204 })
  } catch {
    return NextResponse.json({ error: { message: '撤销授权失败' } }, { status: 500 })
  }
}

export async function PATCH(req: NextRequest, { params }: RouteParams) {
  const { orgName, projectSlug, accessId } = await params
  try {
    const body: unknown = await req.json()
    const { permissionBundle } = body as { permissionBundle?: unknown }
    if (typeof permissionBundle !== 'string') {
      return NextResponse.json(
        { error: { code: 'PARAM_INVALID', message: 'permissionBundle 必须为字符串' } },
        { status: 400 }
      )
    }
    await callGoUpdateEndUserProjectAccess({
      orgName,
      projectSlug,
      accessId,
      permissionBundleId: permissionBundle,
      permissionBundleName: permissionBundle,
    })
    return NextResponse.json({ ok: true })
  } catch {
    return NextResponse.json({ error: { message: '更新权限失败' } }, { status: 500 })
  }
}
