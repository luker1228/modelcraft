// BFF: /api/bff/graphql/org/[orgName]
// 直接透传到网关 /graphql/org/{orgName}

import { NextRequest, NextResponse } from 'next/server'
import { tenantOrgGraphQL } from '@/app/api/bff/gateway-routes'

type Params = { orgName: string }

async function handler(req: NextRequest, { params }: { params: Promise<Params> }) {
  const { orgName } = await params
  const upstreamUrl = tenantOrgGraphQL(orgName)

  console.log('[BFF] org handler hit', { method: req.method, orgName, upstreamUrl })

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)
  const cookieHeader = req.headers.get('cookie')
  if (cookieHeader) headers.set('Cookie', cookieHeader)

  const xRequestId = req.headers.get('X-Request-Id')
  if (xRequestId) headers.set('X-Request-Id', xRequestId)

  const xClientRequestId = req.headers.get('X-Client-Request-Id')
  if (xClientRequestId) headers.set('X-Client-Request-Id', xClientRequestId)

  const traceparent = req.headers.get('traceparent')
  if (traceparent) headers.set('traceparent', traceparent)

  const tracestate = req.headers.get('tracestate')
  if (tracestate) headers.set('tracestate', tracestate)

  const xAction = req.headers.get('X-Action')
  if (xAction) headers.set('X-Action', xAction)
  console.log('[BFF] org forwarding headers', {
    'X-Action': xAction,
    hasAuth: !!authHeader,
    xRequestId,
    xClientRequestId,
    traceparent,
  })

  const body = req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method: req.method, headers, body })
    console.log('[BFF] org upstream response', { status: upstreamRes.status, upstreamUrl })
  } catch (err) {
    console.error('[BFF] org upstream unreachable', { upstreamUrl, err })
    return NextResponse.json(
      { errors: [{ message: '网关不可达' }] },
      { status: 502 }
    )
  }

  const resBody = await upstreamRes.arrayBuffer()
  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })
  upstreamRes.headers.forEach((value, key) => {
    if (['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(key.toLowerCase())) return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
