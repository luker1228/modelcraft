import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { randomUUID } from 'crypto'
import { ModelCraftWorld } from '../support/world'
import { InitOrgResponse, RestResult } from '../support/rest-client'
import { signJWT } from '../support/jwt'
import { uniqueName } from '../fixtures/factory'

const randomPhone = (): string => {
  const suffix = Math.floor(10000000 + Math.random() * 90000000).toString()
  return `199${suffix}`
}

const GET_ROLES = `
  query GetRoles {
    roles {
      id
      name
      isSystem
    }
  }
`

const CREATE_API_KEY = `
  mutation CreateApiKey($input: CreateApiKeyInput!) {
    createApiKey(input: $input) {
      result {
        id
        name
        roleIDs
      }
      error {
        __typename
        ... on ApiKeyLimitExceeded { message }
        ... on InvalidInput { message }
      }
    }
  }
`

const UPDATE_API_KEY = `
  mutation UpdateApiKey($id: ID!, $input: UpdateApiKeyInput!) {
    updateApiKey(id: $id, input: $input) {
      apiKey {
        id
        roleIDs
      }
      error {
        __typename
        ... on ResourceNotFound { message resourceType }
        ... on InvalidInput { message }
      }
    }
  }
`

const REVOKE_API_KEY = `
  mutation RevokeApiKey($id: ID!) {
    revokeApiKey(id: $id) {
      apiKey { id }
      error {
        __typename
        ... on ResourceNotFound { message resourceType }
      }
    }
  }
`

Given('我已登录并持有 access token', async function (this: ModelCraftWorld) {
  const phone = randomPhone()
  const password = `bdd${randomUUID().replace(/-/g, '').slice(0, 10)}`

  const regResult = await this.restClient.register(phone, password)
  if (!regResult.data) {
    throw new Error(`前置条件：注册用户失败 — ${JSON.stringify(regResult.error)}`)
  }

  this.currentUserId = regResult.data.userId
  this.currentOrgName = regResult.data.orgName
  this.orgClient.setOrgName(regResult.data.orgName)
  this.token = signJWT(regResult.data.userId, 3600)
  this.projectClient.setAuth(this.token)
  this.orgClient.setAuth(this.token)
})

Given('当前用户没有任何组织 memberships', async function (this: ModelCraftWorld) {
  if (!this.token) {
    throw new Error('当前无 access token')
  }

  const membershipsResult = await this.restClient.getUserMemberships(this.token)
  if (!membershipsResult.data) {
    throw new Error(`查询 memberships 失败 — ${JSON.stringify(membershipsResult.error)}`)
  }

  this.lastMembershipsCount = membershipsResult.data.memberships.length
  expect(this.lastMembershipsCount).toBe(0)
})

When('我初始化组织 displayName 为 {string}', async function (this: ModelCraftWorld, displayName: string) {
  if (!this.token) {
    throw new Error('当前无 access token')
  }

  this.lastRestResult = await this.restClient.initOrganization(this.token, displayName)

  const result = this.lastRestResult as RestResult<InitOrgResponse>
  if (result.data) {
    this.initDisplayName = displayName
    this.initOrgName = result.data.orgName
    this.initAlreadyExists = result.data.alreadyExists
  }
})

When('我首次初始化组织 displayName 为 {string}', async function (this: ModelCraftWorld, displayName: string) {
  if (!this.token) {
    throw new Error('当前无 access token')
  }

  this.lastRestResult = await this.restClient.initOrganization(this.token, displayName)

  const result = this.lastRestResult as RestResult<InitOrgResponse>
  if (!result.data) {
    throw new Error(`首次初始化失败 — ${JSON.stringify(result.error)}`)
  }

  this.initDisplayName = displayName
  this.firstInitOrgName = result.data.orgName
  this.initOrgName = result.data.orgName
  this.initAlreadyExists = result.data.alreadyExists
})

When('我再次使用相同 displayName 初始化组织', async function (this: ModelCraftWorld) {
  if (!this.token) {
    throw new Error('当前无 access token')
  }
  if (!this.initDisplayName) {
    throw new Error('未记录首次初始化的 displayName')
  }

  this.lastRestResult = await this.restClient.initOrganization(this.token, this.initDisplayName)

  const result = this.lastRestResult as RestResult<InitOrgResponse>
  if (!result.data) {
    throw new Error(`二次初始化失败 — ${JSON.stringify(result.error)}`)
  }

  this.secondInitOrgName = result.data.orgName
  this.initOrgName = result.data.orgName
  this.initAlreadyExists = result.data.alreadyExists
})

Then('组织初始化应该成功', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<InitOrgResponse>
  expect(result).not.toBeNull()
  expect(result.status).toBe(200)
  expect(result.data).toBeDefined()
  expect(result.data!.success).toBe(true)
  expect(result.data!.orgName).toBeTruthy()
  expect(result.data!.displayName).toBeTruthy()
})

Then('初始化结果 should have alreadyExists false', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<InitOrgResponse>
  expect(result.data).toBeDefined()
  expect(result.data!.alreadyExists).toBe(false)
})

Then('第二次初始化结果 should have alreadyExists true', function (this: ModelCraftWorld) {
  const result = this.lastRestResult as RestResult<InitOrgResponse>
  expect(result.data).toBeDefined()
  expect(result.data!.alreadyExists).toBe(true)
})

Then('两次初始化返回的 orgName 应相同', function (this: ModelCraftWorld) {
  expect(this.firstInitOrgName).toBeTruthy()
  expect(this.secondInitOrgName).toBeTruthy()
  expect(this.firstInitOrgName).toBe(this.secondInitOrgName)
})

Then('当前用户 memberships 数量应为 {int}', async function (this: ModelCraftWorld, expectedCount: number) {
  if (!this.token) {
    throw new Error('当前无 access token')
  }

  const membershipsResult = await this.restClient.getUserMemberships(this.token)
  if (!membershipsResult.data) {
    throw new Error(`查询 memberships 失败 — ${JSON.stringify(membershipsResult.error)}`)
  }

  this.lastMembershipsCount = membershipsResult.data.memberships.length
  expect(this.lastMembershipsCount).toBe(expectedCount)
})

Given('存在一个可绑定的组织角色', async function (this: ModelCraftWorld) {
  const res = await this.orgClient.query<{
    roles: Array<{ id: string; name: string; isSystem: boolean }>
  }>(GET_ROLES)

  expect(Array.isArray(res.roles)).toBe(true)
  expect(res.roles.length).toBeGreaterThan(0)
  this.selectedRoleID = res.roles[0].id
})

When('我创建一个绑定该角色的 API Key', async function (this: ModelCraftWorld) {
  if (!this.selectedRoleID) {
    throw new Error('未选中角色 ID，请先执行前置步骤')
  }

  const name = uniqueName('bdd-api-key')
  const res = await this.orgClient.mutate<{
    createApiKey: {
      result: { id: string; name: string; roleIDs: string[] } | null
      error: unknown
    }
  }>(CREATE_API_KEY, {
    input: {
      name,
      roleIDs: [this.selectedRoleID],
    },
  })

  this.lastResponse = { createApiKey: res.createApiKey }
  this.currentAPIKeyID = res.createApiKey.result?.id ?? null
  this.currentAPIKeyRoleIDs = res.createApiKey.result?.roleIDs ?? []
})

Then('API Key 应创建成功且角色包含该角色', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    createApiKey: {
      result: { id: string; roleIDs: string[] } | null
      error: unknown
    }
  }).createApiKey

  expect(payload.error).toBeNull()
  expect(payload.result).not.toBeNull()
  expect(this.currentAPIKeyID).toBeTruthy()
  expect(this.selectedRoleID).toBeTruthy()
  expect(payload.result?.roleIDs).toContain(this.selectedRoleID)
})

When('我将该 API Key 的角色更新为空列表', async function (this: ModelCraftWorld) {
  if (!this.currentAPIKeyID) {
    throw new Error('未记录 API Key ID，请先创建 API Key')
  }

  const res = await this.orgClient.mutate<{
    updateApiKey: {
      apiKey: { id: string; roleIDs: string[] } | null
      error: unknown
    }
  }>(UPDATE_API_KEY, {
    id: this.currentAPIKeyID,
    input: {
      roleIDs: [],
    },
  })

  this.lastResponse = { updateApiKey: res.updateApiKey }
  this.currentAPIKeyRoleIDs = res.updateApiKey.apiKey?.roleIDs ?? []
})

Then('API Key 角色应更新为空', async function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    updateApiKey: {
      apiKey: { id: string; roleIDs: string[] } | null
      error: unknown
    }
  }).updateApiKey

  expect(payload.error).toBeNull()
  expect(payload.apiKey).not.toBeNull()
  expect(this.currentAPIKeyRoleIDs).toEqual([])

  // 回收场景创建的 key，避免污染环境
  if (this.currentAPIKeyID) {
    await this.orgClient.mutate(REVOKE_API_KEY, { id: this.currentAPIKeyID })
  }
})
