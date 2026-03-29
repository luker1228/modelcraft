// tests-bdd/step-definitions/model.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'
import { uniqueName } from '../fixtures/factory'

const CREATE_MODEL = `
  mutation CreateModel($input: CreateModelInput!) {
    createModel(input: $input) {
      model {
        id
        name
        title
        databaseName
      }
      error {
        __typename
        ... on ModelAlreadyExists { message }
        ... on InvalidModelInput { message }
        ... on ProjectNotFound { message }
      }
    }
  }
`

const DELETE_MODEL = `
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      error { __typename }
    }
  }
`

Given('已创建名为 {string} 的模型', async function (this: ModelCraftWorld, baseName: string) {
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createModel: { model: { id: string; name: string } | null; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }

  const model = res.createModel.model
  if (!model) throw new Error(`前置条件：创建模型 ${name} 失败`)

  // 存储到 World（不用模块级变量，避免跨 Scenario 污染）
  this.currentModelId = model.id
  this.lastModelName = model.name
  this.createdModelIds.push(model.id)
  // 供 lfk.steps.ts 使用：baseName → 实际 ID
  this.modelMap[baseName] = model.id
})

When('我创建名为 {string} 的模型', async function (this: ModelCraftWorld, baseName: string) {
  const name = uniqueName(baseName)
  const res = await this.projectClient.mutate<{
    createModel: { model: { id: string } | null; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }

  if (res.createModel.model?.id) {
    this.currentModelId = res.createModel.model.id
    this.createdModelIds.push(res.createModel.model.id)
  }
})

When('我再次创建名为 {string} 的模型', async function (this: ModelCraftWorld, _baseName: string) {
  // 使用 Given 步骤保存的实际名称（已包含唯一后缀）
  const name = this.lastModelName
  if (!name) throw new Error('没有记录到上一个模型名称，请先用 Given 创建')

  const res = await this.projectClient.mutate<{
    createModel: { model: unknown; error: unknown }
  }>(CREATE_MODEL, {
    input: { name, title: name, databaseName: 'test_db' },
  })
  this.lastResponse = { createModel: res.createModel }
})

When('我删除该模型', async function (this: ModelCraftWorld) {
  const id = this.currentModelId
  if (!id) throw new Error('没有可删除的模型（请先用 Given 创建）')

  const res = await this.projectClient.mutate<{ deleteModel: { error: unknown } }>(
    DELETE_MODEL, { id }
  )
  this.lastResponse = { deleteModel: res.deleteModel }

  // 从清理列表中移除（已手动删除）
  this.createdModelIds = this.createdModelIds.filter(mid => mid !== id)
  this.currentModelId = null
})

Then('模型应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { createModel: { model: unknown; error: unknown } }).createModel
  expect(payload.error).toBeNull()
  expect(payload.model).not.toBeNull()
})

Then('模型名称应该是 {string}', function (this: ModelCraftWorld, baseName: string) {
  const payload = (this.lastResponse as { createModel: { model: { name: string } | null } }).createModel
  // 名称带了 uniqueName 后缀，只检查包含原始 baseName
  expect(payload.model?.name).toContain(baseName)
})
