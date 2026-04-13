import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'

// --- localStorage mock ---
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} }),
    get length() { return Object.keys(store).length },
    key: vi.fn((i: number) => Object.keys(store)[i] ?? null),
  }
})()
Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock, writable: true })

// --- Mocks ---
const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

vi.mock('@/api/endpoints/reviews', () => ({
  reviewsApi: {
    create: vi.fn().mockResolvedValue({ data: { id: 'review-uuid' } }),
  },
}))

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
}))

import WriteReviewPage from '@/pages/reviews/WriteReviewPage.vue'
import { reviewsApi } from '@/api/endpoints/reviews'

// --- Helper tests (kept from original) ---

describe('Review submit lock', () => {
  it('prevents double submission', async () => {
    const isLocked = ref(false)
    let submitCount = 0

    async function handleSubmit() {
      if (isLocked.value) return
      isLocked.value = true
      submitCount++
      await new Promise((r) => setTimeout(r, 10))
    }

    handleSubmit()
    expect(isLocked.value).toBe(true)
    await handleSubmit()
    expect(submitCount).toBe(1)
  })
})

describe('Review draft autosave', () => {
  const store: Record<string, string> = {}
  function setItem(key: string, value: string) { store[key] = value }
  function getItem(key: string): string | null { return store[key] ?? null }
  function removeItem(key: string) { delete store[key] }

  beforeEach(() => { for (const k of Object.keys(store)) delete store[k] })

  it('saves and restores draft', () => {
    const draftKey = 'review-draft-item-123'
    setItem(draftKey, JSON.stringify({ rating: 3, body: 'Decent' }))
    const draft = JSON.parse(getItem(draftKey)!)
    expect(draft.rating).toBe(3)
  })

  it('clears draft on successful submit', () => {
    const draftKey = 'review-draft-item-456'
    setItem(draftKey, JSON.stringify({ rating: 5, body: 'Amazing' }))
    removeItem(draftKey)
    expect(getItem(draftKey)).toBeNull()
  })
})

describe('Image upload ordering', () => {
  it('enforces max 6 images client-side', () => {
    const images: string[] = []
    for (let i = 0; i < 8; i++) {
      if (6 - images.length > 0) images.push(`image-${i}.jpg`)
    }
    expect(images.length).toBe(6)
  })
})

// --- Mounted component tests ---

describe('WriteReviewPage — mounted', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockPost.mockResolvedValue({ data: { image_id: 101 } })
  })

  it('renders star rating, textarea, file input, and disabled submit', async () => {
    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    expect(wrapper.find('.stars').exists()).toBe(true)
    expect(wrapper.findAll('.stars button')).toHaveLength(5)
    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.find('input[type="file"]').exists()).toBe(true)

    const submitBtn = wrapper.find('.submit-btn')
    expect(submitBtn.exists()).toBe(true)
    expect((submitBtn.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('enables submit button after selecting a rating', async () => {
    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    // Click the 4th star
    const stars = wrapper.findAll('.stars button')
    await stars[3].trigger('click')
    await flushPromises()

    const submitBtn = wrapper.find('.submit-btn')
    expect((submitBtn.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('shows character count for review body', async () => {
    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    const textarea = wrapper.find('textarea')
    await textarea.setValue('Great product!')
    await flushPromises()

    expect(wrapper.text()).toContain('14/2000')
  })

  it('submits review and navigates on success', async () => {
    ;(reviewsApi.create as any).mockResolvedValue({ data: { id: 'new-review-uuid' } })

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    // Select rating
    await wrapper.findAll('.stars button')[4].trigger('click')
    // Set body
    await wrapper.find('textarea').setValue('Excellent quality')
    // Submit
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(reviewsApi.create).toHaveBeenCalledWith(
      'test-item',
      expect.objectContaining({ rating: 5, body: 'Excellent quality' }),
      expect.any(String), // idempotency key
    )
    expect(mockPush).toHaveBeenCalledWith({ name: 'itemDetail', params: { id: 'test-item' } })
  })

  it('shows error message on failed submission', async () => {
    ;(reviewsApi.create as any).mockRejectedValue({
      response: { data: { msg: 'Content contains prohibited words' } },
    })

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    await wrapper.findAll('.stars button')[2].trigger('click')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    // Wait for the error to appear
    await new Promise(r => setTimeout(r, 10))
    await flushPromises()

    expect(wrapper.find('.error-text').exists()).toBe(true)
    expect(wrapper.text()).toContain('Content contains prohibited words')
  })

  it('locks submit button during submission', async () => {
    let resolveCreate: any
    ;(reviewsApi.create as any).mockImplementation(() => new Promise(r => { resolveCreate = r }))

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'test-item' },
    })

    await wrapper.findAll('.stars button')[0].trigger('click')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.find('.submit-btn').text()).toContain('Submitting')
    expect((wrapper.find('.submit-btn').element as HTMLButtonElement).disabled).toBe(true)

    resolveCreate({ data: { id: 'done' } })
    await flushPromises()
  })

  it('loads draft from localStorage on mount', async () => {
    localStorageMock.setItem('review-draft-draft-item', JSON.stringify({ rating: 4, body: 'Saved draft' }))

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'draft-item' },
    })

    await flushPromises()

    // Draft should be loaded — rating 4 means 4 filled stars
    const filledStars = wrapper.findAll('.stars button.filled')
    expect(filledStars.length).toBe(4)
    expect((wrapper.find('textarea').element as HTMLTextAreaElement).value).toBe('Saved draft')
  })

  it('saves draft to localStorage via autosave interval', async () => {
    vi.useFakeTimers()
    localStorageMock.clear()

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'autosave-item' },
    })

    await flushPromises()

    // Set some content
    await wrapper.findAll('.stars button')[2].trigger('click')
    await wrapper.find('textarea').setValue('Autosave test body')
    await flushPromises()

    // Advance past the 10-second autosave interval
    vi.advanceTimersByTime(11000)

    expect(localStorageMock.setItem).toHaveBeenCalled()
    const savedCalls = localStorageMock.setItem.mock.calls.filter(
      (c: any[]) => c[0] === 'review-draft-autosave-item'
    )
    expect(savedCalls.length).toBeGreaterThanOrEqual(1)
    const parsed = JSON.parse(savedCalls[savedCalls.length - 1][1])
    expect(parsed.rating).toBe(3)
    expect(parsed.body).toBe('Autosave test body')

    vi.useRealTimers()
  })

  it('clears draft from localStorage after successful submit', async () => {
    localStorageMock.setItem('review-draft-clear-item', JSON.stringify({ rating: 5, body: 'Will clear' }))
    ;(reviewsApi.create as any).mockResolvedValue({ data: { id: 'new-review' } })

    const wrapper = mount(WriteReviewPage, {
      props: { itemId: 'clear-item' },
    })

    await flushPromises()
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(localStorageMock.removeItem).toHaveBeenCalledWith('review-draft-clear-item')
  })
})
