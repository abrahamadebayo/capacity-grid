/** Date helpers shared by App range navigation. All calendar math is UTC. */

const DAY_MS = 24 * 60 * 60 * 1000

export function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

export function parseDate(s: string): Date {
  const [y, m, day] = s.split('-').map(Number)
  return new Date(Date.UTC(y, m - 1, day))
}

export function mondayOf(d: Date): Date {
  const wd = d.getUTCDay() || 7
  const mon = new Date(d)
  mon.setUTCDate(d.getUTCDate() - (wd - 1))
  return mon
}

export function addDays(d: Date, n: number): Date {
  return new Date(d.getTime() + n * DAY_MS)
}

/** Snap an arbitrary from/to pair onto ISO week boundaries (Mon–Sun). */
export function snapToIsoWeeks(from: string, to: string): { from: string; to: string } {
  const snappedFrom = formatDate(mondayOf(parseDate(from)))
  const snappedTo = formatDate(addDays(mondayOf(parseDate(to)), 6))
  return { from: snappedFrom, to: snappedTo }
}

export function shiftRangeByWeeks(from: string, to: string, delta: number): { from: string; to: string } {
  return {
    from: formatDate(addDays(parseDate(from), delta * 7)),
    to: formatDate(addDays(parseDate(to), delta * 7)),
  }
}

export function formatHours(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(1)
}

export function isOverAllocated(allocated: number, capacity: number): boolean {
  return allocated > capacity
}
