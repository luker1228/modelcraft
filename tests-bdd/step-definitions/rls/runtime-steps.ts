// RLS Runtime Step Definitions
// 对应 Feature: Runtime RLS 注入和行级过滤

import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'
import { EndUserRestClient } from '../../support/end-user-rest-client'

// 用于存储 EndUser 相关的运行时状态
const endUserRuntimeState = new Map<string, {
  client: EndUserRestClient
  token: string
  userId: string
  lastQueryResult: unknown
  lastMutationResult: unknown
  recordIds: Map<string, string> // name -> id
}>()

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

// 辅助函数：创建 Runtime GraphQL 客户端
function createRuntimeClient(orgName: string, projectSlug: string, token: string) {
  return {
    async query<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
      const res = await fetch(
        `${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/runtime`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({ query: document, variables }),
        }
      )
      return (await res.json()).data as T
    },
    async mutate<T>(document: string, variables?: Record<string, unknown>): Promise<T> {
      const res = await fetch(
        `${API_BASE_URL}/graphql/org/${orgName}/project/${projectSlug}/runtime`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({ query: document, variables }),
        }
      )
      const body = await res.json()
      return { data: body.data, errors: body.errors, status: res.status } as T
    },
  }
}

// Given 步骤

Given('终端用户 {string} 的用户 ID 为 {string}', function (
  this: ModelCraftWorld, username: string, userId: string
) {
  // 存储用户 ID 映射供后续使用
  if (!this.modelMap['endUserIds']) {
    this.modelMap['endUserIds'] = {}
  }
  (this.modelMap['endUserIds'] as Record<string, string>)[username] = userId
})

Given('终端用户 {string} 创建了一条 {string} 记录，name 为 {string}', async function (
  this: ModelCraftWorld, username: string, modelName: string, recordName: string
) {
  // 确保用户已登录
  let state = endUserRuntimeState.get(username)
  if (!state) {
    // 创建用户并登录
    const client = new EndUserRestClient(this.endUserOrgName || this.orgName, this.endUserProjectSlug || this.projectSlug)

    // 先创建用户
    const createRes = await client.createEndUser(
      { username, password: 'Pass1234' },
      this.internalToken || process.env.INTERNAL_TOKEN || 'test-internal-token'
    )
    if (createRes.error) {
      throw new Error(`创建终端用户 ${username} 失败: ${createRes.error.message}`)
    }

    // 登录获取 token
    const loginRes = await client.loginEndUser({ username, password: 'Pass1234' })
    if (loginRes.error || !loginRes.data) {
      throw new Error(`终端用户 ${username} 登录失败`)
    }

    state = {
      client,
      token: loginRes.data.accessToken,
      userId: loginRes.data.userId,
      lastQueryResult: null,
      lastMutationResult: null,
      recordIds: new Map(),
    }
    endUserRuntimeState.set(username, state)
  }

  // 创建记录
  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const CREATE_RECORD = `
    mutation Create${modelName}($data: ${modelName}CreateInput!) {
      createOne${modelName}(data: $data) {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.mutate<{ data?: { [`createOne${string}`]: { id: string } }; errors?: unknown[] }>(
    CREATE_RECORD,
    { data: { name: recordName } }
  )

  const recordId = result.data?.[`createOne${modelName}` as keyof typeof result.data]?.id
  if (recordId) {
    state.recordIds.set(recordName, recordId)
  }
})

Given('{string} 记录的实际 ID 为 {string}', function (
  this: ModelCraftWorld, recordName: string, idAlias: string
) {
  // 存储记录 ID 映射
  for (const state of endUserRuntimeState.values()) {
    const id = state.recordIds.get(recordName)
    if (id) {
      this.modelMap[idAlias] = id
      break
    }
  }
})

// When 步骤

When('终端用户 {string} 查询 {string}', async function (
  this: ModelCraftWorld, username: string, modelName: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const QUERY_RECORDS = `
    query FindMany${modelName} {
      findMany${modelName} {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.query<{ [`findMany${string}`]: Array<{ id: string; name: string; owner: string }> }>(
    QUERY_RECORDS
  )

  state.lastQueryResult = result
  this.lastResponse = { runtimeQuery: result }
})

When('终端用户 {string} 查询 {string}，where 条件为 owner = {string}', async function (
  this: ModelCraftWorld, username: string, modelName: string, ownerIdRef: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  // 解析 owner ID 引用
  let ownerId = ownerIdRef
  if (ownerIdRef.startsWith('"') && ownerIdRef.endsWith('"')) {
    ownerId = ownerIdRef.slice(1, -1)
  }
  // 检查是否是变量引用
  const userIds = this.modelMap['endUserIds'] as Record<string, string> | undefined
  if (userIds && userIds[ownerIdRef]) {
    ownerId = userIds[ownerIdRef]
  }

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const QUERY_WITH_WHERE = `
    query FindMany${modelName}($where: ${modelName}WhereInput) {
      findMany${modelName}(where: $where) {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.query<{ [`findMany${string}`]: Array<{ id: string; name: string; owner: string }> }>(
    QUERY_WITH_WHERE,
    { where: { owner: { _eq: ownerId } } }
  )

  state.lastQueryResult = result
  this.lastResponse = { runtimeQuery: result }
})

When('终端用户 {string} 创建一条 {string} 记录，name 为 {string}', async function (
  this: ModelCraftWorld, username: string, modelName: string, recordName: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const CREATE_RECORD = `
    mutation Create${modelName}($data: ${modelName}CreateInput!) {
      createOne${modelName}(data: $data) {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`createOne${string}`]: { id: string; owner: string } }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(CREATE_RECORD, { data: { name: recordName } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }

  if (result.data) {
    const recordId = result.data[`createOne${modelName}` as keyof typeof result.data]?.id
    if (recordId) {
      state.recordIds.set(recordName, recordId)
    }
  }
})

When('终端用户 {string} 创建一条 {string} 记录，name 为 {string}，owner 为 {string}', async function (
  this: ModelCraftWorld, username: string, modelName: string, recordName: string, ownerIdRef: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  // 解析 owner ID 引用
  let ownerId = ownerIdRef
  if (ownerIdRef.startsWith('"') && ownerIdRef.endsWith('"')) {
    ownerId = ownerIdRef.slice(1, -1)
  }

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const CREATE_RECORD = `
    mutation Create${modelName}($data: ${modelName}CreateInput!) {
      createOne${modelName}(data: $data) {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`createOne${string}`]: { id: string; owner: string } }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(CREATE_RECORD, { data: { name: recordName, owner: ownerId } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }
})

When('终端用户 {string} 尝试更新 ID 为 {string} 的 {string} 记录', async function (
  this: ModelCraftWorld, username: string, idRef: string, modelName: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  // 解析 ID 引用
  const recordId = this.modelMap[idRef] || idRef

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const UPDATE_RECORD = `
    mutation Update${modelName}($where: ${modelName}WhereUniqueInput!, $data: ${modelName}UpdateInput!) {
      updateOne${modelName}(where: $where, data: $data) {
        id
        name
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`updateOne${string}`]: { id: string } | null }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(UPDATE_RECORD, { where: { id: recordId }, data: { name: 'UpdatedName' } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }
})

When('终端用户 {string} 尝试删除 ID 为 {string} 的 {string} 记录', async function (
  this: ModelCraftWorld, username: string, idRef: string, modelName: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  // 解析 ID 引用
  const recordId = this.modelMap[idRef] || idRef

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const DELETE_RECORD = `
    mutation Delete${modelName}($where: ${modelName}WhereUniqueInput!) {
      deleteOne${modelName}(where: $where) {
        id
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`deleteOne${string}`]: { id: string } | null }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(DELETE_RECORD, { where: { id: recordId } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }
})

When('终端用户 {string} 尝试创建一条 {string} 记录', async function (
  this: ModelCraftWorld, username: string, modelName: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  const CREATE_RECORD = `
    mutation Create${modelName}($data: ${modelName}CreateInput!) {
      createOne${modelName}(data: $data) {
        id
        name
        owner
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`createOne${string}`]: { id: string; owner: string } }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(CREATE_RECORD, { data: { name: `TestRecord_${Date.now()}` } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }
})

When('终端用户 {string} 尝试将 {string} 的 owner 改为 {string}', async function (
  this: ModelCraftWorld, username: string, idRef: string, newOwnerRef: string
) {
  const state = endUserRuntimeState.get(username)
  if (!state) {
    throw new Error(`终端用户 ${username} 未登录`)
  }

  // 解析 ID 引用
  const recordId = this.modelMap[idRef] || idRef
  const newOwnerId = this.modelMap[newOwnerRef] || newOwnerRef

  const runtimeClient = createRuntimeClient(
    this.endUserOrgName || this.orgName,
    this.endUserProjectSlug || this.projectSlug,
    state.token
  )

  // 需要动态获取模型名，这里简化处理
  const modelName = 'Orders' // 默认使用 Orders

  const UPDATE_OWNER = `
    mutation UpdateOwner($where: ${modelName}WhereUniqueInput!, $data: ${modelName}UpdateInput!) {
      updateOne${modelName}(where: $where, data: $data) {
        id
        owner
      }
    }
  `

  const result = await runtimeClient.mutate<{
    data?: { [`updateOne${string}`]: { id: string; owner: string } | null }
    errors?: Array<{ message: string; extensions?: { code: string } }>
  }>(UPDATE_OWNER, { where: { id: recordId }, data: { owner: newOwnerId } })

  state.lastMutationResult = result
  this.lastResponse = { runtimeMutation: result }
})

// Then 步骤

Then('查询结果只包含 {string}', function (this: ModelCraftWorld, recordName: string) {
  const result = (this.lastResponse as {
    runtimeQuery: Array<{ name: string }>
  }).runtimeQuery

  expect(result.length).toBe(1)
  expect(result[0].name).toBe(recordName)
})

Then('查询结果不包含 {string}', function (this: ModelCraftWorld, recordName: string) {
  const result = (this.lastResponse as {
    runtimeQuery: Array<{ name: string }>
  }).runtimeQuery

  const found = result.some(r => r.name === recordName)
  expect(found).toBe(false)
})

Then('查询结果包含所有用户的记录', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeQuery: Array<{ name: string; owner: string }>
  }).runtimeQuery

  expect(result.length).toBeGreaterThan(1)
})

Then('查询结果为空', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeQuery: Array<unknown>
  }).runtimeQuery

  expect(result.length).toBe(0)
})

Then('记录创建成功', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeMutation: { errors?: unknown[] }
  }).runtimeMutation

  // 没有 GraphQL errors 或者 errors 为空
  expect(result.errors?.length ?? 0).toBe(0)
})

Then('该记录的 owner 字段值为 {string} 的用户 ID', function (
  this: ModelCraftWorld, username: string
) {
  const result = (this.lastResponse as {
    runtimeMutation: {
      data?: { [key: string]: { owner: string } }
    }
  }).runtimeMutation

  const state = endUserRuntimeState.get(username)
  const expectedUserId = state?.userId

  // 获取创建的记录
  const mutationResult = Object.values(result.data || {})[0] as { owner: string } | undefined
  expect(mutationResult?.owner).toBe(expectedUserId)
})

Then('该记录的 owner 字段值为 {string} 的用户 ID（被强制覆盖）', function (
  this: ModelCraftWorld, username: string
) {
  // 同 Then 该记录的 owner 字段值为 {string} 的用户 ID
  const result = (this.lastResponse as {
    runtimeMutation: {
      data?: { [key: string]: { owner: string } }
    }
  }).runtimeMutation

  const state = endUserRuntimeState.get(username)
  const expectedUserId = state?.userId

  const mutationResult = Object.values(result.data || {})[0] as { owner: string } | undefined
  expect(mutationResult?.owner).toBe(expectedUserId)
})

Then('更新操作返回 0 行受影响（静默失败，无错误）', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeMutation: {
      data?: { [key: string]: unknown | null }
      errors?: unknown[]
    }
  }).runtimeMutation

  // USING 失败时返回 null，不报错
  const mutationResult = Object.values(result.data || {})[0]
  expect(mutationResult).toBeNull()
  expect(result.errors?.length ?? 0).toBe(0)
})

Then('删除操作返回 0 行受影响（静默失败，无错误）', function (this: ModelCraftWorld) {
  const result = (this.lastResponse as {
    runtimeMutation: {
      data?: { [key: string]: unknown | null }
      errors?: unknown[]
    }
  }).runtimeMutation

  // USING 失败时返回 null，不报错
  const mutationResult = Object.values(result.data || {})[0]
  expect(mutationResult).toBeNull()
  expect(result.errors?.length ?? 0).toBe(0)
})

Then('操作失败并返回错误类型 {string}', function (
  this: ModelCraftWorld, errorType: string
) {
  const result = (this.lastResponse as {
    runtimeMutation: {
      errors?: Array<{ extensions?: { code: string }; message: string }>
    }
  }).runtimeMutation

  expect(result.errors?.length ?? 0).toBeGreaterThan(0)
  const error = result.errors?.[0]
  expect(error?.extensions?.code || error?.message).toContain(errorType)
})
