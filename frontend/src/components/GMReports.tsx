import { useState, useEffect } from 'react'

// A player REPORT entry, as stored in the game log.
interface ReportEntry {
  id: string
  timestamp: string
  player: string
  details?: string
  roomNum?: number
  roomName?: string
  resolved?: boolean
  resolvedBy?: string
  resolvedAt?: string
  reply?: string
  repliedBy?: string
  repliedAt?: string
  replyDeliveredAt?: string
}

interface Props {
  authHeaders: () => Record<string, string>
  // Called with the number of unresolved reports whenever it's known, so the
  // GM page tab can show it.
  onUnresolvedCount?: (n: number) => void
}

const formatTimestamp = (dateStr?: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export default function GMReports({ authHeaders, onUnresolvedCount }: Props) {
  const [reports, setReports] = useState<ReportEntry[]>([])
  const [resolvedFilter, setResolvedFilter] = useState('false')
  const [playerFilter, setPlayerFilter] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [replyingTo, setReplyingTo] = useState<string | null>(null)
  const [replyDraft, setReplyDraft] = useState('')

  const load = () => {
    setLoading(true)
    setError(null)
    const params = new URLSearchParams({ limit: '200' })
    if (resolvedFilter) params.set('resolved', resolvedFilter)
    if (playerFilter.trim()) params.set('player', playerFilter.trim())
    fetch(`/api/gm/reports?${params}`, { headers: authHeaders() })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        return r.json()
      })
      .then((data: ReportEntry[]) => {
        setReports(data || [])
        if (resolvedFilter === 'false' && !playerFilter.trim()) onUnresolvedCount?.((data || []).length)
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [resolvedFilter])

  const setResolved = (id: string, resolved: boolean) => {
    fetch(`/api/gm/reports/${id}/resolve`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ resolved }),
    })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        return r.json()
      })
      .then((updated: ReportEntry) => {
        setReports(prev => {
          // Drop an entry that no longer matches the resolved filter.
          if (resolvedFilter === 'true' && !updated.resolved) return prev.filter(e => e.id !== id)
          if (resolvedFilter === 'false' && updated.resolved) {
            const next = prev.filter(e => e.id !== id)
            if (!playerFilter.trim()) onUnresolvedCount?.(next.length)
            return next
          }
          return prev.map(e => e.id === id ? updated : e)
        })
      })
      .catch(err => setError(err.message))
  }

  // Empty text clears the reply.
  const saveReply = (id: string, reply: string) => {
    fetch(`/api/gm/reports/${id}/reply`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ reply }),
    })
      .then(async r => {
        if (!r.ok) throw new Error((await r.text()) || `HTTP ${r.status}`)
        return r.json()
      })
      .then((updated: ReportEntry) => {
        setReports(prev => prev.map(e => e.id === id ? updated : e))
        setReplyingTo(null)
        setReplyDraft('')
      })
      .catch(err => setError(err.message))
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="flex flex-wrap gap-2 p-3 bg-[#111] border-b border-[#333] items-center">
        <select
          value={resolvedFilter}
          onChange={e => setResolvedFilter(e.target.value)}
          className="bg-[#0a0a0a] border border-[#444] rounded px-2 py-1 text-gray-200 focus:border-amber-500 focus:outline-none text-xs"
        >
          <option value="false">Unresolved</option>
          <option value="true">Resolved</option>
          <option value="">All</option>
        </select>
        <input
          type="text"
          value={playerFilter}
          onChange={e => setPlayerFilter(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') load() }}
          placeholder="Filter by player..."
          className="bg-[#0a0a0a] border border-[#444] rounded px-2 py-1 text-gray-200 focus:border-amber-500 focus:outline-none text-xs w-40"
        />
        <button
          onClick={load}
          disabled={loading}
          className="px-3 py-1 bg-amber-700 text-white rounded text-xs hover:bg-amber-600 disabled:opacity-50"
        >
          {loading ? 'Loading...' : 'Refresh'}
        </button>
        <span className="text-gray-600 text-xs ml-auto">{reports.length} report{reports.length === 1 ? '' : 's'}</span>
      </div>

      {error && <div className="px-3 py-2 bg-red-900/30 text-red-400 text-xs">{error}</div>}

      <div className="flex-1 overflow-y-auto">
        {reports.map(entry => (
          <div key={entry.id} className={`px-3 py-2.5 border-b border-[#222] ${entry.resolved ? 'opacity-50' : ''}`}>
            <div className="flex flex-wrap items-baseline gap-x-3 text-xs">
              <span className="text-red-400 font-bold">{entry.player || 'Unknown'}</span>
              <span className="text-gray-500">{formatTimestamp(entry.timestamp)}</span>
              {entry.roomNum ? (
                <span className="text-gray-600">Room {entry.roomNum}{entry.roomName ? ` — ${entry.roomName}` : ''}</span>
              ) : null}
            </div>
            <div className={`mt-1 text-gray-200 whitespace-pre-wrap break-words ${entry.resolved ? 'line-through' : ''}`}>
              {entry.details}
            </div>

            {entry.reply && replyingTo !== entry.id && (
              <div className="mt-1 text-amber-300/80 text-xs">
                ↳ {entry.reply}
                <span className="text-gray-600 ml-2">
                  {entry.replyDeliveredAt ? `(seen ${formatTimestamp(entry.replyDeliveredAt)})` : '(not yet seen)'}
                  {entry.repliedBy ? ` — ${entry.repliedBy}` : ''}
                </span>
              </div>
            )}

            {replyingTo === entry.id ? (
              <div className="mt-2 flex flex-col gap-1">
                <textarea
                  value={replyDraft}
                  onChange={e => setReplyDraft(e.target.value)}
                  maxLength={2000}
                  rows={3}
                  autoFocus
                  placeholder={`Reply to ${entry.player}, shown on their next login (or right away if they're online)`}
                  className="bg-[#0a0a0a] border border-[#444] rounded px-2 py-1 text-gray-200 focus:border-amber-500 focus:outline-none text-xs w-full"
                />
                <div className="flex gap-2">
                  <button
                    onClick={() => saveReply(entry.id, replyDraft)}
                    disabled={!replyDraft.trim()}
                    className="px-2 py-0.5 bg-amber-700 text-white rounded text-xs hover:bg-amber-600 disabled:opacity-40"
                  >
                    Send Reply
                  </button>
                  {entry.reply && (
                    <button
                      onClick={() => { if (confirm('Remove this reply?')) saveReply(entry.id, '') }}
                      className="px-2 py-0.5 rounded border border-red-800 text-red-400 hover:bg-red-900/30 text-xs"
                    >
                      Remove
                    </button>
                  )}
                  <button
                    onClick={() => { setReplyingTo(null); setReplyDraft('') }}
                    className="px-2 py-0.5 rounded border border-[#444] text-gray-400 hover:bg-[#222] text-xs"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <div className="mt-2 flex flex-wrap gap-2 items-center">
                {entry.resolved ? (
                  <button
                    onClick={() => setResolved(entry.id, false)}
                    title="Reopen this report"
                    className="px-2 py-0.5 rounded border border-green-700 text-green-400 hover:bg-green-900/30 text-xs"
                  >
                    ✓ Done
                  </button>
                ) : (
                  <button
                    onClick={() => setResolved(entry.id, true)}
                    className="px-2 py-0.5 rounded border border-[#444] text-gray-300 hover:bg-amber-900/30 hover:border-amber-600 hover:text-amber-300 text-xs"
                  >
                    Mark Done
                  </button>
                )}
                <button
                  onClick={() => { setReplyingTo(entry.id); setReplyDraft(entry.reply || '') }}
                  className="px-2 py-0.5 rounded border border-[#444] text-gray-300 hover:bg-blue-900/30 hover:border-blue-600 hover:text-blue-300 text-xs"
                >
                  {entry.reply ? 'Edit Reply' : 'Reply'}
                </button>
                {entry.resolved && entry.resolvedBy && (
                  <span className="text-gray-600 text-xs">
                    resolved by {entry.resolvedBy}{entry.resolvedAt ? ` ${formatTimestamp(entry.resolvedAt)}` : ''}
                  </span>
                )}
              </div>
            )}
          </div>
        ))}
        {reports.length === 0 && !loading && (
          <div className="px-3 py-8 text-center text-gray-600">
            {resolvedFilter === 'false' ? 'No unresolved reports' : 'No reports'}
          </div>
        )}
      </div>
    </div>
  )
}
