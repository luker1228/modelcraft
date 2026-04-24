// tests-bdd/step-definitions/rbac/permission.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'
import { uniqueName } from '../../fixtures/factory'

// ── GraphQL 文档 ───────────────────────────────────────────────

const CREATE_PERMISSION = `
  mutation CreateEndUserPermission($input: CreateEndUserPermissionInput!) {
    createEndUserPermission(input: $input) {
      permission {
        id
        modelId
        action
        rowScope
        displayName
        columnPolicy {
          defaultMode
          rules { fieldName mode }
        }
      }
      error {
        __typename
        ... on InvalidInput { message }
        ... on ModelNotFound { message }
        ... on RowScopeFieldMissing { message missingField }
        ... on ProjectNotFound { message }
      }
    }
  }
`

const DELETE_PERMISSION = `
  mutation DeleteEndUserPermission($id: ID!) {
    deleteEndUserPermission(id: $id) {
      success
      error {
        __typename
        ... on EndUserPermissionNotFound { message }
        ... on EndUserPermissionInUse { message }
      }
    }
  }
`

const LIST_PERMISSIONS = `
  query ListEndUserPermissions {
    endUserPermissions {
      edges {
        node {
          id
          modelId
          action
          rowScope
          displayName
        }
        cursor
      }
      totalCount
    }
  }
`

// ── 状态追踪（World 扩展） ─────────────────────────────────────

declare module '../../support/world' {
  interface ModelCraftWorld {
    currentPermissionId: string | null
    createdPermissionIds: string[]
    currentBundleId: string | null
    createdBundleIds: string[]
    currentRoleId: string | null
    createdRoleIds: string[]
  }
}

function initRbacState(world: ModelCraftWorld) {
  if (world.currentPermissionId === undefined) world.currentPermissionId = null
  if (world.createdPermissionIds === undefined) world.createdPermissionIds = []
  if (world.currentBundleId === undefined) world.currentBundleId = null
  if (world.createdBundleIds === undefined) world.createdBundleIds = []
  if (world.currentRoleId === undefined) world.currentRoleId = null
  if (world.createdRoleIds === undefined) world.createdRoleIds = []
}

// ── Given ─────────────────────────────────────────────────────

Given(
  '已为该模型创建权限点，名称={string}，动作={word}，行策略={word}',
  async function (
    this: ModelCraftWorld,
    baseName: string,
    action: string,
    rowScope: string
  ) {
    initRbacState(this)
    const modelId = this.currentModelId
    if (!modelId) throw new Error('前置条件：没有可用的模型 ID，请先创建模型')

    const displayName = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEndUserPermission: {
        permission: { id: string } | null
        error: unknown
      }
    }>(CREATE_PERMISSION, {
      input: {
        modelId,
        action,
        rowScope,
        displayName,
        columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
      },
    })

    const perm = res.createEndUserPermission.permission
    if (!perm) throw new Error(`前置条件：创建权限点失败: ${JSON.stringify(res.createEndUserPermission.error)}`)

    this.currentPermissionId = perm.id
    this.createdPermissionIds.push(perm.id)
    this.lastResponse = { createEndUserPermission: res.createEndUserPermission }
  }
)

// ── When ──────────────────────────────────────────────────────

When(
  '我为该模型创建权限点，名称={string}，动作={word}，行策略={word}',
  async function (
    this: ModelCraftWorld,
    baseName: string,
    action: string,
    rowScope: string
  ) {
    initRbacState(this)
    const modelId = this.currentModelId
    if (!modelId) throw new Error('没有可用的模型 ID，请先用 Given 创建模型')

    const displayName = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEndUserPermission: {
        permission: { id: string } | null
        error: unknown
      }
    }>(CREATE_PERMISSION, {
      input: {
        modelId,
        action,
        rowScope,
        displayName,
        columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
      },
    })

    this.lastResponse = { createEndUserPermission: res.createEndUserPermission }
    if (res.createEndUserPermission.permission?.id) {
      this.currentPermissionId = res.createEndUserPermission.permission.id
      this.createdPermissionIds.push(res.createEndUserPermission.permission.id)
    }
  }
)

When(
  '我再次为该模型创建相同动作行策略的权限点，名称={string}，动作={word}，行策略={word}',
  async function (
    this: ModelCraftWorld,
    baseName: string,
    action: string,
    rowScope: string
  ) {
    initRbacState(this)
    const modelId = this.currentModelId
    if (!modelId) throw new Error('没有可用的模型 ID')

    // 使用相同的 uniqueName base，但实际上 DB 约束是 model_id+action+row_scope+name
    // 为了触发重名，使用和 Given 相同的 baseName 生成同名 displayName（注：这里直接用 baseName 不加随机后缀）
    const res = await this.projectClient.mutate<{
      createEndUserPermission: {
        permission: { id: string } | null
        error: unknown
      }
    }>(CREATE_PERMISSION, {
      input: {
        modelId,
        action,
        rowScope,
        displayName: baseName, // 不加随机后缀，与 Given 中的名称冲突（DB unique constraint）
        columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
      },
    })

    this.lastResponse = { createEndUserPermission: res.createEndUserPermission }
  }
)

When('我删除该权限点', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const id = this.currentPermissionId
  if (!id) throw new Error('没有可删除的权限点')

  const res = await this.projectClient.mutate<{
    deleteEndUserPermission: { success: boolean; error: unknown }
  }>(DELETE_PERMISSION, { id })

  this.lastResponse = { deleteEndUserPermission: res.deleteEndUserPermission }
  if (res.deleteEndUserPermission.success) {
    this.createdPermissionIds = this.createdPermissionIds.filter(p => p !== id)
    this.currentPermissionId = null
  }
})

When('我查询项目下所有权限点', async function (this: ModelCraftWorld) {
  const res = await this.projectClient.query<{
    endUserPermissions: {
      edges: Array<{ node: { id: string } }>
      totalCount: number
    }
  }>(LIST_PERMISSIONS)
  this.lastResponse = { endUserPermissions: res.endUserPermissions }
})

// ── Then ──────────────────────────────────────────────────────

Then('权限点创建成功', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserPermission: { permission: unknown; error: unknown }
    }
  ).createEndUserPermission
  expect(payload.error).toBeNull()
  expect(payload.permission).not.toBeNull()
})

Then('响应包含权限点 ID', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserPermission: { permission: { id: string } | null }
    }
  ).createEndUserPermission
  expect(payload.permission?.id).toBeTruthy()
})

Then('权限点列表不为空', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      endUserPermissions: { edges: Array<unknown>; totalCount: number }
    }
  ).endUserPermissions
  expect(payload.edges.length).toBeGreaterThan(0)
  expect(payload.totalCount).toBeGreaterThan(0)
})

Then('删除操作成功', function (this: ModelCraftWorld) {
  // 判断是权限点、权限包还是角色删除
  const r = this.lastResponse as Record<string, { success?: boolean; error?: unknown }>
  const key = Object.keys(r)[0]
  const payload = r[key]
  expect(payload.error).toBeNull()
  expect(payload.success).toBe(true)
})

Then('应该返回权限点错误', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserPermission: { permission: unknown; error: { __typename: string } | null }
    }
  ).createEndUserPermission
  expect(payload.error).not.toBeNull()
  expect(payload.permission).toBeNull()
})
