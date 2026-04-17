import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mockGet = vi.fn()
const mockPut = vi.fn()
vi.mock('@/api/client', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: vi.fn(),
    put: (...args: any[]) => mockPut(...args),
    delete: vi.fn(),
  },
}))

import ModerationQueuePage from '@/pages/moderation/ModerationQueuePage.vue'

function queueResponse(reports: any[]) {
  return { data: { data: reports } }
}

describe('ModerationQueuePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // silence alert() used in error handlers
    ;(globalThis as any).alert = vi.fn()
  })

  it('fetches queue and appeals on mount', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url === '/moderation/queue') return Promise.resolve(queueResponse([]))
      if (url === '/moderation/appeals') return Promise.resolve({ data: { data: [] } })
      return Promise.resolve({ data: { data: [] } })
    })
    mount(ModerationQueuePage)
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/moderation/queue', { params: {} })
    expect(mockGet).toHaveBeenCalledWith('/moderation/appeals', { params: {} })
  })

  it('renders reports with status badges and action buttons', async () => {
    mockGet.mockImplementation((url: string) =>
      url === '/moderation/queue'
        ? Promise.resolve(queueResponse([
            { id: 7, status: 'pending', target_type: 'review', target_id: 1, priority: 'high', category: 'spam', description: 'd', created_at: '2026-01-01' },
          ]))
        : Promise.resolve({ data: { data: [] } })
    )
    const wrapper = mount(ModerationQueuePage)
    await flushPromises()

    expect(wrapper.text()).toContain('pending')
    expect(wrapper.text()).toContain('review')
    expect(wrapper.text()).toContain('Report #7')
    const buttons = wrapper.findAll('.report-card button')
    expect(buttons.length).toBeGreaterThan(0)
    // Resolve/Dismiss are disabled until a note is entered.
    const resolveBtn = buttons.find(b => b.text() === 'Resolve')!
    expect((resolveBtn.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('resolves a report after notes are filled and refetches queue', async () => {
    mockGet.mockImplementation((url: string) =>
      url === '/moderation/queue'
        ? Promise.resolve(queueResponse([
            { id: 42, status: 'pending', target_type: 'review', target_id: 1, priority: 'low', category: 'spam', created_at: '' },
          ]))
        : Promise.resolve({ data: { data: [] } })
    )
    mockPut.mockResolvedValue({ data: {} })
    const wrapper = mount(ModerationQueuePage)
    await flushPromises()

    await wrapper.find('.notes-input').setValue('clearly spam')
    const resolveBtn = wrapper.findAll('.report-card button').find(b => b.text() === 'Resolve')!
    await resolveBtn.trigger('click')
    await flushPromises()

    expect(mockPut).toHaveBeenCalledWith('/moderation/reports/42', {
      status: 'resolved',
      resolution_note: 'clearly spam',
    })
    // mount + refetch after PUT = 2 calls to queue
    expect(mockGet.mock.calls.filter(c => c[0] === '/moderation/queue').length).toBe(2)
  })

  it('applies status filter and re-fetches queue when changed', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ModerationQueuePage)
    await flushPromises()
    mockGet.mockClear()

    const statusSelect = wrapper.findAll('.filter-bar select')[0]
    await statusSelect.setValue('in_review')
    await flushPromises()

    expect(mockGet).toHaveBeenCalledWith('/moderation/queue', { params: { status: 'in_review' } })
  })

  it('shows empty states when there is no data', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })
    const wrapper = mount(ModerationQueuePage)
    await flushPromises()
    expect(wrapper.text()).toContain('No reports in queue')
    expect(wrapper.text()).toContain('No appeals')
  })

  it('handles appeal with a note via PUT /moderation/appeals/:id', async () => {
    mockGet.mockImplementation((url: string) =>
      url === '/moderation/appeals'
        ? Promise.resolve({ data: { data: [{ id: 9, status: 'pending', body: 'please reconsider', created_at: '' }] } })
        : Promise.resolve({ data: { data: [] } })
    )
    mockPut.mockResolvedValue({ data: {} })
    const wrapper = mount(ModerationQueuePage)
    await flushPromises()

    await wrapper.find('.notes-input').setValue('approved: valid')
    const approveBtn = wrapper.findAll('.report-card button').find(b => b.text() === 'Approve')!
    await approveBtn.trigger('click')
    await flushPromises()

    expect(mockPut).toHaveBeenCalledWith('/moderation/appeals/9', {
      status: 'approved',
      note: 'approved: valid',
    })
  })
})
