// src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/refresh
// 转发到后端 POST /internal/v1/end-user/auth/refresh

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params
  return proxyEndUserAuth(req, orgName, 'refresh', 'POST')
}
