// src/app/api/bff/org/[orgName]/end-user/auth/me/route.ts
// BFF: GET /api/bff/org/[orgName]/end-user/auth/me
// 转发到后端 GET /internal/v1/end-user/auth/me

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params
  return proxyEndUserAuth(req, orgName, 'me', 'GET')
}
