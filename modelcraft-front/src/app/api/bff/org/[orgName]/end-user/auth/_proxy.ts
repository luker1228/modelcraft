/**
 * BFF End-User Auth Proxy Helper
 *
 * 将前端 /api/bff/org/[orgName]/end-user/auth/* 请求透传到 gateway
 * /api/end-user/auth/*。
 *
 * Cookie 策略：
 *   - login / register / refresh / select-project 成功时，
 *     从 response body 读取 refreshToken，写入 mc_enduser_refresh_token HttpOnly cookie。
 *   - logout 成功时清除该 cookie。
 *   - /me 等只读接口直接透传，不操作 cookie。
 */

import { NextRequest, NextResponse } from 'next/server'
import { END_USER_REFRESH_COOKIE } from '@/middleware'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

const COOKIE_MAX_AGE = 60 * 60 * 24 * 30 // 30 天

/** 需要写 cookie 的接口 */
const WRITE_COOKIE_PATHS = new Set(['login', 'register', 'refresh', 'select-project'])

/** 需要清除 cookie 的接口 */
const CLEAR_COOKIE_PATHS = new Set(['logout'])

/**
 * 将请求转发到 gateway end-user auth 端点，并管理 mc_enduser_refresh_token cookie。
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

  // 清除 cookie（logout）
  if (CLEAR_COOKIE_PATHS.has(path)) {
    const response = new NextResponse(resBodyBuf, {
      status: upstreamRes.status,
      statusText: upstreamRes.statusText,
    })
    upstreamRes.headers.forEach((value, key) => {
      if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
      response.headers.append(key, value)
    })
    response.cookies.delete(END_USER_REFRESH_COOKIE)
    return response
  }

  // 写 cookie（login / register / refresh / select-project）
  if (WRITE_COOKIE_PATHS.has(path) && upstreamRes.ok) {
    try {
      const json = JSON.parse(Buffer.from(resBodyBuf).toString('utf-8')) as {
        refreshToken?: string
      }
      if (json.refreshToken) {
        const response = new NextResponse(resBodyBuf, {
          status: upstreamRes.status,
          statusText: upstreamRes.statusText,
        })
        upstreamRes.headers.forEach((value, key) => {
          if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
          response.headers.append(key, value)
        })
        response.cookies.set(END_USER_REFRESH_COOKIE, json.refreshToken, {
          httpOnly: true,
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'lax',
          path: '/',
          maxAge: COOKIE_MAX_AGE,
        })
        return response
      }
    } catch {
      // JSON 解析失败时回退到默认透传
    }
  }

  // 默认透传
  const response = new NextResponse(resBodyBuf, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })
  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })
  return response
}
