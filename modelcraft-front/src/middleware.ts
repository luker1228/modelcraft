import { NextRequest, NextResponse } from 'next/server'

/**
 * Next.js Middleware — Single auth gate for all protected routes.
 *
 * Strategy:
 *  - Public routes (/login, /auth/callback, /api/*) are allowed through unconditionally.
 *  - All other routes require the `refresh_token` httpOnly cookie to be present.
 *    If missing, redirect to /login with the original URL as `returnUrl`.
 *  - We do NOT validate the token here (that would require calling the backend on every
 *    request). We only check presence. The actual token exchange happens client-side via
 *    silent refresh (/api/bff/auth/refresh) after the page loads.
 */

const PUBLIC_PATHS = [
  '/login',
  '/auth/callback',
]

const COOKIE_NAME = 'refresh_token'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow all /api/* routes (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/')) {
    return NextResponse.next()
  }

  // Allow public pages
  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next()
  }

  // Check for refresh token cookie
  const hasRefreshToken = request.cookies.has(COOKIE_NAME)
  console.log(`[middleware] ${pathname} — cookie present: ${hasRefreshToken}`)

  if (!hasRefreshToken) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('returnUrl', pathname)
    console.log(`[middleware] No refresh token, redirecting to: ${loginUrl.toString()}`)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  // Run on all routes except static assets and Next.js internals
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
