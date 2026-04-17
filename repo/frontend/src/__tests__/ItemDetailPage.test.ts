import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockGet = vi.fn()
const mockPost = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

const mockItemsGetById = vi.fn()
vi.mock('@/api/endpoints/items', () => ({
  itemsApi: { getById: (...args: any[]) => mockItemsGetById(...args), list: vi.fn() },
}))

const mockReviewsList = vi.fn()
vi.mock('@/api/endpoints/reviews', () => ({
  reviewsApi: { listByItem: (...args: any[]) => mockReviewsList(...args) },
}))

const stubs = { RouterLink: { template: '<a><slot /></a>' } }
import ItemDetailPage from '@/pages/catalog/ItemDetailPage.vue'

describe('ItemDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(globalThis as any).alert = vi.fn()
    // Default: experiment assignment not available
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/experiments/assignment')) return Promise.reject(new Error('no assignment'))
      if (url.includes('/questions')) return Promise.resolve({ data: { data: [] } })
      return Promise.resolve({ data: { data: [] } })
    })
  })

  it('renders item title, category, and description on mount', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 'Coffee Maker', category: 'Kitchen', description: 'A good one.' } })
    mockReviewsList.mockResolvedValue({ data: { data: [] } })

    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()

    expect(wrapper.find('h1').text()).toBe('Coffee Maker')
    expect(wrapper.text()).toContain('Kitchen')
    expect(wrapper.text()).toContain('A good one.')
    expect(mockItemsGetById).toHaveBeenCalledWith('i1')
    expect(mockReviewsList).toHaveBeenCalledWith('i1', {})
  })

  it('renders reviews with star ratings', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 't', category: '', description: '' } })
    mockReviewsList.mockResolvedValue({
      data: { data: [{ id: 'r1', rating: 4, body: 'Nice', created_at: '2026-01-01' }] },
    })

    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()

    const review = wrapper.find('.review-card')
    expect(review.exists()).toBe(true)
    expect(review.find('.review-rating').text()).toBe('★★★★☆')
    expect(review.text()).toContain('Nice')
  })

  it('shows "No reviews yet" when reviews list is empty', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 't', category: '', description: '' } })
    mockReviewsList.mockResolvedValue({ data: { data: [] } })

    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()
    expect(wrapper.text()).toContain('No reviews yet')
  })

  it('switches to Q&A tab and fetches questions', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 't', category: '', description: '' } })
    mockReviewsList.mockResolvedValue({ data: { data: [] } })
    mockGet.mockImplementation((url: string) => {
      if (url === '/items/i1/questions') return Promise.resolve({ data: { data: [{ id: 'q1', body: 'Is it shiny?', created_at: '2026-02-01' }] } })
      if (url.includes('/experiments/assignment')) return Promise.reject(new Error('no'))
      return Promise.resolve({ data: { data: [] } })
    })

    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()

    const tabButtons = wrapper.findAll('.tabs button')
    await tabButtons[1].trigger('click')
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/items/i1/questions')
    expect(wrapper.text()).toContain('Is it shiny?')
  })

  it('disables "Submit Question" button until text is entered', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 't', category: '', description: '' } })
    mockReviewsList.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()
    const tabButtons = wrapper.findAll('.tabs button')
    await tabButtons[1].trigger('click')
    await flushPromises()

    const submitBtn = wrapper.findAll('.ask-form button').find(b => b.text() === 'Submit Question')!
    expect((submitBtn.element as HTMLButtonElement).disabled).toBe(true)

    await wrapper.find('.ask-form textarea').setValue('Is it durable?')
    expect((submitBtn.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('submits a question via POST and clears the input', async () => {
    mockItemsGetById.mockResolvedValue({ data: { id: 'i1', title: 't', category: '', description: '' } })
    mockReviewsList.mockResolvedValue({ data: { data: [] } })
    mockPost.mockResolvedValue({ data: {} })

    const wrapper = mount(ItemDetailPage, { props: { id: 'i1' }, global: { stubs } })
    await flushPromises()
    await wrapper.findAll('.tabs button')[1].trigger('click')
    await flushPromises()

    await wrapper.find('.ask-form textarea').setValue('Is it durable?')
    await wrapper.findAll('.ask-form button').find(b => b.text() === 'Submit Question')!.trigger('click')
    await flushPromises()

    expect(mockPost).toHaveBeenCalledWith('/items/i1/questions', { body: 'Is it durable?' })
    expect((wrapper.find('.ask-form textarea').element as HTMLTextAreaElement).value).toBe('')
  })
})
