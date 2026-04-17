import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUIStore } from '@/stores/ui.store'

describe('UI store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('starts with no toasts and loading off', () => {
    const store = useUIStore()
    expect(store.toasts).toEqual([])
    expect(store.isGlobalLoading).toBe(false)
  })

  it('showToast appends a toast with a unique id and default type info', () => {
    const store = useUIStore()
    store.showToast('hello')
    store.showToast('bye', 'error')

    expect(store.toasts).toHaveLength(2)
    expect(store.toasts[0].message).toBe('hello')
    expect(store.toasts[0].type).toBe('info')
    expect(store.toasts[1].type).toBe('error')
    expect(store.toasts[0].id).not.toBe(store.toasts[1].id)
  })

  it('dismissToast removes a toast by id', () => {
    const store = useUIStore()
    store.showToast('one')
    store.showToast('two')
    const firstId = store.toasts[0].id

    store.dismissToast(firstId)
    expect(store.toasts).toHaveLength(1)
    expect(store.toasts[0].message).toBe('two')
  })

  it('toast auto-dismisses after 5 seconds', () => {
    const store = useUIStore()
    store.showToast('temp')
    expect(store.toasts).toHaveLength(1)

    vi.advanceTimersByTime(4999)
    expect(store.toasts).toHaveLength(1)

    vi.advanceTimersByTime(1)
    expect(store.toasts).toHaveLength(0)
  })

  it('setLoading toggles the global loading flag', () => {
    const store = useUIStore()
    store.setLoading(true)
    expect(store.isGlobalLoading).toBe(true)
    store.setLoading(false)
    expect(store.isGlobalLoading).toBe(false)
  })
})
