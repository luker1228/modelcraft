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
  // GetEndUserPermissions — 返回 3-5 个权限点
  graphql.query('GetEndUserPermissions', () => {
    const permissions = Array.from(
      { length: 4 },
      () => createMockEndUserPermission()
    )
    return HttpResponse.json({
      data: { endUserPermissions: permissions },
    })
  }),

  // GetEndUserBundles — 返回 3 个权限包
  graphql.query('GetEndUserBundles', () => {
    const bundles = Array.from(
      { length: 3 },
      () => createMockEndUserBundle()
    )
    return HttpResponse.json({
      data: { endUserBundles: bundles },
    })
  }),

  // GetEndUserBundle — 返回单个权限包详情
  graphql.query('GetEndUserBundle', ({ variables }) => {
    const bundle = createMockEndUserBundle({ id: variables.id as string })
    return HttpResponse.json({
      data: { endUserBundle: bundle },
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
      data: { endUserRoles: roles },
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

  // GetEndUserEffectivePermissions — 返回空 grants（无权限时的初始状态）
  graphql.query('GetEndUserEffectivePermissions', ({ variables }) => {
    return HttpResponse.json({
      data: {
        endUserEffectivePermissions: {
          endUserId: variables.endUserId,
          modelId: '',
          grants: [],
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

  // CreateEndUserBundle
  graphql.mutation('CreateEndUserBundle', ({ variables }) => {
    const input = variables.input as { name?: string; description?: string }
    return HttpResponse.json({
      data: {
        createEndUserBundle: {
          bundle: createMockEndUserBundle({
            name: input?.name,
            description: input?.description,
          }),
          error: null,
        },
      },
    })
  }),

  // UpdateEndUserBundle
  graphql.mutation('UpdateEndUserBundle', ({ variables }) => {
    const input = variables.input as { name?: string; description?: string }
    return HttpResponse.json({
      data: {
        updateEndUserBundle: {
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

  // DeleteEndUserBundle
  graphql.mutation('DeleteEndUserBundle', () => {
    return HttpResponse.json({
      data: {
        deleteEndUserBundle: {
          success: true,
          error: null,
        },
      },
    })
  }),

  // AddPermissionToBundle
  graphql.mutation('AddPermissionToBundle', ({ variables }) => {
    return HttpResponse.json({
      data: {
        addPermissionToBundle: {
          bundle: createMockEndUserBundle({ id: variables.bundleId as string }),
          error: null,
        },
      },
    })
  }),

  // RemovePermissionFromBundle
  graphql.mutation('RemovePermissionFromBundle', ({ variables }) => {
    return HttpResponse.json({
      data: {
        removePermissionFromBundle: {
          bundle: createMockEndUserBundle({ id: variables.bundleId as string }),
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

  // AssignBundleToRole
  graphql.mutation('AssignBundleToRole', ({ variables }) => {
    return HttpResponse.json({
      data: {
        assignBundleToRole: {
          role: createMockEndUserRole({ id: variables.roleId as string }),
          error: null,
        },
      },
    })
  }),

  // RevokeBundleFromRole
  graphql.mutation('RevokeBundleFromRole', ({ variables }) => {
    return HttpResponse.json({
      data: {
        revokeBundleFromRole: {
          role: createMockEndUserRole({ id: variables.roleId as string }),
          error: null,
        },
      },
    })
  }),

  // AssignEndUserRoleToUser
  graphql.mutation('AssignEndUserRoleToUser', ({ variables }) => {
    return HttpResponse.json({
      data: {
        assignEndUserRoleToUser: {
          assignment: {
            endUserId: variables.endUserId,
            role: createMockEndUserRole({ id: variables.roleId as string }),
            assignedAt: new Date().toISOString(),
          },
          error: null,
        },
      },
    })
  }),

  // RevokeEndUserRoleFromUser
  graphql.mutation('RevokeEndUserRoleFromUser', () => {
    return HttpResponse.json({
      data: {
        revokeEndUserRoleFromUser: {
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
          assignment: {
            endUserId: variables.endUserId,
            bundle: createMockEndUserBundle({ id: variables.bundleId as string }),
            grantedAt: new Date().toISOString(),
          },
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
