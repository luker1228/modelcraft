/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
/**
 * RBAC 模块 MSW mock handlers
 *
 * 页面 key: 'rbac-bundles'
 *
 * 启用方式：
 *   NEXT_PUBLIC_MOCK_PAGES=rbac-bundles
 */

import { graphql, HttpResponse } from 'msw'
import { faker } from '@faker-js/faker'

// ── Mock 数据工厂 ──────────────────────────────────────────────────────────

const PRESET_TYPES = ['READ_WRITE_ALL', 'READ_ALL', 'READ_WRITE_OWNER', 'READ_ALL_WRITE_OWNER'] as const
const ACTION_TYPES = ['SELECT', 'INSERT', 'UPDATE', 'DELETE'] as const
const ROW_SCOPE_TYPES = ['ALL', 'OWNER'] as const

function createMockPermission(override: Record<string, unknown> = {}) {
  const modelId = faker.string.uuid()
  return {
    id: faker.string.uuid(),
    modelId,
    modelDisplayName: faker.commerce.productName(),
    action: faker.helpers.arrayElement(ACTION_TYPES),
    rowScope: faker.helpers.arrayElement(ROW_SCOPE_TYPES),
    preset: faker.datatype.boolean() ? faker.helpers.arrayElement(PRESET_TYPES) : null,
    createdAt: faker.date.recent().toISOString(),
    ...override,
  }
}

function createMockBundle(permissions: any[] = [], override: Record<string, unknown> = {}) {
  return {
    id: faker.string.uuid(),
    name: faker.helpers.slugify(faker.word.noun()),
    description: faker.lorem.sentence(),
    permissions,
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

// ── 内存存储（模拟数据库）────────────────────────────────────────────────
const mockPermissions: any[] = Array.from({ length: 20 }, () => createMockPermission())
const mockBundles: any[] = Array.from({ length: 3 }, (_, i) => {
  const bundlePermissions = mockPermissions.slice(i * 5, i * 5 + 3)
  return createMockBundle(bundlePermissions, { name: `权限包 ${i + 1}` })
})

// ── Handlers ────────────────────────────────────────────────────────────────

export const rbacHandlers = [
  // GetEndUserPermissions
  graphql.query('GetEndUserPermissions', ({ variables }) => {
    const { modelId } = variables
    let filtered = [...mockPermissions]

    if (modelId) {
      filtered = filtered.filter((p) => p.modelId === modelId)
    }

    return HttpResponse.json({
      data: {
        endUserPermissions: {
          edges: filtered.map((node) => ({ node, cursor: node.id })),
          pageInfo: { hasNextPage: false, hasPreviousPage: false, startCursor: null, endCursor: null },
          totalCount: filtered.length,
        },
      },
    })
  }),

  // GetEndUserPermissionBundles
  graphql.query('GetEndUserPermissionBundles', () => {
    return HttpResponse.json({
      data: {
        endUserPermissionBundles: {
          edges: mockBundles.map((node) => ({ node, cursor: node.id })),
          pageInfo: { hasNextPage: false, hasPreviousPage: false, startCursor: null, endCursor: null },
          totalCount: mockBundles.length,
        },
      },
    })
  }),

  // GetEndUserPermissionBundle
  graphql.query('GetEndUserPermissionBundle', ({ variables }) => {
    const { id } = variables
    const bundle = mockBundles.find((b) => b.id === id)

    if (!bundle) {
      return HttpResponse.json({
        errors: [{ message: '权限包不存在', extensions: { code: 'NOT_FOUND' } }],
      })
    }

    return HttpResponse.json({
      data: {
        endUserPermissionBundle: bundle,
      },
    })
  }),

  // AddEndUserPermissionToBundle
  graphql.mutation('AddEndUserPermissionToBundle', ({ variables }) => {
    const { bundleId, permissionId } = variables
    const bundle = mockBundles.find((b) => b.id === bundleId)

    if (!bundle) {
      return HttpResponse.json({
        errors: [{ message: '权限包不存在', extensions: { code: 'NOT_FOUND' } }],
      })
    }

    const permission = mockPermissions.find((p) => p.id === permissionId)
    if (!permission) {
      return HttpResponse.json({
        errors: [{ message: '权限点不存在', extensions: { code: 'NOT_FOUND' } }],
      })
    }

    // Check if already added
    if (!bundle.permissions.some((p: any) => p.id === permissionId)) {
      bundle.permissions.push(permission)
    }

    return HttpResponse.json({
      data: {
        addEndUserPermissionToBundle: {
          success: true,
          permission: permission,
          errorMessage: null,
        },
      },
    })
  }),

  // RemoveEndUserPermissionFromBundle
  graphql.mutation('RemoveEndUserPermissionFromBundle', ({ variables }) => {
    const { bundleId, permissionId } = variables
    const bundle = mockBundles.find((b) => b.id === bundleId)

    if (!bundle) {
      return HttpResponse.json({
        errors: [{ message: '权限包不存在', extensions: { code: 'NOT_FOUND' } }],
      })
    }

    bundle.permissions = bundle.permissions.filter((p: any) => p.id !== permissionId)

    return HttpResponse.json({
      data: {
        removeEndUserPermissionFromBundle: {
          success: true,
          errorMessage: null,
        },
      },
    })
  }),

  // CreateEndUserPermissionBundle
  graphql.mutation('CreateEndUserPermissionBundle', ({ variables }) => {
    const { input } = variables
    const newBundle = createMockBundle([], {
      name: input.name,
      description: input.description,
    })

    mockBundles.push(newBundle)

    return HttpResponse.json({
      data: {
        createEndUserPermissionBundle: {
          success: true,
          bundle: newBundle,
          errorMessage: null,
        },
      },
    })
  }),

  // DeleteEndUserPermissionBundle
  graphql.mutation('DeleteEndUserPermissionBundle', ({ variables }) => {
    const { id } = variables
    const index = mockBundles.findIndex((b) => b.id === id)

    if (index === -1) {
      return HttpResponse.json({
        errors: [{ message: '权限包不存在', extensions: { code: 'NOT_FOUND' } }],
      })
    }

    mockBundles.splice(index, 1)

    return HttpResponse.json({
      data: {
        deleteEndUserPermissionBundle: {
          success: true,
          errorMessage: null,
        },
      },
    })
  }),
]
