// src/bff/end-user/end-user-jwt-utils.ts
// 终端用户 JWT 签发/验证工具（对称 bff/auth/jwt-utils.ts）
// 关键差异：issuer = 'modelcraft-end-user'（区别于开发者的 'modelcraft'），payload 含 role/org_name/project_slug

import { SignJWT, jwtVerify } from 'jose'

const getJWTSecret = () => {
  const secret = process.env.JWT_SECRET
  if (!secret) throw new Error('JWT_SECRET environment variable is required')
  return new TextEncoder().encode(secret)
}

// issuer 与开发者不同，天然隔离
const ISSUER = 'modelcraft-end-user' // 开发者 issuer = 'modelcraft'
const EXPIRY = '1h'

export interface EndUserTokenPayload {
  userId: string
  orgName: string
  projectSlug: string
}

/**
 * 签发终端用户 access token
 * payload 含 sub(userId) + org_name + project_slug + role="end_user"
 * @param payload - 包含 userId, orgName, projectSlug 的 payload
 * @returns 签名后的 JWT
 */
export async function signEndUserAccessToken(payload: EndUserTokenPayload): Promise<string> {
  return new SignJWT({
    sub: payload.userId,
    org_name: payload.orgName,
    project_slug: payload.projectSlug,
    role: 'end_user',
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuer(ISSUER)
    .setIssuedAt()
    .setExpirationTime(EXPIRY)
    .sign(getJWTSecret())
}

/**
 * 验证终端用户 access token
 * 检查 issuer，防止开发者 token 被用于终端用户场景
 * @param token - JWT token
 * @returns 解析后的 payload
 * @throws 验证失败时抛出错误
 */
export async function verifyEndUserAccessToken(token: string): Promise<EndUserTokenPayload> {
  const { payload } = await jwtVerify(token, getJWTSecret(), { issuer: ISSUER })

  return {
    userId: payload.sub as string,
    orgName: payload['org_name'] as string,
    projectSlug: payload['project_slug'] as string,
  }
}
