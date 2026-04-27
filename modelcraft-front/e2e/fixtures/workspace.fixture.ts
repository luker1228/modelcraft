import { test as base } from 'playwright/test'
import path from 'path'

export const AUTH_STORAGE_STATE = path.join(__dirname, '../.auth/user.json')

/**
 * 已登录态测试 fixture。
 * 依赖 auth.setup.ts 已生成 storageState，workspace/project 等测试用此 fixture。
 */
export const test = base.extend({
  storageState: AUTH_STORAGE_STATE,
})

export { expect } from 'playwright/test'
