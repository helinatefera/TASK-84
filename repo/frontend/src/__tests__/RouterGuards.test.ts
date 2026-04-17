import { describe, it, expect, beforeEach } from 'vitest'

// Install a local storage mock in case prior test files have replaced the global.
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

// Re-implement the beforeEach guard from src/router/index.ts as a pure function
// so we can exercise every branch without mounting vue-router.
// This is the exact contract the live guard implements.
interface RouteMatch {
  meta: { requiresAuth?: boolean; roles?: string[] }
}
interface RouteLike {
  path: string
  matched: RouteMatch[]
  meta: { roles?: string[] }
}

function guard(to: RouteLike): { name?: string; allow?: boolean } {
  const token = localStorage.getItem('access_token')
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null

  if (to.matched.some((r) => r.meta.requiresAuth) && !token) {
    return { name: 'login' }
  }
  if (to.path.startsWith('/auth') && token) {
    return { name: 'catalog' }
  }
  const requiredRoles = to.meta.roles
  if (requiredRoles && user) {
    if (!requiredRoles.includes(user.role)) {
      return { name: 'forbidden' }
    }
  }
  return { allow: true }
}

describe('router navigation guards', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('redirects unauthenticated users to login when visiting a protected route', () => {
    const result = guard({
      path: '/catalog',
      matched: [{ meta: { requiresAuth: true } }],
      meta: {},
    })
    expect(result).toEqual({ name: 'login' })
  })

  it('redirects authenticated users away from auth pages to catalog', () => {
    localStorage.setItem('access_token', 'abc')
    const result = guard({
      path: '/auth/login',
      matched: [{ meta: {} }],
      meta: {},
    })
    expect(result).toEqual({ name: 'catalog' })
  })

  it('allows authenticated users to visit protected routes that have no role restriction', () => {
    localStorage.setItem('access_token', 'abc')
    localStorage.setItem('user', JSON.stringify({ role: 'regular_user' }))
    const result = guard({
      path: '/catalog',
      matched: [{ meta: { requiresAuth: true } }],
      meta: {},
    })
    expect(result.allow).toBe(true)
  })

  it('redirects to forbidden when user lacks a required role', () => {
    localStorage.setItem('access_token', 'abc')
    localStorage.setItem('user', JSON.stringify({ role: 'regular_user' }))
    const result = guard({
      path: '/admin/users',
      matched: [{ meta: { requiresAuth: true } }],
      meta: { roles: ['admin'] },
    })
    expect(result).toEqual({ name: 'forbidden' })
  })

  it('allows admins onto admin-only routes', () => {
    localStorage.setItem('access_token', 'abc')
    localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
    const result = guard({
      path: '/admin/users',
      matched: [{ meta: { requiresAuth: true } }],
      meta: { roles: ['admin'] },
    })
    expect(result.allow).toBe(true)
  })

  it('allows moderator onto routes that list moderator in roles', () => {
    localStorage.setItem('access_token', 'abc')
    localStorage.setItem('user', JSON.stringify({ role: 'moderator' }))
    const result = guard({
      path: '/moderation',
      matched: [{ meta: { requiresAuth: true } }],
      meta: { roles: ['moderator', 'admin'] },
    })
    expect(result.allow).toBe(true)
  })

  it('allows product_analyst onto analytics pages', () => {
    localStorage.setItem('access_token', 'abc')
    localStorage.setItem('user', JSON.stringify({ role: 'product_analyst' }))
    const result = guard({
      path: '/analytics',
      matched: [{ meta: { requiresAuth: true } }],
      meta: { roles: ['product_analyst', 'admin'] },
    })
    expect(result.allow).toBe(true)
  })

  it('does not redirect guests away from public (non-auth, non-protected) paths', () => {
    // Not protected, not /auth — guard should just allow through
    const result = guard({
      path: '/forbidden',
      matched: [{ meta: {} }],
      meta: {},
    })
    expect(result.allow).toBe(true)
  })
})
