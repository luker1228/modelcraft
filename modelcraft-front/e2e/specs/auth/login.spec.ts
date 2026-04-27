import { test, expect } from '../../fixtures/auth.fixture'
import { makeAccount } from '../../data/accounts'

const EXISTING_USERNAME = process.env.E2E_LOGIN_USERNAME ?? 'luke'
const EXISTING_PASSWORD = process.env.E2E_LOGIN_PASSWORD ?? 'jmx931228'

test('login with freshly registered account', async ({ page, request }) => {
  const account = makeAccount()

  const registerResponse = await request.post('/api/auth/register', {
    data: {
      phone: account.phone,
      userName: account.userName,
      password: account.password,
    },
  })

  expect(registerResponse.ok()).toBeTruthy()

  await page.goto('/login')
  await page.getByLabel('手机号').fill(account.phone)
  await page.getByPlaceholder('请输入密码').fill(account.password)

  const [loginResponse] = await Promise.all([
    page.waitForResponse(
      response =>
        response.url().includes('/api/auth/login') && response.request().method() === 'POST'
    ),
    page.getByRole('button', { name: '登录' }).click(),
  ])

  expect(loginResponse.ok()).toBeTruthy()

  await expect
    .poll(async () => page.evaluate(() => localStorage.getItem('defaultUserName')))
    .toBe(account.userName)
})

test('login with existing account and enter workspace', async ({ page }) => {
  await page.context().clearCookies()

  await page.goto('/login')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await page.reload()

  await page.getByRole('tab', { name: '用户名登录' }).click()
  await page.getByLabel('用户名').fill(EXISTING_USERNAME)
  await page.getByPlaceholder('请输入密码').fill(EXISTING_PASSWORD)

  const [loginResponse] = await Promise.all([
    page.waitForResponse(
      response =>
        response.url().includes('/api/auth/login') && response.request().method() === 'POST'
    ),
    page.getByRole('button', { name: '登录' }).click(),
  ])

  expect(loginResponse.ok()).toBeTruthy()

  await page.waitForURL(/\/org\/[^/]+\/workspace/)
  await expect(page).toHaveURL(/\/org\/[^/]+\/workspace/)

  await expect
    .poll(async () => page.evaluate(() => localStorage.getItem('defaultUserName')))
    .toBe(EXISTING_USERNAME)
})
