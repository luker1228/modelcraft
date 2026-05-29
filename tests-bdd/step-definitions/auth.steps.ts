// tests-bdd/step-definitions/auth.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'
import {
  RegisterResponse,
  LoginResponse,
  RefreshResponse,
  RestResult,
} from '../support/rest-client'

// ──────────────── 工具函数 ────────────────

/**
 * 为避免并行测试手机号冲突，将 feature 中的固定手机号映射为随机手机号。
 * 同一个 Scenario 内，相同原始手机号始终映射到同一个随机手机号。
 */
const phoneMap = new Map<string, Map<string, string>>()
const userNameMap = new Map<string, Map<string, string>>()
let scenarioKey = ''

function buildUserNameFromPhone(phone: string): string {
  return `u${phone.slice(-10)}`
}

function getOrCreatePhone(world: ModelCraftWorld, rawPhone: string): string {
  // 空字符串或非法格式直接传递给后端，让后端做校验
  if (!rawPhone || rawPhone.length !== 11 || !/^1[3-9]\d{9}$/.test(rawPhone)) {
    return rawPhone
  }
  if (!phoneMap.has(scenarioKey)) {
    phoneMap.set(scenarioKey, new Map())
  }
  const map = phoneMap.get(scenarioKey)!
  if (!map.has(rawPhone)) {
    // 保留前三位（运营商号段），随机生成后八位
    const prefix = rawPhone.slice(0, 3)
    const suffix = Math.floor(10000000 + Math.random() * 90000000).toString()
    map.set(rawPhone, prefix + suffix)
  }
  return map.get(rawPhone)!
}

function getOrCreateUserName(rawUserName: string): string {
  if (!rawUserName) {
    return rawUserName
  }
  if (!userNameMap.has(scenarioKey)) {
    userNameMap.set(scenarioKey, new Map())
  }
  const map = userNameMap.get(scenarioKey)!
  if (!map.has(rawUserName)) {
    const suffix = Math.floor(100000 + Math.random() * 900000).toString()
    map.set(rawUserName, `${rawUserName}_${suffix}`)
  }
  return map.get(rawUserName)!
}

function resolveUserName(rawUserName: string): string {
  return userNameMap.get(scenarioKey)?.get(rawUserName) ?? rawUserName
}

// 每个 Scenario 重置映射
import { Before } from '@cucumber/cucumber'
Before(function () {
  scenarioKey = `${Date.now()}_${Math.random()}`
  phoneMap.set(scenarioKey, new Map())
  userNameMap.set(scenarioKey, new Map())
})

// ──────────────── Given ────────────────

/**
 * 以管理员身份登录（复用现有 token）。
 * 优先使用 .env.test 中的 TEST_ACCESS_TOKEN。
 */
Given('我以管理员身份登录', function (this: ModelCraftWorld) {
  if (!this.token) {
    throw new Error(
      '未找到 TEST_ACCESS_TOKEN。请在 tests-bdd/.env.test 中设置:\n' +
      'TEST_ACCESS_TOKEN=<your-token>\n' +
      '获取方式：在 modelcraft-backend/ 目录运行 just test-user-setup'
    )
  }
})

/**
 * 前置条件：注册一个用户。
 */
Given(
  '已注册手机号 {string} 密码 {string}',
  async function (this: ModelCraftWorld, rawPhone: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    const result = await this.restClient.register(phone, password)
    if (!result.data) {
      throw new Error(`前置条件：注册用户 ${phone} 失败 — ${JSON.stringify(result.error)}`)
    }
    this.registeredPhone = phone
    this.registeredUserName = buildUserNameFromPhone(phone)
    this.registeredPassword = password
    this.currentUserId = result.data.userId
    this.currentOrgName = result.data.orgName
    this.orgClient.setOrgName(result.data.orgName)
  }
)

Given(
  '已注册手机号 {string} 用户名 {string} 密码 {string}',
  async function (this: ModelCraftWorld, rawPhone: string, userName: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    const resolvedUserName = getOrCreateUserName(userName)
    const result = await this.restClient.register(phone, password, resolvedUserName)
    if (!result.data) {
      throw new Error(`前置条件：注册用户 ${phone}/${resolvedUserName} 失败 — ${JSON.stringify(result.error)}`)
    }
    this.registeredPhone = phone
    this.registeredUserName = resolvedUserName
    this.registeredPassword = password
    this.currentUserId = result.data.userId
    this.currentOrgName = result.data.orgName
    this.orgClient.setOrgName(result.data.orgName)
  }
)

/**
 * 前置条件：注册并登录一个用户（获取 refreshToken）。
 */
Given(
  '已注册并登录手机号 {string} 密码 {string}',
  async function (this: ModelCraftWorld, rawPhone: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    // 注册
    const regResult = await this.restClient.register(phone, password)
    if (!regResult.data) {
      throw new Error(`前置条件：注册用户 ${phone} 失败 — ${JSON.stringify(regResult.error)}`)
    }
    // 登录
    const loginResult = await this.restClient.login(phone, password, 'PHONE')
    if (!loginResult.data) {
      throw new Error(`前置条件：登录用户 ${phone} 失败 — ${JSON.stringify(loginResult.error)}`)
    }
    this.registeredPhone = phone
    this.registeredUserName = buildUserNameFromPhone(phone)
    this.registeredPassword = password
    this.currentUserId = loginResult.data.userId
    this.currentOrgName = loginResult.data.orgName
    this.orgClient.setOrgName(loginResult.data.orgName)
    this.currentRefreshToken = loginResult.data.refreshToken
  }
)

// ──────────────── When: 注册 ────────────────

When(
  '我用手机号 {string} 和密码 {string} 注册',
  async function (this: ModelCraftWorld, rawPhone: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    this.lastRestResult = await this.restClient.register(phone, password)
    if (this.lastRestResult.data) {
      this.registeredPhone = phone
      this.registeredUserName = buildUserNameFromPhone(phone)
      this.registeredPassword = password
      this.currentOrgName = (this.lastRestResult.data as RegisterResponse).orgName
      this.orgClient.setOrgName((this.lastRestResult.data as RegisterResponse).orgName)
    }
  }
)

When(
  '我用手机号 {string} 和用户名 {string} 和密码 {string} 注册',
  async function (this: ModelCraftWorld, rawPhone: string, userName: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    const resolvedUserName = resolveUserName(userName)
    this.lastRestResult = await this.restClient.register(phone, password, resolvedUserName)
    if (this.lastRestResult.data) {
      this.registeredPhone = phone
      this.registeredUserName = resolvedUserName
      this.registeredPassword = password
      this.currentOrgName = (this.lastRestResult.data as RegisterResponse).orgName
      this.orgClient.setOrgName((this.lastRestResult.data as RegisterResponse).orgName)
    }
  }
)

// ──────────────── When: 登录 ────────────────

When(
  '我用手机号 {string} 和密码 {string} 登录',
  async function (this: ModelCraftWorld, rawPhone: string, password: string) {
    const phone = getOrCreatePhone(this, rawPhone)
    this.lastRestResult = await this.restClient.login(phone, password, 'PHONE')
    if ((this.lastRestResult as RestResult<LoginResponse>).data?.refreshToken) {
      this.currentRefreshToken = (this.lastRestResult as RestResult<LoginResponse>).data!.refreshToken
      this.currentOrgName = (this.lastRestResult as RestResult<LoginResponse>).data!.orgName
      this.orgClient.setOrgName((this.lastRestResult as RestResult<LoginResponse>).data!.orgName)
    }
  }
)

When(
  '我用用户名 {string} 和密码 {string} 登录',
  async function (this: ModelCraftWorld, userName: string, password: string) {
    const resolvedUserName = resolveUserName(userName)
    this.lastRestResult = await this.restClient.login(resolvedUserName, password, 'USERNAME')
    if ((this.lastRestResult as RestResult<LoginResponse>).data?.refreshToken) {
      this.currentRefreshToken = (this.lastRestResult as RestResult<LoginResponse>).data!.refreshToken
      this.currentOrgName = (this.lastRestResult as RestResult<LoginResponse>).data!.orgName
      this.orgClient.setOrgName((this.lastRestResult as RestResult<LoginResponse>).data!.orgName)
    }
  }
)

// ──────────────── When: Token 刷新 ────────────────

When('我使用 refreshToken 刷新令牌', async function (this: ModelCraftWorld) {
  if (!this.currentRefreshToken) {
    throw new Error('当前无 refreshToken，请先登录')
  }
  this.lastRestResult = await this.restClient.refresh(this.currentRefreshToken)
})

When('我使用无效的 refreshToken 刷新令牌', async function (this: ModelCraftWorld) {
  this.lastRestResult = await this.restClient.refresh('invalid_token_does_not_exist')
})

When('我使用已登出的 refreshToken 刷新令牌', async function (this: ModelCraftWorld) {
  // currentRefreshToken 在登出步骤中保存了已登出的 token
  if (!this.currentRefreshToken) {
    throw new Error('当前无 refreshToken')
  }
  this.lastRestResult = await this.restClient.refresh(this.currentRefreshToken)
})

// ──────────────── When: 登出 ────────────────

When('我使用 refreshToken 登出', async function (this: ModelCraftWorld) {
  if (!this.currentRefreshToken) {
    throw new Error('当前无 refreshToken，请先登录')
  }
  this.lastRestResult = await this.restClient.logout(this.currentRefreshToken)
  // 不清除 currentRefreshToken，后续步骤可能需要验证它已失效
})

// ──────────────── Then: 注册断言 ────────────────

Then('注册应该成功', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RegisterResponse>
  expect(result).not.toBeNull()
  expect(result.status).toBe(201)
  expect(result.data).toBeDefined()
})

// ──────────────── Then: 登录断言 ────────────────

Then('登录应该成功', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<LoginResponse>
  expect(result).not.toBeNull()
  expect(result.status).toBe(200)
  expect(result.data).toBeDefined()
})

Then('登录 JWT 中 is_admin 应为 true', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<LoginResponse>
  expect(result.data?.accessToken).toBeDefined()
  const token = result.data!.accessToken
  // JWT payload 是 base64url 编码的第二段
  const payloadB64 = token.split('.')[1]
  const payload = JSON.parse(Buffer.from(payloadB64, 'base64url').toString('utf-8')) as Record<string, unknown>
  expect(payload.is_admin).toBe(true)
})

// ──────────────── Then: 刷新断言 ────────────────

Then('刷新应该成功', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RefreshResponse>
  expect(result).not.toBeNull()
  expect(result.status).toBe(200)
  expect(result.data).toBeDefined()
})

Then('新 refreshToken 应与旧 refreshToken 不同', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<RefreshResponse>
  expect(result.data).toBeDefined()
  expect(result.data!.refreshToken).not.toBe(this.currentRefreshToken)
})

// ──────────────── Then: 登出断言 ────────────────

Then('登出应该成功', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  expect(this.lastRestResult!.status).toBe(204)
})

// ──────────────── Then: 通用 REST 断言 ────────────────

Then('应该返回 HTTP 状态码 {int}', function (this: ModelCraftWorld, expectedStatus: number) {
  expect(this.lastRestResult).not.toBeNull()
  expect(this.lastRestResult!.status).toBe(expectedStatus)
})

Then('应该返回错误码 {string}', function (this: ModelCraftWorld, expectedCode: string) {
  expect(this.lastRestResult).not.toBeNull()
  expect(this.lastRestResult!.error).toBeDefined()
  expect(this.lastRestResult!.error!.error.code).toBe(expectedCode)
})

Then('响应中应包含 userId', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  const data = this.lastRestResult!.data as Record<string, unknown>
  expect(data).toBeDefined()
  expect(data.userId).toBeDefined()
  expect(typeof data.userId).toBe('string')
  expect((data.userId as string).length).toBeGreaterThan(0)
})

Then('响应中应包含 refreshToken', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  const data = this.lastRestResult!.data as Record<string, unknown>
  expect(data).toBeDefined()
  expect(data.refreshToken).toBeDefined()
  expect(typeof data.refreshToken).toBe('string')
  expect((data.refreshToken as string).length).toBeGreaterThan(0)
})

Then('响应中应包含 userName', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  const data = this.lastRestResult!.data as Record<string, unknown>
  expect(data).toBeDefined()
  expect(data.userName).toBeDefined()
  expect(typeof data.userName).toBe('string')
  expect((data.userName as string).length).toBeGreaterThan(0)
})

Then('响应中应包含 orgName', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  const data = this.lastRestResult!.data as Record<string, unknown>
  expect(data).toBeDefined()
  expect(data.orgName).toBeDefined()
  expect(typeof data.orgName).toBe('string')
  expect((data.orgName as string).length).toBeGreaterThan(0)
})

Then('响应中应包含 expiresAt', function (this: ModelCraftWorld) {
  expect(this.lastRestResult).not.toBeNull()
  const data = this.lastRestResult!.data as Record<string, unknown>
  expect(data).toBeDefined()
  expect(data.expiresAt).toBeDefined()
  expect(typeof data.expiresAt).toBe('string')
  // 验证是合法的日期格式
  const date = new Date(data.expiresAt as string)
  expect(date.getTime()).not.toBeNaN()
})
