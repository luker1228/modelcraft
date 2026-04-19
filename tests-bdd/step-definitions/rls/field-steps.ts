// RLS Field Step Definitions
// 对应 Feature: EndUserRef 字段类型

import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../../support/world'

const GET_MODEL_WITH_FIELDS = `
  query GetModelWithFields($id: ID!) {
    model(id: $id) {
      model {
        id
        name
        fields {
          name
          format
        }
        rlsPolicy {
          modelId
          selectPredicate
          insertCheck
          updatePredicate
          updateCheck
          deletePredicate
          preset
        }
      }
      error {
        __typename
      }
    }
  }
`

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
        rlsPolicy {
          modelId
        }
      }
      error {
        __typename
      }
    }
  }
`

function getCurrentModelId(world: ModelCraftWorld): string {
  const id = world.createdModelIds[world.createdModelIds.length - 1]
  if (!id) throw new Error('没有可用的模型 ID（请先用 Given 创建模型）')
  return id
}

// Given 步骤

Given('模型已有名为 {string} 的 EndUserRef 字段', async function (
  this: ModelCraftWorld, fieldName: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: fieldName, title: fieldName, format: 'END_USER_REF' }],
  })
  this.lastResponse = { addFields: res.addFields }
})

// When 步骤 - RLS 特有

When('我重新为模型添加 EndUserRef 字段', async function (this: ModelCraftWorld) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.mutate<{ addFields: unknown }>(ADD_FIELDS, {
    modelID: modelId,
    input: [{ name: 'owner', title: 'owner', format: 'END_USER_REF' }],
  })
  this.lastResponse = { addFields: res.addFields }
})

When('我查询该模型的 RLS 策略', async function (this: ModelCraftWorld) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: {
      model: {
        rlsPolicy: {
          modelId: string
          selectPredicate: string
          insertCheck: string
          updatePredicate: string
          updateCheck: string
          deletePredicate: string
          preset: string | null
        } | null
      } | null
      error: unknown
    }
  }>(GET_MODEL_WITH_FIELDS, { id: modelId })
  this.lastResponse = { model: res.model }
})

// Then 步骤

Then('模型应该包含名为 {string} 的 EndUserRef 字段', async function (
  this: ModelCraftWorld, fieldName: string
) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: {
      model: {
        fields: Array<{ name: string; format: string }>
      } | null
    }
  }>(GET_MODEL_WITH_FIELDS, { id: modelId })

  const fields = res.model.model?.fields ?? []
  const ownerField = fields.find(f => f.name === fieldName)
  expect(ownerField).toBeDefined()
  expect(ownerField?.format).toBe('END_USER_REF')
})

Then('应该返回默认的 READ_WRITE_OWNER 策略', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    model: { model: { rlsPolicy: { preset: string } | null } | null }
  }).model
  expect(payload.model).not.toBeNull()
  expect(payload.model?.rlsPolicy).not.toBeNull()
  expect(payload.model?.rlsPolicy?.preset).toBe('READ_WRITE_OWNER')
})

Then('应该返回 null（无策略）', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as {
    model: { model: { rlsPolicy: unknown } | null }
  }).model
  expect(payload.model?.rlsPolicy).toBeNull()
})

Then('该模型应该存在默认的 READ_WRITE_OWNER 策略', async function (this: ModelCraftWorld) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: {
      model: {
        rlsPolicy: { preset: string } | null
      } | null
    }
  }>(GET_MODEL_WITH_FIELDS, { id: modelId })

  expect(res.model.model?.rlsPolicy).not.toBeNull()
  expect(res.model.model?.rlsPolicy?.preset).toBe('READ_WRITE_OWNER')
})

Then('该模型应该不存在 RLS 策略', async function (this: ModelCraftWorld) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: {
      model: {
        rlsPolicy: unknown | null
      } | null
    }
  }>(GET_MODEL_WITH_FIELDS, { id: modelId })

  expect(res.model.model?.rlsPolicy).toBeNull()
})

Then('该模型不应该包含 {string} 字段', async function (this: ModelCraftWorld, fieldName: string) {
  const modelId = getCurrentModelId(this)
  const res = await this.projectClient.query<{
    model: {
      model: {
        fields: Array<{ name: string }>
      } | null
    }
  }>(GET_MODEL_WITH_FIELDS, { id: modelId })

  const fields = res.model.model?.fields ?? []
  const field = fields.find(f => f.name === fieldName)
  expect(field).toBeUndefined()
})

// 导入模型相关步骤（待实现）

Given('从数据库表导入名为 {string} 的模型', async function (
  this: ModelCraftWorld, baseName: string
) {
  // 导入模型功能需要单独实现
  // 目前标记为待实现，使用普通创建并手动删除 owner 字段来模拟
  const name = `${baseName}_IMPORTED`
  const res = await this.projectClient.mutate<{
    createModel: { model: { id: string; name: string } | null; error: unknown }
  }>(`
    mutation CreateModel($input: CreateModelInput!) {
      createModel(input: $input) {
        model { id name }
        error { __typename }
      }
    }
  `, {
    input: { name, title: name, databaseName: 'test_db' },
  })

  if (!res.createModel.model?.id) {
    throw new Error(`创建模型 ${name} 失败`)
  }

  this.currentModelId = res.createModel.model.id
  this.createdModelIds.push(res.createModel.model.id)

  // 删除 owner 字段以模拟导入行为
  await this.projectClient.mutate(`
    mutation RemoveField($modelID: ID!, $fieldName: String!) {
      removeField(modelID: $modelID, fieldName: $fieldName) {
        error { __typename }
      }
    }
  `, { modelID: res.createModel.model.id, fieldName: 'owner' })
})
