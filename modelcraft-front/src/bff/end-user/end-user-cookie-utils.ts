// src/bff/end-user/end-user-cookie-utils.ts
// 终端用户 refresh token Cookie 操作（对称 bff/auth/cookie-utils.ts）
// 为确保 /api/bff/end-user/auth/refresh 能读取 Cookie，path 必须覆盖 /api 路径。
// 项目隔离通过 token 内的 org/project claims 校验保证。

import { cookies } from 'next/headers'

const COOKIE_NAME = 'end_user_refresh_token'
const COOKIE_MAX_AGE = 7 * 24 * 60 * 60 // 7 days in seconds

/**
 * 写入终端用户 refresh token Cookie
 * @param token - refresh token
 * @param orgName - 组织名称
 * @param projectSlug - 项目标识
 */
export function setEndUserRefreshTokenCookie(
  token: string,
  orgName: string,
  projectSlug: string
): void {
  void orgName
  void projectSlug
  cookies().set(COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: COOKIE_MAX_AGE,
    path: '/',
  })
}

/**
 * 读取终端用户 refresh token Cookie
 * 用于 BFF Route Handler 内读取
 * @returns refresh token 或 undefined
 */
export function getEndUserRefreshTokenFromCookie(): string | undefined {
  return cookies().get(COOKIE_NAME)?.value
}

/**
 * 清除终端用户 refresh token Cookie
 * 必须传入 orgName 和 projectSlug 以精确匹配 path
 * @param orgName - 组织名称
 * @param projectSlug - 项目标识
 */
export function clearEndUserRefreshTokenCookie(orgName: string, projectSlug: string): void {
  void orgName
  void projectSlug
  cookies().set(COOKIE_NAME, '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: 0, // 立即过期
    path: '/',
  })
}
