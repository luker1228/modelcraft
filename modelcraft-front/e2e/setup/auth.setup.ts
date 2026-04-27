import { test, request as apiRequest } from 'playwright/test'
import path from 'path'
import fs from 'fs'
import { makeAccount } from '../data/accounts'

/**
 * E2E 测试账号策略：
 *
 * 优先使用环境变量 E2E_LOGIN_USERNAME / E2E_LOGIN_PASSWORD 指定的专用账号。
 * 如果未配置，则自动注册一个临时账号并使用。
 *
 * ⚠️ 绝对不要使用开发账号（如 'luke'）作为 fallback——
 *    refresh token rotation 会导致测试运行后开发 session 失效。
 */

export const AUTH_STORAGE_STATE = path.join(__dirname, '../.auth/user.json')

test('prepare auth storage state', async ({ page, baseURL }) => {
  // 确保 .auth 目录存在
  const authDir = path.dirname(AUTH_STORAGE_STATE)
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true })
  }

  let username: string
  let password: string
  let loginTab: 'phone' | 'username'

  if (process.env.E2E_LOGIN_USERNAME && process.env.E2E_LOGIN_PASSWORD) {
    // 使用显式配置的专用 e2e 账号
    username = process.env.E2E_LOGIN_USERNAME
    password = process.env.E2E_LOGIN_PASSWORD
    loginTab = 'username'
    console.log('[auth.setup] Using configured E2E account:', username)
  } else {
    // 自动注册临时账号，避免污染开发 session
    const account = makeAccount()
    const ctx = await apiRequest.newContext({ baseURL })
    const res = await ctx.post('/api/auth/register', {
      data: { phone: account.phone, userName: account.userName, password: account.password },
    })
    await ctx.dispose()

    if (!res.ok()) {
      throw new Error(`Auto-register failed: ${res.status()} ${await res.text()}`)
    }

    username = account.phone   // 用手机号登录
    password = account.password
    loginTab = 'phone'
    console.log('[auth.setup] Auto-registered temp account:', account.userName)
  }

  await page.context().clearCookies()
  await page.goto('/login')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await page.reload()

  if (loginTab === 'username') {
    await page.getByRole('tab', { name: '用户名登录' }).click()
    await page.getByLabel('用户名').fill(username)
  } else {
    // 手机号登录（默认 tab）
    await page.getByLabel('手机号').fill(username)
  }
  await page.getByPlaceholder('请输入密码').fill(password)

  const [loginResponse] = await Promise.all([
    page.waitForResponse(
      (response) =>
        response.url().includes('/api/auth/login') &&
        response.request().method() === 'POST',
    ),
    page.getByRole('button', { name: '登录' }).click(),
  ])

  if (!loginResponse.ok()) {
    throw new Error(`Login failed: ${loginResponse.status()} ${await loginResponse.text()}`)
  }

  await page.waitForURL(/\/org\/[^/]+\/workspace/)
  await page.context().storageState({ path: AUTH_STORAGE_STATE })
  console.log('[auth.setup] storageState saved')
})
