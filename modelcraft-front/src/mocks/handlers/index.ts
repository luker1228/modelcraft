import { graphql, HttpResponse } from 'msw'
import {
  createMockMyUserProfilePayload,
  createMockUpdateMyProfilePayload,
} from '../data/org/profile-factory'
import { endUserAuthHandlers } from './end-user/auth-handlers'

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
  if (!input) {
    return true
  }

  return [input.nickname, input.avatarUrl, input.bio].every((value) => {
    if (typeof value === 'string') {
      return value.trim().length === 0
    }

    return value == null
  })
}

const profileHandlers = [
  graphql.query('MyUserProfile', ({ request }) => {
    const scenario = resolveScenario(request)
    const payload = createMockMyUserProfilePayload({
      type: scenario === 'profileNotFound' ? 'profileNotFound' : 'success',
    })

    return HttpResponse.json({
      data: {
        myUserProfile: payload,
      },
    })
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

    return HttpResponse.json({
      data: {
        updateMyProfile: payload,
      },
    })
  }),
]

// 汇总所有 MSW handlers
// org/ 和 project/ 下的 generated.ts 由 codegen 自动生成，禁止手动编辑
export const handlers = [
  ...profileHandlers,
  ...endUserAuthHandlers,
  // TODO(profile-contract-ready): contract 同步并重新执行 npm run codegen 后，
  // 接入 org/generated.ts 与 project/generated.ts 的自动生成 handlers
]
