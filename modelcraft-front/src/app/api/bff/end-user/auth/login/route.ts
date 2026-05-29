import { NextRequest, NextResponse } from 'next/server'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

export async function POST(req: NextRequest) {
  const upstreamUrl = `${BACKEND_URL}/api/end-user/auth/login`

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const cookieHeader = req.headers.get('cookie')
  if (cookieHeader) headers.set('Cookie', cookieHeader)

  const body = await req.text()

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, {
      method: 'POST',
      headers,
      body,
    })
  } catch {
    return NextResponse.json(
      { error: { code: 'NETWORK_ERROR', message: '后端服务不可达' } },
      { status: 502 }
    )
  }

  const resBodyBuf = await upstreamRes.arrayBuffer()
  const response = new NextResponse(resBodyBuf, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

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

  const skipHeaders = new Set([
    'content-encoding',
    'content-length',
    'transfer-encoding',
    'connection',
    'set-cookie',
  ])
  upstreamRes.headers.forEach((value, key) => {
    if (skipHeaders.has(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  for (const cookieStr of setCookieValues) {
    response.headers.append('Set-Cookie', cookieStr)
  }

  return response
}
