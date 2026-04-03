// Re-export from shared location accessible by both BFF and Web layers
export {
  getMemberships,
  invalidateMembershipsCache,
  getCachedMemberships,
  hasCachedMemberships,
  preloadMemberships,
} from '@/shared/cache/memberships-cache'
export type { MembershipInfo } from '@/shared/cache/memberships-cache'
