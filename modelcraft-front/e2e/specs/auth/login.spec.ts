import { test, expect } from '../../fixtures/auth.fixture'
import { makeAccount } from '../../data/accounts'

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
