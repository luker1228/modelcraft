// src/app/api/bff/org/[orgName]/end-user/auth/select-project/route.ts
// BFF: POST /api/bff/org/[orgName]/end-user/auth/select-project
// 转发到 gateway POST /api/end-user/auth/select-project

import { NextRequest } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'

export async function POST(req: NextRequest) {
  return proxyEndUserAuth(req, 'select-project', 'POST')
}
