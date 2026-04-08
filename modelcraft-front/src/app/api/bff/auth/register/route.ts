// src/app/api/bff/auth/register/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { callGoRegister } from '@/bff/auth/go-auth-client'

export async function POST(req: NextRequest) {
  let body: { phone?: unknown; password?: unknown }
  try {
    body = (await req.json()) as { phone?: unknown; password?: unknown }
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const { phone, password } = body
  if (!phone || typeof phone !== 'string') {
    return NextResponse.json({ error: 'Phone number required' }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: 'Password required' }, { status: 400 })
  }

  try {
    await callGoRegister({ phone, password })
    return NextResponse.json({ success: true })
  } catch (err) {
    console.error('BFF register error:', err)
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Registration failed' },
      { status: 400 },
    )
  }
}
