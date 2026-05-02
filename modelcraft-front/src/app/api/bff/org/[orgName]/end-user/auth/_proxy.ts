/**
 * BFF End-User Auth Proxy Helper
 *
 * 将前端 /api/bff/org/[orgName]/end-user/auth/* 请求透传到 gateway
 * /api/end-user/auth/*。
 *
 * 不注入 X-Org-Name / X-Internal-Token：
 *   - gateway 直接转发到后端，不校验 internal token
 *   - 后端从请求 body（登录/注册/刷新）或 Bearer JWT（/me）中获取 orgName
 */

import { NextRequest, NextResponse } from 'next/server'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

/**
 * 将请求转发到 gateway end-user auth 端点。
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

  // 透传 Authorization 头（/me 端点需要 Bearer token）
  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const body = method !== 'GET' && method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method, headers, body })
  } catch {
    return NextResponse.json({ error: { code: 'NETWORK_ERROR', message: '后端服务不可达' } }, { status: 502 })
  }

  const resBody = await upstreamRes.arrayBuffer()
  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  // 透传响应头（Set-Cookie 等）
  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase()))
      return
    response.headers.append(key, value)
  })

  return response
}
