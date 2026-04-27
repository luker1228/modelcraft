import { test as base } from 'playwright/test'

/**
 * 基础测试 fixture（无登录态）。
 * 用于 auth 相关测试（login、register），从未登录状态开始。
 */
export const test = base.extend({})

export { expect } from 'playwright/test'
