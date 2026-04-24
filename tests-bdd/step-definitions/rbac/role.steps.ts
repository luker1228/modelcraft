// tests-bdd/step-definitions/rbac/role.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'
import { uniqueName } from '../../fixtures/factory'

// ── GraphQL 文档 ───────────────────────────────────────────────

const CREATE_ROLE = `
  mutation CreateEndUserRole($input: CreateEndUserRoleInput!) {
    createEndUserRole(input: $input) {
      role {
        id
        name
        isImplicit
        permissionBundles { bundle { id name } assignedAt }
      }
      error {
        __typename
        ... on EndUserRoleAlreadyExists { message }
        ... on InvalidInput { message }
      }
    }
  }
`

const DELETE_ROLE = `
  mutation DeleteEndUserRole($id: ID!) {
    deleteEndUserRole(id: $id) {
      success
      error {
        __typename
        ... on EndUserRoleNotFound { message }
        ... on EndUserImplicitRoleCannotBeModified { message }
      }
    }
  }
`

const ASSIGN_BUNDLE_TO_ROLE = `
  mutation AssignBundleToEndUserRole($input: AssignBundleToEndUserRoleInput!) {
    assignBundleToEndUserRole(input: $input) {
      role {
        id
        permissionBundles { bundle { id } assignedAt }
      }
      error {
        __typename
        ... on EndUserRoleNotFound { message }
        ... on EndUserPermissionBundleNotFound { message }
      }
    }
  }
`

const REVOKE_BUNDLE_FROM_ROLE = `
  mutation RevokeBundleFromEndUserRole($input: RevokeBundleFromEndUserRoleInput!) {
    revokeBundleFromEndUserRole(input: $input) {
      role {
        id
        permissionBundles { bundle { id } assignedAt }
      }
      error {
        __typename
        ... on EndUserRoleNotFound { message }
        ... on EndUserPermissionBundleNotFound { message }
      }
    }
  }
`

const LIST_ROLES = `
  query ListEndUserRoles {
    endUserRoles {
      edges {
        node {
          id
          name
          isImplicit
        }
        cursor
      }
      totalCount
    }
  }
`

// ── 辅助函数 ──────────────────────────────────────────────────

function initRbacState(world: ModelCraftWorld) {
  if (world.currentBundleId === undefined) world.currentBundleId = null
  if (world.createdBundleIds === undefined) world.createdBundleIds = []
  if (world.currentRoleId === undefined) world.currentRoleId = null
  if (world.createdRoleIds === undefined) world.createdRoleIds = []
}

// ── Given ─────────────────────────────────────────────────────

Given('已创建名为 {string} 的 RBAC 角色', async function (this: ModelCraftWorld, baseName: string) {
  initRbacState(this)
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createEndUserRole: { role: { id: string } | null; error: unknown }
  }>(CREATE_ROLE, { input: { name } })

  const role = res.createEndUserRole.role
  if (!role) throw new Error(`前置条件：创建角色失败: ${JSON.stringify(res.createEndUserRole.error)}`)

  this.currentRoleId = role.id
  this.createdRoleIds.push(role.id)
  this.lastResponse = { createEndUserRole: res.createEndUserRole }
})

Given('已将该权限包关联到该角色', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const roleId = this.currentRoleId
  const bundleId = this.currentBundleId
  if (!roleId) throw new Error('没有可用的角色 ID')
  if (!bundleId) throw new Error('没有可用的权限包 ID')

  const res = await this.projectClient.mutate<{
    assignBundleToEndUserRole: { role: unknown; error: unknown }
  }>(ASSIGN_BUNDLE_TO_ROLE, { input: { roleId, bundleId } })

  if (res.assignBundleToEndUserRole.error) {
    throw new Error(`前置条件：关联权限包到角色失败: ${JSON.stringify(res.assignBundleToEndUserRole.error)}`)
  }
})

// ── When ──────────────────────────────────────────────────────

When('我创建名为 {string} 的 RBAC 角色', async function (this: ModelCraftWorld, baseName: string) {
  initRbacState(this)
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createEndUserRole: { role: { id: string; isImplicit: boolean } | null; error: unknown }
  }>(CREATE_ROLE, { input: { name } })

  this.lastResponse = { createEndUserRole: res.createEndUserRole }
  if (res.createEndUserRole.role?.id) {
    this.currentRoleId = res.createEndUserRole.role.id
    this.createdRoleIds.push(res.createEndUserRole.role.id)
  }
})

When('我将该权限包关联到该角色', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const roleId = this.currentRoleId
  const bundleId = this.currentBundleId
  if (!roleId) throw new Error('没有可用的角色 ID')
  if (!bundleId) throw new Error('没有可用的权限包 ID')

  const res = await this.projectClient.mutate<{
    assignBundleToEndUserRole: { role: unknown; error: unknown }
  }>(ASSIGN_BUNDLE_TO_ROLE, { input: { roleId, bundleId } })
  this.lastResponse = { assignBundleToEndUserRole: res.assignBundleToEndUserRole }
})

When('我从该角色解除该权限包的关联', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const roleId = this.currentRoleId
  const bundleId = this.currentBundleId
  if (!roleId) throw new Error('没有可用的角色 ID')
  if (!bundleId) throw new Error('没有可用的权限包 ID')

  const res = await this.projectClient.mutate<{
    revokeBundleFromEndUserRole: { role: unknown; error: unknown }
  }>(REVOKE_BUNDLE_FROM_ROLE, { input: { roleId, bundleId } })
  this.lastResponse = { revokeBundleFromEndUserRole: res.revokeBundleFromEndUserRole }
})

When('我删除该角色', async function (this: ModelCraftWorld) {
  initRbacState(this)
  const id = this.currentRoleId
  if (!id) throw new Error('没有可删除的角色')

  const res = await this.projectClient.mutate<{
    deleteEndUserRole: { success: boolean; error: unknown }
  }>(DELETE_ROLE, { id })

  this.lastResponse = { deleteEndUserRole: res.deleteEndUserRole }
  if (res.deleteEndUserRole.success) {
    this.createdRoleIds = this.createdRoleIds.filter(r => r !== id)
    this.currentRoleId = null
  }
})

When('我查询项目下所有角色', async function (this: ModelCraftWorld) {
  const res = await this.projectClient.query<{
    endUserRoles: { edges: Array<{ node: { id: string } }>; totalCount: number }
  }>(LIST_ROLES)
  this.lastResponse = { endUserRoles: res.endUserRoles }
})

// ── Then ──────────────────────────────────────────────────────

Then('角色创建成功', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as { createEndUserRole: { role: unknown; error: unknown } }
  ).createEndUserRole
  expect(payload.error).toBeNull()
  expect(payload.role).not.toBeNull()
})

Then('响应包含角色 ID', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as { createEndUserRole: { role: { id: string } | null } }
  ).createEndUserRole
  expect(payload.role?.id).toBeTruthy()
})

Then('角色不是隐式角色', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      createEndUserRole: { role: { isImplicit: boolean } | null }
    }
  ).createEndUserRole
  expect(payload.role?.isImplicit).toBe(false)
})

Then('关联操作成功', function (this: ModelCraftWorld) {
  const r = this.lastResponse as Record<string, { role?: unknown; error?: unknown }>
  const key = Object.keys(r)[0]
  expect(r[key].error).toBeNull()
  expect(r[key].role).not.toBeNull()
})

Then('解除操作成功', function (this: ModelCraftWorld) {
  const r = this.lastResponse as Record<string, { role?: unknown; error?: unknown }>
  const key = Object.keys(r)[0]
  expect(r[key].error).toBeNull()
})

Then('角色列表不为空', function (this: ModelCraftWorld) {
  const payload = (
    this.lastResponse as {
      endUserRoles: { edges: Array<unknown>; totalCount: number }
    }
  ).endUserRoles
  expect(payload.edges.length).toBeGreaterThan(0)
})
