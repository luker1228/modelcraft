// src/app/api/bff/auth/refresh/route.ts
import { NextResponse } from 'next/server'
import { callGoRefresh } from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import {
  getRefreshTokenFromCookie,
  setRefreshTokenCookie,
  clearRefreshTokenCookie,
} from '@/bff/auth/cookie-utils'

export async function POST() {
  const refreshToken = getRefreshTokenFromCookie()
  if (!refreshToken) {
    return NextResponse.json({ error: 'No refresh token' }, { status: 401 })
  }

  try {
    const { userId, refreshToken: newRefreshToken } = await callGoRefresh(refreshToken)

    // Rotate httpOnly Cookie
    setRefreshTokenCookie(newRefreshToken)

    // Issue new Access Token
    const accessToken = await signAccessToken(userId)
    return NextResponse.json({ accessToken, expiresIn: 3600 })
  } catch {
    // Token reuse or expired - clear Cookie, force re-login
    clearRefreshTokenCookie()
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }
}
