// tests-bdd/step-definitions/enum.steps.ts
import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'
import { uniqueName } from '../fixtures/factory'

const CREATE_ENUM = `
  mutation CreateEnum($input: CreateEnumInput!) {
    createEnum(input: $input) {
      enum {
        id
        name
        displayName
        options { code label order }
      }
      error {
        __typename
        ... on EnumAlreadyExists { message }
        ... on InvalidEnumInput { message }
      }
    }
  }
`

function buildOptions(optionCodes: string) {
  return optionCodes.split(',').map((code, idx) => ({
    code: code.trim(),
    label: code.trim(),
    order: idx + 1,
  }))
}

Given(
  '已创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, baseName: string, optionCodes: string) {
    const name = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEnum: { enum: { id: string; name: string } | null; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }

    const enumDef = res.createEnum.enum
    if (!enumDef) throw new Error(`前置条件：创建枚举 ${name} 失败`)

    this.createdEnumNames.push(enumDef.name)
    this.lastEnumName = enumDef.name
  }
)

When(
  '我创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, baseName: string, optionCodes: string) {
    const name = uniqueName(baseName)
    const res = await this.projectClient.mutate<{
      createEnum: { enum: { id: string; name: string } | null; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }

    if (res.createEnum.enum?.name) {
      this.createdEnumNames.push(res.createEnum.enum.name)
    }
  }
)

When(
  '我再次创建名为 {string} 的枚举，选项为 {string}',
  async function (this: ModelCraftWorld, _baseName: string, optionCodes: string) {
    // 使用与 Given 步骤相同的实际名称
    const name = this.lastEnumName
    if (!name) throw new Error('没有记录到上一个枚举名称，请先用 Given 创建')
    const res = await this.projectClient.mutate<{
      createEnum: { enum: unknown; error: unknown }
    }>(CREATE_ENUM, {
      input: { name, displayName: name, options: buildOptions(optionCodes) },
    })
    this.lastResponse = { createEnum: res.createEnum }
  }
)

Then('枚举应该创建成功', function (this: ModelCraftWorld) {
  const payload = (this.lastResponse as { createEnum: { enum: unknown; error: unknown } }).createEnum
  expect(payload.error).toBeNull()
  expect(payload.enum).not.toBeNull()
})

Then('枚举名称应该包含 {string}', function (this: ModelCraftWorld, baseName: string) {
  const payload = (this.lastResponse as { createEnum: { enum: { name: string } | null } }).createEnum
  expect(payload.enum?.name).toContain(baseName)
})
