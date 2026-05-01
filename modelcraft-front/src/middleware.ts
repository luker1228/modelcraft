import { NextRequest, NextResponse } from 'next/server'

/**
 * Next.js Middleware — Single auth gate for all protected routes.
 *
 * Strategy:
 *  - Public routes (/login, /register, /api/*) are allowed through unconditionally.
 *  - All other routes require the `refresh_token` httpOnly cookie to be present.
 *    If missing, redirect to /login with the original URL as `returnUrl`.
 *  - We do NOT validate the token here (that would require calling the backend on every
 *    request). We only check presence. The actual token exchange happens client-side via
 *    silent refresh (/api/bff/auth/refresh) after the page loads.
 *
 * End-User Auth:
 *  - End-user routes use /end-user/{orgName}/{projectSlug}/*
 *  - /end-user/{orgName}/login is public
 *  - /end-user/{orgName}/{projectSlug}/data/* requires end_user_refresh_token
 *
 * Legacy End-User Routes:
 *  - /org/{org}/project/{project}/user/* and /data/* are retired immediately.
 *  - Middleware lets them pass through so Next.js returns 404 directly.
 */

// ============================================
// 开发者认证配置（现有，完整保留，不变）
// ============================================
const PUBLIC_PATHS = [
  '/login',
  '/register',
]

const COOKIE_NAME = 'mc_refresh_token'

// ============================================
// 终端用户认证配置（新增）
// ============================================
const END_USER_COOKIE = 'end_user_refresh_token'

const END_USER_LOGIN_RE = /^\/end-user\/[^/]+\/login$/
const END_USER_DATA_RE = /^\/end-user\/[^/]+\/[^/]+\/data(\/.*)?$/
const LEGACY_END_USER_RE = /^\/org\/[^/]+\/project\/[^/]+\/(user(\/.*)?|data(\/.*)?)$/

function extractEndUserParams(pathname: string): { orgName: string; projectSlug: string } | null {
  const match = pathname.match(/^\/end-user\/([^/]+)\/([^/]+)/)
  if (!match) return null
  return { orgName: match[1], projectSlug: match[2] }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow all /api/* routes (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/')) {
    return NextResponse.next()
  }

  // Allow /auth/* routes (proxied to backend, must not be gated)
  if (pathname.startsWith('/auth/')) {
    return NextResponse.next()
  }

  // ===== END USER AUTH START =====
  // 终端用户路径分支（必须在开发者鉴权前处理）

  // Legacy 路径立即切断：不做兼容跳转，让 Next.js 直接 404
  if (LEGACY_END_USER_RE.test(pathname)) {
    return NextResponse.next()
  }

  if (END_USER_LOGIN_RE.test(pathname)) {
    return NextResponse.next()
  }

  if (END_USER_DATA_RE.test(pathname)) {
    const hasEndUserToken = request.cookies.has(END_USER_COOKIE)
    console.log(`[middleware] end-user route ${pathname} — cookie present: ${hasEndUserToken}`)

    if (!hasEndUserToken) {
      const params = extractEndUserParams(pathname)
      if (params) {
        const loginUrl = new URL(`/end-user/${params.orgName}/login`, request.url)
        loginUrl.searchParams.set('redirect', pathname)
        console.log(`[middleware] No end-user token, redirecting to: ${loginUrl.toString()}`)
        return NextResponse.redirect(loginUrl)
      }
    }

    return NextResponse.next()
  }

  // 非 data/login 的 /end-user 路径不归 developer 鉴权，交给 Next.js 正常路由（通常 404）
  if (pathname.startsWith('/end-user/')) {
    return NextResponse.next()
  }

  // ===== END USER AUTH END =====

  // ============================================
  // 开发者路由守卫（现有逻辑，完整保留，不做任何修改）
  // ============================================

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
