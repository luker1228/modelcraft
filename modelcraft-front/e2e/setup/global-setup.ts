import type { FullConfig } from 'playwright/test'

export default async function globalSetup(config: FullConfig): Promise<void> {
  const resolvedBaseURL =
    process.env.PLAYWRIGHT_BASE_URL ??
    config.projects.find(project => project.name === 'chromium')?.use?.baseURL

  if (!resolvedBaseURL) {
    throw new Error('PLAYWRIGHT_BASE_URL 未配置，且 Playwright 配置中缺少 baseURL。')
  }
}
