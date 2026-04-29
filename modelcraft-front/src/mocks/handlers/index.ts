/**
 * MSW handlers 汇总入口
 *
 * 通过 NEXT_PUBLIC_MOCK_PAGES 环境变量控制哪些页面启用 mock。
 * 只有对应页面的 handlers 才会被注册，其余请求走真实 API。
 *
 * 示例：
 *   NEXT_PUBLIC_MOCK_PAGES=model-editor,enum-list
 *
 * 页面 key 与 buildHandlers() 中的 key 一一对应。
 *
 * ── 注意 ─────────────────────────────────────────────────────────────────────
 * org/ 和 project/ 下的 generated.ts 由 codegen 自动生成，禁止手动编辑。
 */

import { graphql, HttpResponse } from 'msw'
import { isMockPage } from '../page-mock-config'
import {
  createMockMyUserProfilePayload,
  createMockUpdateMyProfilePayload,
} from '../data/org/profile-factory'
import { endUserAuthHandlers } from './end-user/auth-handlers'
import { modelHandlers } from './model/handlers'
import { enumHandlers } from './enum/handlers'
import { rbacHandlers } from './project/rbac-handlers'

type ProfileScenarioType = 'success' | 'profileNotFound' | 'invalidInput'

type UpdateMyProfileInput = {
  nickname?: string | null
  avatarUrl?: string | null
  bio?: string | null
}

interface UpdateMyProfileRequestBody {
  variables?: {
    input?: UpdateMyProfileInput
  }
}

const PROFILE_SCENARIO_HEADER = 'x-mock-profile-scenario'

function resolveScenario(request: Request): ProfileScenarioType {
  const scenario = request.headers.get(PROFILE_SCENARIO_HEADER)
  if (scenario === 'profileNotFound' || scenario === 'invalidInput') {
    return scenario
  }
  return 'success'
}

function isProfileInputEmpty(input?: UpdateMyProfileInput): boolean {
  if (!input) return true
  return [input.nickname, input.avatarUrl, input.bio].every((value) => {
    if (typeof value === 'string') return value.trim().length === 0
    return value == null
  })
}

const profileHandlers = [
  graphql.query('MyUserProfile', ({ request }) => {
    const scenario = resolveScenario(request)
    const payload = createMockMyUserProfilePayload({
      type: scenario === 'profileNotFound' ? 'profileNotFound' : 'success',
    })
    return HttpResponse.json({ data: { myUserProfile: payload } })
  }),
  graphql.mutation('UpdateMyProfile', async ({ request }) => {
    const scenario = resolveScenario(request)
    const rawBody: unknown = await request.json()
    const body: UpdateMyProfileRequestBody | null =
      rawBody && typeof rawBody === 'object' ? (rawBody as UpdateMyProfileRequestBody) : null
    const input = body?.variables?.input

    const payload = createMockUpdateMyProfilePayload({
      type:
        scenario === 'profileNotFound'
          ? 'profileNotFound'
          : scenario === 'invalidInput' || isProfileInputEmpty(input)
            ? 'invalidInput'
            : 'success',
    })
    return HttpResponse.json({ data: { updateMyProfile: payload } })
  }),
]

/**
 * 根据 NEXT_PUBLIC_MOCK_PAGES 动态组合 handlers。
 *
 * 始终激活的 handlers（非页面级控制）：
 *   - profileHandlers  — 用户 profile 操作
 *   - endUserAuthHandlers — 终端用户认证
 *
 * 页面级控制的 handlers（按页面 key 启用）：
 *   - 'model-editor'  → modelHandlers
 *   - 'enum-list'     → enumHandlers
 *   - 'enum-detail'   → enumHandlers
 */
function buildHandlers() {
  const active = [
    ...profileHandlers,
    ...endUserAuthHandlers,
  ]

  if (isMockPage('model-editor')) {
    active.push(...modelHandlers)
  }

  if (isMockPage('enum-list') || isMockPage('enum-detail')) {
    active.push(...enumHandlers)
  }

  if (isMockPage('rbac-bundles')) {
    active.push(...rbacHandlers)
  }

  // TODO(profile-contract-ready): contract 同步并重新执行 npm run codegen 后，
  // 接入 org/generated.ts 与 project/generated.ts 的自动生成 handlers

  return active
}

export const handlers = buildHandlers()
