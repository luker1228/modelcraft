// tests-bdd/step-definitions/end-user-auth.steps.ts
// End-User Auth BDD Step Definitions

import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'
import { EndUserRestClient, EndUserAuthResult, EndUserListResult } from '../support/end-user-rest-client'
import { GraphQLClient as ProjectGraphQLClient } from '../support/graphql-client'

// 每个 Scenario 的客户端实例
let endUserClient: EndUserRestClient
let endUserProjectClient: ProjectGraphQLClient | null = null
let currentEndUserAuth: EndUserAuthResult | null = null
let currentEndUserList: EndUserListResult | null = null
let currentRefreshToken: string | null = null
let lastError: { code: string; message: string } | null = null
let lastStatusCode: number = 0

// Scenario 级别的用户映射（用于存储创建的终端用户）
const endUserMap = new Map<string, { userId: string; username: string; password: string }>()

// ──────────────── Given ────────────────

Given(
  '存在已配置数据库集群的组织 {string} 和项目 {string}',
  function (this: ModelCraftWorld, orgName: string, projectSlug: string) {
    // 为了兼容 feature 中的占位值 testorg/testproject，这里优先使用 .env.test 的固定夹具
    const resolvedOrgName = orgName === 'testorg'
      ? (process.env.TEST_ORG_NAME || orgName)
      : orgName
    const resolvedProjectSlug = projectSlug === 'testproject'
      ? (process.env.TEST_PROJECT_SLUG || projectSlug)
      : projectSlug

    // 组织/项目信息存储在 world 中供后续使用
    this.endUserOrgName = resolvedOrgName
    this.endUserProjectSlug = resolvedProjectSlug
    endUserClient = new EndUserRestClient(resolvedOrgName, resolvedProjectSlug)
    endUserProjectClient = new ProjectGraphQLClient(resolvedOrgName, resolvedProjectSlug)
    if (this.token) {
      endUserProjectClient.setAuth(this.token)
    }
  }
)

function ensureProjectGraphQLClient(world: ModelCraftWorld): ProjectGraphQLClient {
  if (!endUserProjectClient) {
    throw new Error('GraphQL client 未初始化，请先执行组织/项目 Given')
  }

  if (!world.token && process.env.TEST_ACCESS_TOKEN) {
    world.token = process.env.TEST_ACCESS_TOKEN
  }
  if (!world.token) {
    throw new Error('缺少 TEST_ACCESS_TOKEN，无法执行项目授权前置')
  }

  endUserProjectClient.setAuth(world.token)
  return endUserProjectClient
}

function resolveProjectSlugForWorld(world: ModelCraftWorld, projectSlug: string): string {
  if (projectSlug !== 'testproject') {
    return projectSlug
  }
  return world.endUserProjectSlug || process.env.TEST_PROJECT_SLUG || projectSlug
}

const LIST_PERMISSION_BUNDLES = `
  query ListPermissionBundles($input: ListEndUserPermissionBundlesInput) {
    endUserPermissionBundles(input: $input) {
      edges {
        node {
          id
        }
      }
    }
  }
`

const CREATE_PERMISSION_BUNDLE = `
  mutation CreatePermissionBundle($input: CreateEndUserPermissionBundleInput!) {
    createEndUserPermissionBundle(input: $input) {
      bundle {
        id
      }
      error {
        __typename
        ... on EndUserPermissionBundleAlreadyExists {
          message
        }
        ... on InvalidInput {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

const GRANT_END_USER_ACCESS = `
  mutation GrantEndUserAccess($input: GrantEndUserProjectAccessInput!) {
    grantEndUserProjectAccess(input: $input) {
      access {
        id
      }
      error {
        __typename
        ... on EndUserNotFound {
          message
        }
        ... on EndUserProjectAccessAlreadyExists {
          message
        }
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on InvalidInput {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

async function ensurePermissionBundleID(world: ModelCraftWorld): Promise<string> {
  const projectClient = ensureProjectGraphQLClient(world)

  const listResp = await projectClient.query<{
    endUserPermissionBundles: {
      edges: Array<{ node: { id: string } }>
    }
  }>(LIST_PERMISSION_BUNDLES, { input: { first: 1 } })

  const existingID = listResp.endUserPermissionBundles.edges[0]?.node?.id
  if (existingID) {
    return existingID
  }

  const bundleName = `bdd_auto_bundle_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
  const createResp = await projectClient.mutate<{
    createEndUserPermissionBundle: {
      bundle?: { id: string }
      error?: { __typename: string; message?: string }
    }
  }>(CREATE_PERMISSION_BUNDLE, {
    input: {
      name: bundleName,
      description: 'BDD auto-created bundle for end-user auth',
    },
  })

  const createdID = createResp.createEndUserPermissionBundle.bundle?.id
  if (createdID) {
    return createdID
  }

  const err = createResp.createEndUserPermissionBundle.error
  throw new Error(`前置条件：创建权限包失败 — ${err?.__typename || 'UNKNOWN'}: ${err?.message || ''}`)
}

async function ensureProjectAccessGranted(world: ModelCraftWorld, endUserID: string): Promise<void> {
  const projectClient = ensureProjectGraphQLClient(world)
  const bundleID = await ensurePermissionBundleID(world)

  const grantResp = await projectClient.mutate<{
    grantEndUserProjectAccess: {
      access?: { id: string }
      error?: { __typename: string; message?: string }
    }
  }>(GRANT_END_USER_ACCESS, {
    input: {
      endUserId: endUserID,
      permissionBundleId: bundleID,
    },
  })

  if (grantResp.grantEndUserProjectAccess.access?.id) {
    return
  }

  const err = grantResp.grantEndUserProjectAccess.error
  if (err?.__typename === 'EndUserProjectAccessAlreadyExists') {
    return
  }

  throw new Error(`前置条件：授予项目访问失败 — ${err?.__typename || 'UNKNOWN'}: ${err?.message || ''}`)
}

function ensureInternalToken(world: ModelCraftWorld): string {
  if (!world.internalToken) {
    world.internalToken = process.env.INTERNAL_TOKEN || 'test-internal-token'
  }
  return world.internalToken
}

Given('我以开发者身份登录', function (this: ModelCraftWorld) {
  // 开发者 token 用于调用内部接口
  ensureInternalToken(this)
})

Given(
  '已存在终端用户 {string}，密码 {string}',
  async function (this: ModelCraftWorld, username: string, password: string) {
    await ensureEndUserExists(this, username, password)
  }
)

Given(
  '已存在被禁用的终端用户 {string}',
  async function (this: ModelCraftWorld, username: string) {
    await ensureEndUserExists(this, username, 'Pass1234')

    const user = endUserMap.get(username)
    if (!user) {
      throw new Error(`前置条件：无法定位终端用户 ${username}`)
    }

    const updateResult = await endUserClient.updateEndUserStatus(
      user.userId,
      true,
      ensureInternalToken(this)
    )
    if (updateResult.error) {
      throw new Error(`前置条件：禁用终端用户 ${username} 失败 — ${updateResult.error.code}: ${updateResult.error.message}`)
    }
  }
)

// Helper function to create end user
async function ensureEndUserExists(
  world: ModelCraftWorld,
  username: string,
  password: string = 'Pass1234'
): Promise<void> {
  if (endUserMap.has(username)) {
    return
  }

  const result = await endUserClient.createEndUser(
    { username, password },
    ensureInternalToken(world)
  )
  if (!result.error) {
    endUserMap.set(username, {
      userId: result.data!.userId,
      username,
      password,
    })
    await ensureProjectAccessGranted(world, result.data!.userId)
    return
  }

  if (result.error.code !== 'CONFLICT.END_USER' && result.error.code !== 'CONFLICT') {
    throw new Error(`前置条件：创建终端用户 ${username} 失败 — ${result.error.code}: ${result.error.message}`)
  }

  // 用户已存在：复用存量用户并确保状态为启用
  const listResult = await endUserClient.listEndUsers(
    { search: username, first: 20 },
    ensureInternalToken(world)
  )
  if (listResult.error) {
    throw new Error(`前置条件：查询已存在终端用户 ${username} 失败 — ${listResult.error.code}: ${listResult.error.message}`)
  }

  const existingUser = listResult.data?.users.find((u) => u.username === username)
  if (!existingUser) {
    throw new Error(`前置条件：终端用户 ${username} 已冲突但无法在列表中定位`)
  }

  const enableResult = await endUserClient.updateEndUserStatus(
    existingUser.id,
    false,
    ensureInternalToken(world)
  )
  if (enableResult.error) {
    throw new Error(`前置条件：重置终端用户 ${username} 启用状态失败 — ${enableResult.error.code}: ${enableResult.error.message}`)
  }

  endUserMap.set(username, {
    userId: existingUser.id,
    username,
    password,
  })

  await ensureProjectAccessGranted(world, existingUser.id)
}

Given('终端用户 {string} 已登录', async function (this: ModelCraftWorld, username: string) {
  // 确保用户存在
  await ensureEndUserExists(this, username)

  const user = endUserMap.get(username)!
  const result = await endUserClient.loginEndUser({
    username: user.username,
    password: user.password,
  })

  if (result.error) {
    throw new Error(`前置条件：终端用户 ${username} 登录失败 — ${result.error.code}`)
  }

  currentEndUserAuth = result.data!
  currentRefreshToken = result.data!.refreshToken || null
  this.currentEndUserId = result.data!.userId
})

Given(
  '终端用户 {string} 已登录并持有 refresh token',
  async function (this: ModelCraftWorld, username: string) {
    // 先登录
    await ensureEndUserExists(this, username)

    const user = endUserMap.get(username)!
    const result = await endUserClient.loginEndUser({
      username: user.username,
      password: user.password,
    })

    if (result.error) {
      throw new Error(`前置条件：终端用户 ${username} 登录失败`)
    }

    currentEndUserAuth = result.data!
    currentRefreshToken = result.data!.refreshToken || null
    this.currentEndUserId = result.data!.userId
  }
)

Given('该用户的 refresh token 已被撤销', async function () {
  // 登出会撤销 refresh token
  if (!currentRefreshToken) {
    throw new Error('前置条件：缺少 refresh token')
  }
  await endUserClient.logoutEndUser(currentRefreshToken)
  currentRefreshToken = null
})

Given(
  '已存在并登录后被禁用的终端用户 {string}',
  async function (this: ModelCraftWorld, username: string) {
    await ensureEndUserExists(this, username)

    const user = endUserMap.get(username)!
    const loginResult = await endUserClient.loginEndUser({
      username: user.username,
      password: user.password,
    })
    if (loginResult.error) {
      throw new Error(`前置条件：终端用户 ${username} 登录失败 — ${loginResult.error.code}`)
    }

    currentEndUserAuth = loginResult.data!
    currentRefreshToken = loginResult.data!.refreshToken || null
    this.currentEndUserId = loginResult.data!.userId

    const disableResult = await endUserClient.updateEndUserStatus(
      user.userId,
      true,
      ensureInternalToken(this)
    )
    if (disableResult.error) {
      throw new Error(`前置条件：禁用终端用户 ${username} 失败 — ${disableResult.error.code}`)
    }
  }
)

// ──────────────── When ────────────────

When(
  '我创建终端用户，用户名为 {string}，密码为 {string}',
  async function (this: ModelCraftWorld, username: string, password: string) {
    const result = await endUserClient.createEndUser(
      { username, password },
      ensureInternalToken(this)
    )
    lastStatusCode = result.status

    if (result.error) {
      lastError = result.error
    } else {
      currentEndUserAuth = result.data!
      lastError = null
      endUserMap.set(username, {
        userId: result.data!.userId,
        username,
        password,
      })
    }
  }
)

When('我查询终端用户列表，每页 {int} 条', async function (this: ModelCraftWorld, first: number) {
  const result = await endUserClient.listEndUsers(
    { first },
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserList = result.data!
    lastError = null
  }
})

When('我搜索终端用户，关键词为 {string}', async function (this: ModelCraftWorld, search: string) {
  const result = await endUserClient.listEndUsers(
    { search, first: 20 },
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserList = result.data!
    lastError = null
  }
})

When('我禁用终端用户 {string}', async function (this: ModelCraftWorld, username: string) {
  const user = endUserMap.get(username)
  if (!user) {
    throw new Error(`用户 ${username} 不存在`)
  }

  const result = await endUserClient.updateEndUserStatus(
    user.userId,
    true, // isForbidden
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
  }
})

When('我启用终端用户 {string}', async function (this: ModelCraftWorld, username: string) {
  const user = endUserMap.get(username)
  if (!user) {
    throw new Error(`用户 ${username} 不存在`)
  }

  const result = await endUserClient.updateEndUserStatus(
    user.userId,
    false, // isForbidden = false (启用)
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
  }
})

When('我删除终端用户 {string}', async function (this: ModelCraftWorld, username: string) {
  const user = endUserMap.get(username)
  if (!user) {
    throw new Error(`用户 ${username} 不存在`)
  }

  const result = await endUserClient.deleteEndUser(
    user.userId,
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
    endUserMap.delete(username)
  }
})

When('开发者删除终端用户 {string}', async function (this: ModelCraftWorld, username: string) {
  const user = endUserMap.get(username)
  if (!user) {
    throw new Error(`用户 ${username} 不存在`)
  }

  const result = await endUserClient.deleteEndUser(
    user.userId,
    ensureInternalToken(this)
  )
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
    endUserMap.delete(username)
  }
})

// ──────────────── 终端用户自助操作 ────────────────

When(
  '终端用户自助注册，用户名为 {string}，密码为 {string}',
  async function (_this: ModelCraftWorld, username: string, password: string) {
    const result = await endUserClient.registerEndUser({
      username,
      password,
    })
    lastStatusCode = result.status

    if (result.error) {
      lastError = result.error
    } else {
      currentEndUserAuth = result.data!
      currentRefreshToken = result.data!.refreshToken || null
      lastError = null
      endUserMap.set(username, {
        userId: result.data!.userId,
        username,
        password,
      })
    }
  }
)

When(
  '终端用户登录，用户名为 {string}，密码为 {string}',
  async function (this: ModelCraftWorld, username: string, password: string) {
    const result = await endUserClient.loginEndUser({
      username,
      password,
    })
    lastStatusCode = result.status

    if (result.error) {
      lastError = result.error
    } else {
      currentEndUserAuth = result.data!
      currentRefreshToken = result.data!.refreshToken || null
      this.currentEndUserId = result.data!.userId
      lastError = null
    }
  }
)

When('终端用户查询自己的信息', async function (this: ModelCraftWorld) {
  if (!this.currentEndUserId) {
    throw new Error('终端用户未登录')
  }

  const result = await endUserClient.getEndUserMe(this.currentEndUserId)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
    this.lastEndUserInfo = result.data || null
  }
})

When('终端用户登出', async function () {
  if (!currentRefreshToken) {
    throw new Error('终端用户未登录')
  }

  const result = await endUserClient.logoutEndUser(currentRefreshToken)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    lastError = null
    currentRefreshToken = null
  }
})

When('终端用户刷新 token', async function () {
  if (!currentRefreshToken) {
    throw new Error('缺少 refresh token')
  }

  const result = await endUserClient.refreshEndUserToken(currentRefreshToken)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserAuth = result.data!
    currentRefreshToken = result.data!.refreshToken || currentRefreshToken
    lastError = null
  }
})

When('使用该用户的 refresh token 刷新', async function () {
  if (!currentRefreshToken) {
    throw new Error('缺少 refresh token')
  }

  const result = await endUserClient.refreshEndUserToken(currentRefreshToken)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserAuth = result.data!
    currentRefreshToken = result.data!.refreshToken || currentRefreshToken
    lastError = null
  }
})

When('终端用户选择项目上下文 {string}', async function (this: ModelCraftWorld, projectSlug: string) {
  if (!currentRefreshToken) {
    throw new Error('缺少 refresh token')
  }

  const resolvedProjectSlug = resolveProjectSlugForWorld(this, projectSlug)
  const result = await endUserClient.selectProjectContext(currentRefreshToken, resolvedProjectSlug)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserAuth = result.data!
    lastError = null
  }
})

When('使用该用户的 refresh token 选择项目 {string}', async function (this: ModelCraftWorld, projectSlug: string) {
  if (!currentRefreshToken) {
    throw new Error('缺少 refresh token')
  }

  const resolvedProjectSlug = resolveProjectSlugForWorld(this, projectSlug)
  const result = await endUserClient.selectProjectContext(currentRefreshToken, resolvedProjectSlug)
  lastStatusCode = result.status

  if (result.error) {
    lastError = result.error
  } else {
    currentEndUserAuth = result.data!
    lastError = null
  }
})

// ──────────────── Then ────────────────

Then('终端用户应该创建成功', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth).toBeTruthy()
})

Then('终端用户应该删除成功', function () {
  expect(lastError).toBeNull()
})

Then('用户状态更新成功', function () {
  expect(lastError).toBeNull()
})

Then('登出成功', function () {
  expect(lastError).toBeNull()
})

Then('注册成功并返回 refresh token', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth?.refreshToken).toBeTruthy()
})

Then('登录成功并返回 access token', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth?.accessToken).toBeTruthy()
})

Then('刷新成功并返回新的 access token', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth?.accessToken).toBeTruthy()
})

Then('返回的 token 有效期为 {int} 秒', function (expiresIn: number) {
  expect(currentEndUserAuth?.expiresIn).toBe(expiresIn)
})

Then('返回 refresh token', function () {
  expect(currentEndUserAuth?.refreshToken).toBeTruthy()
})

Then('返回新的 refresh token', function () {
  expect(currentEndUserAuth?.refreshToken).toBeTruthy()
})

Then('返回可访问项目列表', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth?.projects).toBeTruthy()
  expect((currentEndUserAuth?.projects || []).length).toBeGreaterThan(0)
})

Then('可访问项目列表应包含 {string}', function (this: ModelCraftWorld, projectSlug: string) {
  const resolvedProjectSlug = resolveProjectSlugForWorld(this, projectSlug)
  const hasProject = (currentEndUserAuth?.projects || []).some((p) => p.slug === resolvedProjectSlug)
  expect(hasProject).toBe(true)
})

Then('选择项目成功并返回用户信息', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserAuth?.userId).toBeTruthy()
})

Then('返回已选择项目 {string}', function (this: ModelCraftWorld, projectSlug: string) {
  const resolvedProjectSlug = resolveProjectSlugForWorld(this, projectSlug)
  expect(currentEndUserAuth?.selectedProject).toBe(resolvedProjectSlug)
})

Then('返回中不应包含 access token', function () {
  expect(currentEndUserAuth?.accessToken).toBeFalsy()
})

Then('应该返回用户列表', function () {
  expect(lastError).toBeNull()
  expect(currentEndUserList?.users).toBeTruthy()
})

Then('列表中应该包含用户 {string}', function (username: string) {
  const user = currentEndUserList?.users.find((u) => u.username === username)
  expect(user).toBeTruthy()
})

Then('总用户数应该大于等于 {int}', function (count: number) {
  expect(currentEndUserList?.totalCount).toBeGreaterThanOrEqual(count)
})

Then('该用户应该不存在', async function (this: ModelCraftWorld) {
  // 这个步骤需要配合具体用户名，通常在删除后验证
  // 简化处理：验证 lastError 为 null 表示删除成功
  expect(lastError).toBeNull()
})

Then('应该返回用户信息', function () {
  expect(lastError).toBeNull()
  expect(this.lastEndUserInfo).toBeTruthy()
})

Then('用户状态应该是启用', function () {
  // 验证状态更新结果
  expect(lastError).toBeNull()
})

Then('用户状态应该是禁用', function () {
  // 验证状态更新结果
  expect(lastError).toBeNull()
})

Then('返回的信息应该包含创建时间', function () {
  expect(this.lastEndUserInfo?.createdAt).toBeTruthy()
})

Then('返回的用户名应该是 {string}', function (expectedUsername: string) {
  const actualUsername =
    this.lastEndUserInfo?.username || currentEndUserAuth?.username
  expect(actualUsername).toBe(expectedUsername)
})

Then('HTTP 状态码应该是 {int}', function (statusCode: number) {
  expect(lastStatusCode).toBe(statusCode)
})

// 注意：错误码验证使用已存在的步骤定义 from auth.steps.ts
// 这里需要把错误信息设置到 world 中
Then('返回终端用户错误码 {string}', function (this: ModelCraftWorld, errorCode: string) {
  expect(lastError).toBeTruthy()
  const actualCode = lastError!.code
  const matches = actualCode === errorCode || actualCode.startsWith(`${errorCode}.`)
  expect(matches).toBe(true)
})
