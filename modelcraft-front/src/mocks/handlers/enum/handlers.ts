/**
 * Enum 模块 MSW mock handlers
 *
 * 页面 key: 'enum-list', 'enum-detail'
 *
 * 启用方式：
 *   NEXT_PUBLIC_MOCK_PAGES=enum-list,enum-detail
 */

import { graphql, HttpResponse } from 'msw'
import { faker } from '@faker-js/faker'

function createMockEnumOption(override: Record<string, unknown> = {}) {
  return {
    code: faker.helpers.slugify(faker.word.noun()).toUpperCase(),
    label: faker.commerce.productAdjective(),
    order: faker.number.int({ min: 0, max: 100 }),
    description: null,
    ...override,
  }
}

function createMockEnum(override: Record<string, unknown> = {}) {
  return {
    id: faker.string.uuid(),
    projectSlug: 'default',
    name: faker.helpers.slugify(faker.word.noun()).toLowerCase() + '_enum',
    displayName: faker.commerce.department(),
    description: faker.lorem.sentence(),
    isMultiSelect: false,
    options: Array.from({ length: 3 }, (_, i) => createMockEnumOption({ order: i })),
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

export const enumHandlers = [
  // GetEnums
  graphql.query('GetEnums', () => {
    const enums = Array.from({ length: 4 }, () => createMockEnum())
    return HttpResponse.json({
      data: { enums },
    })
  }),

  // GetEnum
  graphql.query('GetEnum', ({ variables }) => {
    return HttpResponse.json({
      data: {
        enum: {
          enum: createMockEnum({ name: variables.name as string }),
          error: null,
        },
      },
    })
  }),

  // GetEnumReferences
  graphql.query('GetEnumReferences', () => {
    return HttpResponse.json({
      data: { enumReferences: [] },
    })
  }),

  // CreateEnum
  graphql.mutation('CreateEnum', ({ variables }) => {
    const input = variables.input as { name?: string; displayName?: string }
    return HttpResponse.json({
      data: {
        createEnum: {
          enum: createMockEnum({ name: input?.name, displayName: input?.displayName }),
          error: null,
        },
      },
    })
  }),

  // UpdateEnum
  graphql.mutation('UpdateEnum', ({ variables }) => {
    return HttpResponse.json({
      data: {
        updateEnum: {
          enum: createMockEnum({ name: variables.name as string }),
          error: null,
        },
      },
    })
  }),

  // DeleteEnum
  graphql.mutation('DeleteEnum', () => {
    return HttpResponse.json({
      data: {
        deleteEnum: { success: true, error: null },
      },
    })
  }),
]
