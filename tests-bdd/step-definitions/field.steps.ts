// tests-bdd/step-definitions/field.steps.ts
import { Given, When, Then, DataTable } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'

const ADD_FIELDS = `
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
      model {
        id
        fields {
          name
          format
          enumName
          enumRelationId
        }
      }
      results {
        name
        success
        error {
          __typename
          ... on InvalidInput { message }
        }
      }
      error {
        __typename
        ... on InvalidInput { message }
      }
    }
  }
`

const REMOVE_FIELD = `
  mutation RemoveField($modelID: ID!, $fieldName: String!) {
    removeField(modelID: $modelID, fieldName: $fieldName) {
      model {
        id
        fields { name }
      }
      error {
        __typename
        ... on InvalidInput { message }
        ... on FieldReferenceInUse { code message }
      }
    }
  }
`

const CREATE_FIELD_ENUM_RELATION = `
  mutation CreateFieldEnumRelation($input: CreateFieldEnumRelationInput!) {
    createFieldEnumRelation(input: $input) {
      relation {
        id
        modelId
        sourceFieldName
        labelFieldName
        enumName
      }
      error {
        __typename
        ... on InvalidInput { message }
        ... on FieldEnumSourceConflict { code message }
      }
    }
  }
`

function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID（请先用 Given 创建模型）')
  return id
}

Given('模型已有名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: fieldName, title: fieldName, format }],
  })
  this.lastResponse = { addFields: res.addFields }
})

Given('模型已有名为 {string} 格式为 {string} 且关联最近创建枚举的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  if (!this.lastEnumName) throw new Error('没有最近创建的枚举，请先执行创建枚举步骤')

  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input: [{
      name: fieldName,
      title: fieldName,
      format,
      relateEnumName: this.lastEnumName,
    }],
  })
  this.lastResponse = { addFields: res.addFields }
})

Given('已创建字段枚举关联 source {string} label {string}', async function (
  this: ModelCraftWorld,
  sourceFieldName: string,
  labelFieldName: string
) {
  const modelId = getCurrentModelId(this)
  if (!this.lastEnumName) throw new Error('没有最近创建的枚举，请先执行创建枚举步骤')

  const res = await this.projectClient.mutate<{ createFieldEnumRelation: unknown }>(
    CREATE_FIELD_ENUM_RELATION,
    {
      input: {
        modelId,
        sourceFieldName,
        labelFieldName,
        enumName: this.lastEnumName,
      },
    }
  )
  this.lastResponse = { createFieldEnumRelation: res.createFieldEnumRelation }
})

When('我为该模型添加名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: fieldName, title: fieldName, format }],
  })
  this.lastResponse = { addFields: res.addFields }
})

When('我批量添加字段', async function (this: ModelCraftWorld, table: DataTable) {
  const modelId = getCurrentModelId(this)
  const input = table.hashes().map((row) => {
    const item: Record<string, unknown> = {
      name: row.name,
      title: row.title,
      format: row.format,
    }

    if (row.relateEnumName) {
      item.relateEnumName = row.relateEnumName === '@lastEnum' ? this.lastEnumName : row.relateEnumName
    }
    if (row.enumRelationId) {
      item.enumRelationId = row.enumRelationId
    }
    return item
  })

  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input,
  })
  this.lastResponse = { addFields: res.addFields }
})

When('我创建字段枚举关联 source {string} label {string}', async function (
  this: ModelCraftWorld,
  sourceFieldName: string,
  labelFieldName: string
) {
  const modelId = getCurrentModelId(this)
  if (!this.lastEnumName) throw new Error('没有最近创建的枚举，请先执行创建枚举步骤')

  const res = await this.projectClient.mutate<{ createFieldEnumRelation: unknown }>(
    CREATE_FIELD_ENUM_RELATION,
    {
      input: {
        modelId,
        sourceFieldName,
        labelFieldName,
        enumName: this.lastEnumName,
      },
    }
  )
  this.lastResponse = { createFieldEnumRelation: res.createFieldEnumRelation }
})

When('我删除名为 {string} 的字段', async function (this: ModelCraftWorld, fieldName: string) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ removeField: unknown }>(REMOVE_FIELD, {
    modelID: modelId,
    fieldName,
  })
  this.lastResponse = { removeField: res.removeField }
})

Then('字段应该添加成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { addFields: { error: unknown } }).addFields
  expect(payload.error).toBeNull()
})

Then('模型应该包含名为 {string} 的字段', function (this: ModelCraftWorld, fieldName: string) {
  const payload = (this.lastResponse as {
    addFields: { model: { fields: Array<{ name: string }> } | null }
  }).addFields
  const fieldNames = payload.model?.fields.map(f => f.name) ?? []
  expect(fieldNames).toContain(fieldName)
})

Then('字段应该删除成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    removeField: { model: { id: string } | null; error: unknown }
  }).removeField
  expect(payload.error).toBeNull()
  expect(payload.model).not.toBeNull()
})

Then('addFields 结果中字段 {string} 应该成功', function (this: ModelCraftWorld, fieldName: string) {
  const payload = (this.lastResponse as {
    addFields: { results: Array<{ name: string; success: boolean }> }
  }).addFields
  const item = payload.results.find(r => r.name === fieldName)
  expect(item).toBeDefined()
  expect(item?.success).toBe(true)
})

Then('addFields 结果中字段 {string} 应该失败并返回 {string}', function (
  this: ModelCraftWorld,
  fieldName: string,
  errorType: string
) {
  const payload = (this.lastResponse as {
    addFields: {
      results: Array<{ name: string; success: boolean; error: { __typename: string } | null }>
    }
  }).addFields
  const item = payload.results.find(r => r.name === fieldName)
  expect(item).toBeDefined()
  expect(item?.success).toBe(false)
  expect(item?.error?.__typename).toBe(errorType)
})

Then('错误码应该是 {string}', function (this: ModelCraftWorld, code: string) {
  const payload = Object.values(this.lastResponse!)[0] as { error?: { code?: string } | null }
  expect(payload.error).not.toBeNull()
  expect(payload.error?.code).toBe(code)
})
