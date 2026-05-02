// src/app/api/bff/org/[orgName]/end-user/auth/login/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/login
// 转发到 gateway POST /api/end-user/auth/login

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(req: NextRequest) {
  return proxyEndUserAuth(req, 'login', 'POST')
}
