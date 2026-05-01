// src/app/api/bff/org/[orgName]/end-user/auth/login/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/login
// 转发到后端 POST /internal/v1/end-user/auth/login

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ orgName: string }> }
) {
  const { orgName } = await params
  return proxyEndUserAuth(req, orgName, 'login', 'POST')
}
