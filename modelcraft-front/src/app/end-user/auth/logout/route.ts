import { NextRequest, NextResponse } from 'next/server'

type LogoutBody = {
  orgName?: string
  refreshToken?: string
}

export async function POST(req: NextRequest) {
  let body: LogoutBody
  try {
    body = (await req.json()) as LogoutBody
  } catch {
    return NextResponse.json(
      { error: { code: 'PARAM_INVALID', message: 'Invalid request body' } },
      { status: 400 }
    )
  }

  if (!body.orgName) {
    return NextResponse.json(
      { error: { code: 'PARAM_INVALID', message: 'orgName is required' } },
      { status: 400 }
    )
  }

  const upstreamRes = await fetch(
    `${req.nextUrl.origin}/api/bff/org/${encodeURIComponent(body.orgName)}/end-user/auth/logout`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: body.refreshToken ?? '' }),
      cache: 'no-store',
    }
  )

  const buf = await upstreamRes.arrayBuffer()
  const response = new NextResponse(buf, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  return response
}
