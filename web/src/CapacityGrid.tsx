import {
  useEffect,
  useId,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
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

const ROW_HEIGHT = 44

function weekLabel(w: WeekHeader): string {
  const d = new Date(w.start + 'T00:00:00Z')
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  })
}

type CapacityEditProps = {
  row: CapacityRow
  editing: boolean
  value: string
  saving: boolean
  invalid: boolean
  onStart: () => void
  onChange: (value: string) => void
  onSave: () => void
  onCancel: () => void
}

function CapacityEdit({
  row,
  editing,
  value,
  saving,
  invalid,
  onStart,
  onChange,
  onSave,
  onCancel,
}: CapacityEditProps) {
  const triggerRef = useRef<HTMLButtonElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const wasEditing = useRef(false)
  const baseId = useId()
  const labelId = `${baseId}-label`
  const hintId = `${baseId}-hint`
  const errorId = `${baseId}-error`

  useEffect(() => {
    if (editing) {
      wasEditing.current = true
      inputRef.current?.focus()
      inputRef.current?.select()
      return
    }
    if (wasEditing.current) {
      wasEditing.current = false
      triggerRef.current?.focus()
    }
  }, [editing])

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      e.preventDefault()
      onSave()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      onCancel()
    }
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    onSave()
  }

  if (!editing) {
    return (
      <button
        ref={triggerRef}
        type="button"
        className="capacity-btn"
        onClick={onStart}
        aria-label={`Edit weekly capacity for ${row.name}, currently ${formatHours(row.weeklyHours)} hours`}
        aria-expanded={false}
        aria-haspopup="true"
      >
        <span aria-hidden="true">{formatHours(row.weeklyHours)}h</span>
      </button>
    )
  }

  return (
    <form
      className="edit-capacity"
      onSubmit={onSubmit}
      aria-labelledby={labelId}
      aria-describedby={invalid ? `${hintId} ${errorId}` : hintId}
      aria-busy={saving || undefined}
    >
      <span id={labelId} className="sr-only">
        Weekly capacity for {row.name}
      </span>
      <span id={hintId} className="sr-only">
        Enter hours per week. Press Enter to save, Escape to cancel.
      </span>
      <input
        ref={inputRef}
        type="number"
        min={0}
        max={168}
        step={1}
        inputMode="decimal"
        value={value}
        disabled={saving}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        aria-invalid={invalid || undefined}
        aria-errormessage={invalid ? errorId : undefined}
        aria-required="true"
      />
      {invalid && (
        <span id={errorId} className="sr-only" role="alert">
          Weekly hours must be a number between 0 and 168.
        </span>
      )}
      <button type="submit" disabled={saving} aria-disabled={saving || undefined}>
        {saving ? 'Saving…' : 'Save'}
      </button>
      <button type="button" onClick={onCancel} disabled={saving}>
        Cancel
      </button>
    </form>
  )
}

export function CapacityGrid({ from, to }: Props) {
  const [data, setData] = useState<CapacityResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editValue, setEditValue] = useState('')
  const [savingId, setSavingId] = useState<number | null>(null)
  const [announce, setAnnounce] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  const rows = data?.rows ?? []
  const colCount = 2 + (data?.weeks.length ?? 0)

  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 12,
  })

  useEffect(() => {
    const ctrl = new AbortController()
    setLoading(true)
    setError(null)
    setEditingId(null)

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
        setAnnounce(`Loaded capacity for ${json.rows.length} people.`)
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(err instanceof Error ? err.message : 'Failed to load capacity')
        setLoading(false)
      })

    return () => ctrl.abort()
  }, [from, to])

  const editInvalid = (() => {
    if (editingId === null) return false
    if (editValue.trim() === '') return true
    const n = Number(editValue)
    return Number.isNaN(n) || n < 0 || n > 168
  })()

  function startEdit(row: CapacityRow) {
    setError(null)
    setEditingId(row.personId)
    setEditValue(String(row.weeklyHours))
  }

  function cancelEdit() {
    setEditingId(null)
    setEditValue('')
  }

  async function saveEdit(personId: number) {
    const next = Number(editValue)
    if (Number.isNaN(next) || next < 0 || next > 168) {
      setError('Weekly hours must be a number between 0 and 168')
      return
    }

    const personName = data?.rows.find((r) => r.personId === personId)?.name ?? 'person'
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
      setAnnounce(`Updated ${personName} capacity to ${formatHours(person.weeklyHours)} hours.`)
      cancelEdit()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save')
    } finally {
      setSavingId(null)
    }
  }

  if (loading && !data) {
    return (
      <p className="status" role="status" aria-live="polite">
        Loading capacity…
      </p>
    )
  }
  if (error && !data) {
    return (
      <p className="status error" role="alert">
        {error}
      </p>
    )
  }
  if (!data) {
    return <p className="status">No data</p>
  }

  const virtualRows = virtualizer.getVirtualItems()
  const totalSize = virtualizer.getTotalSize()
  const paddingTop = virtualRows.length > 0 ? virtualRows[0]!.start : 0
  const paddingBottom =
    virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1]!.end : 0

  return (
    <div className="grid-shell">
      <div className="sr-only" aria-live="polite" aria-atomic="true">
        {announce}
      </div>
      {error && (
        <p className="status error" role="alert">
          {error}
        </p>
      )}
      {loading && (
        <p className="status" role="status" aria-live="polite">
          Refreshing…
        </p>
      )}

      <div
        ref={scrollRef}
        className="grid-wrap"
        tabIndex={0}
        role="region"
        aria-label="Team capacity grid"
      >
        <table className="capacity-grid">
          <caption className="sr-only">
            Allocated hours versus weekly capacity for each person, by week from {data.from} to{' '}
            {data.to}. Over-allocated cells are marked.
          </caption>
          <thead>
            <tr>
              <th scope="col" className="sticky-col">
                Person
              </th>
              <th scope="col" className="capacity-col">
                Capacity
              </th>
              {data.weeks.map((w) => (
                <th key={w.start} scope="col" title={`${w.start} → ${w.end}`}>
                  Week of {weekLabel(w)}
                  <span className="sr-only">
                    {' '}
                    ({w.start} to {w.end})
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paddingTop > 0 && (
              <tr aria-hidden="true" className="virt-pad">
                <td colSpan={colCount} style={{ height: paddingTop }} />
              </tr>
            )}
            {virtualRows.map((vRow) => {
              const row = rows[vRow.index]!
              return (
                <tr key={row.personId} data-index={vRow.index}>
                  <th scope="row" className="sticky-col">
                    {row.name}
                  </th>
                  <td className="capacity-col">
                    <CapacityEdit
                      row={row}
                      editing={editingId === row.personId}
                      value={editValue}
                      saving={savingId === row.personId}
                      invalid={editingId === row.personId && editInvalid}
                      onStart={() => startEdit(row)}
                      onChange={setEditValue}
                      onSave={() => void saveEdit(row.personId)}
                      onCancel={cancelEdit}
                    />
                  </td>
                  {row.weeks.map((cell) => (
                    <td
                      key={cell.weekStart}
                      className={cell.over ? 'over' : cell.allocated > 0 ? 'used' : ''}
                    >
                      <span className="alloc">{formatHours(cell.allocated)}</span>
                      <span className="sep" aria-hidden="true">
                        /
                      </span>
                      <span className="cap">{formatHours(cell.capacity)}</span>
                      <span className="sr-only">
                        {' '}
                        hours allocated of {formatHours(cell.capacity)} capacity
                        {cell.over ? ', over allocated' : ''}
                      </span>
                    </td>
                  ))}
                </tr>
              )
            })}
            {paddingBottom > 0 && (
              <tr aria-hidden="true" className="virt-pad">
                <td colSpan={colCount} style={{ height: paddingBottom }} />
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <p className="grid-meta" aria-hidden="true">
        {rows.length} people · showing {virtualRows.length || 0} rows
      </p>
    </div>
  )
}
