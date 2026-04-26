import { print, parse } from 'graphql'
import { describe, expect, it } from 'vitest'
import {
  buildCountQuery,
  buildCreateMutation,
  buildDeleteMutation,
  buildFieldSelections,
  buildFindFirstQuery,
  buildFindManyQuery,
  buildFindUniqueQuery,
  buildModelQueryOperations,
  buildUpdateMutation,
  extractFieldsFromSchema,
  extractWritableFieldNamesFromSchema,
  sanitizeMutationInputData,
  type FieldDefinition,
} from './runtime-query-builder'

// ============================================================================
// Helper: Parse and print to verify valid GraphQL DocumentNode
// ============================================================================

function assertValidDocument(doc: string): void {
  // Should not throw
  const parsed = parse(doc)
  expect(parsed).toBeDefined()
  // Should have at least one definition
  expect(parsed.definitions.length).toBeGreaterThan(0)
}

// ============================================================================
// 1. 可解析断言 - Each builder returns valid DocumentNode
// ============================================================================

describe('runtime-query-builder: DocumentNode validity', () => {
  const modelName = 'User'
  const fields = ['id', 'name', 'email']

  it('buildFindManyQuery returns valid DocumentNode', () => {
    const doc = buildFindManyQuery(modelName, fields)
    const printed = print(doc)
    expect(printed).toContain('findMany')
    assertValidDocument(printed)
  })

  it('buildFindUniqueQuery returns valid DocumentNode', () => {
    const doc = buildFindUniqueQuery(modelName, fields)
    const printed = print(doc)
    expect(printed).toContain('findUnique')
    assertValidDocument(printed)
  })

  it('buildFindFirstQuery returns valid DocumentNode', () => {
    const doc = buildFindFirstQuery(modelName, fields)
    const printed = print(doc)
    expect(printed).toContain('findFirst')
    assertValidDocument(printed)
  })

  it('buildCountQuery returns valid DocumentNode', () => {
    const doc = buildCountQuery(modelName)
    const printed = print(doc)
    expect(printed).toContain('count')
    assertValidDocument(printed)
  })

  it('buildCreateMutation returns valid DocumentNode', () => {
    const doc = buildCreateMutation(modelName)
    const printed = print(doc)
    expect(printed).toContain('create')
    assertValidDocument(printed)
  })

  it('buildUpdateMutation returns valid DocumentNode', () => {
    const doc = buildUpdateMutation(modelName)
    const printed = print(doc)
    expect(printed).toContain('update')
    assertValidDocument(printed)
  })

  it('buildDeleteMutation returns valid DocumentNode', () => {
    const doc = buildDeleteMutation(modelName)
    const printed = print(doc)
    expect(printed).toContain('delete')
    assertValidDocument(printed)
  })

  it('buildModelQueryOperations returns all operations', () => {
    const ops = buildModelQueryOperations(modelName, fields)
    expect(ops.findMany).toBeDefined()
    expect(ops.findUnique).toBeDefined()
    expect(ops.findFirst).toBeDefined()
    expect(ops.count).toBeDefined()

    // All should be valid DocumentNodes
    assertValidDocument(print(ops.findMany))
    assertValidDocument(print(ops.findUnique))
    assertValidDocument(print(ops.findFirst))
    assertValidDocument(print(ops.count))
  })
})

// ============================================================================
// 2. 字段选择结构测试 - Relation vs Scalar field selection
// ============================================================================

describe('buildFieldSelections: relation vs scalar fields', () => {
  it('scalar fields remain as strings', () => {
    const fields: string[] = ['id', 'name', 'email']
    const selections = buildFieldSelections(fields)
    expect(selections).toEqual(['id', 'name', 'email'])
  })

  it('scalar FieldDefinition remains as string', () => {
    const fields: FieldDefinition[] = [
      { name: 'id', type: 'ID' },
      { name: 'name', type: 'String' },
      { name: 'age', type: 'Int', format: 'INT32' },
    ]
    const selections = buildFieldSelections(fields)
    expect(selections).toEqual(['id', 'name', 'age'])
  })

  it('RELATION fields generate nested selection with id and _displayName', () => {
    const fields: FieldDefinition[] = [
      { name: 'id', type: 'ID' },
      { name: 'name', type: 'String' },
      { name: 'owner', type: 'User', format: 'RELATION' },
    ]
    const selections = buildFieldSelections(fields)

    expect(selections).toEqual([
      'id',
      'name',
      { owner: ['id', '_displayName'] },
    ])
  })

  it('mixed scalar and RELATION fields', () => {
    const fields: (string | FieldDefinition)[] = [
      'id',
      { name: 'title', type: 'String' },
      { name: 'createdBy', type: 'User', format: 'RELATION' },
      { name: 'category', type: 'Category', format: 'RELATION' },
    ]
    const selections = buildFieldSelections(fields)

    expect(selections).toEqual([
      'id',
      'title',
      { createdBy: ['id', '_displayName'] },
      { category: ['id', '_displayName'] },
    ])
  })

  it('empty fields defaults to id', () => {
    const selections = buildFieldSelections([])
    expect(selections).toEqual(['id'])
  })
})

// ============================================================================
// 3. 边界输入测试
// ============================================================================

describe('runtime-query-builder: edge cases', () => {
  describe('empty fields handling', () => {
    it('buildFindManyQuery with empty fields defaults to id', () => {
      const doc = buildFindManyQuery('User', [])
      const printed = print(doc)
      expect(printed).toContain('id')
      // Should NOT have name or email since not provided
      expect(printed).not.toContain('name')
      expect(printed).not.toContain('email')
    })

    it('buildFindFirstQuery with empty fields defaults to id', () => {
      const doc = buildFindFirstQuery('User', [])
      const printed = print(doc)
      expect(printed).toContain('id')
    })
  })

  describe('only id field', () => {
    it('buildFindUniqueQuery with only id', () => {
      const doc = buildFindUniqueQuery('User', ['id'])
      const printed = print(doc)
      expect(printed).toContain('id')
      assertValidDocument(printed)
    })
  })

  describe('mixed string and FieldDefinition input', () => {
    it('handles mixed input array', () => {
      const mixed: (string | FieldDefinition)[] = [
        'id',
        { name: 'name', type: 'String' },
        'email',
        { name: 'profile', type: 'Profile', format: 'RELATION' },
      ]
      const doc = buildFindManyQuery('User', mixed)
      const printed = print(doc)

      expect(printed).toContain('id')
      expect(printed).toContain('name')
      expect(printed).toContain('email')
      expect(printed).toContain('profile')
      assertValidDocument(printed)
    })
  })

  describe('variable types are correctly defined', () => {
    it('findMany has where, orderBy, skip, take variables', () => {
      const doc = buildFindManyQuery('User', ['id'])
      const printed = print(doc)

      expect(printed).toContain('$where')
      expect(printed).toContain('$orderBy')
      expect(printed).toContain('$skip')
      expect(printed).toContain('$take')
      expect(printed).toContain('UserWhereInput')
      expect(printed).toContain('UserOrderByInput')
    })

    it('findUnique has required where variable', () => {
      const doc = buildFindUniqueQuery('User', ['id'])
      const printed = print(doc)

      expect(printed).toContain('$where')
      expect(printed).toContain('UserUniqueWhereInput')
    })

    it('create mutation has required data variable', () => {
      const doc = buildCreateMutation('User')
      const printed = print(doc)

      expect(printed).toContain('$data')
      expect(printed).toContain('UserCreateInput')
    })

    it('update mutation has where and data variables', () => {
      const doc = buildUpdateMutation('User')
      const printed = print(doc)

      expect(printed).toContain('$where')
      expect(printed).toContain('$data')
      expect(printed).toContain('UserUniqueWhereInput')
      expect(printed).toContain('UserUpdateInput')
    })

    it('delete mutation has where variable', () => {
      const doc = buildDeleteMutation('User')
      const printed = print(doc)

      expect(printed).toContain('$where')
      expect(printed).toContain('UserUniqueWhereInput')
    })
  })

  describe('response structure', () => {
    it('findMany returns timeCost, reqId, items', () => {
      const doc = buildFindManyQuery('User', ['id'])
      const printed = print(doc)

      expect(printed).toContain('timeCost')
      expect(printed).toContain('reqId')
      expect(printed).toContain('items')
    })

    it('findUnique returns timeCost, reqId, item', () => {
      const doc = buildFindUniqueQuery('User', ['id'])
      const printed = print(doc)

      expect(printed).toContain('timeCost')
      expect(printed).toContain('reqId')
      expect(printed).toContain('item')
    })

    it('count returns count field', () => {
      const doc = buildCountQuery('User')
      const printed = print(doc)

      expect(printed).toContain('count')
    })

    it('create returns id', () => {
      const doc = buildCreateMutation('User')
      const printed = print(doc)

      // Should have id field in selection
      expect(printed).toContain('id')
    })

    it('update returns success', () => {
      const doc = buildUpdateMutation('User')
      const printed = print(doc)

      expect(printed).toContain('success')
    })

    it('delete returns success', () => {
      const doc = buildDeleteMutation('User')
      const printed = print(doc)

      expect(printed).toContain('success')
    })
  })
})

// ============================================================================
// 4. Helper 函数测试
// ============================================================================

describe('extractFieldsFromSchema', () => {
  it('extracts field names from schema properties', () => {
    const schema = {
      properties: {
        id: {},
        name: {},
        email: {},
      },
    }
    const fields = extractFieldsFromSchema(schema)
    expect(fields).toEqual(['id', 'name', 'email'])
  })

  it('always includes id if not present', () => {
    const schema = {
      properties: {
        name: {},
        email: {},
      },
    }
    const fields = extractFieldsFromSchema(schema)
    expect(fields).toEqual(['id', 'name', 'email'])
  })

  it('handles null/undefined schema', () => {
    expect(extractFieldsFromSchema(null)).toEqual(['id'])
    expect(extractFieldsFromSchema(undefined)).toEqual(['id'])
    expect(extractFieldsFromSchema({})).toEqual(['id'])
    expect(extractFieldsFromSchema({ properties: undefined })).toEqual(['id'])
  })
})

describe('extractWritableFieldNamesFromSchema', () => {
  it('filters out readOnly fields', () => {
    const schema = {
      properties: {
        id: { readOnly: true },
        name: {},
        email: { readOnly: true },
        age: {},
      },
    }
    const fields = extractWritableFieldNamesFromSchema(schema)
    expect(fields).toEqual(['name', 'age'])
  })

  it('handles null/undefined schema', () => {
    expect(extractWritableFieldNamesFromSchema(null)).toEqual([])
    expect(extractWritableFieldNamesFromSchema(undefined)).toEqual([])
    expect(extractWritableFieldNamesFromSchema({})).toEqual([])
  })
})

describe('sanitizeMutationInputData', () => {
  it('filters out disallowed fields', () => {
    const data = {
      id: '123',
      name: 'John',
      email: 'john@example.com',
      unknownField: 'should be removed',
    }
    const allowed = ['id', 'name', 'email'] as const

    const result = sanitizeMutationInputData(data, allowed)
    expect(result).toEqual({
      id: '123',
      name: 'John',
      email: 'john@example.com',
    })
  })

  it('preserves explicit null values', () => {
    const data = {
      name: 'John',
      email: null,
    }
    const allowed = ['name', 'email'] as const

    const result = sanitizeMutationInputData(data, allowed)
    expect(result).toEqual({
      name: 'John',
      email: null,
    })
  })

  it('removes undefined values', () => {
    const data = {
      name: 'John',
      email: undefined,
    }
    const allowed = ['name', 'email'] as const

    const result = sanitizeMutationInputData(data, allowed)
    expect(result).toEqual({
      name: 'John',
    })
  })

  it('handles empty allowedFieldNames', () => {
    const data = { name: 'John' }
    const result = sanitizeMutationInputData(data, [])
    expect(result).toEqual({})
  })

  it('handles null/undefined data', () => {
    expect(sanitizeMutationInputData(null, ['name'])).toEqual({})
    expect(sanitizeMutationInputData(undefined, ['name'])).toEqual({})
  })
})

// ============================================================================
// 5. 快照测试 - Key query/mutation strings
// ============================================================================

describe('runtime-query-builder: snapshots', () => {
  it('buildFindManyQuery with relation field snapshot', () => {
    const fields: FieldDefinition[] = [
      { name: 'id', type: 'ID' },
      { name: 'title', type: 'String' },
      { name: 'owner', type: 'User', format: 'RELATION' },
    ]
    const doc = buildFindManyQuery('Task', fields)
    const printed = print(doc)

    // Snapshot: verify key structure
    expect(printed).toContain('findMany')
    expect(printed).toContain('$where: TaskWhereInput')
    expect(printed).toContain('$orderBy: [TaskOrderByInput!]')
    expect(printed).toContain('$skip: Int')
    expect(printed).toContain('$take: Int')
    expect(printed).toContain('timeCost')
    expect(printed).toContain('reqId')
    expect(printed).toContain('items')
    // Relation field selection
    expect(printed).toContain('owner {')
    expect(printed).toContain('id')
    expect(printed).toContain('_displayName')
  })

  it('buildFindUniqueQuery snapshot', () => {
    const doc = buildFindUniqueQuery('User', ['id', 'name', 'email'])
    const printed = print(doc)

    expect(printed).toContain('findUnique')
    expect(printed).toContain('$where: UserUniqueWhereInput!')
    expect(printed).toContain('item {')
  })

  it('buildFindFirstQuery snapshot', () => {
    const doc = buildFindFirstQuery('Post', ['id', 'title'])
    const printed = print(doc)

    expect(printed).toContain('findFirst')
    expect(printed).toContain('$where: PostWhereInput')
  })

  it('buildCountQuery snapshot', () => {
    const doc = buildCountQuery('Order')
    const printed = print(doc)

    expect(printed).toContain('count(')
    expect(printed).toContain('$where: OrderWhereInput')
    expect(printed).toContain('count')
  })

  it('buildCreateMutation snapshot', () => {
    const doc = buildCreateMutation('Article')
    const printed = print(doc)

    expect(printed).toContain('create(')
    expect(printed).toContain('$data: ArticleCreateInput!')
    expect(printed).toContain('id')
  })

  it('buildUpdateMutation snapshot', () => {
    const doc = buildUpdateMutation('Article')
    const printed = print(doc)

    expect(printed).toContain('update(')
    expect(printed).toContain('$where: ArticleUniqueWhereInput!')
    expect(printed).toContain('$data: ArticleUpdateInput!')
    expect(printed).toContain('success')
  })

  it('buildDeleteMutation snapshot', () => {
    const doc = buildDeleteMutation('Article')
    const printed = print(doc)

    expect(printed).toContain('delete(')
    expect(printed).toContain('$where: ArticleUniqueWhereInput!')
    expect(printed).toContain('success')
  })
})
