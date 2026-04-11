'use client'

import { useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@web/components/ui/avatar'
import { Button } from '@web/components/ui/button'
import { Settings, LogOut, ChevronDown, CircleUserRound } from 'lucide-react'
import { cn } from '@/shared/utils'

/**
 * UserMenu Component
 *
 * PUBLIC COMPONENT - DO NOT MODIFY UNLESS ABSOLUTELY NECESSARY
 *
 * This is a core navigation component used across the entire application.
 * Any changes here will affect all pages.
 *
 * Built with shadcn/ui components:
 * - DropdownMenu (navigation)
 * - Avatar (user display)
 * - Button (trigger)
 *
 * Usage:
 * <UserMenu
 *   userName="John Doe"
 *   userEmail="john@example.com"
 *   onLogout={() => handleLogout()}
 * />
 */

interface UserMenuProps {
  /** User's display name */
  userName: string
  /** User's email address */
  userEmail?: string
  /** Optional user avatar image URL */
  userAvatar?: string
  /** Callback when user clicks logout */
  onLogout: () => void
  /** Optional callback when user clicks profile */
  onProfile?: () => void
  /** Optional callback when user clicks settings */
  onSettings?: () => void
}

export function UserMenu({
  userName,
  userEmail,
  userAvatar,
  onLogout,
  onProfile,
  onSettings,
}: UserMenuProps) {
  const params = useParams()
  const router = useRouter()

  const orgName = typeof params.orgName === 'string' ? params.orgName : ''

  // Get initials for avatar fallback (up to 2 characters)
  const getInitials = (name: string) => {
    const parts = name.trim().split(/\s+/)
    if (parts.length === 1) {
      return parts[0].substring(0, 2).toUpperCase()
    }
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
  }

  const initials = userName ? getInitials(userName) : 'U'

  const handleProfile = useCallback(() => {
    if (onProfile) {
      onProfile()
      return
    }

    if (!orgName) {
      return
    }

    router.push(`/org/${orgName}/profile`)
  }, [onProfile, orgName, router])

  const handleSettings = useCallback(() => {
    if (onSettings) {
      onSettings()
      return
    }

    if (!orgName) {
      return
    }

    router.push(`/org/${orgName}/settings`)
  }, [onSettings, orgName, router])

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="relative h-9 gap-2 rounded-lg px-2 transition-colors hover:bg-selected"
        >
          {/* Avatar */}
          <Avatar className="size-7">
            {userAvatar && <AvatarImage src={userAvatar} alt={userName} />}
            <AvatarFallback className={cn(
              "text-xs font-semibold",
              "bg-primary text-primary-foreground",
              "text-white"
            )}>
              {initials}
            </AvatarFallback>
          </Avatar>

          {/* Dropdown indicator */}
          <ChevronDown className="size-3.5 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-56 border border-slate-200 shadow-lg"
        sideOffset={8}
      >
        {/* User Info Header */}
        <DropdownMenuLabel>
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-semibold leading-none text-foreground">{userName}</p>
            {userEmail && (
              <p className="text-xs leading-none text-muted-foreground">
                {userEmail}
              </p>
            )}
          </div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        {/* Profile */}
        <DropdownMenuItem
          className="cursor-pointer focus:bg-selected focus:text-selected-foreground"
          onClick={handleProfile}
        >
          <CircleUserRound className="mr-2 size-4" />
          <span>个人资料</span>
        </DropdownMenuItem>

        {/* Settings */}
        <DropdownMenuItem
          className="cursor-pointer focus:bg-selected focus:text-selected-foreground"
          onClick={handleSettings}
        >
          <Settings className="mr-2 size-4" />
          <span>设置</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* Logout */}
        <DropdownMenuItem
          className="cursor-pointer focus:bg-selected focus:text-selected-foreground"
          onClick={onLogout}
        >
          <LogOut className="mr-2 size-4" />
          <span>退出登录</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
