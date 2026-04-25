import { cn } from '@/shared/utils'

interface PageHeaderProps {
  title: string
  description?: string
  /** Right-aligned actions slot */
  actions?: React.ReactNode
  /** Show bottom border. Default: false */
  bordered?: boolean
  /** Bottom margin. Default: 'default' (mb-8) */
  spacing?: 'default' | 'compact'
  /** Extra className on root div */
  className?: string
}

export function PageHeader({
  title,
  description,
  actions,
  bordered = false,
  spacing = 'default',
  className,
}: PageHeaderProps) {
  return (
    <div
      className={cn(
        spacing === 'compact' ? 'mb-6' : 'mb-8',
        bordered && 'border-b border-border pb-5',
        className,
      )}
    >
      <div className={cn('flex items-start justify-between gap-4', !actions && 'block')}>
        <div>
          <h1 className="text-xl font-semibold tracking-tight text-foreground">
            {title}
          </h1>
          {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
        </div>
        {actions && <div className="shrink-0">{actions}</div>}
      </div>
    </div>
  )
}
