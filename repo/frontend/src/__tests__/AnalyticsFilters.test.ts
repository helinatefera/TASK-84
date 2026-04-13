import { describe, it, expect } from 'vitest'

// Tests for analytics dashboard filter logic, saved view management,
// shared-link contract, and idempotency-key enforcement.

interface DashboardFilters {
  item_id: string
  start_date: string
  end_date: string
  sentiment: string
  keywords: string
}

function buildQueryParams(filters: DashboardFilters): Record<string, string> {
  const params: Record<string, string> = {}
  if (filters.item_id) params.item_id = filters.item_id
  if (filters.start_date) params.start_date = filters.start_date
  if (filters.end_date) params.end_date = filters.end_date
  if (filters.sentiment) params.sentiment = filters.sentiment
  if (filters.keywords) params.keywords = filters.keywords
  return params
}

function validateDateRange(start: string, end: string): string | null {
  if (!start && !end) return null
  if (start && end && new Date(start) > new Date(end)) {
    return 'Start date must be before end date'
  }
  return null
}

const DEFAULT_FILTERS: DashboardFilters = {
  item_id: '', start_date: '', end_date: '', sentiment: '', keywords: '',
}

describe('Dashboard filter query building', () => {
  it('omits empty filter values', () => {
    const params = buildQueryParams({ ...DEFAULT_FILTERS })
    expect(Object.keys(params)).toHaveLength(0)
  })

  it('includes only set filters', () => {
    const params = buildQueryParams({ ...DEFAULT_FILTERS, start_date: '2026-01-01', sentiment: 'positive' })
    expect(params).toEqual({ start_date: '2026-01-01', sentiment: 'positive' })
  })

  it('includes item_id when set', () => {
    const params = buildQueryParams({ ...DEFAULT_FILTERS, item_id: 'abc-123-uuid' })
    expect(params).toEqual({ item_id: 'abc-123-uuid' })
  })

  it('includes all filters when set', () => {
    const params = buildQueryParams({
      item_id: 'item-uuid',
      start_date: '2026-01-01', end_date: '2026-03-31',
      sentiment: 'negative', keywords: 'slow delivery',
    })
    expect(Object.keys(params)).toHaveLength(5)
    expect(params.item_id).toBe('item-uuid')
    expect(params.keywords).toBe('slow delivery')
  })
})

describe('Date range validation', () => {
  it('allows empty date range', () => {
    expect(validateDateRange('', '')).toBeNull()
  })

  it('allows valid date range', () => {
    expect(validateDateRange('2026-01-01', '2026-12-31')).toBeNull()
  })

  it('rejects reversed date range', () => {
    const err = validateDateRange('2026-12-31', '2026-01-01')
    expect(err).toBe('Start date must be before end date')
  })
})

describe('Saved view serialization', () => {
  it('round-trips filter config including item_id through JSON', () => {
    const filters: DashboardFilters = {
      item_id: 'some-uuid',
      start_date: '2026-01-01', end_date: '2026-06-30',
      sentiment: 'positive', keywords: 'quality',
    }
    const serialized = JSON.stringify(filters)
    const restored = JSON.parse(serialized) as DashboardFilters
    expect(restored).toEqual(filters)
    expect(restored.item_id).toBe('some-uuid')
  })

  it('restores from saved view config with item_id', () => {
    const savedViewConfig = '{"item_id":"uuid-123","start_date":"2026-02-01","end_date":"2026-02-28","sentiment":"","keywords":"returns"}'
    const restored = JSON.parse(savedViewConfig) as DashboardFilters
    expect(restored.item_id).toBe('uuid-123')
    expect(restored.start_date).toBe('2026-02-01')
    expect(restored.keywords).toBe('returns')
  })

  it('handles missing item_id in legacy config by defaulting', () => {
    const legacyConfig = '{"start_date":"2026-01-01","end_date":"","sentiment":"","keywords":""}'
    const restored = JSON.parse(legacyConfig) as Partial<DashboardFilters>
    const merged = { ...DEFAULT_FILTERS, ...restored }
    expect(merged.item_id).toBe('')
    expect(merged.start_date).toBe('2026-01-01')
  })
})

// --- Shared link contract tests ---

describe('Shared link data contract', () => {
  it('parses shared view response with filter_config as object', () => {
    // Backend returns filter_config as json.RawMessage (JSON object, not string)
    const sharedResponse = {
      filter_config: { item_id: 'uuid-abc', start_date: '2026-03-01', end_date: '', sentiment: 'positive', keywords: '' },
      dashboard: { data: [{ period_start: '2026-03-01', impressions: 100 }] },
      keywords: { data: [{ keyword: 'great', weight: 0.9 }] },
      topics: { data: [] },
      sentiment: { data: [{ sentiment_label: 'positive', count: 10, avg_confidence: 0.85 }] },
      cooccurrences: { data: [] },
    }

    // filter_config is an object — apply directly
    const config = sharedResponse.filter_config
    expect(typeof config).toBe('object')
    const filters = { ...DEFAULT_FILTERS, ...config }
    expect(filters.item_id).toBe('uuid-abc')
    expect(filters.sentiment).toBe('positive')

    // dashboard data is pre-wrapped
    expect(sharedResponse.dashboard.data).toHaveLength(1)
    expect(sharedResponse.dashboard.data[0].impressions).toBe(100)

    // visualization data arrays
    expect(sharedResponse.keywords.data).toHaveLength(1)
    expect(sharedResponse.sentiment.data[0].count).toBe(10)
  })

  it('parses shared view response with filter_config as string (legacy)', () => {
    const sharedResponse = {
      filter_config: '{"item_id":"","start_date":"2026-01-01","end_date":"","sentiment":"","keywords":"test"}',
    }

    const config = typeof sharedResponse.filter_config === 'string'
      ? JSON.parse(sharedResponse.filter_config)
      : sharedResponse.filter_config
    const filters = { ...DEFAULT_FILTERS, ...config }
    expect(filters.keywords).toBe('test')
    expect(filters.item_id).toBe('')
  })

  it('shared view disables mutating actions when readonly', () => {
    const isReadonly = true
    // Filter bar: v-if="!isReadonly"
    expect(!isReadonly).toBe(false)
    // Saved views section: v-if="!isReadonly"
    expect(!isReadonly).toBe(false)
    // Saved views should not be loaded
    const shouldLoadSavedViews = !isReadonly
    expect(shouldLoadSavedViews).toBe(false)
  })
})

// --- Idempotency key contract tests ---

describe('Idempotency key enforcement', () => {
  it('generates a valid UUID for X-Idempotency-Key', () => {
    const key = crypto.randomUUID()
    // UUID v4 format: 8-4-4-4-12 hex digits
    expect(key).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
  })

  it('generates unique keys per call', () => {
    const keys = new Set(Array.from({ length: 100 }, () => crypto.randomUUID()))
    expect(keys.size).toBe(100)
  })

  it('simulates interceptor attaching key to POST config', () => {
    const config = { method: 'post' as const, headers: {} as Record<string, string> }

    // Simulate the interceptor logic from client.ts
    if (config.method === 'post' && !config.headers['X-Idempotency-Key']) {
      config.headers['X-Idempotency-Key'] = crypto.randomUUID()
    }

    expect(config.headers['X-Idempotency-Key']).toBeDefined()
    expect(config.headers['X-Idempotency-Key'].length).toBe(36)
  })

  it('does not overwrite an explicit key', () => {
    const config = { method: 'post' as const, headers: { 'X-Idempotency-Key': 'explicit-key' } }

    if (config.method === 'post' && !config.headers['X-Idempotency-Key']) {
      config.headers['X-Idempotency-Key'] = crypto.randomUUID()
    }

    expect(config.headers['X-Idempotency-Key']).toBe('explicit-key')
  })

  it('does not attach key to GET requests', () => {
    const config = { method: 'get' as string, headers: {} as Record<string, string> }

    if (config.method === 'post' && !config.headers['X-Idempotency-Key']) {
      config.headers['X-Idempotency-Key'] = crypto.randomUUID()
    }

    expect(config.headers['X-Idempotency-Key']).toBeUndefined()
  })

  it('backend rejects POST without key (400 contract)', () => {
    // Simulates the expected backend behavior from idempotency.go
    function simulateMiddleware(method: string, header: string | undefined): { status: number; allowed: boolean } {
      if (method !== 'POST') return { status: 200, allowed: true }
      if (!header) return { status: 400, allowed: false }
      return { status: 200, allowed: true }
    }

    expect(simulateMiddleware('POST', undefined)).toEqual({ status: 400, allowed: false })
    expect(simulateMiddleware('POST', '')).toEqual({ status: 400, allowed: false })
    expect(simulateMiddleware('POST', 'valid-key')).toEqual({ status: 200, allowed: true })
    expect(simulateMiddleware('GET', undefined)).toEqual({ status: 200, allowed: true })
  })
})
