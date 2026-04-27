import { test, expect } from '../../fixtures/auth.fixture'
import { makeAccount } from '../../data/accounts'

test('register and redirect to workspace', async ({ page }) => {
  const account = makeAccount()

  await page.goto('/register')
  await page.getByLabel('手机号').fill(account.phone)
  await page.getByLabel('用户名').fill(account.userName)
  await page.getByPlaceholder('至少 8 位密码').fill(account.password)
  await page.getByPlaceholder('请再次输入密码').fill(account.password)

  const [registerResponse, loginResponse] = await Promise.all([
    page.waitForResponse(
      response =>
        response.url().includes('/api/auth/register') && response.request().method() === 'POST'
    ),
    page.waitForResponse(
      response =>
        response.url().includes('/api/auth/login') && response.request().method() === 'POST'
    ),
    page.getByRole('button', { name: '注册' }).click(),
  ])

  expect(registerResponse.ok()).toBeTruthy()
  expect(loginResponse.ok()).toBeTruthy()

  await expect
    .poll(async () => page.evaluate(() => localStorage.getItem('defaultUserName')))
    .toBe(account.userName)
})
