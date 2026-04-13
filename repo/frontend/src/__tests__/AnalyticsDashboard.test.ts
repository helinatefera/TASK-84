import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'

// --- Mock apiClient ---
const mockGet = vi.fn()
const mockPost = vi.fn()
const mockDelete = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

// --- Mock VChart (echarts) as a stub ---
vi.mock('vue-echarts', () => ({
  default: defineComponent({ name: 'VChart', props: ['option'], render() { return h('div', { class: 'vchart-stub' }) } }),
}))
vi.mock('echarts', () => ({}))
vi.mock('echarts-wordcloud', () => ({}))

import AnalyticsDashboardPage from '@/pages/analytics/AnalyticsDashboardPage.vue'

function mockSharedDataResponse() {
  return {
    data: {
      filter_config: { item_id: 'uuid-abc', start_date: '2026-01-01', end_date: '2026-06-30', sentiment: 'positive', keywords: '' },
      dashboard: { data: [{ period_start: '2026-01-01', impressions: 50, clicks: 10, avg_dwell_secs: 3.2, favorites: 5, shares: 2, comments: 1 }] },
      keywords: { data: [{ keyword: 'great', weight: 0.8 }] },
      topics: { data: [{ topic: 'quality', confidence: 0.9, count: 5 }] },
      sentiment: { data: [{ sentiment_label: 'positive', count: 20, avg_confidence: 0.85 }] },
      cooccurrences: { data: [] },
    },
  }
}

function mockEmptyResponses() {
  // items endpoint
  mockGet.mockImplementation((url: string) => {
    if (url === '/items') return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/dashboard')) return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/keywords')) return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/topics')) return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/cooccurrences')) return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/sentiment')) return Promise.resolve({ data: { data: [] } })
    if (url.startsWith('/analytics/saved-views')) return Promise.resolve({ data: { data: [] } })
    return Promise.resolve({ data: {} })
  })
}

describe('AnalyticsDashboardPage — shared token mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches /shared/:token/data in shared mode instead of analyst endpoints', async () => {
    const sharedToken = 'test-share-token-abc'
    mockGet.mockImplementation((url: string) => {
      if (url === '/items') return Promise.resolve({ data: { data: [] } })
      if (url === `/shared/${sharedToken}`) {
        return Promise.resolve({
          data: {
            filter_config: { item_id: '', start_date: '2026-03-01', end_date: '', sentiment: '', keywords: '' },
          },
        })
      }
      if (url === `/shared/${sharedToken}/data`) return Promise.resolve(mockSharedDataResponse())
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(AnalyticsDashboardPage, {
      props: { sharedToken, readonly: true },
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    // Should have called /shared/:token (config) and /shared/:token/data
    const urls = mockGet.mock.calls.map((c: any[]) => c[0])
    expect(urls).toContain(`/shared/${sharedToken}`)
    expect(urls).toContain(`/shared/${sharedToken}/data`)

    // Should NOT have called analyst-only endpoints
    expect(urls.some((u: string) => u.includes('/analytics/dashboard'))).toBe(false)
    expect(urls.some((u: string) => u.includes('/analytics/saved-views'))).toBe(false)
  })

  it('populates dashboard data from shared response', async () => {
    const sharedToken = 'data-token'
    mockGet.mockImplementation((url: string) => {
      if (url === '/items') return Promise.resolve({ data: { data: [] } })
      if (url === `/shared/${sharedToken}`) {
        return Promise.resolve({
          data: { filter_config: { item_id: '', start_date: '', end_date: '', sentiment: '', keywords: '' } },
        })
      }
      if (url === `/shared/${sharedToken}/data`) return Promise.resolve(mockSharedDataResponse())
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(AnalyticsDashboardPage, {
      props: { sharedToken, readonly: true },
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    // Dashboard table should show the data row
    const rows = wrapper.findAll('tbody tr')
    expect(rows.length).toBeGreaterThanOrEqual(1)
    expect(wrapper.text()).toContain('50') // impressions
  })
})

describe('AnalyticsDashboardPage — readonly mode UI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('hides filter bar in readonly mode', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/items') return Promise.resolve({ data: { data: [] } })
      if (url.startsWith('/shared/')) return Promise.resolve({ data: { filter_config: {}, dashboard: { data: [] }, keywords: { data: [] }, topics: { data: [] }, sentiment: { data: [] }, cooccurrences: { data: [] } } })
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(AnalyticsDashboardPage, {
      props: { sharedToken: 'ro-token', readonly: true },
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    expect(wrapper.find('.filter-bar').exists()).toBe(false)
  })

  it('hides saved-views section in readonly mode', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/items') return Promise.resolve({ data: { data: [] } })
      if (url.startsWith('/shared/')) return Promise.resolve({ data: { filter_config: {}, dashboard: { data: [] }, keywords: { data: [] }, topics: { data: [] }, sentiment: { data: [] }, cooccurrences: { data: [] } } })
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(AnalyticsDashboardPage, {
      props: { sharedToken: 'ro-token-2', readonly: true },
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    expect(wrapper.find('.saved-views-section').exists()).toBe(false)
  })

  it('shows readonly banner in readonly mode', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/items') return Promise.resolve({ data: { data: [] } })
      if (url.startsWith('/shared/')) return Promise.resolve({ data: { filter_config: {}, dashboard: { data: [] }, keywords: { data: [] }, topics: { data: [] }, sentiment: { data: [] }, cooccurrences: { data: [] } } })
      return Promise.resolve({ data: {} })
    })

    const wrapper = mount(AnalyticsDashboardPage, {
      props: { sharedToken: 'ro-token-3', readonly: true },
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    expect(wrapper.find('.readonly-banner').exists()).toBe(true)
    expect(wrapper.text()).toContain('shared dashboard snapshot')
  })
})

describe('AnalyticsDashboardPage — normal mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows filter bar with item selector in normal mode', async () => {
    mockEmptyResponses()

    const wrapper = mount(AnalyticsDashboardPage, {
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    expect(wrapper.find('.filter-bar').exists()).toBe(true)
    // Item selector should be present
    const selects = wrapper.findAll('select')
    expect(selects.length).toBeGreaterThanOrEqual(1)
    // First select should be the item filter
    const itemSelect = selects[0]
    expect(itemSelect.find('option').text()).toContain('All Items')
  })

  it('shows saved-views section in normal mode', async () => {
    mockEmptyResponses()

    const wrapper = mount(AnalyticsDashboardPage, {
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    expect(wrapper.find('.saved-views-section').exists()).toBe(true)
    expect(wrapper.find('.readonly-banner').exists()).toBe(false)
  })

  it('calls analyst dashboard endpoint in normal mode', async () => {
    mockEmptyResponses()

    mount(AnalyticsDashboardPage, {
      global: { stubs: { VChart: true } },
    })

    await flushPromises()

    const urls = mockGet.mock.calls.map((c: any[]) => c[0])
    expect(urls.some((u: string) => u.includes('/analytics/dashboard'))).toBe(true)
  })
})
