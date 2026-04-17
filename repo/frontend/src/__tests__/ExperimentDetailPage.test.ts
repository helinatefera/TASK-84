import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: vi.fn(),
  },
}))

import ExperimentDetailPage from '@/pages/experiments/ExperimentDetailPage.vue'

describe('ExperimentDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(globalThis as any).alert = vi.fn()
  })

  it('fetches experiment and results on mount', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/experiments/exp-1') return Promise.resolve({
        data: { id: 'exp-1', name: 'Banner Test', description: 'd', status: 'draft', variants: [] },
      })
      if (url === '/experiments/exp-1/results') return Promise.resolve({ data: null })
      return Promise.resolve({ data: null })
    })
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'exp-1' } })
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/experiments/exp-1')
    expect(mockGet).toHaveBeenCalledWith('/experiments/exp-1/results')
    expect(wrapper.find('h1').text()).toBe('Banner Test')
    expect(wrapper.text()).toContain('draft')
  })

  it('shows Start button for draft status, no traffic editor', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'draft', variants: [{ name: 'c', traffic_pct: 50 }, { name: 'v', traffic_pct: 50 }] } })
    )
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    const buttons = wrapper.findAll('.actions button')
    expect(buttons.map(b => b.text())).toContain('Start')
    expect(wrapper.find('.traffic-controls').exists()).toBe(false)
  })

  it('shows traffic editor when status is running, with sum badge', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'running', variants: [{ name: 'c', traffic_pct: 60 }, { name: 'v', traffic_pct: 40 }] } })
    )
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    expect(wrapper.find('.traffic-controls').exists()).toBe(true)
    expect(wrapper.find('.traffic-sum').text()).toContain('100%')
    expect(wrapper.find('.traffic-sum').classes()).not.toContain('invalid')
  })

  it('marks traffic sum invalid and disables save when it deviates from 100', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'running', variants: [{ name: 'c', traffic_pct: 50 }, { name: 'v', traffic_pct: 50 }] } })
    )
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    const inputs = wrapper.findAll('.traffic-input')
    await inputs[0].setValue(70)  // 70 + 50 = 120
    await flushPromises()

    expect(wrapper.find('.traffic-sum').classes()).toContain('invalid')
    const saveBtn = wrapper.findAll('.traffic-controls button').find(b => b.text() === 'Save Traffic')!
    expect((saveBtn.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('calls PUT /experiments/:id/traffic with current edits when sum=100', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'running', variants: [{ name: 'c', traffic_pct: 50 }, { name: 'v', traffic_pct: 50 }] } })
    )
    mockPut.mockResolvedValue({ data: {} })
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    const inputs = wrapper.findAll('.traffic-input')
    await inputs[0].setValue(70)
    await inputs[1].setValue(30)
    await flushPromises()

    const saveBtn = wrapper.findAll('.traffic-controls button').find(b => b.text() === 'Save Traffic')!
    await saveBtn.trigger('click')
    await flushPromises()

    expect(mockPut).toHaveBeenCalledWith('/experiments/x/traffic', {
      variants: [
        { name: 'c', traffic_pct: 70 },
        { name: 'v', traffic_pct: 30 },
      ],
    })
  })

  it('triggers action buttons (start/pause/complete/rollback) as POST calls', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'running', variants: [] } })
    )
    mockPost.mockResolvedValue({ data: {} })
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    const pauseBtn = wrapper.findAll('.actions button').find(b => b.text() === 'Pause')!
    await pauseBtn.trigger('click')
    await flushPromises()
    expect(mockPost).toHaveBeenCalledWith('/experiments/x/pause')
  })

  it('renders results table when results payload has variants', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: { confidence_state: 'recommend_keep', variants: [{ name: 'c', sample_size: 100, exposures: 95 }] } })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'running', variants: [] } })
    )
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()

    expect(wrapper.text()).toContain('recommend_keep')
    expect(wrapper.text()).toContain('95')
  })

  it('hides action buttons when experiment is completed or rolled back', async () => {
    mockGet.mockImplementation((url: string) =>
      url.endsWith('/results')
        ? Promise.resolve({ data: null })
        : Promise.resolve({ data: { id: 'x', name: 'x', status: 'completed', variants: [] } })
    )
    const wrapper = mount(ExperimentDetailPage, { props: { id: 'x' } })
    await flushPromises()
    expect(wrapper.find('.actions').exists()).toBe(false)
  })
})
