// BFF: /end-user/graphql/org/[...path]
// Proxies end-user GraphQL requests (PAT Bearer token) to the backend.
// Used by the CLI (mc) for catalog, describe, and run commands.
//
// Path patterns:
//   /end-user/graphql/org/{orgName}
//   /end-user/graphql/org/{orgName}/project/{projectSlug}
//   /end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}

import { NextRequest, NextResponse } from 'next/server'
import {
  endUserOrgGraphQL,
  endUserProjectGraphQL,
  endUserRuntimeGraphQL,
} from '@/app/api/bff/gateway-routes'

function buildUpstreamUrl(segments: string[]): string {
  // segments = [orgName, ...]
  if (segments.length === 0) {
    throw new Error('end-user GraphQL requires at least an orgName')
  }

  const orgName = segments[0]

  // /end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
  if (
    segments.length === 7 &&
    segments[1] === 'project' &&
    segments[3] === 'db' &&
    segments[5] === 'model'
  ) {
    return endUserRuntimeGraphQL(orgName, segments[2], segments[4], segments[6])
  }

  // /end-user/graphql/org/{orgName}/project/{projectSlug}
  if (segments.length === 3 && segments[1] === 'project') {
    return endUserProjectGraphQL(orgName, segments[2])
  }

  // /end-user/graphql/org/{orgName}
  return endUserOrgGraphQL(orgName)
}

type Params = { path: string[] }

async function handler(
  req: NextRequest,
  { params }: { params: Promise<Params> },
) {
  const { path: segments } = await params
  const upstreamUrl = buildUpstreamUrl(segments)

  console.log('[BFF] end-user GraphQL handler hit', {
    method: req.method,
    segments,
    upstreamUrl,
  })

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

  const body =
    req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method: req.method, headers, body })
  } catch (err) {
    console.error('[BFF] end-user GraphQL upstream unreachable', {
      upstreamUrl,
      err,
    })
    return NextResponse.json(
      { errors: [{ message: '网关不可达' }] },
      { status: 502 },
    )
  }

  const resBody = await upstreamRes.arrayBuffer()
  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  upstreamRes.headers.forEach((value, key) => {
    if (
      ['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(
        key.toLowerCase(),
      )
    )
      return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
