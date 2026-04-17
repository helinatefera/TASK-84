import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockDelete = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: vi.fn(),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

import FavoritesPage from '@/pages/favorites/FavoritesPage.vue'

describe('FavoritesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads favorites and wishlists on mount', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/favorites') return Promise.resolve({ data: { data: [{ item_id: 1 }, { item_id: 2 }] } })
      if (url === '/wishlists') return Promise.resolve({ data: { data: [{ id: 'w1', name: 'Reading list' }] } })
      return Promise.resolve({ data: {} })
    })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/favorites')
    expect(mockGet).toHaveBeenCalledWith('/wishlists')
    const favItems = wrapper.findAll('.fav-item')
    expect(favItems).toHaveLength(2)
    expect(favItems[0].text()).toContain('Item #1')
  })

  it('shows "No favorites yet" when list is empty', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(FavoritesPage)
    await flushPromises()
    expect(wrapper.text()).toContain('No favorites yet')
  })

  it('switches to wishlists tab and renders wishlists', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/favorites') return Promise.resolve({ data: { data: [] } })
      if (url === '/wishlists') return Promise.resolve({ data: { data: [{ id: 'w1', name: 'Reading list' }] } })
      return Promise.resolve({ data: {} })
    })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    const tabs = wrapper.findAll('.tabs button')
    await tabs[1].trigger('click') // Wishlists tab
    await flushPromises()

    expect(wrapper.text()).toContain('Reading list')
    expect(tabs[1].classes()).toContain('active')
  })

  it('removes a favorite and refetches the list', async () => {
    mockGet.mockImplementation((url: string) =>
      url === '/favorites'
        ? Promise.resolve({ data: { data: [{ item_id: 42 }] } })
        : Promise.resolve({ data: { data: [] } })
    )
    mockDelete.mockResolvedValue({ data: {} })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    await wrapper.find('.fav-item .btn-danger').trigger('click')
    await flushPromises()

    expect(mockDelete).toHaveBeenCalledWith('/favorites/42')
    // Called once on mount, once after removal.
    expect(mockGet.mock.calls.filter(c => c[0] === '/favorites').length).toBe(2)
  })

  it('creates a wishlist and clears the input', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    mockPost.mockResolvedValue({ data: { id: 'new-w', name: 'Gifts' } })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    // Switch to wishlists tab
    const tabs = wrapper.findAll('.tabs button')
    await tabs[1].trigger('click')
    await flushPromises()

    await wrapper.find('input.form-input').setValue('Gifts')
    await wrapper.find('.btn-primary').trigger('click')
    await flushPromises()

    expect(mockPost).toHaveBeenCalledWith('/wishlists', { name: 'Gifts' })
    expect((wrapper.find('input.form-input').element as HTMLInputElement).value).toBe('')
  })

  it('disables Create button when name is empty', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    const tabs = wrapper.findAll('.tabs button')
    await tabs[1].trigger('click')
    await flushPromises()

    const btn = wrapper.find('.btn-primary')
    expect((btn.element as HTMLButtonElement).disabled).toBe(true)

    await wrapper.find('input.form-input').setValue('Name')
    expect((btn.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('deletes a wishlist and refetches', async () => {
    mockGet.mockImplementation((url: string) =>
      url === '/wishlists'
        ? Promise.resolve({ data: { data: [{ id: 'w1', name: 'x' }] } })
        : Promise.resolve({ data: { data: [] } })
    )
    mockDelete.mockResolvedValue({ data: {} })
    const wrapper = mount(FavoritesPage)
    await flushPromises()

    const tabs = wrapper.findAll('.tabs button')
    await tabs[1].trigger('click')
    await flushPromises()

    await wrapper.find('.wishlist-item .btn-danger').trigger('click')
    await flushPromises()
    expect(mockDelete).toHaveBeenCalledWith('/wishlists/w1')
  })
})
