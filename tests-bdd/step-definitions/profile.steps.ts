import { Before, Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { randomUUID } from 'crypto'
import { ModelCraftWorld } from '../support/world'
import { RegisterResponse, RestResult } from '../support/rest-client'
import { signJWT } from '../support/jwt'

const UPDATE_MY_PROFILE = `
  mutation UpdateMyProfile($input: UpdateMyProfileInput!) {
    updateMyProfile(input: $input) {
      profile {
        id
        userId
        nickname
        avatarUrl
        bio
      }
      error {
        __typename
        ... on ResourceNotFound { message resourceType }
        ... on InvalidInput { message suggestion }
      }
    }
  }
`

const MY_USER_PROFILE = `
  query MyUserProfile {
    myUserProfile {
      user {
        id
        phone
        userName
        status
        profile {
          id
          userId
          nickname
          avatarUrl
          bio
        }
      }
      error {
        __typename
        ... on ResourceNotFound { message resourceType }
      }
    }
  }
`

const ME_QUERY = `
  query Me {
    me {
      id
      externalID
      email
      name
      permissions
    }
  }
`

interface ProfileDTO {
  id: string
  userId: string
  nickname: string
  avatarUrl: string | null
  bio: string | null
}

interface UpdateMyProfilePayload {
  profile: ProfileDTO | null
  error: { __typename: string; message?: string } | null
}

interface MyUserProfilePayload {
  user:
    | {
        id: string
        profile: ProfileDTO
      }
    | null
  error: { __typename: string; message?: string } | null
}

interface MePayload {
  id: string
  externalID: string
  email: string
  name: string
  permissions: string[]
}

function ensureOrgGraphQLReady(world: ModelCraftWorld): void {
  if (!world.token) {
    throw new Error('当前无访问令牌，请先执行登录/签发令牌步骤')
  }
  if (!world.currentOrgName) {
    throw new Error('当前无 orgName，请先完成注册或组织初始化步骤')
  }

  world.orgClient.setAuth(world.token)
  world.orgClient.setOrgName(world.currentOrgName)
}

const profilePhoneMap = new Map<string, Map<string, string>>()
const profileUserNameMap = new Map<string, Map<string, string>>()
let profileScenarioKey = ''

function randomMainlandPhone(): string {
  const suffix = Math.floor(10000000 + Math.random() * 90000000).toString()
  return `199${suffix}`
}

function getOrCreateProfilePhone(rawPhone: string): string {
  if (!rawPhone || rawPhone.length !== 11 || !/^1[3-9]\d{9}$/.test(rawPhone)) {
    return rawPhone
  }

  const phoneMap = profilePhoneMap.get(profileScenarioKey)
  if (!phoneMap) {
    throw new Error('profile 场景手机号映射未初始化')
  }

  if (!phoneMap.has(rawPhone)) {
    const prefix = rawPhone.slice(0, 3)
    const suffix = Math.floor(10000000 + Math.random() * 90000000).toString()
    phoneMap.set(rawPhone, prefix + suffix)
  }

  return phoneMap.get(rawPhone)!
}

function getOrCreateProfileUserName(rawUserName: string): string {
  const userNameMap = profileUserNameMap.get(profileScenarioKey)
  if (!userNameMap) {
    throw new Error('profile 场景用户名映射未初始化')
  }

  if (!userNameMap.has(rawUserName)) {
    const normalizedBase = rawUserName.replace(/[^a-zA-Z0-9_-]/g, '_') || 'user'
    const suffix = randomUUID().replace(/-/g, '').slice(0, 8)
    const maxBaseLength = 32 - 1 - suffix.length
    const trimmedBase = normalizedBase.slice(0, Math.max(1, maxBaseLength))
    userNameMap.set(rawUserName, `${trimmedBase}_${suffix}`)
  }

  return userNameMap.get(rawUserName)!
}

function applyRegisteredUserState(
  world: ModelCraftWorld,
  result: RegisterResponse,
  phone: string,
  userName: string,
  password: string
): void {
  world.registeredPhone = phone
  world.registeredUserName = userName
  world.registeredPassword = password
  world.currentUserId = result.userId
  world.currentOrgName = result.orgName
  world.orgClient.setOrgName(result.orgName)
}

Before(function () {
  profileScenarioKey = `profile_${randomUUID().replace(/-/g, '').slice(0, 8)}`
  profilePhoneMap.set(profileScenarioKey, new Map())
  profileUserNameMap.set(profileScenarioKey, new Map())
})

Given('我使用当前注册用户的访问令牌', function (this: ModelCraftWorld) {
  if (!this.currentUserId) {
    throw new Error('当前无注册用户 ID，请先执行注册步骤')
  }

  this.token = signJWT(this.currentUserId, 3600)
  this.projectClient.setAuth(this.token)
  this.orgClient.setAuth(this.token)

  if (this.currentOrgName) {
    this.orgClient.setOrgName(this.currentOrgName)
  }
})

Given('已注册唯一手机号 {string} 唯一用户名 {string} 密码 {string}', async function (
  this: ModelCraftWorld,
  rawPhone: string,
  rawUserName: string,
  password: string
) {
  const phone = getOrCreateProfilePhone(rawPhone)
  const userName = getOrCreateProfileUserName(rawUserName)

  const result = await this.restClient.register(phone, password, userName)
  if (!result.data) {
    throw new Error(
      `前置条件：注册用户 ${phone}/${userName} 失败（基准 ${rawPhone}/${rawUserName}）— ${JSON.stringify(result.error)}`
    )
  }

  applyRegisteredUserState(this, result.data, phone, userName, password)
})

When('我用唯一手机号 {string} 和唯一用户名 {string} 和密码 {string} 注册', async function (
  this: ModelCraftWorld,
  rawPhone: string,
  rawUserName: string,
  password: string
) {
  const phone = getOrCreateProfilePhone(rawPhone)
  const userName = getOrCreateProfileUserName(rawUserName)

  const result = await this.restClient.register(phone, password, userName)
  this.lastRestResult = result

  if (result.data) {
    applyRegisteredUserState(this, result.data, phone, userName, password)
  }
})

Given('存在一个仅有 user 无 profile 的用户', async function (this: ModelCraftWorld) {
  const username = `bdd_profile_missing_${randomUUID().replace(/-/g, '').slice(0, 8)}`
  const phone = randomMainlandPhone()
  const password = 'password123'

  const registerResult = await this.restClient.register(phone, password, username)
  if (!registerResult.data) {
    throw new Error(`创建基准用户失败 — ${JSON.stringify(registerResult.error)}`)
  }

  // 运行时已无 auth_provider webhook；这里通过“合法组织 + 不存在用户ID”构造 not-found 场景
  this.currentOrgName = registerResult.data.orgName
  this.orgClient.setOrgName(registerResult.data.orgName)
  this.currentUserId = randomUUID()
})

Given('我使用该用户的访问令牌并初始化组织', async function (this: ModelCraftWorld) {
  if (!this.currentUserId) {
    throw new Error('当前无用户 ID，请先创建测试用户')
  }

  this.token = signJWT(this.currentUserId, 3600)
  this.projectClient.setAuth(this.token)
  this.orgClient.setAuth(this.token)

  if (this.currentOrgName) {
    this.orgClient.setOrgName(this.currentOrgName)
    return
  }

  const displayName = `BDD Profile Missing ${randomUUID().replace(/-/g, '').slice(0, 8)}`
  const initResult = await this.restClient.initOrganization(this.token, displayName)

  if (!initResult.data?.orgName) {
    throw new Error(`初始化组织失败 — ${JSON.stringify(initResult.error)}`)
  }

  this.currentOrgName = initResult.data.orgName
  this.orgClient.setOrgName(initResult.data.orgName)
})

Given(
  '我调用 updateMyProfile 设置完整资料，昵称 {string} 头像 {string} 简介 {string}',
  async function (this: ModelCraftWorld, nickname: string, avatarUrl: string, bio: string) {
    ensureOrgGraphQLReady(this)

    const res = await this.orgClient.mutate<{ updateMyProfile: UpdateMyProfilePayload }>(
      UPDATE_MY_PROFILE,
      {
        input: {
          nickname,
          avatarUrl,
          bio,
        },
      }
    )

    this.lastResponse = { updateMyProfile: res.updateMyProfile }

    if (res.updateMyProfile.error || !res.updateMyProfile.profile) {
      throw new Error(`初始化 profile 失败 — ${JSON.stringify(res.updateMyProfile.error)}`)
    }
  }
)

When('我调用 updateMyProfile 仅更新昵称为 {string}', async function (this: ModelCraftWorld, nickname: string) {
  ensureOrgGraphQLReady(this)

  const res = await this.orgClient.mutate<{ updateMyProfile: UpdateMyProfilePayload }>(UPDATE_MY_PROFILE, {
    input: {
      nickname,
    },
  })

  this.lastResponse = { updateMyProfile: res.updateMyProfile }
})

When('我查询 myUserProfile', async function (this: ModelCraftWorld) {
  ensureOrgGraphQLReady(this)

  const res = await this.orgClient.query<{ myUserProfile: MyUserProfilePayload }>(MY_USER_PROFILE)
  this.lastResponse = { myUserProfile: res.myUserProfile }
})

When('我查询 me', async function (this: ModelCraftWorld) {
  ensureOrgGraphQLReady(this)

  const res = await this.orgClient.query<{ me: MePayload }>(ME_QUERY)
  this.lastResponse = { me: res.me }
})

Then('响应中应包含 profile 快照', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RegisterResponse>
  expect(result).not.toBeNull()
  expect(result.data).toBeDefined()
  expect(result.data!.profile).toBeDefined()
})

Then('响应中的 profile.userId 应与 userId 一致', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RegisterResponse>
  expect(result.data).toBeDefined()
  expect(result.data!.profile).toBeDefined()
  expect(result.data!.profile!.userId).toBe(result.data!.userId)
})

Then('响应中的 profile.nickname 应匹配默认规则', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RegisterResponse>
  expect(result.data).toBeDefined()
  expect(result.data!.profile).toBeDefined()
  expect(result.data!.profile!.nickname).toMatch(/^user_[A-Z0-9]{6}$/)
})

Then('updateMyProfile 应该成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { updateMyProfile: UpdateMyProfilePayload }).updateMyProfile
  expect(payload.error).toBeNull()
  expect(payload.profile).not.toBeNull()
})

Then('返回的 profile.nickname 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { updateMyProfile: UpdateMyProfilePayload }).updateMyProfile
  expect(payload.profile).not.toBeNull()
  expect(payload.profile!.nickname).toBe(expected)
})

Then('返回的 profile.avatarUrl 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { updateMyProfile: UpdateMyProfilePayload }).updateMyProfile
  expect(payload.profile).not.toBeNull()
  expect(payload.profile!.avatarUrl).toBe(expected)
})

Then('返回的 profile.bio 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { updateMyProfile: UpdateMyProfilePayload }).updateMyProfile
  expect(payload.profile).not.toBeNull()
  expect(payload.profile!.bio).toBe(expected)
})

Then('myUserProfile 应返回用户资料', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { myUserProfile: MyUserProfilePayload }).myUserProfile
  expect(payload.error).toBeNull()
  expect(payload.user).not.toBeNull()
  expect(payload.user!.profile).not.toBeNull()
})

Then('myUserProfile 应返回错误类型 {string}', function (this: ModelCraftWorld, expectedTypename: string) {
  const payload = (this.lastResponse as { myUserProfile: MyUserProfilePayload }).myUserProfile
  expect(payload.error).not.toBeNull()
  expect(payload.error!.__typename).toBe(expectedTypename)
})

Then('myUserProfile.user.profile.nickname 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { myUserProfile: MyUserProfilePayload }).myUserProfile
  expect(payload.user).not.toBeNull()
  expect(payload.user!.profile.nickname).toBe(expected)
})

Then('myUserProfile.user.profile.avatarUrl 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { myUserProfile: MyUserProfilePayload }).myUserProfile
  expect(payload.user).not.toBeNull()
  expect(payload.user!.profile.avatarUrl).toBe(expected)
})

Then('myUserProfile.user.profile.bio 应为 {string}', function (this: ModelCraftWorld, expected: string) {
  const payload = (this.lastResponse as { myUserProfile: MyUserProfilePayload }).myUserProfile
  expect(payload.user).not.toBeNull()
  expect(payload.user!.profile.bio).toBe(expected)
})

Then('me 查询应成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(payload).toBeDefined()
  expect(payload.id).toBeTruthy()
})

Then('me.id 应等于当前用户ID', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(this.currentUserId).toBeTruthy()
  expect(payload.id).toBe(this.currentUserId)
})

Then('me 应包含 externalID 字段', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(payload.externalID).toBeDefined()
  expect(typeof payload.externalID).toBe('string')
})

Then('me 应包含 email 字段', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(payload.email).toBeDefined()
  expect(typeof payload.email).toBe('string')
})

Then('me 应包含 name 字段', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(payload.name).toBeDefined()
  expect(typeof payload.name).toBe('string')
})

Then('me.permissions 应为数组', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { me: MePayload }).me
  expect(Array.isArray(payload.permissions)).toBe(true)
})
