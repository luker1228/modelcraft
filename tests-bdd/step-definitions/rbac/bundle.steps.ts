// tests-bdd/step-definitions/rbac/bundle.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'
import { uniqueName } from '../../fixtures/factory'

// ── GraphQL 文档 ───────────────────────────────────────────────

const CREATE_BUNDLE = `
  mutation CreateEndUserPermissionBundle($input: CreateEndUserPermissionBundleInput!) {
    createEndUserPermissionBundle(input: $input) {
      bundle {
        id
        name
        permissions { sortOrder permission { id } }
      }
      error {
        __typename
        ... on EndUserPermissionBundleAlreadyExists { message }
        ... on InvalidInput { message }
      }
    }
  }
`

const DELETE_BUNDLE = `
  mutation DeleteEndUserPermissionBundle($id: ID!) {
    deleteEndUserPermissionBundle(id: $id) {
      success
      error {
        __typename
        ... on EndUserPermissionBundleNotFound { message }
        ... on EndUserPermissionBundleInUse { message }
      }
    }
  }
`

const ADD_PERMISSION_TO_BUNDLE = `
  mutation AddEndUserPermissionToBundle($input: AddEndUserPermissionToBundleInput!) {
    addEndUserPermissionToBundle(input: $input) {
      bundle {
        id
        permissions {
          sortOrder
          permission { id }
        }
      }
      error {
        __typename
        ... on EndUserPermissionBundleNotFound { message }
        ... on EndUserPermissionNotFound { message }
      }
    }
  }
`

const REMOVE_PERMISSION_FROM_BUNDLE = `
  mutation RemoveEndUserPermissionFromBundle($input: RemoveEndUserPermissionFromBundleInput!) {
    removeEndUserPermissionFromBundle(input: $input) {
      bundle {
        id
        permissions { sortOrder permission { id } }
      }
      error {
        __typename
        ... on EndUserPermissionBundleNotFound { message }
        ... on EndUserPermissionNotFound { message }
      }
    }
  }
`

// ── 辅助函数 ──────────────────────────────────────────────────

function initRbacState(world: ModelCraftWorld) {
  if (world.currentPermissionId === undefined) world.currentPermissionId = null
  if (world.createdPermissionIds === undefined) world.createdPermissionIds = []
  if (world.currentBundleId === undefined) world.currentBundleId = null
  if (world.createdBundleIds === undefined) world.createdBundleIds = []
}

// ── Given ─────────────────────────────────────────────────────

Given('已创建名为 {string} 的权限包', async function (this: ModelCraftWorld, baseName: string) {
  initRbacState(this)
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createEndUserPermissionBundle: { bundle: { id: string } | null; error: unknown }
  }>(CREATE_BUNDLE, { input: { name } })

  const bundle = res.createEndUserPermissionBundle.bundle
  if (!bundle) throw new Error(`前置条件：创建权限包失败: ${JSON.stringify(res.createEndUserPermissionBundle.error)}`)

  this.currentBundleId = bundle.id
  this.createdBundleIds.push(bundle.id)
  this.lastResponse = { createEndUserPermissionBundle: res.createEndUserPermissionBundle }
})

Given('已将该权限点添加到该权限包', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const bundleId = this.currentBundleId
  const permissionId = this.currentPermissionId
  if (!bundleId) throw new Error('没有可用的权限包 ID')
  if (!permissionId) throw new Error('没有可用的权限点 ID')

  const res = await this.projectClient.mutate<{
    addEndUserPermissionToBundle: { bundle: unknown; error: unknown }
  }>(ADD_PERMISSION_TO_BUNDLE, {
    input: { bundleId, permissionId, sortOrder: 1 },
  })
  if (res.addEndUserPermissionToBundle.error) {
    throw new Error(`前置条件：添加权限点到权限包失败: ${JSON.stringify(res.addEndUserPermissionToBundle.error)}`)
  }
})

// ── When ──────────────────────────────────────────────────────

When('我创建名为 {string} 的权限包', async function (this: ModelCraftWorld, baseName: string) {
  initRbacState(this)
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createEndUserPermissionBundle: { bundle: { id: string } | null; error: unknown }
  }>(CREATE_BUNDLE, { input: { name } })

  this.lastResponse = { createEndUserPermissionBundle: res.createEndUserPermissionBundle }
  if (res.createEndUserPermissionBundle.bundle?.id) {
    this.currentBundleId = res.createEndUserPermissionBundle.bundle.id
    this.createdBundleIds.push(res.createEndUserPermissionBundle.bundle.id)
  }
})

When(
  '我将该权限点添加到该权限包，排序={int}',
  async function (this: ModelCraftWorld, sortOrder: number) {
    initRbacState(this)
    const bundleId = this.currentBundleId
    const permissionId = this.currentPermissionId
    if (!bundleId) throw new Error('没有可用的权限包 ID')
    if (!permissionId) throw new Error('没有可用的权限点 ID')

    const res = await this.projectClient.mutate<{
      addEndUserPermissionToBundle: {
        bundle: { permissions: Array<unknown> } | null
        error: unknown
      }
    }>(ADD_PERMISSION_TO_BUNDLE, {
      input: { bundleId, permissionId, sortOrder },
    })
    this.lastResponse = { addEndUserPermissionToBundle: res.addEndUserPermissionToBundle }
  }
)

When('我从该权限包移除该权限点', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const bundleId = this.currentBundleId
  const permissionId = this.currentPermissionId
  if (!bundleId) throw new Error('没有可用的权限包 ID')
  if (!permissionId) throw new Error('没有可用的权限点 ID')

  const res = await this.projectClient.mutate<{
    removeEndUserPermissionFromBundle: { bundle: unknown; error: unknown }
  }>(REMOVE_PERMISSION_FROM_BUNDLE, { input: { bundleId, permissionId } })
  this.lastResponse = { removeEndUserPermissionFromBundle: res.removeEndUserPermissionFromBundle }
})

When('我删除该权限包', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const id = this.currentBundleId
  if (!id) throw new Error('没有可删除的权限包')

  const res = await this.projectClient.mutate<{
    deleteEndUserPermissionBundle: { success: boolean; error: unknown }
  }>(DELETE_BUNDLE, { id })

  this.lastResponse = { deleteEndUserPermissionBundle: res.deleteEndUserPermissionBundle }
  if (res.deleteEndUserPermissionBundle.success) {
    this.createdBundleIds = this.createdBundleIds.filter(b => b !== id)
    this.currentBundleId = null
  }
})

// ── Then ──────────────────────────────────────────────────────

Then('权限包创建成功', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserPermissionBundle: { bundle: unknown; error: unknown }
    }
  ).createEndUserPermissionBundle
  expect(payload.error).toBeNull()
  expect(payload.bundle).not.toBeNull()
})

Then('响应包含权限包 ID', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserPermissionBundle: { bundle: { id: string } | null }
    }
  ).createEndUserPermissionBundle
  expect(payload.bundle?.id).toBeTruthy()
})

Then('添加操作成功', function (this: ModelCraftWorld) {
  const r = this.lastResponse as Record<string, { bundle?: unknown; error?: unknown }>
  const key = Object.keys(r)[0]
  expect(r[key].error).toBeNull()
  expect(r[key].bundle).not.toBeNull()
})

Then('权限包内权限点列表不为空', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      addEndUserPermissionToBundle: {
        bundle: { permissions: Array<unknown> } | null
      }
    }
  ).addEndUserPermissionToBundle
  expect(payload.bundle?.permissions.length).toBeGreaterThan(0)
})

Then('移除操作成功', function (this: ModelCraftWorld) {
  const r = this.lastResponse as Record<string, { bundle?: unknown; error?: unknown }>
  const key = Object.keys(r)[0]
  expect(r[key].error).toBeNull()
})
