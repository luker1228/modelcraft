import { NextRequest, NextResponse } from 'next/server'

/**
 * Next.js Middleware — Single auth gate for all protected routes.
 *
 * Strategy:
 *  - Public routes (/tenant/login, /register, /api/*) are allowed through unconditionally.
 *  - All other routes require the `mc_refresh_token` httpOnly cookie to be present.
 *    If missing, redirect to /tenant/login with the original URL as `redirect`.
 *  - We do NOT validate the token here (that would require calling the backend on every
 *    request). We only check presence. The actual token exchange happens client-side via
 *    silent refresh (/api/bff/auth/refresh) after the page loads.
 *
 * End-User Auth:
 *  - All /end-user/* routes are handled separately before developer auth.
 *  - Public end-user paths (login) are allowed through.
 *  - Protected end-user paths (/end-user/[orgName]/workspace, /end-user/[orgName]/[projectSlug]/*)
 *    require the mc_enduser_refresh_token HttpOnly cookie.
 *    If missing, redirect to /end-user/[orgName]/login.
 */

// ============================================
// 开发者认证配置
// ============================================
const DEV_PUBLIC_PATHS = ['/tenant/login', '/register']
const DEV_REFRESH_COOKIE = 'mc_refresh_token'

// ============================================
// 终端用户认证配置
// ============================================
export const END_USER_REFRESH_COOKIE = 'mc_enduser_refresh_token'

/**
 * 终端用户公开路径（精确后缀匹配，带 orgName 动态段）：
 *   /end-user/{orgName}/login
 */
const END_USER_PUBLIC_SUFFIXES_RE = /^\/end-user\/[^/]+\/login(\/.*)?$/

/**
 * 终端用户受保护路径：/end-user/{orgName}/workspace 及 /end-user/{orgName}/{any}/*
 * 需要 mc_enduser_refresh_token cookie。
 */
const END_USER_PROTECTED_RE = /^\/end-user\/([^/]+)\/([^/]+)(\/.*)?$/

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow /api/* and /auth/* unconditionally (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/') || pathname.startsWith('/auth/')) {
    return NextResponse.next()
  }

  // ===== END USER AUTH =====

  if (pathname.startsWith('/end-user/')) {
    // 公开路径（login / select-project / no-project-access）
    if (END_USER_PUBLIC_SUFFIXES_RE.test(pathname)) {
      return NextResponse.next()
    }

    // 受保护路径：/end-user/{orgName}/{projectSlug}/*
    const match = END_USER_PROTECTED_RE.exec(pathname)
    if (match) {
      const hasToken = request.cookies.has(END_USER_REFRESH_COOKIE)
      if (!hasToken) {
        const orgName = match[1]
        const loginUrl = new URL(`/end-user/${orgName}/login`, request.url)
        loginUrl.searchParams.set('redirect', pathname)
        return NextResponse.redirect(loginUrl)
      }
      return NextResponse.next()
    }

    // 其余 /end-user/* 路径直接放行，让 Next.js 返回 404
    return NextResponse.next()
  }

  // ===== DEVELOPER AUTH =====

  if (DEV_PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next()
  }

  const hasRefreshToken = request.cookies.has(DEV_REFRESH_COOKIE)
  if (!hasRefreshToken) {
    const loginUrl = new URL('/tenant/login', request.url)
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
