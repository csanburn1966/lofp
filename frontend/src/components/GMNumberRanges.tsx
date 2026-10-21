import { useState, useEffect, useMemo } from 'react'

interface NumberBlock {
  start: number
  end: number
  source: string
  gm?: boolean
  draft?: boolean
  author?: string
}

interface NumberGap {
  start: number
  end: number
  size: number
}

interface NumberConflict {
  number: number
  draft: string
  others: string[]
}

type Kind = 'room' | 'item' | 'monster' | 'adjective' | 'monadj' | 'noun'

interface KindRanges {
  kind: Kind
  count: number
  min: number
  max: number
  blocks: NumberBlock[]
  gaps: NumberGap[]
  conflicts?: NumberConflict[]
  names?: Record<string, string> // vocabulary kinds only: number -> word
}

interface Report {
  kinds: KindRanges[]
  drafts: string[]
}

// Advisory only — the server never enforces these, so GMs can still share numbers.
interface Reservation {
  id: string
  kind: Kind
  start: number
  end: number
  gm: string
  note?: string
  createdBy: string
  createdAt: string
  updatedBy?: string
  updatedAt?: string
}

interface ResForm {
  id: string | null
  gm: string
  start: string
  end: string
  note: string
}

interface Props {
  authHeaders: () => Record<string, string>
  characterName: string
}

const KINDS: Kind[] = ['room', 'item', 'monster', 'adjective', 'monadj', 'noun']
const KIND_LABEL: Record<Kind, string> = {
  room: 'Rooms', item: 'Items', monster: 'Monsters',
  adjective: 'Adjectives', monadj: 'Monster Adjectives', noun: 'Nouns',
}
// The script directive that defines each vocabulary kind.
const VOCAB_DIRECTIVE: Partial<Record<Kind, string>> = { adjective: 'ADJDEF', monadj: 'MADJDEF', noun: 'NOUNDEF' }
const WORD_LIMIT = 500
const BUCKET = 100

const rangeText = (a: number, b: number) => (a === b ? `${a}` : `${a}–${b}`)

const overlaps = (a1: number, a2: number, b1: number, b2: number) => a1 <= b2 && b1 <= a2

// Remove reserved spans from a list of [start, end] intervals.
function subtractSpans(spans: [number, number][], cut: [number, number][]): [number, number][] {
  let out = spans
  for (const [cs, ce] of cut) {
    const next: [number, number][] = []
    for (const [s, e] of out) {
      if (!overlaps(s, e, cs, ce)) { next.push([s, e]); continue }
      if (s < cs) next.push([s, cs - 1])
      if (e > ce) next.push([ce + 1, e])
    }
    out = next
  }
  return out
}

// A free block of `size` numbers inside [start, end] (end = Infinity for "above max"),
// nudged up to a round start number when it still fits — round blocks are easier to
// hand out and remember.
function placeBlock(start: number, end: number, size: number): number | null {
  if (end - start + 1 < size) return null
  const align = size >= 100 ? 100 : size >= 10 ? 10 : 1
  const aligned = Math.ceil(start / align) * align
  return aligned + size - 1 <= end ? aligned : start
}

export default function GMNumberRanges({ authHeaders, characterName }: Props) {
  const [report, setReport] = useState<Report | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [kind, setKind] = useState<Kind>('room')
  const [wordFilter, setWordFilter] = useState('')
  const [wantSize, setWantSize] = useState(100)
  const [floor, setFloor] = useState(0)
  const [minGap, setMinGap] = useState(10)
  const [gmOnly, setGmOnly] = useState(false)
  const [sourceFilter, setSourceFilter] = useState('')
  const [reservations, setReservations] = useState<Reservation[]>([])
  const [skipReserved, setSkipReserved] = useState(true)
  const [form, setForm] = useState<ResForm | null>(null)
  const [resStatus, setResStatus] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const load = () => {
    setLoading(true)
    setError(null)
    fetch('/api/gm/ranges', { headers: authHeaders() })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        return r.json()
      })
      .then((data: Report) => setReport(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }

  const loadReservations = () => {
    fetch('/api/gm/reservations', { headers: authHeaders() })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        return r.json()
      })
      .then((data: Reservation[]) => setReservations(data || []))
      .catch(err => setResStatus({ type: 'error', message: `Could not load reservations: ${err.message}` }))
  }

  useEffect(() => { load(); loadReservations() }, [])

  const kindRes = reservations.filter(r => r.kind === kind)
  const isMine = (r: Reservation) => r.gm.toLowerCase() === characterName.toLowerCase()

  const openForm = (prefill?: Partial<ResForm>) => {
    setResStatus(null)
    setForm({ id: null, gm: characterName, start: '', end: '', note: '', ...prefill })
  }

  const saveReservation = () => {
    if (!form) return
    const start = parseInt(form.start), end = parseInt(form.end)
    if (isNaN(start) || isNaN(end) || start < 0 || end < start) {
      setResStatus({ type: 'error', message: 'Enter a valid range (end must be at or above start).' })
      return
    }
    const gm = form.gm.trim()
    if (!gm) {
      setResStatus({ type: 'error', message: 'Enter the GM this block is for.' })
      return
    }
    // Overlap is allowed (GMs may share numbers) — just make sure it's intentional.
    const clash = kindRes.filter(r => r.id !== form.id && overlaps(start, end, r.start, r.end))
    if (clash.length > 0 && !confirm(
      `This overlaps ${clash.map(r => `${r.gm}'s ${rangeText(r.start, r.end)}`).join(', ')}. Reservations are only a reference, so this is allowed. Save anyway?`)) return
    fetch(form.id ? `/api/gm/reservations/${form.id}` : '/api/gm/reservations', {
      method: form.id ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ kind, start, end, gm, note: form.note }),
    })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        setResStatus({ type: 'success', message: `Reserved ${KIND_LABEL[kind].toLowerCase()} ${rangeText(start, end)} for ${gm}.` })
        setForm(null)
        loadReservations()
      })
      .catch(err => setResStatus({ type: 'error', message: err.message }))
  }

  const deleteReservation = (res: Reservation) => {
    if (!confirm(`Remove ${res.gm}'s reservation of ${KIND_LABEL[res.kind].toLowerCase()} ${rangeText(res.start, res.end)}? Nothing in the game changes.`)) return
    fetch(`/api/gm/reservations/${res.id}`, { method: 'DELETE', headers: authHeaders() })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        setResStatus(null)
        loadReservations()
      })
      .catch(err => setResStatus({ type: 'error', message: err.message }))
  }

  const data = report?.kinds.find(k => k.kind === kind)

  // Vocabulary kinds: every number with its word and defining file, filtered by
  // number (exact) or word (substring).
  const words = useMemo(() => {
    if (!data?.names) return []
    const sourceOf = (n: number) => data.blocks.find(b => b.start <= n && n <= b.end)?.source || ''
    const q = wordFilter.trim().toLowerCase()
    return Object.entries(data.names)
      .map(([n, word]) => ({ n: Number(n), word }))
      .filter(w => !q || String(w.n) === q || w.word.toLowerCase().includes(q))
      .sort((a, b) => a.n - b.n)
      .map(w => ({ ...w, source: sourceOf(w.n) }))
  }, [data, wordFilter])

  // Used count per 100-number bucket, for the occupancy strip.
  const buckets = useMemo(() => {
    if (!data || data.count === 0) return []
    const top = Math.ceil((data.max + 1) / 1000) * 1000 + 1000
    const counts = new Array(top / BUCKET).fill(0)
    for (const b of data.blocks) {
      for (let i = Math.floor(b.start / BUCKET); i <= Math.floor(b.end / BUCKET); i++) {
        const lo = Math.max(b.start, i * BUCKET)
        const hi = Math.min(b.end, i * BUCKET + BUCKET - 1)
        counts[i] += hi - lo + 1
      }
    }
    return counts as number[]
  }, [data])

  const usedIn = (start: number, end: number) => {
    if (!data) return 0
    let n = 0
    for (const b of data.blocks) {
      if (overlaps(b.start, b.end, start, end)) n += Math.min(b.end, end) - Math.max(b.start, start) + 1
    }
    return n
  }

  const suggestions = useMemo(() => {
    if (!data || wantSize < 1) return []
    const out: { start: number; end: number; gapStart: number; gapEnd: number | null }[] = []
    const gaps: [number, number][] = data.gaps.map(g => [g.start, g.end])
    gaps.push([data.max + 1, Infinity])
    const reserved: [number, number][] = skipReserved
      ? reservations.filter(r => r.kind === kind).map(r => [r.start, r.end])
      : []
    for (const [gs, ge] of gaps) {
      for (const [ps, pe] of subtractSpans([[gs, ge]], reserved)) {
        if (pe < floor) continue
        const s = placeBlock(Math.max(ps, floor), pe, wantSize)
        if (s === null) continue
        out.push({ start: s, end: s + wantSize - 1, gapStart: gs, gapEnd: ge === Infinity ? null : ge })
        if (out.length >= 6) return out
      }
    }
    return out
  }, [data, wantSize, floor, skipReserved, reservations, kind])

  const bucketRes = (i: number) => {
    const names = kindRes.filter(r => overlaps(r.start, r.end, i * BUCKET, i * BUCKET + BUCKET - 1)).map(r => r.gm)
    return names.length ? ` — reserved: ${[...new Set(names)].join(', ')}` : ''
  }
  const gapRes = (start: number, end: number) => kindRes.filter(r => overlaps(r.start, r.end, start, end))

  const visibleGaps = data ? data.gaps.filter(g => g.size >= minGap) : []
  const visibleBlocks = data
    ? data.blocks.filter(b =>
        (!gmOnly || b.gm) &&
        (!sourceFilter || b.source.toLowerCase().includes(sourceFilter.toLowerCase())))
    : []

  const input = 'bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-24'

  return (
    <div className="flex-1 overflow-y-auto p-3 sm:p-4 space-y-5">
      {/* Kind selector */}
      <div className="flex flex-wrap items-center gap-2">
        {KINDS.map(k => {
          const kd = report?.kinds.find(x => x.kind === k)
          return (
            <button
              key={k}
              onClick={() => setKind(k)}
              className={`px-3 py-1.5 rounded text-sm ${kind === k ? 'bg-amber-700 text-white' : 'bg-[#222] text-gray-300 hover:bg-[#333]'}`}
            >
              {KIND_LABEL[k]}{kd ? ` (${kd.count})` : ''}
            </button>
          )
        })}
        <button onClick={load} disabled={loading} className="ml-auto px-3 py-1.5 bg-[#333] hover:bg-[#444] text-gray-300 rounded text-sm disabled:opacity-50">
          {loading ? 'Loading...' : 'Refresh'}
        </button>
      </div>

      {error && <div className="px-3 py-2 bg-red-900/30 text-red-400">{error}</div>}

      {data && (
        <>
          {/* Summary */}
          <div className="text-gray-300 space-y-1">
            <div>
              <span className="text-amber-400">{data.count}</span> {KIND_LABEL[kind].toLowerCase()} defined, numbered{' '}
              <span className="text-amber-400">{data.min}</span> to <span className="text-amber-400">{data.max}</span>.
            </div>
            <div className="text-gray-500 text-xs">
              The engine has no maximum number — everything from {data.max + 1} upward is free.
              Counts include the live world and {report!.drafts.length} draft GM script{report!.drafts.length === 1 ? '' : 's'}
              {report!.drafts.length > 0 ? ` (${report!.drafts.join(', ')})` : ''}.
              {kind === 'item' && ' ANTI.SCR placeholder items count as free.'}
              {VOCAB_DIRECTIVE[kind] && ` Defined with ${VOCAB_DIRECTIVE[kind]} <number> <word>.`}
            </div>
          </div>

          {/* Word list (vocabulary kinds) */}
          {data.names && (
            <div>
              <div className="flex flex-wrap items-center gap-3 mb-2">
                <div className="text-amber-400 font-bold">{KIND_LABEL[kind]}</div>
                <input
                  value={wordFilter}
                  onChange={e => setWordFilter(e.target.value)}
                  placeholder="search word or number"
                  className="bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-48"
                />
                <span className="text-gray-600 text-xs">
                  {words.length > WORD_LIMIT ? `first ${WORD_LIMIT} of ${words.length}` : `${words.length}`} shown
                </span>
              </div>
              <div className="border border-[#333] rounded max-h-96 overflow-y-auto">
                {words.slice(0, WORD_LIMIT).map(w => (
                  <div key={w.n} className="flex flex-wrap items-center gap-x-3 px-3 py-1 border-b border-[#222] text-sm">
                    <span className="text-gray-400 w-16 shrink-0 text-right">{w.n}</span>
                    <span className="text-gray-200">{w.word}</span>
                    <span className="text-gray-600 text-xs ml-auto truncate">{w.source}</span>
                  </div>
                ))}
                {words.length === 0 && <div className="px-3 py-2 text-gray-600">No matching {KIND_LABEL[kind].toLowerCase()}</div>}
              </div>
            </div>
          )}

          {/* Conflicts */}
          {data.conflicts && data.conflicts.length > 0 && (
            <div className="border border-red-800 rounded">
              <div className="px-3 py-1.5 bg-red-900/40 text-red-300 text-xs font-bold">
                {data.conflicts.length} number{data.conflicts.length === 1 ? '' : 's'} in draft scripts already used elsewhere
              </div>
              <div className="max-h-40 overflow-y-auto">
                {data.conflicts.map((c, i) => (
                  <div key={i} className="px-3 py-1 text-xs border-t border-[#222] text-gray-300">
                    <span className="text-red-400">{c.number}</span> — draft {c.draft} vs {c.others.join(', ')}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Occupancy strip */}
          <div>
            <div className="text-gray-500 text-xs mb-1">Occupancy, one cell per {BUCKET} numbers (hover for details)</div>
            <div className="flex flex-wrap gap-px">
              {buckets.map((n, i) => (
                <div
                  key={i}
                  title={`${i * BUCKET}–${i * BUCKET + BUCKET - 1}: ${n} used, ${BUCKET - n} free${bucketRes(i)}`}
                  className={`w-3 h-5 ${n === 0 ? 'bg-[#1d2b1d]' : n === BUCKET ? 'bg-amber-600' : n >= BUCKET / 2 ? 'bg-amber-800' : 'bg-amber-950'} ${bucketRes(i) ? 'border-b-4 border-b-sky-500' : ''}`}
                />
              ))}
            </div>
            <div className="flex flex-wrap gap-3 mt-1 text-[10px] text-gray-500">
              <span><span className="inline-block w-2 h-2 bg-[#1d2b1d] mr-1" />empty</span>
              <span><span className="inline-block w-2 h-2 bg-amber-950 mr-1" />under half used</span>
              <span><span className="inline-block w-2 h-2 bg-amber-800 mr-1" />half or more used</span>
              <span><span className="inline-block w-2 h-2 bg-amber-600 mr-1" />full</span>
              <span><span className="inline-block w-2 h-1 bg-sky-500 mr-1 align-middle" />reserved</span>
            </div>
          </div>

          {/* Reservations */}
          <div className="border border-sky-900 rounded bg-[#0d1218]">
            <div className="flex flex-wrap items-center gap-2 px-3 py-2 border-b border-sky-900">
              <span className="text-sky-400 font-bold">Reserved {KIND_LABEL[kind].toLowerCase()} blocks</span>
              <span className="text-gray-600 text-xs">a shared reference, not enforced — overlapping is fine</span>
              {!form && (
                <button onClick={() => openForm()} className="ml-auto px-2 py-1 bg-sky-800 hover:bg-sky-700 text-white rounded text-xs">
                  + Reserve a block
                </button>
              )}
            </div>
            {resStatus && (
              <div className={`px-3 py-1.5 text-xs ${resStatus.type === 'success' ? 'text-green-400' : 'text-red-400'}`}>{resStatus.message}</div>
            )}
            {form && (
              <div className="flex flex-wrap items-end gap-3 px-3 py-2 border-b border-sky-900 bg-[#101820]">
                <label className="flex flex-col gap-1 text-gray-500 text-xs">
                  For GM
                  <input value={form.gm} onChange={e => setForm({ ...form, gm: e.target.value })} className={`${input} w-32`} />
                </label>
                <label className="flex flex-col gap-1 text-gray-500 text-xs">
                  Start
                  <input type="number" min={0} value={form.start} onChange={e => setForm({ ...form, start: e.target.value })} className={input} />
                </label>
                <label className="flex flex-col gap-1 text-gray-500 text-xs">
                  End
                  <input type="number" min={0} value={form.end} onChange={e => setForm({ ...form, end: e.target.value })} className={input} />
                </label>
                <label className="flex flex-col gap-1 text-gray-500 text-xs flex-1 min-w-40">
                  Note (optional)
                  <input value={form.note} maxLength={500} onChange={e => setForm({ ...form, note: e.target.value })} placeholder="e.g. Nyraine ferry area" className={`${input} w-full`} />
                </label>
                <div className="flex gap-2">
                  <button onClick={saveReservation} className="px-3 py-1.5 bg-sky-700 hover:bg-sky-600 text-white rounded text-sm">
                    {form.id ? 'Save' : 'Reserve'}
                  </button>
                  <button onClick={() => setForm(null)} className="px-3 py-1.5 bg-[#333] hover:bg-[#444] text-gray-300 rounded text-sm">
                    Cancel
                  </button>
                </div>
              </div>
            )}
            {kindRes.length === 0 ? (
              <div className="px-3 py-2 text-gray-600 text-xs">No {KIND_LABEL[kind].toLowerCase()} blocks reserved yet.</div>
            ) : (
              <div className="max-h-64 overflow-y-auto">
                {kindRes.map(r => {
                  const size = r.end - r.start + 1
                  const shared = kindRes.filter(o => o.id !== r.id && overlaps(o.start, o.end, r.start, r.end))
                  return (
                    <div key={r.id} className={`px-3 py-1.5 border-b border-[#1a2430] text-sm ${isMine(r) ? 'bg-sky-950/40' : ''}`}>
                      <div className="flex flex-wrap items-center gap-x-3">
                        <span className="text-gray-200 w-28 shrink-0">{rangeText(r.start, r.end)}</span>
                        <span className={isMine(r) ? 'text-sky-300 font-bold' : 'text-sky-400'}>{r.gm}{isMine(r) ? ' (you)' : ''}</span>
                        <span className="text-gray-500 text-xs">{usedIn(r.start, r.end)} of {size} in use</span>
                        {shared.length > 0 && (
                          <span className="text-amber-500 text-xs">shared with {shared.map(o => o.gm).join(', ')}</span>
                        )}
                        <span className="ml-auto flex gap-3 text-xs">
                          <button
                            onClick={() => { setResStatus(null); setForm({ id: r.id, gm: r.gm, start: String(r.start), end: String(r.end), note: r.note || '' }) }}
                            className="text-amber-500 hover:text-amber-400"
                          >
                            Edit
                          </button>
                          <button onClick={() => deleteReservation(r)} className="text-red-500 hover:text-red-400">Remove</button>
                        </span>
                      </div>
                      {r.note && <div className="text-gray-400 text-xs mt-0.5">{r.note}</div>}
                      <div className="text-gray-600 text-[10px] mt-0.5">
                        set by {r.createdBy} {new Date(r.createdAt).toLocaleDateString()}
                        {r.updatedBy && r.updatedAt ? ` · edited by ${r.updatedBy} ${new Date(r.updatedAt).toLocaleDateString()}` : ''}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          {/* Block finder */}
          <div className="border border-[#333] rounded p-3 bg-[#111] space-y-2">
            <div className="text-amber-400 font-bold">Find a free block</div>
            <div className="flex flex-wrap gap-3 items-end">
              <label className="flex flex-col gap-1 text-gray-500 text-xs">
                Block size
                <input type="number" min={1} value={wantSize} onChange={e => setWantSize(parseInt(e.target.value) || 0)} className={input} />
              </label>
              <label className="flex flex-col gap-1 text-gray-500 text-xs">
                Starting at or above
                <input type="number" min={0} value={floor} onChange={e => setFloor(parseInt(e.target.value) || 0)} className={input} />
              </label>
              <label className="flex items-center gap-1.5 text-gray-400 text-xs pb-1.5">
                <input type="checkbox" checked={skipReserved} onChange={e => setSkipReserved(e.target.checked)} />
                Skip reserved blocks
              </label>
            </div>
            {suggestions.length === 0 ? (
              <div className="text-gray-600 text-xs">Enter a block size.</div>
            ) : (
              <div className="space-y-1">
                {suggestions.map((s, i) => (
                  <div key={i} className="flex flex-wrap items-center gap-x-2 text-sm">
                    <span className="text-green-400">{rangeText(s.start, s.end)}</span>
                    <span className="text-gray-600 text-xs">
                      {s.gapEnd === null ? `above the current max (${data.max})` : `in free gap ${rangeText(s.gapStart, s.gapEnd)}`}
                    </span>
                    <button
                      onClick={() => openForm({ start: String(s.start), end: String(s.end) })}
                      className="text-sky-400 hover:text-sky-300 text-xs"
                    >
                      Reserve
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Free gaps */}
          <div>
            <div className="flex flex-wrap items-end gap-3 mb-2">
              <div className="text-amber-400 font-bold">Free gaps</div>
              <label className="flex items-center gap-2 text-gray-500 text-xs">
                at least
                <input type="number" min={1} value={minGap} onChange={e => setMinGap(parseInt(e.target.value) || 1)} className={input} />
                numbers
              </label>
            </div>
            <div className="border border-[#333] rounded max-h-72 overflow-y-auto">
              {visibleGaps.map((g, i) => (
                <div key={i} className="flex flex-wrap items-center gap-x-3 px-3 py-1 border-b border-[#222] text-sm">
                  <span className="text-green-400">{rangeText(g.start, g.end)}</span>
                  {gapRes(g.start, g.end).map(r => (
                    <span key={r.id} className="text-sky-400 text-xs">{r.gm}: {rangeText(r.start, r.end)}</span>
                  ))}
                  <span className="text-gray-500 ml-auto">{g.size}</span>
                </div>
              ))}
              <div className="flex flex-wrap items-center gap-x-3 px-3 py-1 text-sm">
                <span className="text-green-400">{data.max + 1} and up</span>
                {gapRes(data.max + 1, Number.MAX_SAFE_INTEGER).map(r => (
                  <span key={r.id} className="text-sky-400 text-xs">{r.gm}: {rangeText(r.start, r.end)}</span>
                ))}
                <span className="text-gray-500 ml-auto">unlimited</span>
              </div>
            </div>
          </div>

          {/* Used blocks */}
          <div>
            <div className="flex flex-wrap items-center gap-3 mb-2">
              <div className="text-amber-400 font-bold">Used blocks</div>
              <input
                value={sourceFilter}
                onChange={e => setSourceFilter(e.target.value)}
                placeholder="filter by file"
                className="bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-40"
              />
              <label className="flex items-center gap-1.5 text-gray-400 text-xs">
                <input type="checkbox" checked={gmOnly} onChange={e => setGmOnly(e.target.checked)} />
                GM scripts only
              </label>
              <span className="text-gray-600 text-xs">{visibleBlocks.length} blocks</span>
            </div>
            <div className="border border-[#333] rounded max-h-96 overflow-y-auto">
              {visibleBlocks.map((b, i) => (
                <div key={i} className="flex flex-wrap items-center gap-x-3 px-3 py-1 border-b border-[#222] text-sm">
                  <span className="text-gray-200 w-28 shrink-0">{rangeText(b.start, b.end)}</span>
                  <span className="text-gray-400 truncate">{b.source || '(unknown)'}</span>
                  {b.gm && (
                    <span className={`px-1.5 rounded text-[10px] font-bold ${b.draft ? 'bg-[#333] text-gray-300' : 'bg-green-900 text-green-300'}`}>
                      {b.draft ? 'DRAFT' : 'GM'}
                    </span>
                  )}
                  {b.author && <span className="text-gray-600 text-xs">{b.author}</span>}
                  <span className="text-gray-600 text-xs ml-auto">{b.end - b.start + 1}</span>
                </div>
              ))}
              {visibleBlocks.length === 0 && <div className="px-3 py-2 text-gray-600">No matching blocks</div>}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
