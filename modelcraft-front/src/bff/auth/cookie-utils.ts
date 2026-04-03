// src/bff/auth/cookie-utils.ts
import { cookies } from 'next/headers'

const COOKIE_NAME = 'refresh_token'
const COOKIE_MAX_AGE = 7 * 24 * 60 * 60 // 7 days in seconds

export function setRefreshTokenCookie(token: string): void {
  cookies().set(COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: COOKIE_MAX_AGE,
    path: '/', // Must be '/' so middleware and all BFF routes can read it
  })
}

export function getRefreshTokenFromCookie(): string | undefined {
  return cookies().get(COOKIE_NAME)?.value
}

export function clearRefreshTokenCookie(): void {
  cookies().delete(COOKIE_NAME)
}
