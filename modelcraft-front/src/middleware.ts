import { NextRequest, NextResponse } from 'next/server'
import { TENANT_LOGIN_PATH, TENANT_REGISTER_PATH } from '@shared/constants/routes'

/**
 * Next.js Middleware — Single auth gate for all protected routes.
 *
 * Strategy:
 *  - Public routes (/tenant/login, /tenant/register, /api/*) are allowed through unconditionally.
 *  - All other routes require the `mc_refresh_token` httpOnly cookie to be present.
 *    If missing, redirect to /tenant/login with the original URL as `redirect`.
 *  - We do NOT validate the token here (that would require calling the backend on every
 *    request). We only check presence. The actual token exchange happens client-side via
 *    silent refresh (/api/bff/auth/refresh) after the page loads.
 */

// ============================================
// 开发者认证配置
// ============================================
const DEV_PUBLIC_PATHS = [TENANT_LOGIN_PATH, TENANT_REGISTER_PATH, '/login']
const DEV_REFRESH_COOKIE = 'mc_refresh_token'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow /api/* and /auth/* unconditionally (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/') || pathname.startsWith('/auth/')) {
    return NextResponse.next()
  }

  // ===== DEVELOPER AUTH =====

  if (pathname === '/login') {
    const tenantLoginUrl = new URL(TENANT_LOGIN_PATH, request.url)
    request.nextUrl.searchParams.forEach((value, key) => {
      tenantLoginUrl.searchParams.set(key, value)
    })
    return NextResponse.redirect(tenantLoginUrl)
  }

  if (
    pathname === '/' ||
    DEV_PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`))
  ) {
    return NextResponse.next()
  }

  const hasRefreshToken = request.cookies.has(DEV_REFRESH_COOKIE)
  if (!hasRefreshToken) {
    const loginUrl = new URL(TENANT_LOGIN_PATH, request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
