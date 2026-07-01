/**
 * Centralized Memberships Cache Management
 *
 * Provides a single source of truth for user memberships with:
 * - In-memory cache with TTL (Time To Live)
 * - localStorage persistence
 * - Singleton request pattern (prevents duplicate API calls)
 * - Automatic cache invalidation
 *
 * Usage:
 * ```ts
 * import { getMemberships, invalidateMembershipsCache } from '@web/cache/memberships-cache'
 *
 * // Get memberships (automatically fetches if needed)
 * const memberships = await getMemberships(token)
 *
 * // Force refresh
 * await getMemberships(token, true)
 *
 * // Clear cache
 * invalidateMembershipsCache()
 * ```
 */

export interface MembershipInfo {
  orgId: string
  orgName: string
  displayName: string
  role: string
  joinedAt: string
}

interface CachedMemberships {
  data: MembershipInfo[]
  timestamp: number
  expiresAt: number
}

// Cache configuration
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes
const CACHE_KEY = 'org_memberships_cache'
const TIMESTAMP_KEY = 'org_memberships_timestamp'

// In-memory cache (faster than localStorage)
let memoryCache: CachedMemberships | null = null

// Singleton request promise (prevents duplicate concurrent requests)
let ongoingRequest: Promise<MembershipInfo[]> | null = null

/**
 * Check if cached data is still valid
 */
function isCacheValid(cached: CachedMemberships | null): boolean {
  if (!cached) return false
  return Date.now() < cached.expiresAt
}

/**
 * Load cache from localStorage
 */
function loadFromLocalStorage(): CachedMemberships | null {
  try {
    const dataStr = localStorage.getItem(CACHE_KEY)
    const timestampStr = localStorage.getItem(TIMESTAMP_KEY)

    if (!dataStr || !timestampStr) return null

    const data = JSON.parse(dataStr) as MembershipInfo[]
    const timestamp = parseInt(timestampStr, 10)
    const expiresAt = timestamp + CACHE_TTL

    const cached: CachedMemberships = { data, timestamp, expiresAt }

    // Validate cache
    if (!isCacheValid(cached)) {
      // Cache expired, clean up
      localStorage.removeItem(CACHE_KEY)
      localStorage.removeItem(TIMESTAMP_KEY)
      return null
    }

    return cached
  } catch (error) {
    console.error('[MembershipsCache] Failed to load from localStorage:', error)
    return null
  }
}

/**
 * Save cache to localStorage
 */
function saveToLocalStorage(data: MembershipInfo[]): void {
  try {
    const timestamp = Date.now()
    localStorage.setItem(CACHE_KEY, JSON.stringify(data))
    localStorage.setItem(TIMESTAMP_KEY, timestamp.toString())
  } catch (error) {
    console.error('[MembershipsCache] Failed to save to localStorage:', error)
  }
}

/**
 * Fetch memberships from whoami API
 */
async function fetchMembershipsFromAPI(token: string): Promise<MembershipInfo[]> {
  const response = await fetch(`/api/auth/whoami`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch memberships: ${response.status} ${response.statusText}`)
  }

  const data = (await response.json()) as { memberships?: Array<Record<string, unknown>> }
  const rawMemberships = data.memberships ?? []
  const memberships: MembershipInfo[] = rawMemberships.map((m) => ({
    orgId: String(m.orgId),
    orgName: String(m.orgName),
    displayName: String(m.displayName),
    role: String(m.role),
    joinedAt: (() => { try { const d = new Date(String(m.joinedAt)); return isNaN(d.getTime()) ? '' : d.toISOString() } catch { return '' } })(),
  }))

  return memberships
}

/**
 * Get user memberships with intelligent caching
 *
 * @param token - Authentication token
 * @param forceRefresh - Force bypass cache and fetch fresh data
 * @returns Promise<MembershipInfo[]>
 */
export async function getMemberships(
  token: string | null,
  forceRefresh = false
): Promise<MembershipInfo[]> {
  if (!token) {
    throw new Error('No authentication token provided')
  }

  // If force refresh, clear cache
  if (forceRefresh) {
    invalidateMembershipsCache()
  }

  // Check memory cache first (fastest)
  if (!forceRefresh && memoryCache && isCacheValid(memoryCache)) {
    console.log('[MembershipsCache] Using memory cache')
    return memoryCache.data
  }

  // Check localStorage cache (faster than network)
  if (!forceRefresh) {
    const localCache = loadFromLocalStorage()
    if (localCache && isCacheValid(localCache)) {
      console.log('[MembershipsCache] Using localStorage cache')
      // Update memory cache
      memoryCache = localCache
      return localCache.data
    }
  }

  // Prevent duplicate concurrent requests (singleton pattern)
  if (ongoingRequest) {
    console.log('[MembershipsCache] Reusing ongoing request')
    return ongoingRequest
  }

  // Fetch from API
  console.log('[MembershipsCache] Fetching from API')
  ongoingRequest = fetchMembershipsFromAPI(token)
    .then((data) => {
      // Update caches
      const timestamp = Date.now()
      const cached: CachedMemberships = {
        data,
        timestamp,
        expiresAt: timestamp + CACHE_TTL,
      }

      memoryCache = cached
      saveToLocalStorage(data)

      return data
    })
    .finally(() => {
      // Clear ongoing request
      ongoingRequest = null
    })

  return ongoingRequest
}

/**
 * Invalidate all caches (memory + localStorage)
 *
 * Call this when:
 * - User logs out
 * - User creates/joins/leaves an organization
 * - Manual refresh is needed
 */
export function invalidateMembershipsCache(): void {
  console.log('[MembershipsCache] Invalidating cache')
  memoryCache = null
  ongoingRequest = null

  try {
    localStorage.removeItem(CACHE_KEY)
    localStorage.removeItem(TIMESTAMP_KEY)
  } catch (error) {
    console.error('[MembershipsCache] Failed to clear localStorage:', error)
  }
}

/**
 * Get cached memberships synchronously (without fetching)
 *
 * Returns null if no valid cache exists.
 * Useful for components that want to show cached data immediately.
 */
export function getCachedMemberships(): MembershipInfo[] | null {
  // Check memory cache
  if (memoryCache && isCacheValid(memoryCache)) {
    return memoryCache.data
  }

  // Check localStorage
  const localCache = loadFromLocalStorage()
  if (localCache && isCacheValid(localCache)) {
    memoryCache = localCache
    return localCache.data
  }

  return null
}

/**
 * Check if memberships cache is available
 */
export function hasCachedMemberships(): boolean {
  return getCachedMemberships() !== null
}

/**
 * Preload memberships (useful for login page)
 *
 * Starts fetching in the background without waiting for result.
 * Subsequent calls to getMemberships() will reuse this request.
 */
export function preloadMemberships(token: string): void {
  if (!token) return

  // Only preload if no valid cache exists
  if (hasCachedMemberships()) {
    console.log('[MembershipsCache] Preload skipped - valid cache exists')
    return
  }

  console.log('[MembershipsCache] Preloading memberships')
  getMemberships(token).catch((error) => {
    console.error('[MembershipsCache] Preload failed:', error)
  })
}
