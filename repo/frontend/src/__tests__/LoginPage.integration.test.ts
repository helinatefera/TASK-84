// Integration-style test for LoginPage: uses the REAL Pinia store and a
// REAL vue-router instance. Only the HTTP transport (apiClient) is faked,
// so the store's state transitions, router navigation, and component
// interactions are all exercised end-to-end in jsdom.

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

// Local storage mock — jsdom's default was stripped by another test file.
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

// Fake apiClient only — everything else (store, router, component) is real.
const mockGet = vi.fn()
const mockPost = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...a: any[]) => mockGet(...a),
    post: (...a: any[]) => mockPost(...a),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))
// The endpoint modules under @/api/endpoints are thin wrappers over apiClient;
// mock them too so the store's login / register functions route through our
// fakes. We intentionally do NOT mock @/stores/auth.store — that's the whole
// point of this integration test.
vi.mock('@/api/endpoints/auth', () => ({
  authApi: {
    login: (payload: any) => mockPost('/auth/login', payload).then((r: any) => ({ data: r.data ?? r })),
    register: (payload: any) => mockPost('/auth/register', payload).then((r: any) => ({ data: r.data ?? r })),
    logout: (refreshToken: string) => mockPost('/auth/logout', { refresh_token: refreshToken }),
    refresh: (refreshToken: string) => mockPost('/auth/refresh', { refresh_token: refreshToken }),
  },
}))

import LoginPage from '@/pages/auth/LoginPage.vue'
import { useAuthStore } from '@/stores/auth.store'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>home</div>' } },
      { path: '/catalog', name: 'catalog', component: { template: '<div>catalog</div>' } },
      { path: '/auth/login', name: 'login', component: LoginPage },
      { path: '/auth/register', name: 'register', component: { template: '<div>register</div>' } },
    ],
  })
}

describe('LoginPage (integration with real store + router)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    sessionStorage.clear()
    setActivePinia(createPinia())
  })

  it('login success updates the real auth store and navigates to catalog', async () => {
    mockPost.mockResolvedValueOnce({
      data: {
        access_token: 'real-access',
        refresh_token: 'real-refresh',
        user: { id: '1', username: 'alice', email: 'a@a', role: 'regular_user', is_active: true },
      },
    })

    const router = makeRouter()
    await router.push('/auth/login')
    await router.isReady()

    const wrapper = mount(LoginPage, { global: { plugins: [router] } })

    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('Secret1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    // Real store side effects:
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.userRole).toBe('regular_user')
    expect(localStorage.getItem('access_token')).toBe('real-access')
    expect(localStorage.getItem('refresh_token')).toBe('real-refresh')

    // Real router navigation:
    expect(router.currentRoute.value.name).toBe('catalog')
  })

  it('login failure increments loginFailureCount in the real store', async () => {
    mockPost.mockRejectedValueOnce({ response: { data: { msg: 'Invalid credentials' } } })

    const router = makeRouter()
    await router.push('/auth/login')
    await router.isReady()

    const wrapper = mount(LoginPage, { global: { plugins: [router] } })
    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('wrong')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const auth = useAuthStore()
    expect(auth.loginFailureCount).toBe(1)
    expect(auth.isAuthenticated).toBe(false)
    expect(sessionStorage.getItem('login_failures')).toBe('1')
    expect(wrapper.find('.error-text').text()).toBe('Invalid credentials')
    // Router stays on login page, not redirected.
    expect(router.currentRoute.value.name).toBe('login')
  })

  it('shows CAPTCHA field once the real store hits 5 failures', async () => {
    // Seed failure count via the real store before mount.
    setActivePinia(createPinia())
    const auth = useAuthStore()
    for (let i = 0; i < 5; i++) auth.incrementLoginFailures()
    expect(auth.loginFailureCount).toBe(5)

    mockGet.mockResolvedValue({ data: { captcha_id: 'c-1', captcha_image: 'data:image/png;base64,xxx' } })

    const router = makeRouter()
    await router.push('/auth/login')
    await router.isReady()

    const wrapper = mount(LoginPage, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.text()).toContain('CAPTCHA')
    expect(mockGet).toHaveBeenCalledWith('/captcha/generate')
  })

  it('login success resets loginFailureCount to 0', async () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.incrementLoginFailures()
    auth.incrementLoginFailures()
    expect(auth.loginFailureCount).toBe(2)

    mockPost.mockResolvedValueOnce({
      data: {
        access_token: 'acc', refresh_token: 'ref',
        user: { id: '1', username: 'alice', email: 'a@a', role: 'regular_user', is_active: true },
      },
    })

    const router = makeRouter()
    await router.push('/auth/login')
    await router.isReady()

    const wrapper = mount(LoginPage, { global: { plugins: [router] } })
    await wrapper.find('input[type="text"]').setValue('alice')
    await wrapper.find('input[type="password"]').setValue('Secret1')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(auth.loginFailureCount).toBe(0)
    expect(sessionStorage.getItem('login_failures')).toBe('0')
  })

  it('logout via the real store clears tokens and user', async () => {
    setActivePinia(createPinia())
    mockPost.mockResolvedValueOnce({
      data: {
        access_token: 'acc', refresh_token: 'ref',
        user: { id: '1', username: 'alice', email: 'a@a', role: 'regular_user', is_active: true },
      },
    })
    mockPost.mockResolvedValueOnce({ data: {} }) // logout
    const auth = useAuthStore()

    await auth.login('alice', 'Secret1')
    expect(auth.isAuthenticated).toBe(true)
    expect(localStorage.getItem('access_token')).toBe('acc')

    await auth.logout()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
    expect(localStorage.getItem('access_token')).toBeNull()
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })
})
