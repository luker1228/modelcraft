import { NextRequest, NextResponse } from 'next/server'
import {
  callGoEndUserModelCatalog,
  EndUserAccountDisabledError,
  EndUserInvalidCredentialsError,
  EndUserParamInvalidError,
  EndUserPrivateDBNotInitializedError,
  EndUserUnauthorizedError,
  EndUserUpstreamError,
} from '@/bff/end-user/end-user-go-client'
import { verifyEndUserAccessToken } from '@/bff/end-user/end-user-jwt-utils'
import type { EndUserBffError, EndUserErrorCode } from '@/types/end-user-auth'

export async function GET(req: NextRequest) {
  const authHeader = req.headers.get('Authorization')
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    const errorRes: EndUserBffError = {
      error: { code: 'UNAUTHORIZED', message: 'Missing or invalid Authorization header' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }

  const token = authHeader.slice(7)

  try {
    const { userId, orgName, projectSlug } = await verifyEndUserAccessToken(token)

    const databaseName = req.nextUrl.searchParams.get('databaseName') ?? ''
    if (!databaseName) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: 'databaseName is required' },
      }
      return NextResponse.json(errorRes, { status: 400 })
    }

    const search = req.nextUrl.searchParams.get('search') ?? ''
    const pageRaw = req.nextUrl.searchParams.get('page')
    const pageSizeRaw = req.nextUrl.searchParams.get('pageSize')

    const page = pageRaw ? Number(pageRaw) : 1
    const pageSize = pageSizeRaw ? Number(pageSizeRaw) : 200

    if (!Number.isInteger(page) || page <= 0) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: 'page must be a positive integer' },
      }
      return NextResponse.json(errorRes, { status: 400 })
    }
    if (!Number.isInteger(pageSize) || pageSize <= 0) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: 'pageSize must be a positive integer' },
      }
      return NextResponse.json(errorRes, { status: 400 })
    }

    const result = await callGoEndUserModelCatalog({
      orgName,
      projectSlug,
      userId,
      databaseName,
      search,
      page,
      pageSize,
    })

    return NextResponse.json(result)
  } catch (err) {
    console.error('[BFF] end-user model catalog error:', err)

    if (err instanceof EndUserParamInvalidError) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: err.message },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 400 })
    }

    if (err instanceof EndUserInvalidCredentialsError) {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: '用户不存在或会话已失效' },
      }
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserAccountDisabledError) {
      const errorRes: EndUserBffError = {
        error: { code: 'ACCOUNT_DISABLED', message: '该账号已被禁用' },
      }
      return NextResponse.json(errorRes, { status: 403 })
    }

    if (err instanceof EndUserUnauthorizedError) {
      const errorRes: EndUserBffError = {
        error: { code: 'UNAUTHORIZED', message: err.message || '未授权访问内部数据服务' },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 401 })
    }

    if (err instanceof EndUserPrivateDBNotInitializedError) {
      const errorRes: EndUserBffError = {
        error: { code: 'PRIVATE_DB_NOT_INITIALIZED', message: err.message || '私有库未初始化' },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: 409 })
    }

    if (err instanceof EndUserUpstreamError) {
      const errorRes: EndUserBffError = {
        error: {
          code: err.code === 'UNAUTHORIZED' ? 'UNAUTHORIZED' : (err.code as EndUserErrorCode) ?? 'PARAM_INVALID',
          message: err.message || '上游服务错误',
        },
      }
      if (err.requestId) errorRes.requestId = err.requestId
      return NextResponse.json(errorRes, { status: err.status || 500 })
    }

    const errorRes: EndUserBffError = {
      error: { code: 'UNAUTHORIZED', message: 'Invalid token' },
    }
    return NextResponse.json(errorRes, { status: 401 })
  }
}
