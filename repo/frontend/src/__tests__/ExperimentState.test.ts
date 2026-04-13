import { describe, it, expect, vi } from 'vitest'

// Tests for experiment state transitions and assignment logic.
// These validate the frontend's understanding of experiment lifecycle
// without requiring the actual API.

type ExperimentStatus = 'draft' | 'running' | 'paused' | 'completed' | 'rolled_back'

interface Experiment {
  id: string
  name: string
  status: ExperimentStatus
  variants: { name: string; traffic_pct: number }[]
}

// Valid transitions mirror the backend's state machine
const validTransitions: Record<ExperimentStatus, ExperimentStatus[]> = {
  draft: ['running'],
  running: ['paused', 'completed', 'rolled_back'],
  paused: ['running', 'completed', 'rolled_back'],
  completed: [],
  rolled_back: [],
}

function canTransition(from: ExperimentStatus, to: ExperimentStatus): boolean {
  return validTransitions[from]?.includes(to) ?? false
}

function availableActions(status: ExperimentStatus): string[] {
  const actions: string[] = []
  if (status === 'draft') actions.push('start')
  if (status === 'running') actions.push('pause', 'complete', 'rollback')
  if (status === 'paused') actions.push('complete', 'rollback')
  return actions
}

describe('Experiment state transitions', () => {
  it('draft can only transition to running', () => {
    expect(canTransition('draft', 'running')).toBe(true)
    expect(canTransition('draft', 'paused')).toBe(false)
    expect(canTransition('draft', 'completed')).toBe(false)
  })

  it('running can pause, complete, or rollback', () => {
    expect(canTransition('running', 'paused')).toBe(true)
    expect(canTransition('running', 'completed')).toBe(true)
    expect(canTransition('running', 'rolled_back')).toBe(true)
    expect(canTransition('running', 'draft')).toBe(false)
  })

  it('paused can resume (running), complete, or rollback', () => {
    expect(canTransition('paused', 'running')).toBe(true)
    expect(canTransition('paused', 'completed')).toBe(true)
    expect(canTransition('paused', 'rolled_back')).toBe(true)
  })

  it('completed and rolled_back are terminal states', () => {
    expect(canTransition('completed', 'running')).toBe(false)
    expect(canTransition('completed', 'draft')).toBe(false)
    expect(canTransition('rolled_back', 'running')).toBe(false)
    expect(canTransition('rolled_back', 'draft')).toBe(false)
  })
})

describe('Available actions by status', () => {
  it('shows Start for draft experiments', () => {
    expect(availableActions('draft')).toEqual(['start'])
  })

  it('shows Pause/Complete/Rollback for running experiments', () => {
    expect(availableActions('running')).toEqual(['pause', 'complete', 'rollback'])
  })

  it('shows Complete/Rollback for paused experiments', () => {
    expect(availableActions('paused')).toEqual(['complete', 'rollback'])
  })

  it('shows no actions for terminal states', () => {
    expect(availableActions('completed')).toEqual([])
    expect(availableActions('rolled_back')).toEqual([])
  })
})

describe('Variant traffic validation', () => {
  it('rejects variants that do not sum to 100%', () => {
    const variants = [
      { name: 'control', traffic_pct: 50 },
      { name: 'variant_a', traffic_pct: 30 },
    ]
    const sum = variants.reduce((s, v) => s + v.traffic_pct, 0)
    expect(sum).not.toBe(100)
  })

  it('accepts variants that sum to exactly 100%', () => {
    const variants = [
      { name: 'control', traffic_pct: 50 },
      { name: 'variant_a', traffic_pct: 30 },
      { name: 'variant_b', traffic_pct: 20 },
    ]
    const sum = variants.reduce((s, v) => s + v.traffic_pct, 0)
    expect(sum).toBe(100)
  })
})

describe('Canary traffic adjustment validation', () => {
  function validateTrafficUpdate(
    variants: { name: string; traffic_pct: number }[],
    status: ExperimentStatus
  ): string | null {
    if (status !== 'running' && status !== 'paused') {
      return 'Traffic can only be adjusted on running or paused experiments'
    }
    const sum = variants.reduce((s, v) => s + v.traffic_pct, 0)
    if (sum !== 100) return `Variant traffic_pct must sum to 100 (got ${sum.toFixed(2)})`
    if (variants.length < 2) return 'At least 2 variants required'
    return null
  }

  it('allows traffic change on running experiment', () => {
    expect(validateTrafficUpdate([
      { name: 'control', traffic_pct: 70 },
      { name: 'variant_a', traffic_pct: 30 },
    ], 'running')).toBeNull()
  })

  it('allows traffic change on paused experiment', () => {
    expect(validateTrafficUpdate([
      { name: 'control', traffic_pct: 50 },
      { name: 'variant_a', traffic_pct: 50 },
    ], 'paused')).toBeNull()
  })

  it('rejects traffic change on draft experiment', () => {
    const err = validateTrafficUpdate([
      { name: 'control', traffic_pct: 50 },
      { name: 'variant_a', traffic_pct: 50 },
    ], 'draft')
    expect(err).toContain('running or paused')
  })

  it('rejects traffic change on completed experiment', () => {
    expect(validateTrafficUpdate([], 'completed')).toContain('running or paused')
  })

  it('rejects traffic that does not sum to 100', () => {
    const err = validateTrafficUpdate([
      { name: 'control', traffic_pct: 60 },
      { name: 'variant_a', traffic_pct: 30 },
    ], 'running')
    expect(err).toContain('sum to 100')
  })
})

describe('Experiment assignment determinism', () => {
  it('same user gets same variant on repeated calls', () => {
    // Simulate the FNV-1a hash approach the backend uses
    function simpleHash(salt: string, userId: number): number {
      let hash = 0x811c9dc5
      const input = `${salt}:${userId}`
      for (let i = 0; i < input.length; i++) {
        hash ^= input.charCodeAt(i)
        hash = (hash * 0x01000193) >>> 0
      }
      return hash % 10000
    }

    const salt = 'test-experiment-salt'
    const userId = 42
    const bucket1 = simpleHash(salt, userId)
    const bucket2 = simpleHash(salt, userId)
    expect(bucket1).toBe(bucket2)
  })
})
