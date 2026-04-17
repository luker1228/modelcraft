// src/mocks/data/end-user/end-user-factory.ts
// 终端用户 Mock 数据工厂（对称 org/profile-factory.ts）

import { faker } from '@faker-js/faker'
import type {
  EndUserAuthResponse,
  EndUserMeResponse,
  EndUserBffError,
  EndUserErrorCode,
} from '@/types/end-user-auth'

// ============================================================================
// Mock 场景类型
// ============================================================================

export type EndUserLoginScenario =
  | 'success'
  | 'invalidCredentials'
  | 'accountDisabled'
  | 'clusterNotConfigured'

export type EndUserRegisterScenario =
  | 'success'
  | 'conflict'
  | 'paramInvalid'
  | 'clusterNotConfigured'

export type EndUserRefreshScenario =
  | 'success'
  | 'invalidToken'
  | 'accountDisabled'

export type EndUserMeScenario =
  | 'success'
  | 'unauthorized'
  | 'accountDisabled'

// ============================================================================
// Mock 用户数据
// ============================================================================

export interface MockEndUser {
  id: string
  username: string
  isForbidden: boolean
  createdAt: string
  updatedAt: string
}

/**
 * 创建 Mock 终端用户
 */
export function createMockEndUser(override: Partial<MockEndUser> = {}): MockEndUser {
  const createdAt = override.createdAt ?? faker.date.recent().toISOString()

  return {
    id: faker.string.uuid(),
    username: faker.internet.username().toLowerCase(),
    isForbidden: false,
    createdAt,
    updatedAt: override.updatedAt ?? createdAt,
    ...override,
  }
}

// ============================================================================
// Mock 响应 Payload 工厂
// ============================================================================

/**
 * 创建 Mock 登录响应
 */
export function createMockLoginPayload(
  scenario: EndUserLoginScenario
): { status: number; body: EndUserAuthResponse | EndUserBffError } {
  switch (scenario) {
    case 'invalidCredentials':
      return {
        status: 401,
        body: createMockErrorPayload('INVALID_CREDENTIALS', '用户名或密码错误'),
      }
    case 'accountDisabled':
      return {
        status: 403,
        body: createMockErrorPayload('ACCOUNT_DISABLED', '该账号已被禁用'),
      }
    case 'clusterNotConfigured':
      return {
        status: 503,
        body: createMockErrorPayload('CLUSTER_NOT_CONFIGURED', '服务暂时不可用'),
      }
    case 'success':
    default:
      return {
        status: 200,
        body: {
          accessToken: createMockAccessToken(),
          expiresIn: 3600,
        },
      }
  }
}

/**
 * 创建 Mock 注册响应
 */
export function createMockRegisterPayload(
  scenario: EndUserRegisterScenario
): { status: number; body: EndUserAuthResponse | EndUserBffError } {
  switch (scenario) {
    case 'conflict':
      return {
        status: 409,
        body: createMockErrorPayload('CONFLICT', '该用户名已被使用'),
      }
    case 'paramInvalid':
      return {
        status: 400,
        body: createMockErrorPayload('PARAM_INVALID', '密码强度不足'),
      }
    case 'clusterNotConfigured':
      return {
        status: 503,
        body: createMockErrorPayload('CLUSTER_NOT_CONFIGURED', '服务暂时不可用'),
      }
    case 'success':
    default:
      return {
        status: 200,
        body: {
          accessToken: createMockAccessToken(),
          expiresIn: 3600,
        },
      }
  }
}

/**
 * 创建 Mock Refresh 响应
 */
export function createMockRefreshPayload(
  scenario: EndUserRefreshScenario
): { status: number; body: EndUserAuthResponse | EndUserBffError } {
  switch (scenario) {
    case 'invalidToken':
      return {
        status: 401,
        body: createMockErrorPayload('INVALID_REFRESH_TOKEN', 'Token 无效或已过期'),
      }
    case 'accountDisabled':
      return {
        status: 403,
        body: createMockErrorPayload('ACCOUNT_DISABLED', '该账号已被禁用'),
      }
    case 'success':
    default:
      return {
        status: 200,
        body: {
          accessToken: createMockAccessToken(),
          expiresIn: 3600,
        },
      }
  }
}

/**
 * 创建 Mock /me 响应
 */
export function createMockMePayload(
  scenario: EndUserMeScenario
): { status: number; body: EndUserMeResponse | EndUserBffError } {
  switch (scenario) {
    case 'unauthorized':
      return {
        status: 401,
        body: createMockErrorPayload('UNAUTHORIZED', '未授权，请重新登录'),
      }
    case 'accountDisabled':
      return {
        status: 403,
        body: createMockErrorPayload('ACCOUNT_DISABLED', '该账号已被禁用'),
      }
    case 'success':
    default:
      return {
        status: 200,
        body: {
          id: 'mock-user-001',
          username: 'alice',
          createdAt: '2026-04-10T08:00:00Z',
        },
      }
  }
}

// ============================================================================
// 工具函数
// ============================================================================

/**
 * 创建 Mock 错误响应
 */
export function createMockErrorPayload(
  code: EndUserErrorCode,
  message: string
): EndUserBffError {
  return {
    error: {
      code,
      message,
    },
  }
}

/**
 * 创建 Mock Access Token（非真实 JWT，仅用于测试）
 */
export function createMockAccessToken(): string {
  // 生成一个看起来像 JWT 的 mock token（base64 编码的 JSON）
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const payload = btoa(
    JSON.stringify({
      sub: 'mock-user-001',
      org_name: 'test',
      project_slug: 'demo',
      role: 'end_user',
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000),
    })
  )
  const signature = 'mock-signature'

  return `${header}.${payload}.${signature}`
}

/**
 * 预设的 Mock 用户列表（用于 GraphQL listEndUsers mock）
 */
export function createMockEndUserList(): MockEndUser[] {
  return [
    createMockEndUser({
      id: 'user-001',
      username: 'alice',
      isForbidden: false,
      createdAt: '2026-04-10T08:00:00Z',
      updatedAt: '2026-04-10T08:00:00Z',
    }),
    createMockEndUser({
      id: 'user-002',
      username: 'bob',
      isForbidden: true,
      createdAt: '2026-04-12T10:30:00Z',
      updatedAt: '2026-04-14T09:00:00Z',
    }),
  ]
}
