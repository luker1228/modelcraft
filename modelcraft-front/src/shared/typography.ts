/**
 * Typography System - ModelCraft Design System
 * 
 * 字体统一管理文件
 * 
 * 字体家族：
 * - Inter (font-sans): 正文、UI元素、表单
 * - Space Grotesk (font-heading): 标题、显示文本
 * - Fira Code (font-mono): 代码、技术标识符
 * 
 * 字重规范：
 * - font-normal (400): 正文、正常文本
 * - font-medium (500): 仅用于技术标识符（代码、枚举名、API标识等）
 * - font-semibold (600): 标题、重要UI元素
 * - font-bold (700): 页面主标题、强调文本
 */

/**
 * 基础字体类组合
 * 用于需要精确控制的场景
 */
export const FONT_FAMILIES = {
  /** 正文字体 - Inter */
  sans: 'font-sans',
  /** 标题字体 - Space Grotesk */
  heading: 'font-heading',
  /** 代码字体 - Fira Code */
  mono: 'font-mono',
} as const;

/**
 * 字重类
 * 遵循严格的使用规范
 */
export const FONT_WEIGHTS = {
  /** 正常 (400) - 正文、正常文本 */
  normal: 'font-normal',
  /** 中等 (500) - 仅用于技术标识符 */
  medium: 'font-medium',
  /** 半粗 (600) - 标题、重要UI元素 */
  semibold: 'font-semibold',
  /** 粗体 (700) - 页面主标题、强调文本 */
  bold: 'font-bold',
} as const;

/**
 * 文字大小类
 * 对应 Tailwind 的 text-* 类
 */
export const TEXT_SIZES = {
  xs: 'text-xs',     // 12px
  sm: 'text-sm',     // 14px
  base: 'text-base', // 16px
  lg: 'text-lg',     // 18px
  xl: 'text-xl',     // 20px
  '2xl': 'text-2xl', // 24px
  '3xl': 'text-3xl', // 30px
} as const;

/**
 * 语义化排版组合
 * 直接用于常见UI场景
 */
export const TYPOGRAPHY = {
  // ========== 页面级标题 ==========
  /** 页面主标题 (h1) */
  pageTitle: `${FONT_FAMILIES.heading} ${FONT_WEIGHTS.bold} ${TEXT_SIZES['2xl']}`,
  /** 页面副标题 */
  pageSubtitle: `${FONT_FAMILIES.heading} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.lg}`,

  // ========== 区块级标题 ==========
  /** 章节标题 (h2) */
  sectionTitle: `${FONT_FAMILIES.heading} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.xl}`,
  /** 子章节标题 (h3) */
  subsectionTitle: `${FONT_FAMILIES.heading} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.lg}`,
  /** 卡片标题 */
  cardTitle: `${FONT_FAMILIES.heading} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.base}`,
  /** 小标题 */
  smallTitle: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.sm}`,

  // ========== 正文文本 ==========
  /** 正文（默认） */
  body: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.base}`,
  /** 小正文 */
  bodySmall: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,
  /** 说明文字 */
  caption: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.xs}`,
  /** 辅助文字（带颜色） */
  muted: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm} text-muted-foreground`,

  // ========== 表单元素 ==========
  /** 表单标签 */
  label: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.sm}`,
  /** 输入框文本 */
  input: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,
  /** 占位符文本 */
  placeholder: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm} text-muted-foreground`,
  /** 错误提示 */
  error: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.xs} text-destructive`,
  /** 帮助文字 */
  helper: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.xs} text-muted-foreground`,

  // ========== 按钮 ==========
  /** 按钮文本（默认） */
  button: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,
  /** 按钮文本（小） */
  buttonSmall: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.xs}`,
  /** 按钮文本（大） */
  buttonLarge: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.base}`,

  // ========== 表格 ==========
  /** 表头文本 */
  tableHeader: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.sm}`,
  /** 表格单元格 */
  tableCell: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,

  // ========== 技术标识符（特殊规则：使用 font-medium） ==========
  /** 代码块 */
  code: `${FONT_FAMILIES.mono} ${FONT_WEIGHTS.medium} ${TEXT_SIZES.sm}`,
  /** 内联代码 */
  codeInline: `${FONT_FAMILIES.mono} ${FONT_WEIGHTS.medium} ${TEXT_SIZES.xs}`,
  /** 枚举名/API标识 */
  identifier: `${FONT_FAMILIES.mono} ${FONT_WEIGHTS.medium} ${TEXT_SIZES.sm}`,
  /** 技术标签 */
  tag: `${FONT_FAMILIES.mono} ${FONT_WEIGHTS.medium} ${TEXT_SIZES.xs}`,

  // ========== 特殊用途 ==========
  /** 徽章 */
  badge: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.semibold} ${TEXT_SIZES.xs}`,
  /** 工具提示 */
  tooltip: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.xs}`,
  /** 面包屑 */
  breadcrumb: `${FONT_FAMILIES.sans} ${FONT_WEIGHTS.normal} ${TEXT_SIZES.sm}`,
} as const;

/**
 * 工具函数：组合多个排版类
 */
export function cn(...classes: (string | undefined | null | false)[]): string {
  return classes.filter(Boolean).join(' ');
}

/**
 * 使用示例：
 * 
 * ```tsx
 * import { TYPOGRAPHY, FONT_FAMILIES, FONT_WEIGHTS } from '@/shared/typography';
 * 
 * // 使用语义化组合（推荐）
 * <h1 className={TYPOGRAPHY.pageTitle}>ModelCraft</h1>
 * <p className={TYPOGRAPHY.body}>正文内容</p>
 * <code className={TYPOGRAPHY.code}>UserRole</code>
 * 
 * // 自定义组合
 * <span className={cn(FONT_FAMILIES.sans, FONT_WEIGHTS.semibold, 'text-lg')}>
 *   自定义文本
 * </span>
 * ```
 * 
 * 注意事项：
 * 1. 优先使用 TYPOGRAPHY 中的语义化组合
 * 2. font-medium 仅用于技术标识符（代码、枚举名等）
 * 3. 正文默认使用 font-normal (400)
 * 4. 标题使用 font-heading + font-semibold/font-bold
 */
