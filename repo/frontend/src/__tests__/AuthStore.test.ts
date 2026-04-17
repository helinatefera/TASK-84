import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Ensure localStorage / sessionStorage exist even when other test files have
// replaced the globals. We install our own keyed-store mocks per file.
function makeStorageMock() {
  let store: Record<string, string> = {}
  return {
    getItem: (k: string) => (k in store ? store[k] : null),
    setItem: (k: string, v: string) => { store[k] = v },
    removeItem: (k: string) => { delete store[k] },
    clear: () => { store = {} },
    get length() { return Object.keys(store).length },
    key: (i: number) => Object.keys(store)[i] ?? null,
  }
}
Object.defineProperty(globalThis, 'localStorage', { value: makeStorageMock(), writable: true, configurable: true })
Object.defineProperty(globalThis, 'sessionStorage', { value: makeStorageMock(), writable: true, configurable: true })

// Mock the auth API so the store tests stay pure-logic (no network).
vi.mock('@/api/endpoints/auth', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
}))

import { useAuthStore } from '@/stores/auth.store'
import { authApi } from '@/api/endpoints/auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    sessionStorage.clear()
    vi.clearAllMocks()
  })

  it('starts unauthenticated when no token is in localStorage', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.userRole).toBe('')
  })

  it('login persists tokens and user, and resets failure counter', async () => {
    const userPayload = {
      id: '1',
      username: 'alice',
      email: 'a@example.com',
      role: 'regular_user',
      is_active: true,
    }
    ;(authApi.login as any).mockResolvedValue({
      data: {
        access_token: 'acc',
        refresh_token: 'ref',
        user: userPayload,
      },
    })

    const store = useAuthStore()
    store.incrementLoginFailures()
    expect(store.loginFailureCount).toBe(1)

    await store.login('alice', 'pw')

    expect(store.token).toBe('acc')
    expect(store.refreshToken).toBe('ref')
    expect(store.user).toEqual(userPayload)
    expect(store.isAuthenticated).toBe(true)
    expect(store.userRole).toBe('regular_user')
    expect(localStorage.getItem('access_token')).toBe('acc')
    expect(localStorage.getItem('refresh_token')).toBe('ref')
    expect(store.loginFailureCount).toBe(0)
  })

  it('register forwards to the API without changing auth state', async () => {
    ;(authApi.register as any).mockResolvedValue({ data: {} })
    const store = useAuthStore()

    await store.register('bob', 'b@example.com', 'pw')

    expect(authApi.register).toHaveBeenCalledWith({
      username: 'bob',
      email: 'b@example.com',
      password: 'pw',
    })
    expect(store.isAuthenticated).toBe(false)
  })

  it('logout clears tokens and user, even if the API call fails', async () => {
    ;(authApi.login as any).mockResolvedValue({
      data: {
        access_token: 'acc',
        refresh_token: 'ref',
        user: { id: '1', username: 'x', email: 'x@x', role: 'regular_user', is_active: true },
      },
    })
    ;(authApi.logout as any).mockRejectedValue(new Error('network down'))

    const store = useAuthStore()
    await store.login('x', 'pw')
    expect(store.isAuthenticated).toBe(true)

    await store.logout()
    expect(store.token).toBeNull()
    expect(store.refreshToken).toBeNull()
    expect(store.user).toBeNull()
    expect(localStorage.getItem('access_token')).toBeNull()
  })

  it('hasRole matches any listed role', async () => {
    ;(authApi.login as any).mockResolvedValue({
      data: {
        access_token: 'a',
        refresh_token: 'r',
        user: { id: '1', username: 'mod', email: 'm@m', role: 'moderator', is_active: true },
      },
    })
    const store = useAuthStore()
    await store.login('mod', 'pw')

    expect(store.hasRole('moderator')).toBe(true)
    expect(store.hasRole('admin', 'moderator')).toBe(true)
    expect(store.hasRole('admin')).toBe(false)
  })

  it('incrementLoginFailures persists counter across store re-creation via sessionStorage', () => {
    const store1 = useAuthStore()
    store1.incrementLoginFailures()
    store1.incrementLoginFailures()
    expect(sessionStorage.getItem('login_failures')).toBe('2')

    // Re-create Pinia to simulate a fresh store load
    setActivePinia(createPinia())
    const store2 = useAuthStore()
    expect(store2.loginFailureCount).toBe(2)
  })
})
