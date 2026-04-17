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

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import ExperimentsListPage from '@/pages/experiments/ExperimentsListPage.vue'

describe('ExperimentsListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches experiments on mount', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    mount(ExperimentsListPage)
    await flushPromises()
    expect(mockGet).toHaveBeenCalledWith('/experiments')
  })

  it('renders experiment rows with status badges', async () => {
    mockGet.mockResolvedValue({
      data: {
        data: [
          { id: 'e1', name: 'Homepage Banner', status: 'running', created_at: '2026-01-15T00:00:00Z' },
          { id: 'e2', name: 'Checkout Test', status: 'draft', created_at: '2026-02-01T00:00:00Z' },
        ],
      },
    })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Homepage Banner')
    expect(rows[0].text()).toContain('running')
    expect(rows[0].text()).toContain('2026-01-15')
    expect(rows[1].text()).toContain('draft')
  })

  it('shows empty state when no experiments exist', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()
    expect(wrapper.text()).toContain('No experiments yet')
  })

  it('toggles the create form via the button', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    expect(wrapper.find('.create-form').exists()).toBe(false)
    await wrapper.find('.page-header .btn-primary').trigger('click')
    expect(wrapper.find('.create-form').exists()).toBe(true)
    expect(wrapper.find('.page-header .btn-primary').text()).toBe('Cancel')

    await wrapper.find('.page-header .btn-primary').trigger('click')
    expect(wrapper.find('.create-form').exists()).toBe(false)
  })

  it('rejects creation when variant traffic does not sum to 100', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    await wrapper.find('.page-header .btn-primary').trigger('click')
    // Change first variant percentage to 30 (second stays 50) → sum = 80.
    const pctInputs = wrapper.findAll('input[type="number"]')
    // first is min_sample_size, then variant percentages
    await pctInputs[1].setValue(30)
    await wrapper.find('.create-form .btn-primary').trigger('click')
    await flushPromises()

    expect(wrapper.find('.error-text').text()).toContain('Traffic must sum to 100%')
    expect(mockPost).not.toHaveBeenCalled()
  })

  it('posts experiment payload when traffic sums to 100', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    mockPost.mockResolvedValue({ data: { id: 'new' } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    await wrapper.find('.page-header .btn-primary').trigger('click')
    const textInputs = wrapper.findAll('input[type="text"], input:not([type])')
    // Name is the first text input
    await textInputs[0].setValue('My Experiment')

    await wrapper.find('.create-form .btn-primary').trigger('click')
    await flushPromises()

    expect(mockPost).toHaveBeenCalledWith('/experiments', expect.objectContaining({
      name: 'My Experiment',
      min_sample_size: 100,
      variants: expect.arrayContaining([
        expect.objectContaining({ name: 'control', traffic_pct: 50 }),
        expect.objectContaining({ name: 'variant_a', traffic_pct: 50 }),
      ]),
    }))
    // Form should close after successful create
    expect(wrapper.find('.create-form').exists()).toBe(false)
  })

  it('allows adding and removing variants', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    await wrapper.find('.page-header .btn-primary').trigger('click')
    const initialVariantCount = wrapper.findAll('.variant-row').length
    expect(initialVariantCount).toBe(2)

    // Click "+ Add Variant"
    const addBtn = wrapper.findAll('.create-form .btn-sm').find(b => b.text().includes('Add Variant'))!
    await addBtn.trigger('click')
    expect(wrapper.findAll('.variant-row')).toHaveLength(3)

    // Remove the third variant (only visible when more than 2)
    const removeBtns = wrapper.findAll('.variant-row .btn-sm.btn-danger')
    expect(removeBtns.length).toBeGreaterThan(0)
    await removeBtns[0].trigger('click')
    expect(wrapper.findAll('.variant-row')).toHaveLength(2)
  })

  it('surfaces server error when POST fails', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    mockPost.mockRejectedValue({ response: { data: { msg: 'slug taken' } } })
    const wrapper = mount(ExperimentsListPage)
    await flushPromises()

    await wrapper.find('.page-header .btn-primary').trigger('click')
    await wrapper.find('.create-form .btn-primary').trigger('click')
    await flushPromises()

    expect(wrapper.find('.error-text').text()).toBe('slug taken')
  })
})
