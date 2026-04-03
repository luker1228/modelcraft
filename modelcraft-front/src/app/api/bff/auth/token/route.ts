// src/app/api/bff/auth/token/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { callGoLogin } from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import { setRefreshTokenCookie } from '@/bff/auth/cookie-utils'

async function exchangeCodeWithCasdoor(code: string): Promise<{
  externalId: string
  email: string
  name: string
}> {
  const casdoorUrl = process.env.NEXT_PUBLIC_CASDOOR_URL || ''
  const clientId = process.env.NEXT_PUBLIC_CASDOOR_CLIENT_ID || ''
  const clientSecret = process.env.CASDOOR_CLIENT_SECRET || ''
  const redirectUri =
    process.env.NEXT_PUBLIC_CASDOOR_REDIRECT_URI ||
    `${process.env.NEXTAUTH_URL || 'http://localhost:3000'}/auth/callback`

  // Exchange code for access token with Casdoor
  const tokenRes = await fetch(`${casdoorUrl}/api/login/oauth/access_token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: clientId,
      client_secret: clientSecret,
      code,
      redirect_uri: redirectUri,
    }),
  })

  if (!tokenRes.ok) {
    throw new Error(`Casdoor token exchange failed: ${tokenRes.status}`)
  }

  const tokenData = (await tokenRes.json()) as { access_token?: string }
  const accessToken: string = tokenData.access_token ?? ''

  if (!accessToken) {
    throw new Error('No access token from Casdoor')
  }

  // Decode Casdoor JWT to get user info (no verification needed for server-side internal call)
  const parts = accessToken.split('.')
  if (parts.length < 2) throw new Error('Invalid Casdoor JWT format')

  const payload = JSON.parse(
    Buffer.from(parts[1], 'base64').toString('utf8')
  ) as Record<string, string | undefined>

  return {
    externalId: payload.id || payload.sub || payload.name || '',
    email: payload.email || '',
    name: payload.name || payload.displayName || '',
  }
}

export async function POST(req: NextRequest) {
  let body: { code?: unknown }
  try {
    body = await req.json() as { code?: unknown }
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const { code } = body
  if (!code || typeof code !== 'string') {
    return NextResponse.json({ error: 'Authorization code required' }, { status: 400 })
  }

  try {
    // 1. Exchange code with Casdoor to get user info
    const { externalId, email, name } = await exchangeCodeWithCasdoor(code)

    // 2. Call Go Backend login (user sync + generate Refresh Token)
    const { userId, refreshToken } = await callGoLogin({ externalId, email, name })

    // 3. BFF signs Access Token
    const accessToken = await signAccessToken(userId)

    // 4. Set httpOnly Cookie with Refresh Token
    setRefreshTokenCookie(refreshToken)

    return NextResponse.json({ accessToken, expiresIn: 3600 })
  } catch (err) {
    console.error('BFF token exchange error:', err)
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Authentication failed' },
      { status: 401 }
    )
  }
}
