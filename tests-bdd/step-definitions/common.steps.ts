// tests-bdd/step-definitions/common.steps.ts
import { Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'

/**
 * 断言最近操作返回了指定 __typename 的错误。
 *
 * 错误位于 payload 的 error 字段（不在顶层 errors 数组）：
 *   { data: { createModel: { error: { __typename: "ModelAlreadyExists" } } } }
 *
 * 用法：Then 应该返回错误类型 "ModelAlreadyExists"
 */
Then('应该返回错误类型 {string}', function (this: ModelCraftWorld, expectedTypename: string) {
  expect(this.lastResponse).not.toBeNull()

  // lastResponse 的结构：{ <mutationName>: { error: { __typename: "..." } } }
  // 取第一个 key（只有一个 mutation）
  const payload = Object.values(this.lastResponse!)[0] as Record<string, unknown>
  const error = payload?.error as Record<string, unknown> | null

  expect(error).not.toBeNull()
  expect(error?.__typename).toBe(expectedTypename)
})

/**
 * 断言最近操作成功（error 为 null）。
 */
Then('操作应该成功', function (this: ModelCraftWorld) {
  expect(this.lastResponse).not.toBeNull()
  const payload = Object.values(this.lastResponse!)[0] as Record<string, unknown>
  expect(payload?.error).toBeNull()
})
