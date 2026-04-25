import { cn } from '@/shared/utils'

const maxWidthMap = {
  full: 'w-full',
  '5xl': 'max-w-5xl',
  '6xl': 'max-w-6xl',
  '7xl': 'max-w-7xl',
} as const

const paddingMap = {
  default: 'px-16 py-8',     // 64px H / 32px V — breathing room
  compact: 'px-16 py-4',
  spacious: 'px-16 py-10 xl:px-20',
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
      <div className={cn(maxWidthMap[maxWidth], paddingMap[padding], className)}>
        {children}
      </div>
    </div>
  )
}
