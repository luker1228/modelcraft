// src/app/api/bff/auth/login/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { callGoLogin } from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import { setRefreshTokenCookie } from '@/bff/auth/cookie-utils'

export async function POST(req: NextRequest) {
  let body: { phone?: unknown; password?: unknown }
  try {
    body = (await req.json()) as { phone?: unknown; password?: unknown }
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const { phone, password } = body
  if (!phone || typeof phone !== 'string') {
    return NextResponse.json({ error: 'Phone number required' }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: 'Password required' }, { status: 400 })
  }

  try {
    // 1. Call Go Backend login with phone + password
    const { userId, refreshToken } = await callGoLogin({ phone, password })

    // 2. BFF signs Access Token
    const accessToken = await signAccessToken(userId)

    // 3. Set httpOnly Cookie with Refresh Token
    setRefreshTokenCookie(refreshToken)

    return NextResponse.json({ accessToken, expiresIn: 3600 })
  } catch (err) {
    console.error('BFF login error:', err)
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Authentication failed' },
      { status: 401 },
    )
  }
}
