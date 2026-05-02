import { NextRequest, NextResponse } from 'next/server'

function decodeOrgNameFromBearer(req: NextRequest): string | null {
  const auth = req.headers.get('authorization')
  if (!auth?.startsWith('Bearer ')) return null

  try {
    const token = auth.slice('Bearer '.length)
    const payload = token.split('.')[1]
    if (!payload) return null

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const json = Buffer.from(normalized, 'base64').toString('utf-8')
    const decoded = JSON.parse(json) as { org_name?: string }
    return decoded.org_name ?? null
  } catch {
    return null
  }
}

export async function GET(req: NextRequest) {
  const orgName = decodeOrgNameFromBearer(req) ?? req.nextUrl.searchParams.get('orgName')
  if (!orgName) {
    return NextResponse.json(
      { error: { code: 'PARAM_INVALID', message: 'orgName is required' } },
      { status: 400 }
    )
  }

  const upstreamRes = await fetch(
    `${req.nextUrl.origin}/api/bff/org/${encodeURIComponent(orgName)}/end-user/auth/me`,
    {
      method: 'GET',
      headers: {
        Authorization: req.headers.get('authorization') ?? '',
      },
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
