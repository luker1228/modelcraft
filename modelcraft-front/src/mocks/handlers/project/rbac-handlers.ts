// MSW handlers for EndUser RBAC GraphQL operations
// 使用 graphql.query / graphql.mutation from msw

import { graphql, HttpResponse } from 'msw'
import {
  createMockEndUserPermission,
  createMockEndUserBundle,
  createMockEndUserRole,
  MOCK_IMPLICIT_ROLE,
} from '../../data/project/rbac-factory'

// ── Query Handlers ─────────────────────────────────────────────────────────────

export const rbacHandlers = [
  // GetEndUserPermissions — 返回 3-5 个权限点（现在返回 Connection 类型）
  graphql.query('GetEndUserPermissions', () => {
    const permissions = Array.from(
      { length: 4 },
      () => createMockEndUserPermission()
    )
    return HttpResponse.json({
      data: {
        endUserPermissions: {
          edges: permissions.map(node => ({ node, cursor: node.id })),
          pageInfo: {
            hasNextPage: false,
            endCursor: permissions[permissions.length - 1]?.id,
          },
          totalCount: permissions.length,
        },
      },
    })
  }),

  // GetEndUserBundles — 返回 3 个权限包
  graphql.query('GetEndUserBundles', () => {
    const bundles = Array.from(
      { length: 3 },
      () => createMockEndUserBundle()
    )
    return HttpResponse.json({
      data: {
        endUserPermissionBundles: {
          edges: bundles.map(node => ({ node, cursor: node.id })),
          pageInfo: {
            hasNextPage: false,
            endCursor: bundles[bundles.length - 1]?.id,
          },
          totalCount: bundles.length,
        },
      },
    })
  }),

  // GetEndUserBundle — 返回单个权限包详情
  graphql.query('GetEndUserBundle', ({ variables }) => {
    const bundle = createMockEndUserBundle({ id: variables.id as string })
    return HttpResponse.json({
      data: { endUserPermissionBundle: bundle },
    })
  }),

  // GetEndUserRoles — 返回隐式角色 + 2 个普通角色
  graphql.query('GetEndUserRoles', () => {
    const roles = [
      MOCK_IMPLICIT_ROLE,
      createMockEndUserRole(),
      createMockEndUserRole(),
    ]
    return HttpResponse.json({
      data: {
        endUserRoles: {
          edges: roles.map(node => ({ node, cursor: node.id })),
          pageInfo: {
            hasNextPage: false,
            endCursor: roles[roles.length - 1]?.id,
          },
          totalCount: roles.length,
        },
      },
    })
  }),

  // GetEndUserRole — 返回单个角色详情
  graphql.query('GetEndUserRole', ({ variables }) => {
    const isImplicit = variables.id === MOCK_IMPLICIT_ROLE.id
    const role = isImplicit
      ? MOCK_IMPLICIT_ROLE
      : createMockEndUserRole({ id: variables.id as string })
    return HttpResponse.json({
      data: { endUserRole: role },
    })
  }),

  // GetEndUserEffectivePermissions — 返回有效权限
  graphql.query('GetEndUserEffectivePermissions', ({ variables }) => {
    return HttpResponse.json({
      data: {
        effectivePermissions: {
          effectivePermissions: {
            endUserId: variables.endUserId,
            modelId: variables.modelId,
            grants: [],
          },
          error: null,
        },
      },
    })
  }),

  // ── Mutation Handlers ────────────────────────────────────────────────────────

  // CreateEndUserPermission
  graphql.mutation('CreateEndUserPermission', () => {
    return HttpResponse.json({
      data: {
        createEndUserPermission: {
          permission: createMockEndUserPermission(),
          error: null,
        },
      },
    })
  }),

  // DeleteEndUserPermission
  graphql.mutation('DeleteEndUserPermission', () => {
    return HttpResponse.json({
      data: {
        deleteEndUserPermission: {
          success: true,
          error: null,
        },
      },
    })
  }),

  // CreateEndUserBundle (formerly CreateEndUserBundle)
  graphql.mutation('CreateEndUserBundle', ({ variables }) => {
    const input = variables.input as { name?: string; description?: string }
    return HttpResponse.json({
      data: {
        createEndUserPermissionBundle: {
          bundle: createMockEndUserBundle({
            name: input?.name,
            description: input?.description,
          }),
          error: null,
        },
      },
    })
  }),

  // UpdateEndUserBundle (formerly UpdateEndUserBundle)
  graphql.mutation('UpdateEndUserBundle', ({ variables }) => {
    const input = variables.input as { name?: string; description?: string }
    return HttpResponse.json({
      data: {
        updateEndUserPermissionBundle: {
          bundle: createMockEndUserBundle({
            id: variables.id as string,
            name: input?.name,
            description: input?.description,
          }),
          error: null,
        },
      },
    })
  }),

  // DeleteEndUserBundle (formerly DeleteEndUserBundle)
  graphql.mutation('DeleteEndUserBundle', () => {
    return HttpResponse.json({
      data: {
        deleteEndUserPermissionBundle: {
          success: true,
          error: null,
        },
      },
    })
  }),

  // AddPermissionToBundle (formerly AddPermissionToBundle)
  graphql.mutation('AddPermissionToBundle', ({ variables }) => {
    return HttpResponse.json({
      data: {
        addEndUserPermissionToBundle: {
          bundle: createMockEndUserBundle({ id: variables.input.bundleId as string }),
          error: null,
        },
      },
    })
  }),

  // RemovePermissionFromBundle (formerly RemovePermissionFromBundle)
  graphql.mutation('RemovePermissionFromBundle', ({ variables }) => {
    return HttpResponse.json({
      data: {
        removeEndUserPermissionFromBundle: {
          bundle: createMockEndUserBundle({ id: variables.input.bundleId as string }),
          error: null,
        },
      },
    })
  }),

  // CreateEndUserRole
  graphql.mutation('CreateEndUserRole', ({ variables }) => {
    const input = variables.input as { name?: string; description?: string }
    return HttpResponse.json({
      data: {
        createEndUserRole: {
          role: createMockEndUserRole({
            name: input?.name,
            description: input?.description,
          }),
          error: null,
        },
      },
    })
  }),

  // DeleteEndUserRole
  graphql.mutation('DeleteEndUserRole', () => {
    return HttpResponse.json({
      data: {
        deleteEndUserRole: {
          success: true,
          error: null,
        },
      },
    })
  }),

  // AssignBundleToRole (formerly AssignBundleToRole)
  graphql.mutation('AssignBundleToRole', ({ variables }) => {
    return HttpResponse.json({
      data: {
        assignBundleToEndUserRole: {
          role: createMockEndUserRole({ id: variables.input.roleId as string }),
          error: null,
        },
      },
    })
  }),

  // RevokeBundleFromRole (formerly RevokeBundleFromRole)
  graphql.mutation('RevokeBundleFromRole', ({ variables }) => {
    return HttpResponse.json({
      data: {
        revokeBundleFromEndUserRole: {
          role: createMockEndUserRole({ id: variables.input.roleId as string }),
          error: null,
        },
      },
    })
  }),

  // AssignEndUserRoleToUser (formerly AssignEndUserRoleToUser)
  graphql.mutation('AssignEndUserRoleToUser', ({ variables }) => {
    return HttpResponse.json({
      data: {
        assignEndUserRole: {
          endUserId: variables.input.endUserId,
          role: createMockEndUserRole({ id: variables.input.roleId as string }),
          error: null,
        },
      },
    })
  }),

  // RevokeEndUserRoleFromUser (formerly RevokeEndUserRoleFromUser)
  graphql.mutation('RevokeEndUserRoleFromUser', () => {
    return HttpResponse.json({
      data: {
        revokeEndUserRole: {
          success: true,
          error: null,
        },
      },
    })
  }),

  // AssignBundleToEndUser
  graphql.mutation('AssignBundleToEndUser', ({ variables }) => {
    return HttpResponse.json({
      data: {
        assignBundleToEndUser: {
          endUserId: variables.input.endUserId,
          bundle: createMockEndUserBundle({ id: variables.input.bundleId as string }),
          error: null,
        },
      },
    })
  }),

  // RevokeBundleFromEndUser
  graphql.mutation('RevokeBundleFromEndUser', () => {
    return HttpResponse.json({
      data: {
        revokeBundleFromEndUser: {
          success: true,
          error: null,
        },
      },
    })
  }),
]
