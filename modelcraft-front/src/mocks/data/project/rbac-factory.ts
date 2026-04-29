// Mock 数据工厂 - EndUser RBAC
// 用于 MSW handlers 提供测试数据

import { faker } from '@faker-js/faker'
import type {
  EndUserPermission,
  EndUserPermissionBundle,
  EndUserRole,
  ColumnPolicy,
  EndUserPermissionAction,
  EndUserRowScope,
} from '@/types'

// ── 常量 ───────────────────────────────────────────────────────────────────────

const ACTIONS: EndUserPermissionAction[] = ['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'EXPORT']
const ROW_SCOPES: EndUserRowScope[] = ['ALL', 'SELF', 'DEPT', 'DEPT_AND_CHILDREN']

// ── ColumnPolicy 工厂 ─────────────────────────────────────────────────────────

export function createMockColumnPolicy(override: Partial<ColumnPolicy> = {}): ColumnPolicy {
  return {
    defaultMode: 'VISIBLE',
    rules: [],
    ...override,
  }
}

// ── EndUserPermission 工厂 ────────────────────────────────────────────────────

export function createMockEndUserPermission(
  override: Partial<EndUserPermission> = {}
): EndUserPermission {
  const action = faker.helpers.arrayElement(ACTIONS)

  return {
    id: faker.string.uuid(),
    modelId: faker.string.uuid(),
    action,
    rowScope: faker.helpers.arrayElement(ROW_SCOPES),
    columnPolicy: createMockColumnPolicy(),
    displayName: `${action} 权限`,
    description: faker.lorem.sentence(),
    createdAt: faker.date.recent({ days: 30 }).toISOString(),
    updatedAt: faker.date.recent({ days: 7 }).toISOString(),
    ...override,
  }
}

// ── EndUserPermissionBundle 工厂 ──────────────────────────────────────────────

export function createMockEndUserBundle(
  override: Partial<EndUserPermissionBundle> = {}
): EndUserPermissionBundle {
  return {
    id: faker.string.uuid(),
    name: `${faker.commerce.department()} 权限包`,
    description: faker.lorem.sentence(),
    permissions: Array.from(
      { length: faker.number.int({ min: 1, max: 3 }) },
      () => createMockEndUserPermission()
    ),
    currentVersion: 0,
    snapshots: [],
    createdAt: faker.date.recent({ days: 30 }).toISOString(),
    updatedAt: faker.date.recent({ days: 7 }).toISOString(),
    ...override,
  }
}

// ── 内置隐式角色（固定数据，不可删除）────────────────────────────────────────

export const MOCK_IMPLICIT_ROLE: EndUserRole = {
  id: 'implicit-authenticated-user',
  name: 'SYSTEM_AUTHENTICATED_USER',
  description: '所有已登录终端用户自动获得的基础权限，系统内置，不可删除',
  isImplicit: true,
  permissionBundles: [],
  createdAt: new Date('2024-01-01T00:00:00Z').toISOString(),
  updatedAt: new Date('2024-01-01T00:00:00Z').toISOString(),
}

// ── EndUserRole 工厂 ──────────────────────────────────────────────────────────

export function createMockEndUserRole(
  override: Partial<EndUserRole> = {}
): EndUserRole {
  return {
    id: faker.string.uuid(),
    name: `${faker.person.jobTitle()} 角色`,
    description: faker.lorem.sentence(),
    isImplicit: false,
    permissionBundles: [],
    createdAt: faker.date.recent({ days: 30 }).toISOString(),
    updatedAt: faker.date.recent({ days: 7 }).toISOString(),
    ...override,
  }
}
