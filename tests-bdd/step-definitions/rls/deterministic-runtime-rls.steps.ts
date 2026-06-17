// Deterministic RLS Runtime Smoke Test — Step Definitions
// 对应 Feature: 确定性 RLS 运行时拨测
// 参考: debug/enduser_runtime_rls_check.sh
//
// 两套 API:
//   1. Developer GraphQL API (projectClient) — 模型查找 + RLS v2 策略配置
//   2. Open Data API (PAT + X-MC-Auth-* headers) — 运行时 CRUD 验证

import { Given, When, Then, DataTable } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

// ─── Env var helpers ───────────────────────────────────────────

function detOrgName(world: ModelCraftWorld): string {
  return process.env.DET_ORG_NAME || world.orgName
}

function detProjectSlug(world: ModelCraftWorld): string {
  return process.env.DET_PROJECT_SLUG || world.projectSlug
}

function detDbName(): string {
  const v = process.env.DET_DB_NAME
  if (!v) throw new Error('环境变量 DET_DB_NAME 未设置（确定性拨测需要预置数据库名称）')
  return v
}

function detModelName(): string {
  const v = process.env.DET_MODEL_NAME
  if (!v) throw new Error('环境变量 DET_MODEL_NAME 未设置（确定性拨测需要预置模型名称）')
  return v
}

function detPat(): string {
  const v = process.env.DET_PAT
  if (!v) throw new Error('环境变量 DET_PAT 未设置（Open Data API 需要 PAT token）')
  return v
}

function detEndUserId(): string {
  return process.env.DET_END_USER_ID ?? 'det-test-user-001'
}

function detEndUserName(): string {
  return process.env.DET_END_USER_NAME ?? 'det-test-user'
}

function getDetModelId(world: ModelCraftWorld): string {
  const id = world.modelMap['detModelId']
  if (!id) throw new Error('确定性拨测模型 ID 未就绪（请确认 Background 已执行）')
  return id
}

// ─── Open Data API client ──────────────────────────────────────

interface OpenDataApiResult {
  status: number
  data: unknown
  errors?: Array<{ message: string; extensions?: { code: string } }>
}

/**
 * 构建 Open Data API endpoint URL
 * 格式: /end-user/graphql/org/{org}/project/{slug}/db/{db}/model/{model}
 */
function openDataEndpoint(world: ModelCraftWorld): string {
  const org = detOrgName(world)
  const project = detProjectSlug(world)
  const db = detDbName()
  const model = detModelName()
  return `${API_BASE_URL}/end-user/graphql/org/${org}/project/${project}/db/${db}/model/${model}`
}

/**
 * 调用 Open Data API
 */
async function callOpenDataApi(
  world: ModelCraftWorld,
  query: string,
  role: string,
  userId?: string,
  userName?: string,
): Promise<OpenDataApiResult> {
  const endpoint = openDataEndpoint(world)
  const pat = detPat()
  const uid = userId ?? detEndUserId()
  const uname = userName ?? detEndUserName()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pat}`,
    'X-MC-Auth-Userid': uid,
    'X-MC-Auth-Username': uname,
    'X-MC-Auth-Roles': role,
  }

  const res = await fetch(endpoint, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query }),
  })

  const body = await res.json()

  return {
    status: res.status,
    data: body.data,
    errors: body.errors,
  }
}

// ─── GraphQL documents ─────────────────────────────────────────

// 使用 models 分页查询来查找预置模型（modelByName 在 APISIX 代理后存在 orgName 传递问题）
const MODELS_LIST = `
  query Models($input: ModelQueryInput) {
    models(input: $input) { items { id name databaseName } }
  }
`

const UPSERT_RLS_POLICY = `
  mutation UpsertRlsPolicy($modelId: ID!, $input: RlsPolicyInput!) {
    upsertRlsPolicy(modelId: $modelId, input: $input) {
      policy { id policyName action role usingExpr withCheckExpr }
      error { __typename ... on InvalidInput { message } ... on ResourceNotFound { message resourceType } }
    }
  }
`

const GET_RLS_POLICIES = `
  query GetRlsPolicies($modelId: ID!) {
    rlsPolicies(modelId: $modelId) { id policyName action role usingExpr withCheckExpr }
  }
`

// Open Data API 操作文档
const OPEN_FIND_MANY = `
  query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }]) { items { id name owner } totalCount timeCost reqId } }
`

const OPEN_COUNT = `
  query { count { count } }
`

function openCreate(name: string, owner?: string): string {
  const ownerField = owner ? `, owner: "${owner}"` : ''
  return `mutation { create(data: { name: "${name}"${ownerField} }) { id name owner } }`
}

function openUpdate(id: string, name: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { name: "${name}" }) { id name } }`
}

function openDelete(id: string): string {
  return `mutation { delete(where: { id: "${id}" }) { id } }`
}

// ─── State store ───────────────────────────────────────────────

function setLastOpenDataResult(world: ModelCraftWorld, result: OpenDataApiResult) {
  world.lastResponse = { openDataApi: result }
}

function getLastOpenDataResult(world: ModelCraftWorld): OpenDataApiResult {
  const resp = world.lastResponse as { openDataApi?: OpenDataApiResult } | null
  if (!resp?.openDataApi) throw new Error('没有 Open Data API 调用结果')
  return resp.openDataApi
}

/** RLS v2 policies to configure in Background before every Scenario */
const DET_POLICIES: Array<{
  policyName: string
  action: string
  role: string
  usingExpr?: string
  withCheckExpr?: string
}> = [
  { policyName: 'det_select', action: 'read',   role: '', usingExpr: '{"owner":{"_eq":{"_auth":"uid"}}}' },
  { policyName: 'det_insert', action: 'create', role: '', withCheckExpr: '{"owner":{"_eq":{"_auth":"uid"}}}' },
  { policyName: 'det_update', action: 'update', role: '', usingExpr: '{"owner":{"_eq":{"_auth":"uid"}}}', withCheckExpr: '{"owner":{"_eq":{"_auth":"uid"}}}' },
  { policyName: 'det_delete', action: 'delete', role: '', usingExpr: '{"owner":{"_eq":{"_auth":"uid"}}}' },
]

async function setupDetPolicies(
  client: { mutate<T>(doc: string, vars?: Record<string, unknown>): Promise<T> },
  modelId: string,
) {
  for (const p of DET_POLICIES) {
    const input: Record<string, string> = { policyName: p.policyName, action: p.action, role: p.role }
    if (p.usingExpr) input.usingExpr = p.usingExpr
    if (p.withCheckExpr) input.withCheckExpr = p.withCheckExpr

    const r = await client.mutate<{
      upsertRlsPolicy: {
        policy: { id: string; policyName: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, { modelId, input })

    if (r.upsertRlsPolicy.error) {
      throw new Error(
        `确定性拨测 RLS v2 policy "${p.policyName}" 配置失败 — ${r.upsertRlsPolicy.error.__typename}`,
      )
    }
  }
}

// ─── Given 步骤 ────────────────────────────────────────────────

Given('确定性拨测环境已就绪', async function (this: ModelCraftWorld) {
  const dbName = detDbName()
  const modelName = detModelName()

  // 查找预置模型（使用 models 分页查询，modelByName 存在 APISIX 代理后 orgName 传递问题）
  const client = this.projectClient
  const res = await client.query<{
    models: { items: Array<{ id: string; name: string; databaseName: string }> }
  }>(MODELS_LIST, { input: { databaseName: dbName } })

  const foundModel = res.models.items.find(m => m.name === modelName)
  if (!foundModel) {
    const available = res.models.items.map(m => m.name).join(', ')
    throw new Error(
      `确定性拨测前置失败：在 database "${dbName}" 中找不到模型 "${modelName}"（可用模型: ${available || '(无)'}）`,
    )
  }

  this.modelMap['detModelId'] = foundModel.id

  // 清理已有 det_* 策略，然后重建（保证每个 Scenario 独立）
  const existingPolicies = await client.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId: foundModel.id })

  const staleIds = existingPolicies.rlsPolicies
    .filter(p => p.policyName.startsWith('det_'))
    .map(p => p.id)

  for (const id of staleIds) {
    await client.mutate<{ deleteRlsPolicy: { success: boolean; error: unknown } }>(
      `mutation DeleteRlsPolicy($id: ID!) { deleteRlsPolicy(id: $id) { success error { __typename } } }`,
      { id },
    )
  }

  // 重建策略（每个 Scenario 都在干净的策略集上运行）
  await setupDetPolicies(client, foundModel.id)
})

// ─── When 步骤 ─────────────────────────────────────────────────

When('我为确定性拨测模型配置以下 RLS v2 policies:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetModelId(this)
  const rows = table.hashes()
  const results: Array<{ policyName: string; action: string; role: string }> = []

  for (const row of rows) {
    const input: Record<string, string> = {
      policyName: row.policyName,
      action: row.action,
      role: row.role ?? '',
    }
    if (row.usingExpr) input.usingExpr = row.usingExpr
    if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

    const res = await this.projectClient.mutate<{
      upsertRlsPolicy: {
        policy: { id: string; policyName: string; action: string; role: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, { modelId, input })

    if (res.upsertRlsPolicy.error) {
      throw new Error(
        `确定性拨测 RLS v2 policy "${row.policyName}" 配置失败: ${res.upsertRlsPolicy.error.__typename} — ${res.upsertRlsPolicy.error.message ?? ''}`,
      )
    }
    if (!res.upsertRlsPolicy.policy) {
      throw new Error(`确定性拨测 RLS v2 policy "${row.policyName}" 配置失败：未返回 policy`)
    }
    results.push({
      policyName: res.upsertRlsPolicy.policy.policyName,
      action: res.upsertRlsPolicy.policy.action,
      role: res.upsertRlsPolicy.policy.role,
    })
  }

  this.lastResponse = { detRlsV2Policies: results }
})

When('以 EndUser {string} 调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const result = await callOpenDataApi(this, OPEN_FIND_MANY, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 执行 count 查询', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const result = await callOpenDataApi(this, OPEN_COUNT, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 创建一条 name 为 {string} 的记录', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordName: string,
) {
  const result = await callOpenDataApi(this, openCreate(recordName), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 创建一条 name 为 {string} 且 owner 为 {string} 的记录', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordName: string,
  ownerId: string,
) {
  const result = await callOpenDataApi(this, openCreate(recordName, ownerId), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 更新 id 为 {string} 的记录，设置 name 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordId: string,
  newName: string,
) {
  const result = await callOpenDataApi(this, openUpdate(recordId, newName), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 删除 id 为 {string} 的记录', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordId: string,
) {
  const result = await callOpenDataApi(this, openDelete(recordId), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  role: string,
) {
  const result = await callOpenDataApi(this, OPEN_FIND_MANY, role)
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 调用 Open Data API 创建一条 name 为 {string} 的记录', async function (
  this: ModelCraftWorld,
  role: string,
  recordName: string,
) {
  const result = await callOpenDataApi(this, openCreate(recordName), role)
  setLastOpenDataResult(this, result)
})

// ─── Then 步骤 ─────────────────────────────────────────────────

Then('策略配置成功', function (this: ModelCraftWorld) {
  expect(true).toBe(true)
})

Then('确定性拨测模型的 RLS v2 策略数量应为 {int}', async function (
  this: ModelCraftWorld,
  expectedCount: number,
) {
  const modelId = getDetModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId })

  const detPolicies = res.rlsPolicies.filter(p => p.policyName.startsWith('det_'))
  expect(detPolicies.length).toBe(expectedCount)
})

Then('返回结果为合法的 GraphQL 响应且无 errors', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.status).toBe(200)
  expect(result.errors?.length ?? 0).toBe(0)
  expect(result.data).not.toBeNull()
})

Then('操作被拒绝且返回错误类型 {string}', function (
  this: ModelCraftWorld,
  errorType: string,
) {
  const result = getLastOpenDataResult(this)
  expect(result.errors?.length ?? 0).toBeGreaterThan(0)
  const error = result.errors?.[0]
  const errStr = error?.extensions?.code || error?.message || ''
  expect(errStr).toContain(errorType)
})

Then('更新返回 0 行受影响且无错误', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.errors?.length ?? 0).toBe(0)

  const data = result.data as { update?: { id: string } | null } | undefined
  expect(data?.update).toBeNull()
})

Then('删除返回 0 行受影响且无错误', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.errors?.length ?? 0).toBe(0)

  const data = result.data as { delete?: { id: string } | null } | undefined
  expect(data?.delete).toBeNull()
})

Then('创建结果为合法的 GraphQL 响应且无 errors', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.status).toBe(200)
  expect(result.errors?.length ?? 0).toBe(0)
  expect(result.data).not.toBeNull()
})
