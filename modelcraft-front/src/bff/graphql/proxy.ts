// src/bff/graphql/proxy.ts
// Shared GraphQL proxy utility for BFF route handlers.
// Forwards the request body to Go backend with X-Internal-Token authentication.

import { type NextRequest, NextResponse } from 'next/server'

const GO_BACKEND_INTERNAL_URL = process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'
const INTERNAL_TOKEN = process.env.INTERNAL_TOKEN ?? process.env.INTERNAL_SERVICE_TOKEN ?? ''

/**
 * Proxy a GraphQL POST request to the Go backend using X-Internal-Token.
 * @param req - The incoming NextRequest
 * @param backendPath - The backend path to forward to (e.g. /graphql/org/myorg/)
 */
export async function proxyGraphQL(req: NextRequest, backendPath: string): Promise<NextResponse> {
  const url = `${GO_BACKEND_INTERNAL_URL}${backendPath}`

  const body = await req.text()

  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Internal-Token': INTERNAL_TOKEN,
      'X-Request-ID': req.headers.get('x-request-id') ?? `bff-gql-${Date.now()}`,
    },
    body,
    cache: 'no-store',
  })

  const data = await res.text()

  return new NextResponse(data, {
    status: res.status,
    headers: { 'Content-Type': 'application/json' },
  })
}
