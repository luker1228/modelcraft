// BFF: /api/bff/graphql/end-user/org/[orgName]
// 直接透传到网关 /end-user/graphql/org/{orgName}

import { NextRequest, NextResponse } from 'next/server'
import { endUserOrgGraphQL } from '@/app/api/bff/gateway-routes'

type Params = { orgName: string }

async function handler(req: NextRequest, { params }: { params: Promise<Params> }) {
  const { orgName } = await params
  const upstreamUrl = endUserOrgGraphQL(orgName)

  console.log('[BFF] end-user/org handler hit', { method: req.method, orgName, upstreamUrl })

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

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
  console.log('[BFF] end-user/org forwarding headers', {
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
    console.log('[BFF] end-user/org upstream response', { status: upstreamRes.status, upstreamUrl })
  } catch (err) {
    console.error('[BFF] end-user/org upstream unreachable', { upstreamUrl, err })
    return NextResponse.json(
      { errors: [{ message: '网关不可达' }] },
      { status: 502 }
    )
  }

  const resBody = await upstreamRes.arrayBuffer()

  if (upstreamRes.status >= 400) {
    const bodyText = Buffer.from(resBody).toString('utf-8')
    console.error('[BFF] end-user/org upstream error body:', bodyText)
  }

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
