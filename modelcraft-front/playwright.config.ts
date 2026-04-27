import { defineConfig, devices } from 'playwright/test'

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:3000'

export default defineConfig({
  testDir: './e2e/specs',
  globalSetup: './e2e/setup/global-setup.ts',
  outputDir: 'test-results',
  reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  timeout: 45000,
  use: {
    baseURL,
    headless: false,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testDir: './e2e/setup',
      testMatch: '*.setup.ts',
    },
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // 不在此处全局注入 storageState，由各 spec 的 fixture 自行控制：
        // - auth spec 使用 auth.fixture.ts（无登录态）
        // - workspace spec 使用 workspace.fixture.ts（注入 storageState）
      },
      dependencies: ['setup'],
    },
  ],
})
