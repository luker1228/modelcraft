/**
 * BFF End-User Auth Proxy Helper
 *
 * 将前端 /api/bff/org/[orgName]/end-user/auth/* 请求转发到后端
 * /internal/v1/end-user/auth/*，注入 X-Org-Name 和 X-Internal-Token 头。
 */

import { NextRequest, NextResponse } from 'next/server'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'
const INTERNAL_TOKEN = process.env.INTERNAL_TOKEN ?? ''

/**
 * 将请求转发到后端 end-user auth 端点。
 *
 * @param req - 原始 Next.js 请求
 * @param orgName - 从路由参数提取的组织名
 * @param path - 后端路径（如 "login"、"register"、"refresh"）
 * @param method - HTTP 方法（POST / GET）
 */
export async function proxyEndUserAuth(
  req: NextRequest,
  orgName: string,
  path: string,
  method: string = 'POST'
): Promise<NextResponse> {
  const upstreamUrl = `${BACKEND_URL}/api/end-user/auth/${path}`

  const headers = new Headers()
  headers.set('Content-Type', 'application/json')
  headers.set('X-Org-Name', orgName)
  headers.set('X-Internal-Token', INTERNAL_TOKEN)

  // 转发 Authorization 头（/me 端点需要）
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
