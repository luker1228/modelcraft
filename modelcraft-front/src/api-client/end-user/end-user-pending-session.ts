// src/bff/end-user/end-user-pending-session.ts
// v1 pending session：登录成功但需选择 Project 时，签发短期临时 token
// 区别于正式 access token：不含 projectSlug，issuer = 'modelcraft-end-user-pending'

import { SignJWT, jwtVerify } from 'jose'
import { cookies } from 'next/headers'
import type { EndUserAccessibleProject, EndUserPendingSessionPayload } from '@/types/end-user-auth'

const getJWTSecret = () => {
  const secret = process.env.JWT_SECRET
  if (!secret) throw new Error('JWT_SECRET environment variable is required')
  return new TextEncoder().encode(secret)
}

const PENDING_ISSUER = 'modelcraft-end-user-pending'
const PENDING_COOKIE = 'end_user_pending_session'
const PENDING_EXPIRY = '15m' // 15 分钟内完成 project 选择

export async function signPendingSessionToken(
  userId: string,
  orgName: string,
  projects: EndUserAccessibleProject[]
): Promise<string> {
  return new SignJWT({ sub: userId, org_name: orgName, projects })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuer(PENDING_ISSUER)
    .setIssuedAt()
    .setExpirationTime(PENDING_EXPIRY)
    .sign(getJWTSecret())
}

export async function verifyPendingSessionToken(
  token: string
): Promise<EndUserPendingSessionPayload> {
  const { payload } = await jwtVerify(token, getJWTSecret(), { issuer: PENDING_ISSUER })
  return {
    userId: payload.sub as string,
    orgName: payload['org_name'] as string,
    projects: (payload['projects'] as EndUserAccessibleProject[]) ?? [],
  }
}

export function setPendingSessionCookie(token: string): void {
  cookies().set(PENDING_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: 15 * 60, // 15 分钟
    path: '/',
  })
}

export function getPendingSessionFromCookie(): string | undefined {
  return cookies().get(PENDING_COOKIE)?.value
}

export function clearPendingSessionCookie(): void {
  cookies().set(PENDING_COOKIE, '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: 0,
    path: '/',
  })
}
