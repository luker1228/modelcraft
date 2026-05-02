// src/app/api/bff/org/[orgName]/end-user/auth/logout/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/logout
// 转发到 gateway POST /api/end-user/auth/logout

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(req: NextRequest) {
  return proxyEndUserAuth(req, 'logout', 'POST')
}
