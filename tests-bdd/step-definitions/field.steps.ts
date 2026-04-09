// tests-bdd/step-definitions/field.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
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
        }
      }
      error {
        __typename
        ... on InvalidModelInput { message }
      }
    }
  }
`

const REMOVE_FIELD = `
  mutation RemoveField($modelID: ID!, $fieldName: String!) {
    removeField(modelID: $modelID, fieldName: $fieldName) {
      id
      fields { name }
    }
  }
`

// 当前 Scenario 的 model ID（从 world 的 createdModelIds 取最后一个）
function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID（请先用 Given 创建模型）')
  return id
}

Given('模型已有名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: { model: { id: string; fields: Array<{ name: string }> } | null } }>(
    ADD_FIELDS,
    { modelID: modelId, input: [{ name: fieldName, title: fieldName, format }] }
  )
  this.lastResponse = { addFields: res.addFields }
  if (!res.addFields?.model) throw new Error(`前置条件：添加字段 ${fieldName} 失败`)
})

When('我为该模型添加名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, fieldName: string, format: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: { model: { id: string; fields: Array<{ name: string }> } | null; error: { __typename: string; message?: string } | null } }>(
    ADD_FIELDS,
    { modelID: modelId, input: [{ name: fieldName, title: fieldName, format }] }
  )
  this.lastResponse = { addFields: res.addFields }
})

When('我删除名为 {string} 的字段', async function (this: ModelCraftWorld, fieldName: string) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ removeField: { id: string } | null }>(
    REMOVE_FIELD,
    { modelID: modelId, fieldName }
  )
  this.lastResponse = { removeField: res.removeField }
})

Then('字段应该添加成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { addFields: { model: { id: string } | null } }).addFields
  expect(payload.model).not.toBeNull()
})

Then('模型应该包含名为 {string} 的字段', function (this: ModelCraftWorld, fieldName: string) {
  const payload = (this.lastResponse as { addFields: { model: { fields: Array<{ name: string }> } | null } }).addFields
  const fieldNames = payload.model?.fields.map(f => f.name) ?? []
  expect(fieldNames).toContain(fieldName)
})

Then('字段应该删除成功', function (this: ModelCraftWorld) {
  // removeField 直接返回更新后的 Model 对象（无 error 字段）
  // graphql-request 在请求失败时会 throw，能到这一步说明成功
  const payload = (this.lastResponse as { removeField: { id: string } | null }).removeField
  expect(payload).not.toBeNull()
})
