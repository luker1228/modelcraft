import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { getMemberships, invalidateMembershipsCache, type MembershipInfo } from '@/shared/cache/memberships-cache'

interface OrganizationState {
  // Current selected organization
  currentOrg: string | null

  // All organizations user has access to
  organizations: string[]

  // Full membership data (includes roles, displayNames, etc.)
  memberships: MembershipInfo[]

  // Loading state
  isLoadingMemberships: boolean

  // Actions
  setCurrentOrg: (orgName: string) => void
  setOrganizations: (orgs: string[]) => void
  setMemberships: (memberships: MembershipInfo[]) => void
  switchOrganization: (orgName: string) => void
  clearOrganization: () => void

  // Async actions
  loadMemberships: (token: string, forceRefresh?: boolean) => Promise<MembershipInfo[]>
  refreshMemberships: (token: string) => Promise<MembershipInfo[]>
}

/**
 * Organization store using Zustand
 * Manages current organization context for multi-tenant features
 * Integrates with memberships-cache for optimized API calls
 */
export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set, get) => ({
      currentOrg: null,
      organizations: [],
      memberships: [],
      isLoadingMemberships: false,

      setCurrentOrg: (orgName: string) => {
        set({ currentOrg: orgName })
      },

      setOrganizations: (orgs: string[]) => {
        set({ organizations: orgs })

        // If no current org is selected, set the first one
        const { currentOrg } = get()
        if (!currentOrg && orgs.length > 0) {
          set({ currentOrg: orgs[0] })
        }
      },

      setMemberships: (memberships: MembershipInfo[]) => {
        const orgNames = memberships.map(m => m.orgName)
        set({
          memberships,
          organizations: orgNames,
        })

        // If no current org is selected, set the first one
        const { currentOrg } = get()
        if (!currentOrg && orgNames.length > 0) {
          set({ currentOrg: orgNames[0] })
        }
      },

      switchOrganization: (orgName: string) => {
        const { organizations } = get()

        // Verify user has access to this organization
        if (organizations.length > 0 && !organizations.includes(orgName)) {
          console.error(
            `User does not have access to organization: ${orgName}`
          )
          return
        }

        set({ currentOrg: orgName })

        // Navigation is handled by the OrganizationSwitcher component
        // which uses Next.js router to navigate to the new org URL
      },

      clearOrganization: () => {
        set({
          currentOrg: null,
          organizations: [],
          memberships: [],
        })
        // Clear memberships cache
        invalidateMembershipsCache()
      },

      // Load memberships with caching
      loadMemberships: async (token: string, forceRefresh = false) => {
        set({ isLoadingMemberships: true })

        try {
          const memberships = await getMemberships(token, forceRefresh)

          // Update store
          get().setMemberships(memberships)

          return memberships
        } catch (error) {
          console.error('[OrganizationStore] Failed to load memberships:', error)
          throw error
        } finally {
          set({ isLoadingMemberships: false })
        }
      },

      // Force refresh memberships
      refreshMemberships: async (token: string) => {
        console.log('[OrganizationStore] Refreshing memberships')
        return get().loadMemberships(token, true)
      },
    }),
    {
      name: 'organization-storage', // localStorage key
      // Only persist currentOrg and organizations (not memberships - they come from cache)
      partialize: (state) => ({
        currentOrg: state.currentOrg,
        organizations: state.organizations,
      }),
    }
  )
)

/**
 * Hook to get current organization name
 */
export function useCurrentOrg(): string | null {
  return useOrganizationStore((state) => state.currentOrg)
}

/**
 * Hook to get all organizations
 */
export function useOrganizations(): string[] {
  return useOrganizationStore((state) => state.organizations)
}

/**
 * Hook to get full membership data
 */
export function useMemberships(): MembershipInfo[] {
  return useOrganizationStore((state) => state.memberships)
}

/**
 * Hook to get memberships loading state
 */
export function useIsMembershipsLoading(): boolean {
  return useOrganizationStore((state) => state.isLoadingMemberships)
}

/**
 * Hook to switch organization
 */
export function useSwitchOrganization() {
  return useOrganizationStore((state) => state.switchOrganization)
}

/**
 * Hook to load memberships
 */
export function useLoadMemberships() {
  return useOrganizationStore((state) => state.loadMemberships)
}

/**
 * Hook to refresh memberships
 */
export function useRefreshMemberships() {
  return useOrganizationStore((state) => state.refreshMemberships)
}
