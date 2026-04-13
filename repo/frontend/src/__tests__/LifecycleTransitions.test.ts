import { describe, it, expect } from 'vitest'

// Tests for appeal lifecycle transitions, fraud boundary matrices,
// and sensitive-word rule enforcement logic.

// --- Appeal Lifecycle ---

type AppealStatus = 'pending' | 'approved' | 'rejected' | 'needs_edit'

const appealTransitions: Record<AppealStatus, { moderator: AppealStatus[]; user: AppealStatus[] }> = {
  pending:    { moderator: ['approved', 'rejected', 'needs_edit'], user: [] },
  approved:   { moderator: [], user: [] },                // terminal
  rejected:   { moderator: [], user: [] },                // terminal
  needs_edit: { moderator: [], user: ['pending'] },       // user resubmits → pending
}

function canModeratorTransition(from: AppealStatus, to: AppealStatus): boolean {
  return appealTransitions[from]?.moderator.includes(to) ?? false
}

function canUserResubmit(from: AppealStatus): boolean {
  return appealTransitions[from]?.user.includes('pending') ?? false
}

describe('Appeal lifecycle transitions', () => {
  it('moderator can approve, reject, or request edit on pending appeal', () => {
    expect(canModeratorTransition('pending', 'approved')).toBe(true)
    expect(canModeratorTransition('pending', 'rejected')).toBe(true)
    expect(canModeratorTransition('pending', 'needs_edit')).toBe(true)
  })

  it('approved and rejected are terminal — no moderator transitions', () => {
    expect(canModeratorTransition('approved', 'pending')).toBe(false)
    expect(canModeratorTransition('approved', 'rejected')).toBe(false)
    expect(canModeratorTransition('rejected', 'pending')).toBe(false)
    expect(canModeratorTransition('rejected', 'approved')).toBe(false)
  })

  it('user can resubmit only from needs_edit', () => {
    expect(canUserResubmit('needs_edit')).toBe(true)
    expect(canUserResubmit('pending')).toBe(false)
    expect(canUserResubmit('approved')).toBe(false)
    expect(canUserResubmit('rejected')).toBe(false)
  })

  it('full lifecycle: pending → needs_edit → pending (resubmit) → approved', () => {
    let status: AppealStatus = 'pending'

    // Moderator requests edit
    expect(canModeratorTransition(status, 'needs_edit')).toBe(true)
    status = 'needs_edit'

    // User resubmits
    expect(canUserResubmit(status)).toBe(true)
    status = 'pending'

    // Moderator approves
    expect(canModeratorTransition(status, 'approved')).toBe(true)
    status = 'approved'

    // Terminal
    expect(canModeratorTransition(status, 'pending')).toBe(false)
    expect(canUserResubmit(status)).toBe(false)
  })

  it('direct approval path: pending → approved', () => {
    expect(canModeratorTransition('pending', 'approved')).toBe(true)
  })

  it('direct rejection path: pending → rejected', () => {
    expect(canModeratorTransition('pending', 'rejected')).toBe(true)
  })
})

// --- Fraud Status Boundary Matrix ---

type ReviewFraudStatus = 'normal' | 'suspected_fraud' | 'confirmed_fraud' | 'cleared'
type UserFraudStatus = 'clean' | 'suspected' | 'confirmed'

interface FraudTransition {
  reviewFrom: ReviewFraudStatus
  reviewTo: ReviewFraudStatus
  userEffect: UserFraudStatus | null  // null = no change
}

const moderatorFraudActions: FraudTransition[] = [
  { reviewFrom: 'suspected_fraud', reviewTo: 'confirmed_fraud', userEffect: 'confirmed' },
  { reviewFrom: 'suspected_fraud', reviewTo: 'cleared', userEffect: 'clean' },
]

const jobFraudActions: FraudTransition[] = [
  { reviewFrom: 'normal', reviewTo: 'suspected_fraud', userEffect: 'suspected' },
]

describe('Fraud status boundary matrix', () => {
  it('moderator confirm escalates review AND user', () => {
    const action = moderatorFraudActions.find(a => a.reviewTo === 'confirmed_fraud')!
    expect(action.reviewFrom).toBe('suspected_fraud')
    expect(action.userEffect).toBe('confirmed')
  })

  it('moderator clear de-escalates review AND user', () => {
    const action = moderatorFraudActions.find(a => a.reviewTo === 'cleared')!
    expect(action.reviewFrom).toBe('suspected_fraud')
    expect(action.userEffect).toBe('clean')
  })

  it('automated scan flags normal reviews to suspected AND user to suspected', () => {
    const action = jobFraudActions.find(a => a.reviewTo === 'suspected_fraud')!
    expect(action.reviewFrom).toBe('normal')
    expect(action.userEffect).toBe('suspected')
  })

  it('confirmed_fraud review cannot be auto-scanned back to suspected', () => {
    // The job only targets fraud_status = 'normal', so confirmed is untouched
    const targeted = jobFraudActions.filter(a => a.reviewFrom === 'confirmed_fraud')
    expect(targeted).toHaveLength(0)
  })

  it('cleared review is not re-flagged by automated scan', () => {
    const targeted = jobFraudActions.filter(a => a.reviewFrom === 'cleared')
    expect(targeted).toHaveLength(0)
  })
})

// --- Sensitive Word Rule Enforcement Logic ---

interface WordRule {
  pattern: string
  action: 'block' | 'flag' | 'replace'
  replacement?: string
  is_active: boolean
}

interface FilterResult {
  blocked: boolean
  flagged: boolean
  text: string
}

function applyRules(text: string, rules: WordRule[]): FilterResult {
  const result: FilterResult = { blocked: false, flagged: false, text }

  for (const rule of rules) {
    if (!rule.is_active) continue
    const re = new RegExp(rule.pattern, 'gi')
    if (!re.test(result.text)) continue

    switch (rule.action) {
      case 'block':
        result.blocked = true
        return result  // short-circuit on block
      case 'flag':
        result.flagged = true
        break
      case 'replace':
        result.text = result.text.replace(new RegExp(rule.pattern, 'gi'), rule.replacement || '***')
        break
    }
  }

  return result
}

describe('Sensitive word rule enforcement', () => {
  const rules: WordRule[] = [
    { pattern: 'spam', action: 'block', is_active: true },
    { pattern: 'darn', action: 'replace', replacement: '****', is_active: true },
    { pattern: 'suspicious', action: 'flag', is_active: true },
    { pattern: 'oldword', action: 'block', is_active: false },  // inactive
  ]

  it('blocks content matching a block rule', () => {
    const result = applyRules('This is spam content', rules)
    expect(result.blocked).toBe(true)
    expect(result.flagged).toBe(false)
  })

  it('replaces content matching a replace rule', () => {
    const result = applyRules('Oh darn, that is bad', rules)
    expect(result.blocked).toBe(false)
    expect(result.text).toBe('Oh ****, that is bad')
  })

  it('flags content matching a flag rule', () => {
    const result = applyRules('This is suspicious activity', rules)
    expect(result.blocked).toBe(false)
    expect(result.flagged).toBe(true)
    expect(result.text).toContain('suspicious')  // text unchanged, just flagged
  })

  it('passes safe content through unchanged', () => {
    const result = applyRules('Great product, highly recommend!', rules)
    expect(result.blocked).toBe(false)
    expect(result.flagged).toBe(false)
    expect(result.text).toBe('Great product, highly recommend!')
  })

  it('inactive rules are not applied', () => {
    const result = applyRules('This contains oldword text', rules)
    expect(result.blocked).toBe(false)  // rule is inactive
  })

  it('block takes precedence — short-circuits before flag/replace', () => {
    const result = applyRules('spam and suspicious mixed', rules)
    expect(result.blocked).toBe(true)
    expect(result.flagged).toBe(false)  // block short-circuited
  })

  it('replace and flag can coexist', () => {
    const result = applyRules('darn suspicious activity', rules)
    expect(result.blocked).toBe(false)
    expect(result.flagged).toBe(true)
    expect(result.text).toBe('**** suspicious activity')
  })

  it('multiple replace matches are all replaced', () => {
    const result = applyRules('darn it, darn it again', rules)
    expect(result.text).toBe('**** it, **** it again')
  })
})
