/**
 * Workspace 项目列表 E2E 测试
 *
 * 覆盖场景：
 * - 已登录用户访问 workspace，项目列表正常渲染
 * - 搜索过滤项目
 * - 切换网格/列表视图
 * - 空状态（搜索无匹配时）
 * - 点击项目卡片跳转到项目详情
 */

import { test, expect } from '../../fixtures/workspace.fixture'
import { routes } from '../../helpers/routes'
import { waitForWorkspaceReady } from '../../helpers/waiters'

const ORG_NAME = process.env.E2E_ORG_NAME ?? 'luke2q'

test.describe('Workspace 项目列表', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)
  })

  test('页面标题和核心元素可见', async ({ page }) => {
    // 页面标题（waitForWorkspaceReady 已等待其可见）
    await expect(page.getByRole('heading', { name: '所有项目' })).toBeVisible()

    // 新建项目按钮
    await expect(page.getByRole('button', { name: '新建项目' })).toBeVisible()

    // 搜索框
    await expect(page.getByPlaceholder('搜索项目...')).toBeVisible()
  })

  test('默认以网格视图显示项目', async ({ page }) => {
    // 至少有一张项目卡片，或显示空状态
    const projectCards = page.locator('[class*="grid"] > div, [class*="space-y"] > div').first()
    const emptyState = page.getByText('暂无项目')

    // 网格或空状态必须有一个可见
    const hasContent = await Promise.race([
      projectCards.waitFor({ state: 'visible', timeout: 5000 }).then(() => true).catch(() => false),
      emptyState.waitFor({ state: 'visible', timeout: 5000 }).then(() => true).catch(() => false),
    ])

    expect(hasContent).toBeTruthy()
  })

  test('切换到列表视图再切回网格视图', async ({ page }) => {
    // ViewToggle 通常有 list/grid 两个切换按钮
    // 切到列表视图
    const listToggle = page.getByRole('button', { name: /list|列表/i })
    if (await listToggle.isVisible()) {
      await listToggle.click()
      // 列表视图的行有 space-y-1.5 容器
      await expect(page.locator('[class*="space-y-1"]').first()).toBeVisible()

      // 切回网格
      const gridToggle = page.getByRole('button', { name: /grid|网格/i })
      await gridToggle.click()
      await expect(page.locator('[class*="grid"]').first()).toBeVisible()
    }
  })

  test('搜索框输入过滤项目', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索项目...')

    // 输入一个不可能存在的字符串
    await searchInput.fill('zzz_no_match_xyz')

    // 应显示"未找到匹配的项目"
    await expect(page.getByText('未找到匹配的项目')).toBeVisible({ timeout: 3000 })
  })

  test('清空搜索框恢复项目列表', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索项目...')
    await searchInput.fill('zzz_no_match_xyz')
    await expect(page.getByText('未找到匹配的项目')).toBeVisible({ timeout: 3000 })

    // 点击清空按钮（clearable 属性）
    const clearBtn = page.getByRole('button', { name: /clear|清空/i })
    if (await clearBtn.isVisible()) {
      await clearBtn.click()
    } else {
      await searchInput.clear()
    }

    // 空状态消失
    await expect(page.getByText('未找到匹配的项目')).not.toBeVisible({ timeout: 3000 })
  })
})

test.describe('Workspace 项目导航', () => {
  test('点击项目卡片跳转到项目页', async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)

    // 找到第一个项目卡片（网格 or 列表）
    const firstCard = page
      .locator('[class*="cursor-pointer"]')
      .filter({ hasText: /.+/ })
      .first()

    const cardVisible = await firstCard.isVisible().catch(() => false)
    if (!cardVisible) {
      test.skip()
      return
    }

    await firstCard.click()

    // 应导航到 /org/{orgName}/project/{slug}
    await page.waitForURL(/\/org\/[^/]+\/project\/[^/]+/, { timeout: 10000 })
    await expect(page).toHaveURL(/\/org\/[^/]+\/project\/[^/]+/)
  })
})
