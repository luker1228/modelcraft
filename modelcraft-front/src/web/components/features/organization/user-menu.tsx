'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { Button } from '@web/components/ui/button'
import { Avatar, AvatarFallback } from '@web/components/ui/avatar'
import {
  getUserInfoFromToken,
  getToken,
  removeToken,
  isAuthenticated,
} from '@bff/auth/public'
import type { AuthUser } from '@/types/auth'
import { useOrganizationStore } from '@shared/stores/organization'
import { useCurrentOrg } from '@web/stores'
import { LogOut, Settings } from 'lucide-react'

export function UserMenu() {
  const router = useRouter()
  const [userInfo, setUserInfo] = useState<AuthUser | null>(null)
  const clearOrganization = useOrganizationStore((state) => state.clearOrganization)
  const currentOrg = useCurrentOrg()

  useEffect(() => {
    // Load user info from token
    if (isAuthenticated()) {
      const token = getToken()
      if (token) {
        const info = getUserInfoFromToken(token)
        setUserInfo(info)
      }
    }
  }, [])

  const handleLogout = async () => {
    try {
      // Call backend logout endpoint (optional)
      await fetch('/api/auth/logout', { method: 'POST' })
    } catch (error) {
      console.error('Logout API call failed:', error)
    }

    // Clear local storage
    removeToken()
    clearOrganization()

    // Redirect to login page
    router.push('/login')
  }

  if (!userInfo) {
    return null
  }

  // Get initials for avatar
  const initials = userInfo.name
    ? userInfo.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : (userInfo.phone || 'U')[0].toUpperCase()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="relative size-10 rounded-full">
          <Avatar className="size-10">
            <AvatarFallback className="bg-primary/10 text-sm font-semibold text-primary">
              {initials}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="end" forceMount>
        <DropdownMenuLabel className="font-normal">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-semibold leading-none">{userInfo.name}</p>
            <p className="text-xs leading-none text-muted-foreground">
              {userInfo.phone}
            </p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {currentOrg && (
          <DropdownMenuItem onClick={() => router.push(`/org/${currentOrg}/settings`)}>
            <Settings className="mr-2 size-4" />
            <span>Settings</span>
          </DropdownMenuItem>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleLogout} className="text-red-600">
          <LogOut className="mr-2 size-4" />
          <span>Log out</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
