// src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/refresh
//
// 从 mc_enduser_refresh_token HttpOnly cookie 读取 refreshToken，
// 注入到请求 body 后转发到后端（后端只从 body 读 refreshToken）。

import { NextRequest, NextResponse } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'
import { END_USER_REFRESH_COOKIE } from '@/middleware'

type RouteParams = { params: Promise<{ orgName: string }> }

export async function POST(req: NextRequest, { params }: RouteParams): Promise<NextResponse> {
  const { orgName } = await params

  // 从 HttpOnly cookie 读取 refreshToken
  const cookieToken = req.cookies.get(END_USER_REFRESH_COOKIE)?.value

  // DEBUG: log all cookies
  const allCookies = req.cookies.getAll()
  console.log(`[refresh-bff] orgName=${orgName} cookieToken=${cookieToken ? cookieToken.substring(0, 20) + '...' : 'MISSING'} allCookies=${JSON.stringify(allCookies.map(c => c.name))}`)

  // 解析原始 body（可能含 projectSlug 等额外字段）
  let bodyObj: Record<string, unknown> = {}
  try {
    const raw = await req.text()
    if (raw) bodyObj = JSON.parse(raw) as Record<string, unknown>
  } catch {
    // body 为空或无效 JSON，忽略
  }

  // cookie 优先覆盖 body 里的 refreshToken；orgName 兜底补充
  const mergedBody = {
    ...bodyObj,
    orgName: bodyObj.orgName ?? orgName,
    ...(cookieToken ? { refreshToken: cookieToken } : {}),
  }

  // 构造新 Request 替换 body
  const newReq = new NextRequest(req.url, {
    method: 'POST',
    headers: req.headers,
    body: JSON.stringify(mergedBody),
  })

  return proxyEndUserAuth(newReq, 'refresh', 'POST')
}
