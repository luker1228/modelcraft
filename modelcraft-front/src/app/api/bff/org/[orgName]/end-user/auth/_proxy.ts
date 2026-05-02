/**
 * BFF End-User Auth Proxy Helper
 *
 * 将前端 /api/bff/org/[orgName]/end-user/auth/* 请求透传到 gateway
 * /api/end-user/auth/*。
 *
 * Cookie 策略：透传 gateway 返回的 Set-Cookie header，前端不主动写入或清除 cookie。
 */

import { NextRequest, NextResponse } from 'next/server'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

/**
 * 将请求转发到 gateway end-user auth 端点，透传所有响应 header（含 Set-Cookie）。
 *
 * @param req    - 原始 Next.js 请求
 * @param path   - 后端路径段（如 "login"、"me"）
 * @param method - HTTP 方法（POST / GET）
 */
export async function proxyEndUserAuth(
  req: NextRequest,
  path: string,
  method: string = 'POST'
): Promise<NextResponse> {
  const upstreamUrl = `${BACKEND_URL}/api/end-user/auth/${path}`

  const headers = new Headers()
  headers.set('Content-Type', 'application/json')

  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const body = method !== 'GET' && method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method, headers, body })
  } catch {
    return NextResponse.json({ error: { code: 'NETWORK_ERROR', message: '后端服务不可达' } }, { status: 502 })
  }

  const resBodyBuf = await upstreamRes.arrayBuffer()

  const response = new NextResponse(resBodyBuf, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  // 透传所有响应 header，包括 Set-Cookie
  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  return response
}
