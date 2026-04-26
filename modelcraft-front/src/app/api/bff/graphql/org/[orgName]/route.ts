// src/app/api/bff/graphql/org/[orgName]/route.ts
// BFF GraphQL proxy for Org-scoped operations.
// Forwards requests to Go backend using X-Internal-Token — no Design JWT required.

import { type NextRequest, NextResponse } from 'next/server'
import { proxyGraphQL } from '@/bff/graphql/proxy'

interface RouteParams {
  params: Promise<{ orgName: string }>
}

export async function POST(req: NextRequest, { params }: RouteParams) {
  const { orgName } = await params
  return proxyGraphQL(req, `/graphql/org/${orgName}/`)
}

export async function GET(_req: NextRequest, { params }: RouteParams) {
  const { orgName } = await params
  return NextResponse.redirect(
    new URL(`/graphql/org/${orgName}/`, process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080')
  )
}
