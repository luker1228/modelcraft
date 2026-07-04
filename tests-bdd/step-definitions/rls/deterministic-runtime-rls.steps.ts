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

function detProductsModelName(): string {
  return process.env.DET_PRODUCTS_MODEL_NAME ?? 'products'
}

function detPmDbName(): string {
  return process.env.DET_PM_DB_NAME ?? 'demo_pm'
}

function detTaskModelName(): string {
  return process.env.DET_TASK_MODEL ?? 'tasks'
}

function detCommentsModelName(): string {
  return process.env.DET_COMMENTS_MODEL ?? 'task_comments'
}

function detReporterUserId(): string {
  return process.env.DET_REPORTER_USER_ID ?? 'rls-reporter-001'
}

function detAssigneeUserId(): string {
  return process.env.DET_ASSIGNEE_USER_ID ?? 'rls-assignee-002'
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

function getDetProductsModelId(world: ModelCraftWorld): string {
  const id = world.modelMap['detProductsModelId']
  if (!id) throw new Error('products 模型 ID 未就绪（请确认 Background 已执行）')
  return id
}

function getDetTasksModelId(world: ModelCraftWorld): string {
  const id = world.modelMap['detTasksModelId']
  if (!id) throw new Error('tasks 模型 ID 未就绪（请确认 Background 已执行）')
  return id
}

function getDetCommentsModelId(world: ModelCraftWorld): string {
  const id = world.modelMap['detCommentsModelId']
  if (!id) throw new Error('task_comments 模型 ID 未就绪（请确认 Background 已执行）')
  return id
}

// ─── Open Data API client ──────────────────────────────────────

interface OpenDataApiResult {
  status: number
  data: unknown
  errors?: Array<{ message: string; extensions?: { code: string } }>
  clientRequestId: string
  requestId?: string
}

function genClientRequestId(): string {
  return `bdd-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`
}

function logRequestIds(_label: string, _clientReqId: string, _requestId?: string) {
  // 由 setLastOpenDataResult 统一打印，此处不再重复输出
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

function openDataEndpointForModel(world: ModelCraftWorld, modelName: string): string {
  const org = detOrgName(world)
  const project = detProjectSlug(world)
  const db = detDbName()
  return `${API_BASE_URL}/end-user/graphql/org/${org}/project/${project}/db/${db}/model/${modelName}`
}

function openDataEndpointForModelInDb(world: ModelCraftWorld, dbName: string, modelName: string): string {
  const org = detOrgName(world)
  const project = detProjectSlug(world)
  return `${API_BASE_URL}/end-user/graphql/org/${org}/project/${project}/db/${dbName}/model/${modelName}`
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
  const clientRequestId = genClientRequestId()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pat}`,
    'X-MC-Auth-Userid-Str': uid,
    'X-MC-Auth-Username': uname,
    'X-MC-Auth-Roles': role,
    'X-Client-Request-Id': clientRequestId,
  }

  const res = await fetch(endpoint, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query }),
  })

  const body = await res.json()
  const requestId = (body.extensions as { requestId?: string } | undefined)?.requestId
  logRequestIds(`role=${role} uid=${uid}`, clientRequestId, requestId)

  return {
    status: res.status,
    data: body.data,
    errors: body.errors,
    clientRequestId,
    requestId,
  }
}

async function callOpenDataApiForModel(
  world: ModelCraftWorld,
  modelName: string,
  query: string,
  role: string,
  userId?: string,
  userName?: string,
): Promise<OpenDataApiResult> {
  const endpoint = openDataEndpointForModel(world, modelName)
  const pat = detPat()
  const uid = userId ?? detEndUserId()
  const uname = userName ?? detEndUserName()
  const clientRequestId = genClientRequestId()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pat}`,
    'X-MC-Auth-Userid-Str': uid,
    'X-MC-Auth-Username': uname,
    'X-MC-Auth-Roles': role,
    'X-Client-Request-Id': clientRequestId,
  }

  const res = await fetch(endpoint, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query }),
  })

  const body = await res.json()
  const requestId = (body.extensions as { requestId?: string } | undefined)?.requestId
  logRequestIds(`model=${modelName} role=${role} uid=${uid}`, clientRequestId, requestId)

  return {
    status: res.status,
    data: body.data,
    errors: body.errors,
    clientRequestId,
    requestId,
  }
}

async function callOpenDataApiForModelInDb(
  world: ModelCraftWorld,
  dbName: string,
  modelName: string,
  query: string,
  role: string,
  userId?: string,
  userName?: string,
): Promise<OpenDataApiResult> {
  const endpoint = openDataEndpointForModelInDb(world, dbName, modelName)
  const pat = detPat()
  const uid = userId ?? detEndUserId()
  const uname = userName ?? detEndUserName()
  const clientRequestId = genClientRequestId()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pat}`,
    'X-MC-Auth-Userid-Str': uid,
    'X-MC-Auth-Username': uname,
    'X-MC-Auth-Roles': role,
    'X-Client-Request-Id': clientRequestId,
  }

  const res = await fetch(endpoint, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query }),
  })

  const body = await res.json()
  const requestId = (body.extensions as { requestId?: string } | undefined)?.requestId
  logRequestIds(`db=${dbName} model=${modelName} role=${role} uid=${uid}`, clientRequestId, requestId)

  return {
    status: res.status,
    data: body.data,
    errors: body.errors,
    clientRequestId,
    requestId,
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
  query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }]) { items { id } totalCount timeCost reqId } }
`

const OPEN_COUNT = `
  query { count { count } }
`

function openCreate(orderNo: string, userId?: string): string {
  const ts = Date.now().toString(36)
  const id = `bdd-${ts}`.slice(0, 36)
  const uniqueOrderNo = `${orderNo.slice(0, 20)}-${ts}`.slice(0, 32)
  const userIdField = userId ? `, user_id: "${userId}"` : ''
  return `mutation { create(data: { id: "${id}", order_no: "${uniqueOrderNo}", address_id: "addr-bdd", total_amount: 0, paid_amount: 0${userIdField} }) { id } }`
}

function openCreateWithUserId(orderNo: string, userId?: string): string {
  return openCreate(orderNo, userId)
}

function openUpdate(id: string, remark: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { remark: "${remark}" }) { success updatedObj { id } } }`
}

function openDelete(id: string): string {
  return `mutation { delete(where: { id: "${id}" }) { success deletedObj { id } } }`
}

function productsCreate(name: string): string {
  const ts = Date.now().toString(36)
  const id = `prod-bdd-${ts}`.slice(0, 36)
  const uniqueName = `${name.slice(0, 24)}-${ts}`.slice(0, 64)
  return `mutation { create(data: { id: "${id}", name: "${uniqueName}", price: 0, stock: 0, category_id: 1 }) { id error { __typename ... on PermissionDenied { message } ... on DuplicateKey { message } } } }`
}

function productsUpdate(id: string, name: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { name: "${name}" }) { success updatedObj { id } } }`
}

// tasks mutation builders
function tasksCreate(title: string, reporterId: string, projectId?: string): string {
  const ts = Date.now().toString(36)
  const id = `task-bdd-${ts}`.slice(0, 36)
  const uniqueTitle = `${title.slice(0, 24)}-${ts}`.slice(0, 64)
  const pid = projectId ?? 'proj-allowed-1'
  return `mutation { create(data: { id: "${id}", title: "${uniqueTitle}", reporter_id: "${reporterId}", project_id: "${pid}" }) { id } }`
}

function tasksUpdate(id: string, title: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { title: "${title}" }) { success updatedObj { id } } }`
}

function tasksDelete(id: string): string {
  return `mutation { delete(where: { id: "${id}" }) { success deletedObj { id } } }`
}

// task_comments mutation builders
function commentsCreate(content: string, authorId: string, taskId?: string): string {
  const ts = Date.now().toString(36)
  const id = `cmt-bdd-${ts}`.slice(0, 36)
  const taskField = taskId ? `, task_id: "${taskId}"` : ', task_id: "task-bdd-default"'
  return `mutation { create(data: { id: "${id}", content: "${content}", author_id: "${authorId}"${taskField} }) { id } }`
}

function commentsUpdate(id: string, content: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { content: "${content}" }) { success updatedObj { id } } }`
}

function commentsDelete(id: string): string {
  return `mutation { delete(where: { id: "${id}" }) { success deletedObj { id } } }`
}

// ─── State store ───────────────────────────────────────────────

function setLastOpenDataResult(world: ModelCraftWorld, result: OpenDataApiResult) {
  world.lastResponse = { openDataApi: result }
  const rid = result.requestId ?? '(none)'
  console.log(`  [rls] ${world.scenarioName}`)
  console.log(`        requestId=${rid}`)
  if (result.requestId) {
    console.log(`        └─ just log-cat ${result.requestId}`)
  }
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
  // orders_* — user_id 字段
  { policyName: 'orders_user_read',    action: 'read',   role: '*',     usingExpr: 'row.user_id == auth.userid' },
  { policyName: 'orders_user_create',  action: 'create', role: '*',     withCheckExpr: 'input.user_id == auth.userid' },
  { policyName: 'orders_user_update',  action: 'update', role: '*',     usingExpr: 'row.user_id == auth.userid', withCheckExpr: 'input.user_id == auth.userid' },
  { policyName: 'orders_user_delete',  action: 'delete', role: '*',     usingExpr: 'row.user_id == auth.userid' },
  { policyName: 'orders_admin_read',   action: 'read',   role: 'admin', usingExpr: 'true' },
  { policyName: 'orders_admin_create', action: 'create', role: 'admin', withCheckExpr: 'true' },
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

  // 查找 products 模型（用于 products 专项测试）
  const productsModelName = detProductsModelName()
  const productsRes = await client.query<{
    models: { items: Array<{ id: string; name: string; databaseName: string }> }
  }>(MODELS_LIST, { input: { databaseName: detDbName() } })
  const productsModel = productsRes.models.items.find(m => m.name === productsModelName)
  if (productsModel) {
    this.modelMap['detProductsModelId'] = productsModel.id
  }

  // 查找 demo_pm 下的 tasks 和 task_comments 模型
  const pmRes = await client.query<{
    models: { items: Array<{ id: string; name: string; databaseName: string }> }
  }>(MODELS_LIST, { input: { databaseName: detPmDbName() } })
  const tasksModel = pmRes.models.items.find(m => m.name === detTaskModelName())
  if (tasksModel) {
    this.modelMap['detTasksModelId'] = tasksModel.id
    // 清理旧 tasks_ 策略
    const tp = await client.query<{ rlsPolicies: Array<{ id: string; policyName: string }> }>(GET_RLS_POLICIES, { modelId: tasksModel.id })
    for (const p of tp.rlsPolicies.filter(p => p.policyName.startsWith('tasks_'))) {
      await client.mutate(`mutation DeleteRlsPolicy($id: ID!) { deleteRlsPolicy(id: $id) { success error { __typename } } }`, { id: p.id })
    }
  }
  const commentsModel = pmRes.models.items.find(m => m.name === detCommentsModelName())
  if (commentsModel) {
    this.modelMap['detCommentsModelId'] = commentsModel.id
    // 清理旧 comments_ 策略
    const cp = await client.query<{ rlsPolicies: Array<{ id: string; policyName: string }> }>(GET_RLS_POLICIES, { modelId: commentsModel.id })
    for (const p of cp.rlsPolicies.filter(p => p.policyName.startsWith('comments_'))) {
      await client.mutate(`mutation DeleteRlsPolicy($id: ID!) { deleteRlsPolicy(id: $id) { success error { __typename } } }`, { id: p.id })
    }
  }

  // 清理已有 det_* 策略，然后重建（保证每个 Scenario 独立）
  const existingPolicies = await client.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId: foundModel.id })

  const staleIds = existingPolicies.rlsPolicies
    .filter(p => p.policyName.startsWith('orders_'))
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
      role: row.role ?? '*',
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

When('以 EndUser {string} 调用 Open Data API 创建一条 user_id 为当前用户的记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  orderNoBase: string,
) {
  const userId = detEndUserId()
  const result = await callOpenDataApi(this, openCreate(orderNoBase, userId), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 创建一条 user_id 为 {string} 的记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  userId: string,
  recordName: string,
) {
  const result = await callOpenDataApi(this, openCreateWithUserId(recordName, userId), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 调用 Open Data API 创建一条 user_id 为当前用户的记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  role: string,
  orderNoBase: string,
) {
  const userId = detEndUserId()
  const result = await callOpenDataApi(this, openCreate(orderNoBase, userId), role)
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 调用 Open Data API 创建一条 user_id 为 {string} 的记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  role: string,
  userId: string,
  orderNoBase: string,
) {
  const result = await callOpenDataApi(this, openCreate(orderNoBase, userId), role)
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

  const relevantPolicies = res.rlsPolicies.filter(p => p.policyName.startsWith('orders_'))
  expect(relevantPolicies.length).toBe(expectedCount)
})

Then('确定性拨测模型的 RLS v2 策略总数（det_ 和 orders_ 前缀）应为 {int}', async function (
  this: ModelCraftWorld,
  expectedCount: number,
) {
  const modelId = getDetModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId })

  const relevantPolicies = res.rlsPolicies.filter(
    p => p.policyName.startsWith('det_') || p.policyName.startsWith('orders_'),
  )
  expect(relevantPolicies.length).toBe(expectedCount)
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
  
  // Map old error codes to new type names
  const errorTypeMap: Record<string, string> = {
    'OPERATION_FAILED.PERMISSION': 'PermissionDenied',
    'NOT_FOUND.RECORD': 'RecordNotFound',
    'CONFLICT.RECORD': 'DuplicateKey',
  }
  const expectedType = errorTypeMap[errorType] || errorType
  
  // Check GraphQL errors first
  if (result.errors && result.errors.length > 0) {
    const error = result.errors[0]
    const errStr = error?.extensions?.code || error?.message || ''
    expect(errStr).toContain(errorType)
    return
  }
  
  // Check structured errors in data payload
  const data = result.data as any
  const operations = ['create', 'createMany', 'update', 'updateMany', 'delete', 'deleteMany', 'findUnique']
  for (const op of operations) {
    if (data?.[op]?.error) {
      const error = data[op].error
      const errStr = error?.__typename || error?.message || ''
      expect(errStr).toContain(expectedType)
      return
    }
  }
  
  throw new Error(`Expected error type ${errorType} but found no errors in response`)
})

Then('更新返回 0 行受影响且无错误', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.errors?.length ?? 0).toBe(0)
  const updatedObj = (result.data as any)?.update?.updatedObj
  expect(updatedObj).toBeNull()
})

Then('更新操作未生效', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  const hasErrors = (result.errors?.length ?? 0) > 0
  const updatedObj = (result.data as any)?.update?.updatedObj
  expect(hasErrors || updatedObj == null).toBe(true)
})

Then('删除返回 0 行受影响且无错误', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.errors?.length ?? 0).toBe(0)
  const deletedObj = (result.data as any)?.delete?.deletedObj
  expect(deletedObj).toBeNull()
})

Then('删除操作未生效', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  const hasErrors = (result.errors?.length ?? 0) > 0
  const deletedObj = (result.data as any)?.delete?.deletedObj
  expect(hasErrors || deletedObj == null).toBe(true)
})

Then('创建结果为合法的 GraphQL 响应且无 errors', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  expect(result.status).toBe(200)
  expect(result.errors?.length ?? 0).toBe(0)
  expect(result.data).not.toBeNull()
})

// ─── Products 步骤 ─────────────────────────────────────────────

When('我为 products 模型配置以下 RLS v2 policies:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetProductsModelId(this)
  const rows = table.hashes()

  for (const row of rows) {
    const input: Record<string, string> = {
      policyName: row.policyName,
      action: row.action,
      role: row.role ?? '*',
    }
    if (row.usingExpr) input.usingExpr = row.usingExpr
    if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

    const res = await this.projectClient.mutate<{
      upsertRlsPolicy: {
        policy: { id: string; policyName: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, { modelId, input })

    if (res.upsertRlsPolicy.error) {
      throw new Error(
        `products RLS v2 policy "${row.policyName}" 配置失败: ${res.upsertRlsPolicy.error.__typename} — ${res.upsertRlsPolicy.error.message ?? ''}`,
      )
    }
  }

  this.lastResponse = { productsRlsConfigured: true }
})

Then('products 模型的 RLS v2 策略数量应为 {int}', async function (
  this: ModelCraftWorld,
  expectedCount: number,
) {
  const modelId = getDetProductsModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId })

  const relevant = res.rlsPolicies.filter(p => p.policyName.startsWith('products_'))
  expect(relevant.length).toBe(expectedCount)
})

When('以 EndUser {string} 对 products 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), OPEN_FIND_MANY, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 products 模型调用 Open Data API 执行 count 查询', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), OPEN_COUNT, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 products 模型调用 Open Data API 创建一条记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  name: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), productsCreate(name), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 对 products 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  role: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), OPEN_FIND_MANY, role)
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 对 products 模型调用 Open Data API 创建一条记录，name 为 {string}', async function (
  this: ModelCraftWorld,
  role: string,
  name: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), productsCreate(name), role)
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 对 products 模型调用 Open Data API 更新 id 为 {string} 的记录，设置 name 为 {string}', async function (
  this: ModelCraftWorld,
  role: string,
  recordId: string,
  newName: string,
) {
  const result = await callOpenDataApiForModel(this, detProductsModelName(), productsUpdate(recordId, newName), role)
  setLastOpenDataResult(this, result)
})

// ─── Tasks / Task_comments 通用 helper ─────────────────────────

async function callPmModel(
  world: ModelCraftWorld,
  modelName: string,
  query: string,
  role: string,
  userId?: string,
): Promise<OpenDataApiResult> {
  return callOpenDataApiForModelInDb(world, detPmDbName(), modelName, query, role, userId)
}

// ─── Tasks 步骤 ────────────────────────────────────────────────

When('我为 tasks 模型配置以下 RLS v2 policies:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetTasksModelId(this)
  const rows = table.hashes()

  for (const row of rows) {
    const input: Record<string, string> = {
      policyName: row.policyName,
      action: row.action,
      role: row.role ?? '*',
    }
    if (row.usingExpr) input.usingExpr = row.usingExpr
    if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

    const res = await this.projectClient.mutate<{
      upsertRlsPolicy: {
        policy: { id: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, { modelId, input })

    if (res.upsertRlsPolicy.error) {
      throw new Error(`tasks RLS policy "${row.policyName}" 失败: ${res.upsertRlsPolicy.error.__typename} — ${res.upsertRlsPolicy.error.message ?? ''}`)
    }
  }
  this.lastResponse = { tasksRlsConfigured: true }
})

When('我为 tasks 模型追加以下 RLS v2 policy:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetTasksModelId(this)
  const row = table.hashes()[0]
  const input: Record<string, string> = {
    policyName: row.policyName,
    action: row.action,
    role: row.role ?? '*',
  }
  if (row.usingExpr) input.usingExpr = row.usingExpr
  if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

  const res = await this.projectClient.mutate<{
    upsertRlsPolicy: {
      policy: { id: string } | null
      error: { __typename: string; message?: string } | null
    }
  }>(UPSERT_RLS_POLICY, { modelId, input })

  if (res.upsertRlsPolicy.error) {
    throw new Error(`tasks RLS policy "${row.policyName}" 追加失败: ${res.upsertRlsPolicy.error.__typename}`)
  }
  this.lastResponse = { tasksRlsAppended: true }
})

Then('tasks 模型的 RLS v2 策略数量应为 {int}', async function (
  this: ModelCraftWorld,
  expectedCount: number,
) {
  const modelId = getDetTasksModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId })
  const relevant = res.rlsPolicies.filter(p => p.policyName.startsWith('tasks_'))
  expect(relevant.length).toBe(expectedCount)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
) {
  const result = await callPmModel(this, detTaskModelName(), OPEN_FIND_MANY, 'viewer', detReporterUserId())
  setLastOpenDataResult(this, result)
})

When('以 assignee 用户身份对 tasks 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
) {
  const result = await callPmModel(this, detTaskModelName(), OPEN_FIND_MANY, 'viewer', detAssigneeUserId())
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户的 task，title 为 {string}', async function (
  this: ModelCraftWorld,
  title: string,
) {
  const userId = detReporterUserId()
  const result = await callPmModel(this, detTaskModelName(), tasksCreate(title, userId), 'viewer', userId)
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为 {string} 的 task，title 为 {string}', async function (
  this: ModelCraftWorld,
  reporterId: string,
  title: string,
) {
  const userId = detReporterUserId()
  const result = await callPmModel(this, detTaskModelName(), tasksCreate(title, reporterId), 'viewer', userId)
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户且 project_id 为 {string} 的 task，title 为 {string}', async function (
  this: ModelCraftWorld,
  projectId: string,
  title: string,
) {
  const userId = detReporterUserId()
  const result = await callPmModel(this, detTaskModelName(), tasksCreate(title, userId, projectId), 'viewer', userId)
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 更新 id 为 {string} 的 task，设置 title 为 {string}', async function (
  this: ModelCraftWorld,
  recordId: string,
  title: string,
) {
  const result = await callPmModel(this, detTaskModelName(), tasksUpdate(recordId, title), 'viewer', detReporterUserId())
  setLastOpenDataResult(this, result)
})

When('以 assignee 用户身份对 tasks 模型调用 Open Data API 更新 id 为 {string} 的 task，设置 title 为 {string}', async function (
  this: ModelCraftWorld,
  recordId: string,
  title: string,
) {
  const result = await callPmModel(this, detTaskModelName(), tasksUpdate(recordId, title), 'viewer', detAssigneeUserId())
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 删除 id 为 {string} 的 task', async function (
  this: ModelCraftWorld,
  recordId: string,
) {
  const result = await callPmModel(this, detTaskModelName(), tasksDelete(recordId), 'viewer', detReporterUserId())
  setLastOpenDataResult(this, result)
})

When('以 assignee 用户身份对 tasks 模型调用 Open Data API 删除 id 为 {string} 的 task', async function (
  this: ModelCraftWorld,
  recordId: string,
) {
  const result = await callPmModel(this, detTaskModelName(), tasksDelete(recordId), 'viewer', detAssigneeUserId())
  setLastOpenDataResult(this, result)
})

When('以角色 {string} 对 tasks 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  role: string,
) {
  const result = await callPmModel(this, detTaskModelName(), OPEN_FIND_MANY, role)
  setLastOpenDataResult(this, result)
})

// ─── Task_comments 步骤 ────────────────────────────────────────

When('我为 task_comments 模型配置以下 RLS v2 policies:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetCommentsModelId(this)
  const rows = table.hashes()

  for (const row of rows) {
    const input: Record<string, string> = {
      policyName: row.policyName,
      action: row.action,
      role: row.role ?? '*',
    }
    if (row.usingExpr) input.usingExpr = row.usingExpr
    if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

    const res = await this.projectClient.mutate<{
      upsertRlsPolicy: {
        policy: { id: string } | null
        error: { __typename: string; message?: string } | null
      }
    }>(UPSERT_RLS_POLICY, { modelId, input })

    if (res.upsertRlsPolicy.error) {
      throw new Error(`task_comments RLS policy "${row.policyName}" 失败: ${res.upsertRlsPolicy.error.__typename}`)
    }
  }
  this.lastResponse = { commentsRlsConfigured: true }
})

When('我为 task_comments 模型追加以下 RLS v2 policy:', async function (
  this: ModelCraftWorld,
  table: DataTable,
) {
  const modelId = getDetCommentsModelId(this)
  const row = table.hashes()[0]
  const input: Record<string, string> = {
    policyName: row.policyName,
    action: row.action,
    role: row.role ?? '*',
  }
  if (row.usingExpr) input.usingExpr = row.usingExpr
  if (row.withCheckExpr) input.withCheckExpr = row.withCheckExpr

  const res = await this.projectClient.mutate<{
    upsertRlsPolicy: {
      policy: { id: string } | null
      error: { __typename: string; message?: string } | null
    }
  }>(UPSERT_RLS_POLICY, { modelId, input })

  if (res.upsertRlsPolicy.error) {
    throw new Error(`task_comments RLS policy "${row.policyName}" 追加失败: ${res.upsertRlsPolicy.error.__typename}`)
  }
  this.lastResponse = { commentsRlsAppended: true }
})

Then('task_comments 模型的 RLS v2 策略数量应为 {int}', async function (
  this: ModelCraftWorld,
  expectedCount: number,
) {
  const modelId = getDetCommentsModelId(this)
  const res = await this.projectClient.query<{
    rlsPolicies: Array<{ id: string; policyName: string }>
  }>(GET_RLS_POLICIES, { modelId })
  const relevant = res.rlsPolicies.filter(p => p.policyName.startsWith('comments_'))
  expect(relevant.length).toBe(expectedCount)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 执行 findMany', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const result = await callPmModel(this, detCommentsModelName(), OPEN_FIND_MANY, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 创建一条 author_id 为当前用户的评论，content 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  content: string,
) {
  const userId = detEndUserId()
  const result = await callPmModel(this, detCommentsModelName(), commentsCreate(content, userId), 'viewer', userId)
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 创建一条 author_id 为 {string} 的评论，content 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  authorId: string,
  content: string,
) {
  const result = await callPmModel(this, detCommentsModelName(), commentsCreate(content, authorId), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 创建一条 author_id 为当前用户且 task_id 为 {string} 的评论，content 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  taskId: string,
  content: string,
) {
  const userId = detEndUserId()
  const result = await callPmModel(this, detCommentsModelName(), commentsCreate(content, userId, taskId), 'viewer', userId)
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 更新 id 为 {string} 的评论，设置 content 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordId: string,
  content: string,
) {
  const result = await callPmModel(this, detCommentsModelName(), commentsUpdate(recordId, content), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 task_comments 模型调用 Open Data API 删除 id 为 {string} 的评论', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  recordId: string,
) {
  const result = await callPmModel(this, detCommentsModelName(), commentsDelete(recordId), 'viewer')
  setLastOpenDataResult(this, result)
})

// ─── 跨步骤 ID 传递 ────────────────────────────────────────────

Then('保存上次创建的记录 id', function (this: ModelCraftWorld) {
  const result = getLastOpenDataResult(this)
  const data = result.data as { create?: { id: string } | null } | undefined
  const id = data?.create?.id
  if (!id) throw new Error('上次创建操作未返回 id，无法保存')
  this.modelMap['lastCreatedId'] = id
})

function getLastCreatedId(world: ModelCraftWorld): string {
  const id = world.modelMap['lastCreatedId']
  if (!id) throw new Error('lastCreatedId 未就绪（请确认"保存上次创建的记录 id"步骤已执行）')
  return id
}

When('以 EndUser {string} 调用 Open Data API 更新上次保存的记录，设置 name 为 {string}', async function (
  this: ModelCraftWorld,
  _userLabel: string,
  newName: string,
) {
  const id = getLastCreatedId(this)
  const result = await callOpenDataApi(this, openUpdate(id, newName), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 调用 Open Data API 删除上次保存的记录', async function (
  this: ModelCraftWorld,
  _userLabel: string,
) {
  const id = getLastCreatedId(this)
  const result = await callOpenDataApi(this, openDelete(id), 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 assignee 用户身份对 tasks 模型调用 Open Data API 更新上次保存的 task，设置 title 为 {string}', async function (
  this: ModelCraftWorld,
  title: string,
) {
  const id = getLastCreatedId(this)
  const result = await callPmModel(this, detTaskModelName(), tasksUpdate(id, title), 'viewer', detAssigneeUserId())
  setLastOpenDataResult(this, result)
})

When('以 assignee 用户身份对 tasks 模型调用 Open Data API 删除上次保存的 task', async function (
  this: ModelCraftWorld,
) {
  const id = getLastCreatedId(this)
  const result = await callPmModel(this, detTaskModelName(), tasksDelete(id), 'viewer', detAssigneeUserId())
  setLastOpenDataResult(this, result)
})

When('以 reporter 用户身份对 tasks 模型调用 Open Data API 删除上次保存的 task', async function (
  this: ModelCraftWorld,
) {
  const id = getLastCreatedId(this)
  const result = await callPmModel(this, detTaskModelName(), tasksDelete(id), 'viewer', detReporterUserId())
  setLastOpenDataResult(this, result)
})

// ─── useAdmin 查询条件专项步骤 ─────────────────────────────────────
// X-MC-Auth-Useadmin: true — 设计者以 admin 权限操作 runtime，跳过 RLS 用户过滤

async function callOpenDataApiUseAdmin(
  world: ModelCraftWorld,
  query: string,
  variables?: Record<string, unknown>,
): Promise<OpenDataApiResult> {
  const endpoint = openDataEndpoint(world)
  const pat = detPat()
  const clientRequestId = genClientRequestId()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pat}`,
    'X-MC-Auth-Useadmin': 'true',
    'X-Client-Request-Id': clientRequestId,
  }

  const res = await fetch(endpoint, {
    method: 'POST',
    headers,
    body: JSON.stringify(variables ? { query, variables } : { query }),
  })

  const body = await res.json()
  const requestId = (body.extensions as { requestId?: string } | undefined)?.requestId
  logRequestIds(`useAdmin`, clientRequestId, requestId)

  return { status: res.status, data: body.data, errors: body.errors, clientRequestId, requestId }
}

When('以 useAdmin 方式调用 Open Data API 执行 findMany', async function (this: ModelCraftWorld) {
  const result = await callOpenDataApiUseAdmin(this, OPEN_FIND_MANY)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，take {int} skip {int}', async function (
  this: ModelCraftWorld,
  take: number,
  skip: number,
) {
  const query = `query { findMany(take: ${take}, skip: ${skip}, orderBy: [{ id: asc }]) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，orderBy id desc', async function (this: ModelCraftWorld) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: desc }]) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，orderBy total_amount asc', async function (this: ModelCraftWorld) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ total_amount: asc }]) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where user_id eq {string}', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { user_id: { equals: "${userId}" } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where user_id in {string}', async function (
  this: ModelCraftWorld,
  userIds: string,
) {
  const ids = userIds.split(',').map(s => `"${s.trim()}"`).join(', ')
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { user_id: { in: [${ids}] } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，take {int} skip {int} orderBy id desc', async function (
  this: ModelCraftWorld,
  take: number,
  skip: number,
) {
  const query = `query { findMany(take: ${take}, skip: ${skip}, orderBy: [{ id: desc }]) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 count', async function (this: ModelCraftWorld) {
  const result = await callOpenDataApiUseAdmin(this, OPEN_COUNT)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 count，where user_id eq {string}', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const query = `query { count(where: { user_id: { equals: "${userId}" } }) { count } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where total_amount gt {int}', async function (
  this: ModelCraftWorld,
  value: number,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { total_amount: { gt: ${value} } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where total_amount lte {int}', async function (
  this: ModelCraftWorld,
  value: number,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { total_amount: { lte: ${value} } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where order_no startsWith {string}', async function (
  this: ModelCraftWorld,
  prefix: string,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { order_no: { startsWith: "${prefix}" } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where address_id ne {string}', async function (
  this: ModelCraftWorld,
  value: string,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { address_id: { not: "${value}" } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where remark not {string}', async function (
  this: ModelCraftWorld,
  value: string,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { remark: { not: "${value}" } }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，多字段排序 total_amount asc id desc', async function (this: ModelCraftWorld) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ total_amount: asc }, { id: desc }]) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式调用 Open Data API 执行 findMany，where paid_amount gte {int} lte {int}', async function (
  this: ModelCraftWorld,
  gte: number,
  lte: number,
) {
  const query = `query { findMany(take: 20, skip: 0, orderBy: [{ id: asc }], where: { AND: [{ paid_amount: { gte: ${gte} } }, { paid_amount: { lte: ${lte} } }] }) { items { id } totalCount timeCost reqId } }`
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

// ─── 结构化错误测试步骤 ─────────────────────────────────────────

function openCreateWithFixedId(id: string, orderNo: string, userId: string): string {
  return `mutation { create(data: { id: "${id}", order_no: "${orderNo}", address_id: "addr-bdd", total_amount: 0, paid_amount: 0, user_id: "${userId}" }) { id error { __typename ... on DuplicateKey { message } } } }`
}

function openUpdateNonExistent(id: string): string {
  return `mutation { update(where: { id: "${id}" }, data: { remark: "test" }) { success error { __typename ... on RecordNotFound { message } } } }`
}

function openDeleteNonExistent(id: string): string {
  return `mutation { delete(where: { id: "${id}" }) { success error { __typename ... on RecordNotFound { message } } } }`
}

function openCreateManyWithDuplicateIds(id1: string, id2: string, userId: string): string {
  const ts = Date.now().toString(36)
  return `mutation { createMany(data: [
    { id: "${id1}", order_no: "batch-${ts}-1", address_id: "addr-bdd", total_amount: 0, paid_amount: 0, user_id: "${userId}" },
    { id: "${id2}", order_no: "batch-${ts}-2", address_id: "addr-bdd", total_amount: 0, paid_amount: 0, user_id: "${userId}" }
  ]) { count error { __typename ... on DuplicateKey { message } } } }`
}

function openUpdateManyNonExistent(): string {
  return `mutation { updateMany(where: { AND: [{ order_no: { equals: "non-existent-batch-id" } }] }, data: { remark: "batch-update" }, take: 10) { count error { __typename ... on RecordNotFound { message } } } }`
}

function openDeleteManyNonExistent(): string {
  return `mutation { deleteMany(where: { AND: [{ order_no: { equals: "non-existent-batch-id" } }] }, take: 10) { count error { __typename ... on RecordNotFound { message } } } }`
}

function openCreateManyPermissionDenied(): string {
  const ts = Date.now().toString(36)
  return `mutation { createMany(data: [
    { id: "perm-denied-${ts}-1", order_no: "perm-test-${ts}-1", address_id: "addr-bdd", total_amount: 0, paid_amount: 0, user_id: "other-user-id" }
  ]) { count error { __typename ... on PermissionDenied { message } } } }`
}

function openUpdateManyPermissionDenied(): string {
  return `mutation { updateMany(where: { AND: [{ order_no: { equals: "some-order" } }] }, data: { user_id: "other-user-id" }, take: 10) { count error { __typename ... on PermissionDenied { message } } } }`
}

function openDeleteManyPermissionDenied(): string {
  return `mutation { deleteMany(where: { AND: [{ name: { equals: "some-product" } }] }, take: 10) { count error { __typename ... on PermissionDenied { message } } } }`
}

When('以 EndUser {string} 对 orders 模型创建一条固定 id 的记录', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const ts = Date.now().toString(36)
  const fixedId = `dup-key-${ts}`.slice(0, 36)
  const orderNo = `dup-test-${ts}`
  const query = openCreateWithFixedId(fixedId, orderNo, userId)
  const result = await callOpenDataApi(this, query, 'viewer')
  setLastOpenDataResult(this, result)
  // 保存 fixedId 供后续步骤使用
  if (!this.lastResponse) this.lastResponse = {}
  this.lastResponse.fixedId = fixedId
})

When('以 EndUser {string} 再次对 orders 模型创建相同 id 的记录', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const fixedId = this.lastResponse?.fixedId
  if (!fixedId) throw new Error('需要先创建一条记录以获取 fixedId')
  const orderNo = `dup-test-${Date.now().toString(36)}`
  const query = openCreateWithFixedId(fixedId, orderNo, userId)
  const result = await callOpenDataApi(this, query, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 orders 模型批量创建包含重复 id 的记录', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const fixedId = this.lastResponse?.fixedId
  if (!fixedId) throw new Error('需要先创建一条记录以获取 fixedId')
  const newId = `batch-dup-${Date.now().toString(36)}`.slice(0, 36)
  const query = openCreateManyWithDuplicateIds(fixedId, newId, userId)
  const result = await callOpenDataApi(this, query, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式对 orders 模型更新一条不存在的记录 id 为 {string}', async function (
  this: ModelCraftWorld,
  nonExistentId: string,
) {
  const query = openUpdateNonExistent(nonExistentId)
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式对 orders 模型删除一条不存在的记录 id 为 {string}', async function (
  this: ModelCraftWorld,
  nonExistentId: string,
) {
  const query = openDeleteNonExistent(nonExistentId)
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式对 orders 模型批量更新不存在的记录', async function (
  this: ModelCraftWorld,
) {
  const query = openUpdateManyNonExistent()
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 useAdmin 方式对 orders 模型批量删除不存在的记录', async function (
  this: ModelCraftWorld,
) {
  const query = openDeleteManyNonExistent()
  const result = await callOpenDataApiUseAdmin(this, query)
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 orders 模型批量创建尝试修改 user_id', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const query = openCreateManyPermissionDenied()
  const result = await callOpenDataApi(this, query, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 orders 模型批量更新尝试修改 user_id', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const query = openUpdateManyPermissionDenied()
  const result = await callOpenDataApi(this, query, 'viewer')
  setLastOpenDataResult(this, result)
})

When('以 EndUser {string} 对 products 模型批量删除尝试绕过权限', async function (
  this: ModelCraftWorld,
  userId: string,
) {
  const query = openDeleteManyPermissionDenied()
  const result = await callOpenDataApiForModel(this, detProductsModelName(), query, 'viewer')
  setLastOpenDataResult(this, result)
})

