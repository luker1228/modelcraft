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
 * End-User Auth Extension (2026-04-16):
 *  - End-user data routes (/org/{orgName}/project/{projectSlug}/data/*) require
 *    `end_user_refresh_token` cookie. If missing, redirect to end-user login page.
 *  - End-user login page (/org/{orgName}/project/{projectSlug}/user/login) is public.
 *  - This branch is checked BEFORE developer auth to avoid false captures.
 */

// ============================================
// 开发者认证配置（现有，完整保留，不变）
// ============================================
const PUBLIC_PATHS = [
  '/login',
  '/register',
]

const COOKIE_NAME = 'refresh_token'

// ============================================
// 终端用户认证配置（新增）
// ============================================
const END_USER_COOKIE = 'end_user_refresh_token'

/**
 * 判断路径是否属于终端用户数据路由（需要守卫）。
 * 匹配：/org/{orgName}/project/{projectSlug}/data 及其子路径
 */
function isEndUserDataRoute(pathname: string): boolean {
  return /^\/org\/[^/]+\/project\/[^/]+\/data(\/.*)?$/.test(pathname)
}

/**
 * 判断路径是否为终端用户登录页（公开，无需守卫）。
 * 匹配：/org/{orgName}/project/{projectSlug}/user/login
 */
function isEndUserLoginPage(pathname: string): boolean {
  return /^\/org\/[^/]+\/project\/[^/]+\/user\/login$/.test(pathname)
}

/**
 * 从路径中提取 orgName 和 projectSlug，用于构造重定向 URL。
 */
function extractProjectParams(
  pathname: string
): { orgName: string; projectSlug: string } | null {
  const match = pathname.match(/^\/org\/([^/]+)\/project\/([^/]+)/)
  if (!match) return null
  return { orgName: match[1], projectSlug: match[2] }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow all /api/* routes (BFF endpoints, rewrites, etc.)
  if (pathname.startsWith('/api/')) {
    return NextResponse.next()
  }

  // ===== END USER AUTH START =====
  // 终端用户路由守卫（新增分支，在开发者守卫之前判断）

  // 终端用户登录页本身：公开，直接放行
  if (isEndUserLoginPage(pathname)) {
    return NextResponse.next()
  }

  // 终端用户数据管理路由：需要 end_user_refresh_token cookie
  if (isEndUserDataRoute(pathname)) {
    const hasEndUserToken = request.cookies.has(END_USER_COOKIE)
    console.log(`[middleware] end-user route ${pathname} — cookie present: ${hasEndUserToken}`)

    if (!hasEndUserToken) {
      const params = extractProjectParams(pathname)
      if (params) {
        const loginUrl = new URL(
          `/org/${params.orgName}/project/${params.projectSlug}/user/login`,
          request.url
        )
        // 注意：终端用户用 redirect，开发者用 returnUrl
        loginUrl.searchParams.set('redirect', pathname)
        console.log(`[middleware] No end-user token, redirecting to: ${loginUrl.toString()}`)
        return NextResponse.redirect(loginUrl)
      }
    }

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
