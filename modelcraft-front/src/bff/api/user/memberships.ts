import { NextRequest, NextResponse } from 'next/server'

/**
 * GET /api/user/memberships
 * Get current user's organization memberships
 */
export async function GET(request: NextRequest) {
  // Get authorization header from request
  const authHeader = request.headers.get('authorization')

  if (!authHeader) {
    return NextResponse.json(
      { error: 'Authorization header is required' },
      { status: 401 }
    )
  }

  if (!authHeader.startsWith('Bearer ')) {
    return NextResponse.json(
      { error: 'Authorization header must use Bearer scheme' },
      { status: 401 }
    )
  }

  // Forward request to backend API
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
  const membershipsUrl = `${backendUrl}/api/user/memberships`

  const response = await fetch(membershipsUrl, {
    method: 'GET',
    headers: {
      'Authorization': authHeader,
      'Content-Type': 'application/json',
    },
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Failed to get memberships' })) as Record<string, unknown>
    return NextResponse.json(errorData, { status: response.status })
  }

  const data = await response.json().catch(() => null) as Record<string, unknown> | null
  if (!data || typeof data !== 'object') {
    return NextResponse.json(
      { error: 'Invalid response from memberships service' },
      { status: 502 }
    )
  }

  return NextResponse.json(data)
}
