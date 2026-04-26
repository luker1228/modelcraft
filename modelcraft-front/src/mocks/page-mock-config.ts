/**
 * 页面级 mock 控制。
 *
 * 通过环境变量 NEXT_PUBLIC_MOCK_PAGES 控制哪些页面使用 MSW mock 数据。
 * 多个页面用逗号分隔，例如：
 *
 *   NEXT_PUBLIC_MOCK_PAGES=model-editor,enum-list
 *
 * 页面 key 与 buildHandlers() 中的 key 对应，也与各页面目录名保持一致。
 */

const MOCK_PAGES_ENV = process.env.NEXT_PUBLIC_MOCK_PAGES ?? ''

const MOCK_PAGES: Set<string> = new Set(
  MOCK_PAGES_ENV.split(',')
    .map((s) => s.trim())
    .filter(Boolean)
)

/**
 * 判断指定页面是否启用 mock 模式。
 *
 * @param pageKey 页面标识符，与 buildHandlers() 中的 key 对应
 * @example
 * if (isMockPage('model-editor')) { ... }
 */
export function isMockPage(pageKey: string): boolean {
  return MOCK_PAGES.has(pageKey)
}

/**
 * 返回所有已启用 mock 的页面 key 列表。
 */
export function getMockPages(): string[] {
  return Array.from(MOCK_PAGES)
}
