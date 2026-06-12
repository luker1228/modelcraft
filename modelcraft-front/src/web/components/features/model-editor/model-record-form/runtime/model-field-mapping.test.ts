import { describe, expect, it } from 'vitest'
import {
  buildEditFormData,
  mapModelFieldsToRuntimeFields,
  mapModelFieldsToTableFieldInfos,
  type ModelField,
} from './model-field-mapping'

describe('mapModelFieldsToRuntimeFields', () => {
  it('maps fields and applies type fallback chain', () => {
    const fields: ModelField[] = [
      { name: 'age', schemaType: 'INTEGER', format: 'INT32', storageHint: 'NUMBER' },
      { name: 'nickname', type: 'string', format: null, schemaType: null, storageHint: null },
      { name: 'note' },
    ]

    expect(mapModelFieldsToRuntimeFields(fields)).toEqual([
      {
        name: 'age',
        type: 'INTEGER',
        format: 'INT32',
        schemaType: 'INTEGER',
        storageHint: 'NUMBER',
      },
      {
        name: 'nickname',
        type: 'string',
        format: undefined,
        schemaType: undefined,
        storageHint: undefined,
      },
      {
        name: 'note',
        type: 'string',
        format: undefined,
        schemaType: undefined,
        storageHint: undefined,
      },
    ])
  })
})

describe('mapModelFieldsToTableFieldInfos', () => {
  it('maps title/name related metadata and normalizes booleans', () => {
    const fields: ModelField[] = [
      {
        name: 'code',
        title: 'Code',
        isPrimary: true,
        storageHint: 'TEXT',
        schemaType: 'STRING',
        format: 'ENUM',
      },
      { name: 'desc', title: null, isPrimary: false, storageHint: null, schemaType: null },
    ]

    expect(mapModelFieldsToTableFieldInfos(fields)).toEqual([
      {
        name: 'code',
        title: 'Code',
        isPrimary: true,
        isDeprecated: false,
        format: 'ENUM',
        storageHint: 'TEXT',
        schemaType: 'STRING',
      },
      {
        name: 'desc',
        title: null,
        isPrimary: false,
        isDeprecated: false,
        format: null,
        storageHint: null,
        schemaType: null,
      },
    ])
  })
})

describe('buildEditFormData', () => {
  it('filters primary fields and converts nullish values to empty string', () => {
    const fields: ModelField[] = [
      { name: 'id', isPrimary: true },
      { name: 'name', isPrimary: false },
      { name: 'score' },
      { name: 'enabled' },
      { name: 'description' },
      { name: 'note' },
    ]

    const item = {
      id: 'p_1',
      name: null,
      score: 0,
      enabled: false,
      description: undefined,
      note: 'ok',
    }

    expect(buildEditFormData(fields, item)).toEqual({
      name: '',
      score: 0,
      enabled: false,
      description: '',
      note: 'ok',
    })
  })
})
