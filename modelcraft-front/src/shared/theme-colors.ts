/**
 * Unified Theme Colors and Styles
 * 
 * This file exports all the theme colors and style constants used across the application.
 * Update these values to change colors globally throughout the app.
 * 
 * Related CSS Variables in src/app/globals.css:
 * - --container-bg: Main container background color
 * - --sidebar-background: Sidebar background color
 */

/**
 * Container Background Color
 * Used for: sidebar, cards, main content areas
 * CSS Variable: --container-bg
 */
export const CONTAINER_BG_CLASS = 'bg-sidebar'

/**
 * Container Background with Blur Effect
 * Used for: glass morphism cards and transparent overlays
 */
export const CONTAINER_BG_BLUR_CLASS = 'bg-sidebar backdrop-blur-sm'

/**
 * Hover State for Container Background
 * Used for: hover effects on cards
 */
export const CONTAINER_BG_HOVER_CLASS = 'hover:bg-sidebar/95'

/**
 * Selected/Accent Color
 * CSS Variable: --selected
 */
export const SELECTED_COLOR_CLASS = 'bg-selected'
export const SELECTED_HOVER_CLASS = 'hover:bg-selected'

/**
 * Theme Color Descriptions
 * Light Mode:
 * - Container Background: hsl(0 0% 100%) - Pure white
 * 
 * Dark Mode:
 * - Container Background: hsl(240 5.9% 10%) - Dark background
 * 
 * Note: These values are automatically adjusted based on the current theme (light/dark)
 */

/**
 * Common Component Classes
 * Use these combinations for consistent styling
 */
export const CARD_CONTAINER_CLASS = `${CONTAINER_BG_CLASS} backdrop-blur-sm border-0 shadow-md ${CONTAINER_BG_HOVER_CLASS} transition-all duration-300 cursor-pointer`

export const INPUT_CONTAINER_CLASS = `${CONTAINER_BG_CLASS} backdrop-blur-sm border-slate-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all`

export const BUTTON_CONTAINER_CLASS = `h-10 ${CONTAINER_BG_CLASS} backdrop-blur-sm ${CONTAINER_BG_HOVER_CLASS} border border-slate-200 transition-all`

/**
 * Usage Examples:
 * 
 * 1. In TSX files (import from this file):
 *    import { CARD_CONTAINER_CLASS } from '@/shared/theme-colors'
 *    <Card className={CARD_CONTAINER_CLASS}>...</Card>
 * 
 * 2. In CSS/globals.css:
 *    Use CSS variables: hsl(var(--container-bg))
 *    or use Tailwind classes: bg-sidebar
 * 
 * 3. To update all containers globally:
 *    - Change --container-bg and --sidebar-background in globals.css
 *    - All components using CONTAINER_BG_CLASS will update automatically
 */
