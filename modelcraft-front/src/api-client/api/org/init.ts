import { NextRequest, NextResponse } from 'next/server'

/**
 * POST /api/org/init
 * Initialize (or re-initialize) the current user's organization.
 * Idempotent: returns the existing organization if the user already has one.
 */
export async function POST(request: NextRequest) {
  const contentType = request.headers.get('content-type')
  if (!contentType?.includes('application/json')) {
    return NextResponse.json(
      { error: 'Content-Type must be application/json' },
      { status: 415 }
    )
  }

  // Get authorization header
  const authHeader = request.headers.get('authorization')
  if (!authHeader) {
    return NextResponse.json(
      { error: 'Authorization token is required' },
      { status: 401 }
    )
  }

  if (!authHeader.startsWith('Bearer ')) {
    return NextResponse.json(
      { error: 'Authorization header must use Bearer scheme' },
      { status: 401 }
    )
  }

  let body: { organizationName?: unknown; displayName?: unknown }
  try {
    body = await request.json() as { organizationName?: unknown; displayName?: unknown }
  } catch {
    return NextResponse.json(
      { error: 'Invalid JSON body' },
      { status: 400 }
    )
  }

  const { organizationName, displayName } = body

  // Forward request to backend API
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
  const orgInitUrl = `${backendUrl}/api/org/init`

  const response = await fetch(orgInitUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': authHeader,
    },
    body: JSON.stringify({
      organizationName: typeof organizationName === 'string' ? organizationName : '',
      displayName: typeof displayName === 'string' ? displayName : '',
    }),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Failed to initialize organization' })) as Record<string, unknown>
    return NextResponse.json(errorData, { status: response.status })
  }

  const data = await response.json().catch(() => null) as Record<string, unknown> | null
  if (!data || typeof data !== 'object') {
    return NextResponse.json(
      { error: 'Invalid response from organization service' },
      { status: 502 }
    )
  }

  return NextResponse.json(data)
}
