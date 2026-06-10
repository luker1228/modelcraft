import { createHmac } from 'crypto'
import 'dotenv/config'

const JWT_SECRET = process.env.JWT_SECRET ?? 'modelcraft_dev_secret_please_change_in_production'

function base64url(input: string | Buffer): string {
  const b64 = Buffer.from(input).toString('base64')
  return b64.replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_')
}

/**
 * Generate a minimal test-only HMAC-SHA256 JWT for BDD startup fallback.
 * This is not the production auth format. Claims: { user_id, iat, exp }
 */
export function signJWT(userId: string, expiresInSeconds = 3600): string {
  const header = base64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))

  const now = Math.floor(Date.now() / 1000)
  const payload = base64url(
    JSON.stringify({
      user_id: userId,
      iss: 'modelcraft',
      iat: now,
      exp: now + expiresInSeconds,
    })
  )

  const data = `${header}.${payload}`
  const signature = base64url(
    createHmac('sha256', JWT_SECRET).update(data).digest()
  )

  return `${data}.${signature}`
}
