import { cn } from '@/shared/utils'

const maxWidthMap = {
  full: 'w-full',
  '5xl': 'max-w-5xl',
  '6xl': 'max-w-6xl',
  '7xl': 'max-w-7xl',
} as const

const paddingMap = {
  default: 'px-6 py-8',
  compact: 'p-6',
  spacious: 'px-6 pb-12 pt-10 xl:px-10',
  none: '',
} as const

interface PageLayoutProps {
  children: React.ReactNode
  /** Content max-width. Default: '7xl' */
  maxWidth?: keyof typeof maxWidthMap
  /** Background. Default: 'default' (inherits bg-muted from AppLayout) */
  background?: 'default' | 'card'
  /** Padding preset. Default: 'default' (px-6 py-8) */
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
