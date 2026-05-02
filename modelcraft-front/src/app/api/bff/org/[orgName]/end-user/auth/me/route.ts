// src/app/api/bff/org/[orgName]/end-user/auth/me/route.ts
// BFF: GET /api/bff/org/[orgName]/end-user/auth/me
// 转发到 gateway GET /api/end-user/auth/me（后端从 Bearer JWT 解析身份）

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function GET(req: NextRequest) {
  return proxyEndUserAuth(req, 'me', 'GET')
}
