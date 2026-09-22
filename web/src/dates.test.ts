import { describe, expect, it } from 'vitest'
import {
  formatHours,
  isOverAllocated,
  mondayOf,
  parseDate,
  formatDate,
  shiftRangeByWeeks,
  snapToIsoWeeks,
} from './dates'

describe('mondayOf', () => {
  it('returns the same day for a Monday', () => {
    expect(formatDate(mondayOf(parseDate('2025-12-29')))).toBe('2025-12-29')
  })

  it('snaps mid-week dates back to Monday', () => {
    expect(formatDate(mondayOf(parseDate('2026-01-01')))).toBe('2025-12-29')
  })

  it('snaps Sunday to the prior Monday', () => {
    expect(formatDate(mondayOf(parseDate('2026-01-04')))).toBe('2025-12-29')
  })
})

describe('snapToIsoWeeks', () => {
  it('expands a mid-week to to Sunday of that week', () => {
    expect(snapToIsoWeeks('2025-12-30', '2026-01-16')).toEqual({
      from: '2025-12-29',
      to: '2026-01-18',
    })
  })
})

describe('shiftRangeByWeeks', () => {
  it('moves both ends by seven days', () => {
    expect(shiftRangeByWeeks('2025-12-29', '2026-01-18', 1)).toEqual({
      from: '2026-01-05',
      to: '2026-01-25',
    })
  })
})

describe('formatHours', () => {
  it('keeps integers clean and rounds fractions to one decimal', () => {
    expect(formatHours(40)).toBe('40')
    expect(formatHours(3.5)).toBe('3.5')
  })
})

describe('isOverAllocated', () => {
  it('is strict greater-than', () => {
    expect(isOverAllocated(40, 40)).toBe(false)
    expect(isOverAllocated(40.1, 40)).toBe(true)
    expect(isOverAllocated(0, 0)).toBe(false)
  })
})
