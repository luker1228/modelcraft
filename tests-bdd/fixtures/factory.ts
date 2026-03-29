import { randomUUID } from 'crypto'

/**
 * 生成并发安全的唯一名称，避免多次运行之间冲突。
 * 不使用连字符分隔，因为 model/enum 名称不允许包含连字符。
 * @example uniqueName('User') → 'Usera3f2b1c0'
 */
export const uniqueName = (prefix: string): string =>
  `${prefix}${randomUUID().replace(/-/g, '').slice(0, 8)}`
