import { NextRequest, NextResponse } from 'next/server'

/**
 * POST /api/auth/token
 * Proxy token exchange request to backend
 */
export async function POST(request: NextRequest) {
  const contentType = request.headers.get('content-type')
  if (!contentType?.includes('application/json')) {
    return NextResponse.json(
      { error: 'Content-Type must be application/json' },
      { status: 415 }
    )
  }

  let body: { code?: unknown }
  try {
    body = await request.json() as { code?: unknown }
  } catch {
    return NextResponse.json(
      { error: 'Invalid JSON body' },
      { status: 400 }
    )
  }

  const { code } = body

  if (!code || typeof code !== 'string') {
    return NextResponse.json(
      { error: 'Authorization code is required' },
      { status: 400 }
    )
  }

  // Forward request to backend API
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
  const tokenUrl = `${backendUrl}/api/auth/token`

  const response = await fetch(tokenUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ code }),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Token exchange failed' })) as Record<string, unknown>
    return NextResponse.json(errorData, { status: response.status })
  }

  const data = await response.json().catch(() => null) as Record<string, unknown> | null
  if (!data || typeof data !== 'object' || !('accessToken' in data || 'access_token' in data)) {
    return NextResponse.json(
      { error: 'Invalid response from authentication service' },
      { status: 502 }
    )
  }

  return NextResponse.json(data)
}
