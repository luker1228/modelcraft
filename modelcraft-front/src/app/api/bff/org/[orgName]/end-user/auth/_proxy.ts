/**
 * BFF End-User Auth Proxy Helper
 *
 * 将前端 /api/bff/org/[orgName]/end-user/auth/* 请求透传到 gateway
 * /api/end-user/auth/*。
 *
 * Cookie 策略：
 * - 后端通过 Set-Cookie header 直接设置 mc_enduser_refresh_token（HttpOnly）
 * - BFF 将上游 Set-Cookie header 原样透传给浏览器
 * - logout 路径：追加 Set-Cookie header 清除 cookie
 */

import { NextRequest, NextResponse } from 'next/server'
import { END_USER_REFRESH_COOKIE } from '@/middleware'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

/** 需要清除 cookie 的路径 */
const CLEAR_COOKIE_PATHS = new Set(['logout'])

/**
 * 将请求转发到 gateway end-user auth 端点。
 * 透传上游 Set-Cookie header（含 mc_enduser_refresh_token）。
 * 对 logout 路径：额外写入清除 cookie 的 Set-Cookie header。
 *
 * @param req    - 原始 Next.js 请求
 * @param path   - 后端路径段（如 "login"、"refresh"、"logout"）
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

  // Gateway 的 refresh/logout/select-project 依赖 HttpOnly cookie，
  // BFF 转发时必须把浏览器原始 Cookie 透传给上游。
  const cookieHeader = req.headers.get('cookie')
  if (cookieHeader) headers.set('Cookie', cookieHeader)

  const body = method !== 'GET' && method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method, headers, body })
  } catch {
    return NextResponse.json({ error: { code: 'NETWORK_ERROR', message: '后端服务不可达' } }, { status: 502 })
  }

  const resBodyBuf = await upstreamRes.arrayBuffer()

  // 收集上游所有 Set-Cookie header（getSetCookie 避免被合并）
  const setCookieValues: string[] = []
  try {
    const raw = upstreamRes.headers as unknown as { getSetCookie?: () => string[] }
    if (typeof raw.getSetCookie === 'function') {
      setCookieValues.push(...raw.getSetCookie())
    } else {
      const single = upstreamRes.headers.get('set-cookie')
      if (single) setCookieValues.push(single)
    }
  } catch {
    const single = upstreamRes.headers.get('set-cookie')
    if (single) setCookieValues.push(single)
  }

  // logout：追加清除 cookie 的指令
  if (CLEAR_COOKIE_PATHS.has(path)) {
    setCookieValues.push(
      `${END_USER_REFRESH_COOKIE}=; Path=/; Max-Age=0; HttpOnly; SameSite=Strict`
    )
  }

  // 构造响应（不在构造函数里传 headers，避免 Next.js 内部去重）
  const response = new NextResponse(resBodyBuf, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  // 透传响应 header（跳过 hop-by-hop 和 set-cookie，set-cookie 单独处理）
  const skipHeaders = new Set(['content-encoding', 'content-length', 'transfer-encoding', 'connection', 'set-cookie'])
  upstreamRes.headers.forEach((value, key) => {
    if (skipHeaders.has(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  // 写入所有 Set-Cookie（用 append 支持多个 cookie）
  for (const cookieStr of setCookieValues) {
    response.headers.append('Set-Cookie', cookieStr)
  }

  return response
}
