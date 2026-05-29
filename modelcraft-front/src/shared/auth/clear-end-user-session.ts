import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

const END_USER_SESSION_PREFIXES = [
  'eu_token_',
  'eu_token_expires_at_',
  'eu_accessible_projects_',
]

export function clearEndUserSessionArtifacts(): void {
  useEndUserAuthStore.getState().clearSession()

  if (typeof sessionStorage === 'undefined') {
    return
  }

  const keysToRemove: string[] = []
  for (let i = 0; i < sessionStorage.length; i += 1) {
    const key = sessionStorage.key(i)
    if (key && END_USER_SESSION_PREFIXES.some((prefix) => key.startsWith(prefix))) {
      keysToRemove.push(key)
    }
  }

  keysToRemove.forEach((key) => sessionStorage.removeItem(key))
}
