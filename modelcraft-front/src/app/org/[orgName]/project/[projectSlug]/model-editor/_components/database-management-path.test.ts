import { describe, expect, it } from 'vitest'
import { buildDatabaseManagementPath } from './database-management-path'

describe('buildDatabaseManagementPath', () => {
  it('builds the project database management route', () => {
    expect(buildDatabaseManagementPath('acme', 'order-center')).toBe(
      '/org/acme/project/order-center/databases'
    )
  })
})
