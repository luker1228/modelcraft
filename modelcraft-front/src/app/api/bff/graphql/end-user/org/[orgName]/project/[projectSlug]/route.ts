// BFF: /api/bff/graphql/end-user/org/[orgName]/project/[projectSlug]
// 直接透传到网关 /end-user/graphql/org/{orgName}/project/{projectSlug}

import { NextRequest, NextResponse } from 'next/server'
import { endUserProjectGraphQL } from '@/app/api/bff/gateway-routes'

type Params = { orgName: string; projectSlug: string }

async function handler(req: NextRequest, { params }: { params: Promise<Params> }) {
  const { orgName, projectSlug } = await params
  const upstreamUrl = endUserProjectGraphQL(orgName, projectSlug)

  console.log('[BFF] end-user/org/project handler hit', { method: req.method, orgName, projectSlug, upstreamUrl })

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  const xAction = req.headers.get('X-Action')
  if (xAction) headers.set('X-Action', xAction)
  console.log('[BFF] end-user/org/project forwarding headers', { 'X-Action': xAction, hasAuth: !!authHeader })

  const body = req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method: req.method, headers, body })
    console.log('[BFF] end-user/org/project upstream response', { status: upstreamRes.status, upstreamUrl })
  } catch (err) {
    console.error('[BFF] end-user/org/project upstream unreachable', { upstreamUrl, err })
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
