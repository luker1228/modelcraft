// src/app/api/bff/org/[orgName]/end-user/auth/register/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/register
// 转发到 gateway POST /api/end-user/auth/register

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(req: NextRequest) {
  return proxyEndUserAuth(req, 'register', 'POST')
}
