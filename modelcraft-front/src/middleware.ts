import { NextRequest, NextResponse } from 'next/server'
import { END_USER_LOGIN_PATH, TENANT_LOGIN_PATH, TENANT_REGISTER_PATH } from '@shared/constants/routes'

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
 *    require the mc_refresh_token HttpOnly cookie.
 *    If missing, redirect to /end-user/[orgName]/login.
 */

// ============================================
// 开发者认证配置
// ============================================
const DEV_PUBLIC_PATHS = [TENANT_LOGIN_PATH, TENANT_REGISTER_PATH, '/login']
const DEV_REFRESH_COOKIE = 'mc_refresh_token'

// ============================================
// 终端用户认证配置
// ============================================
export const END_USER_REFRESH_COOKIE = 'mc_refresh_token'

/**
 * 终端用户公开路径（仅这两类）：
 *   /end-user/{orgName}/login
 *   /end-user/{orgName}/no-project-access
 */
const END_USER_PUBLIC_PATH_RE = /^\/end-user(?:\/login|\/[^/]+\/(login|no-project-access))\/?$/

/**
 * 终端用户受保护路径（仅真实业务路由）：
 *   /end-user/{orgName}/workspace
 *   /end-user/{orgName}/projects/{projectSlug}/...
 */
const END_USER_WORKSPACE_RE = /^\/end-user\/([^/]+)\/workspace\/?$/
const END_USER_PROJECT_RE = /^\/end-user\/([^/]+)\/projects\/[^/]+(\/.*)?$/

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow /api/* and /auth/* unconditionally (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/') || pathname.startsWith('/auth/')) {
    return NextResponse.next()
  }

  // ===== END USER AUTH =====

  if (pathname.startsWith('/end-user/')) {
    // 公开路径（login / no-project-access）
    if (END_USER_PUBLIC_PATH_RE.test(pathname)) {
      return NextResponse.next()
    }

    // 受保护路径：仅 workspace 与 projects/*
    const workspaceMatch = END_USER_WORKSPACE_RE.exec(pathname)
    const projectMatch = END_USER_PROJECT_RE.exec(pathname)
    const protectedOrgName = workspaceMatch?.[1] ?? projectMatch?.[1]
    if (protectedOrgName) {
      const hasToken = request.cookies.has(END_USER_REFRESH_COOKIE)
      if (!hasToken) {
        const loginUrl = new URL(END_USER_LOGIN_PATH, request.url)
        loginUrl.searchParams.set('redirect', pathname)
        return NextResponse.redirect(loginUrl)
      }
      return NextResponse.next()
    }

    // 其余 /end-user/* 路径直接放行，让 Next.js 返回 404
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
