/**
 * Auth Proxy Route — /api/auth/[...path]
 *
 * Forwards all auth requests to the gateway and transparently passes back
 * Set-Cookie headers so mc_refresh_token is stored under the localhost domain.
 */

import { NextRequest, NextResponse } from 'next/server'

const GATEWAY_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

async function handler(req: NextRequest, { params }: { params: { path: string[] } }) {
  const pathSegments = (await params).path
  const upstreamUrl = `${GATEWAY_URL}/auth/${pathSegments.join('/')}`

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')
  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const body = req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  console.log(`[auth-proxy] ${req.method} ${upstreamUrl}`, body?.substring(0, 200))

  const upstreamRes = await fetch(upstreamUrl, {
    method: req.method,
    headers,
    body,
  })

  console.log(`[auth-proxy] upstream response: ${upstreamRes.status}`)

  const resBody = await upstreamRes.arrayBuffer()

  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  // Transparently forward all response headers, especially Set-Cookie
  upstreamRes.headers.forEach((value, key) => {
    // Skip headers that Next.js manages
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
export const PUT = handler
export const PATCH = handler
export const DELETE = handler
