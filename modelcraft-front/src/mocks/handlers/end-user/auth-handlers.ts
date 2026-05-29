// src/mocks/handlers/end-user/auth-handlers.ts
// 终端用户 BFF REST Mock Handlers（MSW）
// 对称现有 profile handlers 风格，通过 x-mock-scenario header 切换场景

import { http, HttpResponse } from 'msw'
import {
  createMockLoginPayload,
  createMockRegisterPayload,
  createMockRefreshPayload,
  createMockMePayload,
  type EndUserLoginScenario,
  type EndUserRegisterScenario,
  type EndUserRefreshScenario,
  type EndUserMeScenario,
} from '../../data/end-user/end-user-factory'

// ============================================================================
// 场景解析
// ============================================================================

const SCENARIO_HEADER = 'x-mock-end-user-scenario'

function resolveLoginScenario(request: Request): EndUserLoginScenario {
  const scenario = request.headers.get(SCENARIO_HEADER)
  if (
    scenario === 'invalidCredentials' ||
    scenario === 'accountDisabled' ||
    scenario === 'clusterNotConfigured'
  ) {
    return scenario
  }
  return 'success'
}

function resolveRegisterScenario(request: Request): EndUserRegisterScenario {
  const scenario = request.headers.get(SCENARIO_HEADER)
  if (
    scenario === 'conflict' ||
    scenario === 'paramInvalid' ||
    scenario === 'clusterNotConfigured'
  ) {
    return scenario
  }
  return 'success'
}

function resolveRefreshScenario(request: Request): EndUserRefreshScenario {
  const scenario = request.headers.get(SCENARIO_HEADER)
  if (scenario === 'invalidToken' || scenario === 'accountDisabled') {
    return scenario
  }
  return 'success'
}

function resolveMeScenario(request: Request): EndUserMeScenario {
  const scenario = request.headers.get(SCENARIO_HEADER)
  if (scenario === 'unauthorized' || scenario === 'accountDisabled') {
    return scenario
  }
  return 'success'
}

// ============================================================================
// Handlers
// ============================================================================

export const endUserAuthHandlers = [
  // POST /api/bff/end-user/auth/login
  http.post('/api/bff/end-user/auth/login', async ({ request }) => {
    const scenario = resolveLoginScenario(request)
    const { status, body } = createMockLoginPayload(scenario)

    return HttpResponse.json(body, { status })
  }),

  // POST /api/bff/org/:orgName/end-user/auth/login
  http.post('/api/bff/org/:orgName/end-user/auth/login', async ({ request }) => {
    const scenario = resolveLoginScenario(request)
    const { status, body } = createMockLoginPayload(scenario)

    return HttpResponse.json(body, { status })
  }),

  // POST /api/bff/org/:orgName/end-user/auth/register
  http.post('/api/bff/org/:orgName/end-user/auth/register', async ({ request }) => {
    const scenario = resolveRegisterScenario(request)
    const { status, body } = createMockRegisterPayload(scenario)

    return HttpResponse.json(body, { status })
  }),

  // POST /api/bff/org/:orgName/end-user/auth/logout
  http.post('/api/bff/org/:orgName/end-user/auth/logout', async () => {
    // Logout always succeeds (best-effort)
    return new HttpResponse(null, { status: 204 })
  }),

  // POST /api/bff/org/:orgName/end-user/auth/refresh
  http.post('/api/bff/org/:orgName/end-user/auth/refresh', async ({ request }) => {
    const scenario = resolveRefreshScenario(request)
    const { status, body } = createMockRefreshPayload(scenario)

    return HttpResponse.json(body, { status })
  }),

  // GET /api/bff/org/:orgName/end-user/auth/me
  http.get('/api/bff/org/:orgName/end-user/auth/me', async ({ request }) => {
    const scenario = resolveMeScenario(request)
    const { status, body } = createMockMePayload(scenario)

    return HttpResponse.json(body, { status })
  }),
]
