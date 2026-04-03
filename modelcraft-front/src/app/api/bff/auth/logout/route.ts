// src/app/api/bff/auth/logout/route.ts
import { NextResponse } from 'next/server'
import { callGoLogout } from '@/bff/auth/go-auth-client'
import { getRefreshTokenFromCookie, clearRefreshTokenCookie } from '@/bff/auth/cookie-utils'

export async function POST() {
  const refreshToken = getRefreshTokenFromCookie()

  if (refreshToken) {
    await callGoLogout(refreshToken)
  }

  clearRefreshTokenCookie()
  return new NextResponse(null, { status: 204 })
}
