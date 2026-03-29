import { After, Before } from '@cucumber/cucumber'
import { ModelCraftWorld } from './world'

const DELETE_MODEL = `
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      error { __typename }
    }
  }
`

const DELETE_ENUM = `
  mutation DeleteEnum($name: String!) {
    deleteEnum(name: $name) {
      error { __typename }
    }
  }
`

// 每个 Scenario 前重置追踪列表
Before(function (this: ModelCraftWorld) {
  this.createdModelIds = []
  this.createdEnumNames = []
  this.currentModelId = null
  this.modelMap = {}
  this.lastModelName = null
  this.lastEnumName = null
  this.lastResponse = null
  this.lastError = null
})

// 每个 Scenario 后通过 API 清理创建的数据（@smoke 除外保留数据方便调试）
After({ tags: 'not @smoke' }, async function (this: ModelCraftWorld) {
  // 逆序删除 model（field 随 model 级联删除）
  for (const id of [...this.createdModelIds].reverse()) {
    try {
      await this.projectClient.mutate(DELETE_MODEL, { id })
    } catch {
      // 清理失败不影响测试结果，静默处理
    }
  }

  // 删除 enum
  for (const name of this.createdEnumNames) {
    try {
      await this.projectClient.mutate(DELETE_ENUM, { name })
    } catch {
      // 静默处理
    }
  }
})
