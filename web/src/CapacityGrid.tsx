import { useEffect, useState, type FormEvent, type KeyboardEvent } from 'react'
import { formatHours } from './dates'

type Props = {
  from: string
  to: string
}

type WeekHeader = {
  start: string
  end: string
}

type CapacityCell = {
  weekStart: string
  allocated: number
  capacity: number
  over: boolean
}

type CapacityRow = {
  personId: number
  name: string
  weeklyHours: number
  weeks: CapacityCell[]
}

type CapacityResponse = {
  from: string
  to: string
  weeks: WeekHeader[]
  rows: CapacityRow[]
}

type PersonResponse = {
  id: number
  name: string
  weeklyHours: number
}

function weekLabel(w: WeekHeader): string {
  const d = new Date(w.start + 'T00:00:00Z')
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  })
}

export function CapacityGrid({ from, to }: Props) {
  const [data, setData] = useState<CapacityResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editValue, setEditValue] = useState('')
  const [savingId, setSavingId] = useState<number | null>(null)

  useEffect(() => {
    const ctrl = new AbortController()
    setLoading(true)
    setError(null)

    fetch(`/api/capacity?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`, {
      signal: ctrl.signal,
    })
      .then(async (res) => {
        if (!res.ok) throw new Error(await res.text())
        return res.json() as Promise<CapacityResponse>
      })
      .then((json) => {
        setData(json)
        setLoading(false)
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(err instanceof Error ? err.message : 'Failed to load capacity')
        setLoading(false)
      })

    return () => ctrl.abort()
  }, [from, to])

  function startEdit(row: CapacityRow) {
    setEditingId(row.personId)
    setEditValue(String(row.weeklyHours))
  }

  function cancelEdit() {
    setEditingId(null)
    setEditValue('')
  }

  async function saveEdit(personId: number) {
    const next = Number(editValue)
    if (Number.isNaN(next) || next < 0) {
      setError('Weekly hours must be a non-negative number')
      return
    }

    setSavingId(personId)
    setError(null)
    try {
      const res = await fetch(`/api/people/${personId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ weeklyHours: next }),
      })
      if (!res.ok) throw new Error(await res.text())
      const person = (await res.json()) as PersonResponse

      setData((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          rows: prev.rows.map((row) => {
            if (row.personId !== person.id) return row
            return {
              ...row,
              weeklyHours: person.weeklyHours,
              weeks: row.weeks.map((cell) => ({
                ...cell,
                capacity: person.weeklyHours,
                over: cell.allocated > person.weeklyHours,
              })),
            }
          }),
        }
      })
      cancelEdit()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save')
    } finally {
      setSavingId(null)
    }
  }

  function onEditKey(e: KeyboardEvent<HTMLInputElement>, personId: number) {
    if (e.key === 'Enter') {
      e.preventDefault()
      void saveEdit(personId)
    } else if (e.key === 'Escape') {
      cancelEdit()
    }
  }

  function onEditSubmit(e: FormEvent, personId: number) {
    e.preventDefault()
    void saveEdit(personId)
  }

  if (loading && !data) {
    return <p className="status">Loading capacity…</p>
  }
  if (error && !data) {
    return <p className="status error">{error}</p>
  }
  if (!data) {
    return <p className="status">No data</p>
  }

  return (
    <div className="grid-wrap">
      {error && <p className="status error">{error}</p>}
      {loading && <p className="status">Refreshing…</p>}

      <table className="capacity-grid">
        <thead>
          <tr>
            <th className="sticky-col">Person</th>
            <th className="capacity-col">Capacity</th>
            {data.weeks.map((w) => (
              <th key={w.start} title={`${w.start} → ${w.end}`}>
                {weekLabel(w)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.rows.map((row) => (
            <tr key={row.personId}>
              <td className="sticky-col">{row.name}</td>
              <td className="capacity-col">
                {editingId === row.personId ? (
                  <form
                    className="edit-capacity"
                    onSubmit={(e) => onEditSubmit(e, row.personId)}
                  >
                    <input
                      type="number"
                      min={0}
                      step={1}
                      value={editValue}
                      autoFocus
                      disabled={savingId === row.personId}
                      onChange={(e) => setEditValue(e.target.value)}
                      onKeyDown={(e) => onEditKey(e, row.personId)}
                      aria-label={`Weekly hours for ${row.name}`}
                    />
                    <button type="submit" disabled={savingId === row.personId}>
                      Save
                    </button>
                    <button type="button" onClick={cancelEdit}>
                      Cancel
                    </button>
                  </form>
                ) : (
                  <button
                    type="button"
                    className="capacity-btn"
                    onClick={() => startEdit(row)}
                    title="Edit weekly hours"
                  >
                    {formatHours(row.weeklyHours)}h
                  </button>
                )}
              </td>
              {row.weeks.map((cell) => (
                <td
                  key={cell.weekStart}
                  className={cell.over ? 'over' : cell.allocated > 0 ? 'used' : ''}
                  title={`${formatHours(cell.allocated)} / ${formatHours(cell.capacity)}h`}
                >
                  <span className="alloc">{formatHours(cell.allocated)}</span>
                  <span className="sep">/</span>
                  <span className="cap">{formatHours(cell.capacity)}</span>
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
