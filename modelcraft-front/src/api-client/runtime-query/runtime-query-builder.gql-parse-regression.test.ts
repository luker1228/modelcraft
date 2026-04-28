import { gql } from '@apollo/client'
import { describe, expect, it, vi } from 'vitest'

vi.mock('@apollo/client', async () => {
  const actual = await vi.importActual<typeof import('@apollo/client')>(
    '@apollo/client'
  )

  return {
    ...actual,
    gql: vi.fn((...args: unknown[]) => {
      const firstArg = args[0]
      const isTemplateLiteralCall =
        Array.isArray(firstArg) &&
        Object.prototype.hasOwnProperty.call(firstArg, 'raw')

      if (isTemplateLiteralCall) {
        throw new Error(
          'Regression: runtime query builder must call gql(query), not gql`${query}`.'
        )
      }

      if (args.length !== 1 || typeof firstArg !== 'string') {
        throw new Error(
          `Regression: gql expects exactly one string argument, got ${args.length}.`
        )
      }

      return actual.gql(firstArg)
    }),
  }
})

import {
  buildCountQuery,
  buildCreateMutation,
  buildDeleteMutation,
  buildFindFirstQuery,
  buildFindManyQuery,
  buildFindUniqueQuery,
  buildUpdateMutation,
} from './runtime-query-builder'

describe('runtime-query-builder: gql(query) regression', () => {
  it('all runtime builders pass plain query string into gql()', () => {
    buildFindManyQuery('User', ['id', 'name'])
    buildFindUniqueQuery('User', ['id'])
    buildFindFirstQuery('User', ['id'])
    buildCountQuery('User')
    buildCreateMutation('User')
    buildUpdateMutation('User')
    buildDeleteMutation('User')

    const gqlMock = vi.mocked(gql)
    expect(gqlMock).toHaveBeenCalled()

    for (const call of gqlMock.mock.calls) {
      expect(call).toHaveLength(1)
      expect(typeof call[0]).toBe('string')
    }
  })
})
