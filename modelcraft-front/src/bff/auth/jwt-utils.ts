// src/bff/auth/jwt-utils.ts
import { SignJWT, jwtVerify } from 'jose'

const getJWTSecret = () => {
  const secret = process.env.JWT_SECRET
  if (!secret) throw new Error('JWT_SECRET environment variable is required')
  return new TextEncoder().encode(secret)
}

const ISSUER = 'modelcraft'
const EXPIRY = '1h'

export async function signAccessToken(userId: string): Promise<string> {
  return new SignJWT({ user_id: userId })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuer(ISSUER)
    .setIssuedAt()
    .setExpirationTime(EXPIRY)
    .sign(getJWTSecret())
}

export async function verifyAccessToken(token: string): Promise<{ userId: string }> {
  const { payload } = await jwtVerify(token, getJWTSecret(), { issuer: ISSUER })
  return { userId: payload['user_id'] as string }
}
