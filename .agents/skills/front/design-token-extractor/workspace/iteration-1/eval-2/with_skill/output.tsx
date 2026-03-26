import { LayoutDashboard, Database, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SidebarNavItemProps {
  href: string
  icon: React.ReactNode
  label: string
  active?: boolean
}

function SidebarNavItem({ href, icon, label, active }: SidebarNavItemProps) {
  return (
    <a
      href={href}
      className={cn(
        'flex items-center gap-3 py-2 px-3 rounded-[6px] text-sm font-medium no-underline transition-all duration-200 whitespace-nowrap',
        active
          ? 'bg-[#dadee5] text-gray-900 font-semibold'
          : 'text-gray-500 hover:bg-gray-50 hover:text-gray-900'
      )}
    >
      {icon}
      {label}
    </a>
  )
}

export function Sidebar() {
  return (
    <div className="w-[200px] bg-white border-r border-gray-200 flex flex-col overflow-hidden flex-shrink-0">
      <div className="flex-1 overflow-y-auto overflow-x-hidden p-2">
        <div className="px-3 pt-2 pb-1 text-[10px] font-medium uppercase tracking-[0.05em] text-gray-400 whitespace-nowrap">
          工作区
        </div>
        <SidebarNavItem
          href="#"
          icon={<LayoutDashboard className="w-4 h-4 flex-shrink-0" strokeWidth={1.5} />}
          label="总览"
          active
        />
        <SidebarNavItem
          href="#"
          icon={<Database className="w-4 h-4 flex-shrink-0" strokeWidth={1.5} />}
          label="数据模型"
        />
        <SidebarNavItem
          href="#"
          icon={<Settings className="w-4 h-4 flex-shrink-0" strokeWidth={1.5} />}
          label="设置"
        />
      </div>
    </div>
  )
}
