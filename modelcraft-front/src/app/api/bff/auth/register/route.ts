// src/app/api/bff/auth/register/route.ts
import { NextRequest, NextResponse } from 'next/server'
import {
  callGoRegister,
  AuthParamInvalidError,
  UserConflictError,
} from '@/bff/auth/go-auth-client'
import type { RegisterProfileSnapshot, RegisterResponse } from '@/types/auth'

interface RegisterRequestBody {
  phone?: unknown
  userName?: unknown
  password?: unknown
}

function buildMockAvatarUrl(userName: string): string {
  return `https://api.dicebear.com/9.x/initials/svg?seed=${encodeURIComponent(userName)}`
}

function buildFallbackProfileSnapshot(
  userId: string,
  userName: string
): RegisterProfileSnapshot {
  return {
    id: `mock-profile-${userId}`,
    userId,
    nickname: userName,
    avatarUrl: buildMockAvatarUrl(userName),
  }
}

export async function POST(req: NextRequest) {
  let body: RegisterRequestBody
  try {
    body = (await req.json()) as RegisterRequestBody
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const { phone, userName, password } = body

  // 参数校验
  if (!phone || typeof phone !== 'string') {
    return NextResponse.json({ error: '请输入手机号' }, { status: 400 })
  }
  if (!userName || typeof userName !== 'string') {
    return NextResponse.json({ error: '请输入用户名' }, { status: 400 })
  }
  if (!password || typeof password !== 'string') {
    return NextResponse.json({ error: '请输入密码' }, { status: 400 })
  }

  try {
    const { userId, orgName, profile } = await callGoRegister({
      phone,
      userName,
      password,
    })

    const response: RegisterResponse = {
      userId,
      orgName,
      profile: profile ?? buildFallbackProfileSnapshot(userId, userName),
    }

    return NextResponse.json(response, { status: 201 })
  } catch (err) {
    console.error('BFF register error:', err)

    // 区分错误类型返回不同的状态码
    if (err instanceof UserConflictError) {
      // 根据消息内容判断是手机号还是用户名冲突
      const message = err.message
      let userFriendlyMessage = '注册失败'

      if (
        message.includes('phone') ||
        message.includes('手机号')
      ) {
        userFriendlyMessage = '该手机号已被注册'
      } else if (
        message.includes('userName') ||
        message.includes('用户名')
      ) {
        userFriendlyMessage = '该用户名已被占用'
      } else {
        userFriendlyMessage = message
      }

      return NextResponse.json({ error: userFriendlyMessage }, { status: 409 })
    }

    if (err instanceof AuthParamInvalidError) {
      return NextResponse.json({ error: err.message }, { status: 400 })
    }

    return NextResponse.json({ error: '注册失败，请稍后重试' }, { status: 500 })
  }
}
