// tests-bdd/step-definitions/lfk.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'

const ADD_FIELDS = `
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
      id
      name
    }
  }
`

const CREATE_LFK = `
  mutation CreateLogicalForeignKey($input: CreateLogicalForeignKeyInput!) {
    createLogicalForeignKey(input: $input) {
      result {
        __typename
        ... on LogicalForeignKey {
          id
          pairId
          sourceFields
          targetFields
        }
        ... on FKColumnsNotFoundError { message }
        ... on FKFieldCountMismatchError { message }
      }
    }
  }
`

Given('{string} 已有名为 {string} 格式为 {string} 的字段', async function (
  this: ModelCraftWorld, modelBaseName: string, fieldName: string, format: string
) {
  const modelId = this.modelMap[modelBaseName]
  if (!modelId) throw new Error(`模型 ${modelBaseName} 尚未创建，请先用 Given 创建`)

  await this.projectClient.mutate(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: fieldName, title: fieldName, format }],
  })
})

When('我创建从 {string} 到 {string} 的逻辑外键', async function (
  this: ModelCraftWorld, source: string, target: string
) {
  // source/target 格式: "ModelBaseName.fieldName"
  const [srcModel, srcField] = source.split('.')
  const [tgtModel, tgtField] = target.split('.')

  const srcId = this.modelMap[srcModel]
  const tgtId = this.modelMap[tgtModel]

  if (!srcId || !tgtId) {
    throw new Error(`模型 ID 未找到: ${srcModel}=${srcId}, ${tgtModel}=${tgtId}`)
  }

  const res = await this.projectClient.mutate<{
    createLogicalForeignKey: { result: { __typename: string; id?: string; pairId?: string } }
  }>(CREATE_LFK, {
    input: {
      modelId: srcId,
      refModelId: tgtId,
      sourceFields: [srcField],
      targetFields: [tgtField],
    },
  })
  this.lastResponse = { createLogicalForeignKey: res.createLogicalForeignKey }
})

Then('逻辑外键应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    createLogicalForeignKey: { result: { __typename: string } }
  }).createLogicalForeignKey
  expect(payload.result.__typename).toBe('LogicalForeignKey')
})

Then('应该返回 LFK 错误类型 {string}', function (this: ModelCraftWorld, expectedTypename: string) {
  const payload = (this.lastResponse as {
    createLogicalForeignKey: { result: { __typename: string } }
  }).createLogicalForeignKey
  expect(payload.result.__typename).toBe(expectedTypename)
})
