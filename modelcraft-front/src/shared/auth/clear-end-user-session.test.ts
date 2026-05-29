import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { clearEndUserSessionArtifacts } from './clear-end-user-session'

const createStorageMock = () => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value
    },
    removeItem: (key: string) => {
      delete store[key]
    },
    clear: () => {
      store = {}
    },
    key: (index: number) => Object.keys(store)[index] ?? null,
    get length() {
      return Object.keys(store).length
    },
  }
}

const sessionStorageMock = createStorageMock()

vi.stubGlobal('sessionStorage', sessionStorageMock)

describe('clearEndUserSessionArtifacts', () => {
  beforeEach(() => {
    sessionStorage.clear()
    useEndUserAuthStore.setState({
      accessToken: null,
      expiresAt: null,
      isAdmin: null,
      userInfo: null,
    })
  })

  it('clears end-user store state and sessionStorage artifacts', () => {
    useEndUserAuthStore.getState().setAccessToken('header.payload.signature', 3600)
    useEndUserAuthStore.getState().setUserInfo({
      id: 'eu_1',
      username: 'luke',
      orgName: 'luke4_n3j6',
      projectSlug: 'proj',
    })

    sessionStorage.setItem('eu_token_luke4_n3j6', 'token')
    sessionStorage.setItem('eu_token_expires_at_luke4_n3j6', '123')
    sessionStorage.setItem('eu_accessible_projects_luke4_n3j6', '[]')
    sessionStorage.setItem('unrelated_key', 'keep')

    clearEndUserSessionArtifacts()

    const state = useEndUserAuthStore.getState()
    expect(state.accessToken).toBeNull()
    expect(state.expiresAt).toBeNull()
    expect(state.userInfo).toBeNull()
    expect(sessionStorage.getItem('eu_token_luke4_n3j6')).toBeNull()
    expect(sessionStorage.getItem('eu_token_expires_at_luke4_n3j6')).toBeNull()
    expect(sessionStorage.getItem('eu_accessible_projects_luke4_n3j6')).toBeNull()
    expect(sessionStorage.getItem('unrelated_key')).toBe('keep')
  })
})
