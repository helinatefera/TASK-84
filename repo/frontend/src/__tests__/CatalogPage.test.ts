import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockList = vi.fn()
vi.mock('@/api/endpoints/items', () => ({
  itemsApi: { list: (...args: any[]) => mockList(...args) },
}))

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
}))

import CatalogPage from '@/pages/catalog/CatalogPage.vue'

describe('CatalogPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  it('calls itemsApi.list on mount and renders items', async () => {
    mockList.mockResolvedValue({
      data: {
        data: [
          { id: '1', title: 'Book', category: 'Books', description: 'A short tale.' },
          { id: '2', title: 'Phone', category: 'Tech', description: 'x'.repeat(200) },
        ],
      },
    })
    const wrapper = mount(CatalogPage)
    await flushPromises()

    expect(mockList).toHaveBeenCalledWith({ search: undefined })
    const cards = wrapper.findAll('.item-card')
    expect(cards).toHaveLength(2)
    expect(cards[0].text()).toContain('Book')
    expect(cards[0].text()).toContain('Books')
    // Long descriptions are truncated with an ellipsis.
    expect(cards[1].text()).toContain('...')
  })

  it('renders "No items found" when list is empty', async () => {
    mockList.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(CatalogPage)
    await flushPromises()
    expect(wrapper.text()).toContain('No items found')
  })

  it('shows "Uncategorized" when an item has no category', async () => {
    mockList.mockResolvedValue({ data: { data: [{ id: '1', title: 't', description: 'd' }] } })
    const wrapper = mount(CatalogPage)
    await flushPromises()
    expect(wrapper.text()).toContain('Uncategorized')
  })

  it('navigates to itemDetail when a card is clicked', async () => {
    mockList.mockResolvedValue({ data: { data: [{ id: 'abc', title: 't', description: '' }] } })
    const wrapper = mount(CatalogPage)
    await flushPromises()
    await wrapper.find('.item-card').trigger('click')
    expect(mockPush).toHaveBeenCalledWith({ name: 'itemDetail', params: { id: 'abc' } })
  })

  it('debounces search input and only fetches after 300ms', async () => {
    mockList.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(CatalogPage)
    await flushPromises()
    mockList.mockClear()

    const input = wrapper.find('.search-input')
    await input.setValue('book')
    await input.trigger('input')
    expect(mockList).not.toHaveBeenCalled()

    vi.advanceTimersByTime(299)
    expect(mockList).not.toHaveBeenCalled()

    vi.advanceTimersByTime(1)
    await flushPromises()
    expect(mockList).toHaveBeenCalledWith({ search: 'book' })
  })

  it('handles fetch errors gracefully', async () => {
    mockList.mockRejectedValue(new Error('network'))
    const wrapper = mount(CatalogPage)
    await flushPromises()
    expect(wrapper.text()).toContain('No items found')
  })
})
