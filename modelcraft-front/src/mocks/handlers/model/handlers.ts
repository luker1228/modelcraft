/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
/**
 * Model 模块 MSW mock handlers
 *
 * 页面 key: 'model-editor'
 *
 * 启用方式：
 *   NEXT_PUBLIC_MOCK_PAGES=model-editor
 */

import { graphql, HttpResponse } from 'msw'
import { faker } from '@faker-js/faker'

function createMockField(override: Record<string, unknown> = {}) {
  return {
    name: faker.helpers.slugify(faker.word.noun()).toLowerCase(),
    title: faker.commerce.productName(),
    format: faker.helpers.arrayElement(['STRING', 'INTEGER', 'BOOLEAN', 'DATETIME']),
    schemaType: 'string',
    storageHint: null,
    nonNull: false,
    required: false,
    isPrimary: false,
    isUnique: false,
    isDeprecated: false,
    isArray: false,
    description: null,
    relateFkId: null,
    belongsToFkId: null,
    enum: null,
    validationConfig: null,
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

function createMockGroup(override: Record<string, unknown> = {}) {
  return {
    id: faker.string.uuid(),
    name: faker.word.noun(),
    isVirtual: false,
    displayOrder: 0,
    ...override,
  }
}

function createMockModel(override: Record<string, unknown> = {}) {
  return {
    id: faker.string.uuid(),
    projectSlug: 'default',
    name: faker.helpers.slugify(faker.word.noun()).toLowerCase(),
    title: faker.commerce.productName(),
    description: faker.lorem.sentence(),
    displayField: null,
    databaseName: 'mc_default',
    storageType: 'TABLE',
    jsonSchema: null,
    dbTable: faker.helpers.slugify(faker.word.noun()).toLowerCase(),
    fields: [
      createMockField({ name: 'id', title: 'ID', format: 'STRING', isPrimary: true }),
      createMockField({ name: 'name', title: '名称', format: 'STRING' }),
    ],
    group: createMockGroup(),
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

export const modelHandlers = [
  // GetModels
  graphql.query('GetModels', () => {
    const models = Array.from({ length: 5 }, () => createMockModel())
    return HttpResponse.json({
      data: {
        models: {
          edges: models.map((node) => ({ node, cursor: node.id })),
          pageInfo: { hasNextPage: false, hasPreviousPage: false, startCursor: null, endCursor: null },
          totalCount: models.length,
        },
      },
    })
  }),

  // GetModelsForRelation
  graphql.query('GetModelsForRelation', () => {
    const models = Array.from({ length: 3 }, () => ({
      id: faker.string.uuid(),
      name: faker.helpers.slugify(faker.word.noun()).toLowerCase(),
      title: faker.commerce.productName(),
      databaseName: 'mc_default',
    }))
    return HttpResponse.json({
      data: {
        models: {
          edges: models.map((node) => ({ node, cursor: node.id })),
        },
      },
    })
  }),

  // GetModel
  graphql.query('GetModel', ({ variables }) => {
    return HttpResponse.json({
      data: {
        model: {
          model: createMockModel({ id: variables.id as string }),
          error: null,
        },
      },
    })
  }),

  // GetModelRecordWorkspace
  graphql.query('GetModelRecordWorkspace', ({ variables }) => {
    return HttpResponse.json({
      data: {
        model: {
          model: {
            id: variables.id,
            name: 'mock_model',
            title: 'Mock Model',
            description: null,
            databaseName: 'mc_default',
            jsonSchema: null,
            fields: [
              { name: 'id', isDeprecated: false },
              { name: 'name', isDeprecated: false },
            ],
          },
          error: null,
        },
      },
    })
  }),

  // GetModelGroups
  graphql.query('GetModelGroups', () => {
    return HttpResponse.json({
      data: {
        modelGroups: [
          {
            ...createMockGroup({ name: 'default', isVirtual: true }),
            models: Array.from({ length: 3 }, () => createMockModel()),
          },
        ],
      },
    })
  }),

  // GetLogicalForeignKeys
  graphql.query('GetLogicalForeignKeys', () => {
    return HttpResponse.json({
      data: { logicalForeignKeys: [] },
    })
  }),

  // CreateModel
  graphql.mutation('CreateModel', ({ variables }) => {
    const input = variables.input as { name?: string; title?: string }
    return HttpResponse.json({
      data: {
        createModel: {
          model: createMockModel({ name: input?.name, title: input?.title }),
          error: null,
        },
      },
    })
  }),

  // UpdateModelMeta
  graphql.mutation('UpdateModelMeta', ({ variables }) => {
    return HttpResponse.json({
      data: {
        updateModelMeta: {
          success: true,
          model: createMockModel({ id: variables.id as string }),
          error: null,
        },
      },
    })
  }),

  // DeleteModel
  graphql.mutation('DeleteModel', () => {
    return HttpResponse.json({
      data: {
        deleteModel: { success: true, error: null },
      },
    })
  }),

  // AddFields
  graphql.mutation('AddFields', ({ variables }) => {
    return HttpResponse.json({
      data: {
        addFields: {
          model: createMockModel({ id: variables.modelID as string }),
          error: null,
        },
      },
    })
  }),

  // UpdateField
  graphql.mutation('UpdateField', ({ variables }) => {
    return HttpResponse.json({
      data: {
        updateField: {
          model: createMockModel({ id: variables.modelID as string }),
          error: null,
        },
      },
    })
  }),

  // RemoveField
  graphql.mutation('RemoveField', ({ variables }) => {
    return HttpResponse.json({
      data: {
        removeField: {
          model: createMockModel({ id: variables.modelID as string }),
          error: null,
        },
      },
    })
  }),

  // CreateGroup
  graphql.mutation('CreateGroup', ({ variables }) => {
    const input = variables.input as { name?: string }
    return HttpResponse.json({
      data: {
        createGroup: {
          group: { ...createMockGroup({ name: input?.name }), models: [] },
          error: null,
        },
      },
    })
  }),

  // MoveModelToGroup
  graphql.mutation('MoveModelToGroup', () => {
    return HttpResponse.json({
      data: {
        moveModelToGroup: { success: true, error: null },
      },
    })
  }),
]
