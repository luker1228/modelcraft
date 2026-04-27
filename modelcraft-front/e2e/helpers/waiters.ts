import type { Page } from 'playwright/test'
import path from 'path'

export const AUTH_STORAGE_STATE = path.join(__dirname, '../.auth/user.json')

export async function waitForAppReady(page: Page): Promise<void> {
  await page.waitForLoadState('domcontentloaded')
}

/**
 * 等待 workspace 页面加载完成。
 *
 * 应用用 refresh token cookie 换取 access token（silent refresh，单次使用 rotation）。
 * 成功后后端会 set-cookie 写入新的 refresh token，必须将最新 storageState 持久化，
 * 否则下一个测试的 browser context 拿到的是已被 revoke 的旧 cookie。
 */
export async function waitForWorkspaceReady(page: Page): Promise<void> {
  // 等待 URL 确认在 workspace（silent refresh 完成后才会停在此 URL）
  await page.waitForURL(/\/org\/[^/]+\/workspace/, { timeout: 30000 })
  // 等待页面主标题渲染，表示组件挂载完毕
  await page.getByRole('heading', { name: '所有项目' }).waitFor({ state: 'visible', timeout: 20000 })
  // 持久化最新 storageState（包含 rotation 后的新 refresh token cookie）
  await page.context().storageState({ path: AUTH_STORAGE_STATE })
}
