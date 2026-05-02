// src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/refresh
// 转发到 gateway POST /api/end-user/auth/refresh

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(req: NextRequest) {
  return proxyEndUserAuth(req, 'refresh', 'POST')
}
