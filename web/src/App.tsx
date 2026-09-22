import { useState, type FormEvent } from 'react'
import { CapacityGrid } from './CapacityGrid'
import { shiftRangeByWeeks, snapToIsoWeeks } from './dates'

/** Default window: three ISO weeks covering the seeded Dec/Jan slice. */
const DEFAULT_FROM = '2025-12-29'
const DEFAULT_TO = '2026-01-18'

export function App() {
  const [from, setFrom] = useState(DEFAULT_FROM)
  const [to, setTo] = useState(DEFAULT_TO)
  const [draftFrom, setDraftFrom] = useState(DEFAULT_FROM)
  const [draftTo, setDraftTo] = useState(DEFAULT_TO)

  function shiftWeeks(delta: number) {
    const next = shiftRangeByWeeks(from, to, delta)
    setFrom(next.from)
    setTo(next.to)
    setDraftFrom(next.from)
    setDraftTo(next.to)
  }

  function applyRange(e: FormEvent) {
    e.preventDefault()
    if (!draftFrom || !draftTo) return
    if (draftTo < draftFrom) return
    const snapped = snapToIsoWeeks(draftFrom, draftTo)
    setFrom(snapped.from)
    setTo(snapped.to)
    setDraftFrom(snapped.from)
    setDraftTo(snapped.to)
  }

  return (
    <main>
      <h1>Team capacity</h1>

      <div className="toolbar">
        <div className="nav">
          <button type="button" onClick={() => shiftWeeks(-1)} aria-label="Previous week">
            ← Prev week
          </button>
          <button type="button" onClick={() => shiftWeeks(1)} aria-label="Next week">
            Next week →
          </button>
        </div>

        <form className="range-form" onSubmit={applyRange}>
          <label>
            From
            <input
              type="date"
              value={draftFrom}
              onChange={(e) => setDraftFrom(e.target.value)}
            />
          </label>
          <label>
            To
            <input
              type="date"
              value={draftTo}
              onChange={(e) => setDraftTo(e.target.value)}
            />
          </label>
          <button type="submit">Apply</button>
        </form>

        <p className="range">
          Showing {from} → {to}
        </p>
      </div>

      <CapacityGrid from={from} to={to} />
    </main>
  )
}
