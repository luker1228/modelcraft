import { cn } from '@/shared/utils'

const maxWidthMap = {
  full: 'w-full',
  '5xl': 'max-w-5xl',
  '6xl': 'max-w-6xl',
  '7xl': 'max-w-7xl',
} as const

const paddingMap = {
  default: 'px-8 py-6',      // 32px H / 24px V — Stripe standard
  compact: 'px-8 py-4',
  spacious: 'px-8 py-8 xl:px-10',
  none: '',
} as const

interface PageLayoutProps {
  children: React.ReactNode
  /** Content max-width. Default: '7xl' */
  maxWidth?: keyof typeof maxWidthMap
  /** Background. Default: 'default' (inherits bg-background from AppLayout) */
  background?: 'default' | 'card'
  /** Padding preset. Default: 'default' (px-8 py-6) */
  padding?: keyof typeof paddingMap
  /** Extra className on the inner container */
  className?: string
}

export function PageLayout({
  children,
  maxWidth = '7xl',
  background = 'default',
  padding = 'default',
  className,
}: PageLayoutProps) {
  return (
    <div className={cn('h-full overflow-auto', background === 'card' && 'bg-card')}>
      <div className={cn('mx-auto', maxWidthMap[maxWidth], paddingMap[padding], className)}>
        {children}
      </div>
    </div>
  )
}
