import { NextRequest, NextResponse } from 'next/server'
import type { EndUserBffError } from '@/types/end-user-auth'

export async function POST(req: NextRequest) {
  try {
    const body = (await req.json().catch(() => ({}))) as {
      orgName?: string
      projectSlug?: string
    }

    const orgName = body.orgName?.trim() ?? ''
    const projectSlug = body.projectSlug?.trim() ?? ''

    if (!orgName || !projectSlug) {
      const errorRes: EndUserBffError = {
        error: { code: 'PARAM_INVALID', message: 'orgName and projectSlug are required' },
      }
      return NextResponse.json(errorRes, { status: 400 })
    }

    // Deprecated: keep endpoint for backward compatibility only.
    // Frontend should call GraphQL mutation initPrivateDB with project client instead.
    return NextResponse.json({
      success: false,
      error: {
        code: 'PARAM_INVALID',
        message: 'Please use GraphQL mutation initPrivateDB via project client',
      },
    }, { status: 400 })
  } catch {
    const errorRes: EndUserBffError = {
      error: { code: 'PARAM_INVALID', message: 'Invalid request body' },
    }
    return NextResponse.json(errorRes, { status: 400 })
  }
}
