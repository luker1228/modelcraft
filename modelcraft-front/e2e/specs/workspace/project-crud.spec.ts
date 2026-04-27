/**
 * Workspace 项目 CRUD E2E 测试
 *
 * 覆盖场景：
 * - 打开"创建新项目"对话框
 * - Step 1 表单验证（必填项）
 * - Step 1 → Step 2 slug 自动生成
 * - Step 2 连接测试按钮可交互
 * - 编辑项目（修改标题/描述）
 * - 删除项目确认对话框
 */

import { test, expect } from '../../fixtures/workspace.fixture'
import { routes } from '../../helpers/routes'
import { waitForWorkspaceReady } from '../../helpers/waiters'

const ORG_NAME = process.env.E2E_ORG_NAME ?? 'luke2q'

// ──────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────

async function openCreateDialog(page: import('playwright/test').Page) {
  await page.getByRole('button', { name: '新建项目' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await expect(page.getByRole('heading', { name: '创建新项目' })).toBeVisible()
}

// ──────────────────────────────────────────────
// 创建项目 — 对话框交互
// ──────────────────────────────────────────────

test.describe('创建项目对话框', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)
    await openCreateDialog(page)
  })

  test('Step 1 — 对话框初始状态正确', async ({ page }) => {
    const dialog = page.getByRole('dialog')

    // 步骤指示器（文本节点）
    await expect(dialog.locator('text=步骤 1/3')).toBeVisible()
    // 步骤描述文字
    await expect(dialog.locator('text=填写项目的基础信息')).toBeVisible()

    // 表单字段
    await expect(dialog.getByLabel('项目名称 *')).toBeVisible()
    await expect(dialog.getByLabel('项目标识 *')).toBeVisible()
    await expect(dialog.getByLabel('项目描述')).toBeVisible()

    // 底部按钮
    await expect(dialog.getByRole('button', { name: '取消' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: '下一步' })).toBeVisible()
  })

  test('Step 1 — 项目名称为空时下一步触发校验', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: '下一步' }).click()

    // Zod 验证消息应出现
    await expect(dialog.getByText('项目名称不能为空')).toBeVisible()
  })

  test('Step 1 — 输入项目名称自动生成 slug', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const titleInput = dialog.getByLabel('项目名称 *')
    const slugInput = dialog.getByLabel('项目标识 *')

    // 输入英文名称，slug 应该自动生成
    await titleInput.fill('Test Project')
    // 等待 slug 被填充（debounce 或同步）
    await expect(slugInput).not.toHaveValue('', { timeout: 2000 })
  })

  test('Step 1 — slug 格式不合法时显示错误', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('项目名称 *').fill('test')
    const slugInput = dialog.getByLabel('项目标识 *')
    await slugInput.clear()
    await slugInput.fill('1invalid-slug')  // 不以字母开头
    await dialog.getByRole('button', { name: '下一步' }).click()

    await expect(dialog.getByText('必须以字母开头，只能包含小写字母、数字和下划线')).toBeVisible()
  })

  test('Step 1 → Step 2 — 正确填写后进入下一步', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('项目名称 *').fill('E2E Test Project')

    // 等待 slug 自动填充
    await expect(dialog.getByLabel('项目标识 *')).not.toHaveValue('', { timeout: 2000 })

    await dialog.getByRole('button', { name: '下一步' }).click()

    // 进入 Step 2 — 验证步骤计数和步骤描述文字
    await expect(dialog.locator('text=步骤 2/3')).toBeVisible()
    await expect(dialog.locator('text=配置项目默认数据库连接')).toBeVisible()
    await expect(dialog.getByRole('button', { name: '测试连接' })).toBeVisible()
  })

  test('Step 2 — 数据库字段可见', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('项目名称 *').fill('E2E Test Project')
    await expect(dialog.getByLabel('项目标识 *')).not.toHaveValue('', { timeout: 2000 })
    await dialog.getByRole('button', { name: '下一步' }).click()

    await expect(dialog.locator('text=步骤 2/3')).toBeVisible()

    // 数据库连接字段
    await expect(dialog.getByLabel('集群名称 *')).toBeVisible()
    await expect(dialog.getByPlaceholder('localhost 或 db.example.com')).toBeVisible()  // 主机
    await expect(dialog.getByRole('button', { name: '测试连接' })).toBeVisible()
    // 下一步按钮在未测试连接时应禁用
    await expect(dialog.getByRole('button', { name: '下一步' })).toBeDisabled()
  })

  test('Step 2 → Step 1 — 上一步返回', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('项目名称 *').fill('E2E Test Project')
    await expect(dialog.getByLabel('项目标识 *')).not.toHaveValue('', { timeout: 2000 })
    await dialog.getByRole('button', { name: '下一步' }).click()
    await expect(dialog.locator('text=步骤 2/3')).toBeVisible()

    await dialog.getByRole('button', { name: '上一步' }).click()
    await expect(dialog.locator('text=步骤 1/3')).toBeVisible()
  })

  test('取消按钮关闭对话框', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: '取消' }).click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })
})

// ──────────────────────────────────────────────
// 编辑项目
// ──────────────────────────────────────────────

test.describe('编辑项目对话框', () => {
  test('通过项目卡片菜单打开编辑对话框', async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)

    // 检查是否有项目卡片
    const card = page.locator('[class*="cursor-pointer"]').first()
    const cardVisible = await card.isVisible().catch(() => false)
    if (!cardVisible) {
      test.skip()
      return
    }

    // hover 触发 MoreHorizontal 按钮显示
    await card.hover()
    const menuBtn = card.getByRole('button', { name: '打开菜单' })
    await menuBtn.click()

    // 点击"编辑"
    await page.getByRole('menuitem', { name: '编辑' }).click()

    // 编辑对话框打开
    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByRole('heading', { name: '编辑项目' })).toBeVisible()
  })

  test('编辑模式下 slug 字段禁用', async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)

    const card = page.locator('[class*="cursor-pointer"]').first()
    const cardVisible = await card.isVisible().catch(() => false)
    if (!cardVisible) {
      test.skip()
      return
    }

    await card.hover()
    await card.getByRole('button', { name: '打开菜单' }).click()
    await page.getByRole('menuitem', { name: '编辑' }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()

    // slug 字段在编辑模式下应该是禁用的
    await expect(dialog.getByLabel('项目标识 *')).toBeDisabled()

    // 编辑模式只有保存/取消，没有多步流程
    await expect(dialog.getByRole('button', { name: '保存修改' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: '取消' })).toBeVisible()
  })
})

// ──────────────────────────────────────────────
// 删除项目
// ──────────────────────────────────────────────

test.describe('删除项目确认对话框', () => {
  test('通过项目卡片菜单打开删除确认', async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)

    const card = page.locator('[class*="cursor-pointer"]').first()
    const cardVisible = await card.isVisible().catch(() => false)
    if (!cardVisible) {
      test.skip()
      return
    }

    await card.hover()
    await card.getByRole('button', { name: '打开菜单' }).click()
    await page.getByRole('menuitem', { name: '删除' }).click()

    // 删除确认对话框应出现
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('删除对话框可以取消', async ({ page }) => {
    await page.goto(routes.workspace(ORG_NAME))
    await waitForWorkspaceReady(page)

    const card = page.locator('[class*="cursor-pointer"]').first()
    const cardVisible = await card.isVisible().catch(() => false)
    if (!cardVisible) {
      test.skip()
      return
    }

    await card.hover()
    await card.getByRole('button', { name: '打开菜单' }).click()
    await page.getByRole('menuitem', { name: '删除' }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()

    // 点击取消，对话框消失，不执行删除
    const cancelBtn = dialog.getByRole('button', { name: /取消/ })
    await cancelBtn.click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })
})
