// src/app/api/bff/end-user/auth/me/route.ts
// 获取当前终端用户信息 BFF Route Handler
// 验证 end-user access token → 解析 payload → 调用 Go /me → 返回用户信息

import { NextRequest, NextResponse } from 'next/server'
import {
  callGoEndUserMe,
  EndUserInvalidCredentialsError,
  EndUserAccountDisabledError,
} from '@/bff/end-user/end-user-go-client'
import { verifyEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import type { EndUserMeResponse, EndUserBffError } from '@/types/end-user-auth'

export async function GET(req: NextRequest) {
  // 从 Authorization header 获取 access token
  const authHeader = req.headers.get('Authorization')

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    const errorRes: EndUserBffError = {
      error: { code: 'UNAUTHORIZED', message: 'Missing or invalid Authorization header' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }

  const token = authHeader.slice(7) // 去掉 "Bearer " 前缀

  try {
    // 1. 验证 JWT（检查 issuer = modelcraft-end-user）
    const { userId, orgName, projectSlug } = await verifyEndUserAccessToken(token)

    // 2. 调用 Go Backend /me（传递 X-End-User-Id 等 Header）
    const meResult = await callGoEndUserMe({
      orgName,
      projectSlug,
      userId,
    })

    const response: EndUserMeResponse = {
      id: meResult.id,
      username: meResult.username,
      createdAt: meResult.createdAt,
    }

    return NextResponse.json(response)
  } catch (err) {
    console.error('[BFF] end-user me error:', err)

    // JWT 验证失败
    if (err instanceof Error && err.name === 'JWTExpired') {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: 'Token expired' },
      }
      return NextResponse.json(errorRes, { status: 401 })
    }

    // 用户不存在
    if (err instanceof EndUserInvalidCredentialsError) {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: '用户不存在' },
      }
      return NextResponse.json(errorRes, { status: 401 })
    }

    // 账号被禁用
    if (err instanceof EndUserAccountDisabledError) {
      const errorRes: EndUserBffError = {
        error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用' },
      }
      return NextResponse.json(errorRes, { status: 403 })
    }

    // 其他验证错误
    const errorRes: EndUserBffError = {
      error: { code: 'UNAUTHORIZED', message: 'Invalid token' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }
}
