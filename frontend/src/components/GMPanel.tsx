import { useState, useEffect, useRef } from 'react'
import { useAuth } from '../App'
import GMNumberRanges from './GMNumberRanges'
import GMReports from './GMReports'

interface ScriptMeta {
  filename: string
  name: string
  priority: number
  size: number
  uploadedBy: string
  uploadedAt: string
  draft?: boolean
  parseStats: { rooms: number; items: number; monsters: number; nouns: number; variables: number }
}

interface WithdrawStats {
  rooms: number
  items: number
  monsters: number
  macros: number
  cevents: number
  despawned: number
}

const withdrawnText = (w?: WithdrawStats | null) => {
  if (!w) return ''
  const parts = [`${w.rooms} rooms`, `${w.items} items`, `${w.monsters} monsters`]
  if (w.cevents) parts.push(`${w.cevents} timed events`)
  let text = ` Restored from the previous live version: ${parts.join(', ')}.`
  if (w.despawned) text += ` ${w.despawned} creature(s) despawned.`
  return text
}

interface ScriptVersion {
  content: string
  uploadedBy: string
  uploadedAt: string
}

interface ScriptFull extends ScriptMeta {
  content: string
  history: ScriptVersion[]
}

interface Props {
  characterName: string
}

export default function GMPanel({ characterName }: Props) {
  const { user } = useAuth()
  const [scripts, setScripts] = useState<ScriptMeta[]>([])
  const [selected, setSelected] = useState<ScriptFull | null>(null)
  const [editing, setEditing] = useState(false)
  const [editorContent, setEditorContent] = useState('')
  const [editorName, setEditorName] = useState('')
  const [editorPriority, setEditorPriority] = useState(100)
  const [editorFilename, setEditorFilename] = useState('')
  const [isNew, setIsNew] = useState(false)
  const [status, setStatus] = useState<{ type: 'success' | 'error'; message: string } | null>(null)
  const [saving, setSaving] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [view, setView] = useState<'scripts' | 'reports' | 'ranges'>('scripts')
  const [unresolvedReports, setUnresolvedReports] = useState<number | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const authHeaders = (): Record<string, string> => {
    const headers: Record<string, string> = {}
    if (user?.token) headers['Authorization'] = `Bearer ${user.token}`
    headers['X-Character'] = characterName
    return headers
  }

  const loadScripts = () => {
    fetch('/api/gm/scripts', { headers: authHeaders() })
      .then(r => r.json())
      .then((data: ScriptMeta[]) => setScripts(data || []))
      .catch(() => setScripts([]))
  }

  useEffect(() => {
    loadScripts()
    // Unresolved player REPORT count for the Reports tab badge.
    fetch('/api/gm/reports?resolved=false&limit=500', { headers: authHeaders() })
      .then(r => r.ok ? r.json() : null)
      .then((data: unknown[] | null) => { if (data) setUnresolvedReports(data.length) })
      .catch(() => {})
  }, [])

  const selectScript = (filename: string, keepStatus = false) => {
    fetch(`/api/gm/scripts/${encodeURIComponent(filename)}`, { headers: authHeaders() })
      .then(r => r.json())
      .then((data: ScriptFull) => {
        setSelected(data)
        setEditorContent(data.content)
        setEditorName(data.name)
        setEditorPriority(data.priority)
        setEditorFilename(data.filename)
        setEditing(true)
        setIsNew(false)
        if (!keepStatus) setStatus(null)
        setSidebarOpen(false)
      })
  }

  const newScript = () => {
    setSelected(null)
    setEditorContent('')
    setEditorName('')
    setEditorPriority(100)
    setEditorFilename('')
    setEditing(true)
    setIsNew(true)
    setStatus(null)
    setSidebarOpen(false)
  }

  const isLive = !!selected && !selected.draft

  const saveScript = (draft: boolean) => {
    const filename = editorFilename.trim()
    if (!filename) {
      setStatus({ type: 'error', message: 'Filename is required' })
      return
    }
    if (!editorContent.trim()) {
      setStatus({ type: 'error', message: 'Script content is required' })
      return
    }
    if (draft && isLive && !confirm(`"${filename}" is live. Saving it as a draft removes it from the game until it is activated again. Continue?`)) return
    if (!draft && !isLive && !confirm(`Apply "${filename}" to the live game now? Players will see the changes immediately.`)) return
    setSaving(true)
    setStatus(null)
    fetch(`/api/gm/scripts/${encodeURIComponent(filename)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({
        name: editorName || filename,
        content: editorContent,
        priority: editorPriority,
        draft,
      }),
    })
      .then(async r => {
        const data = await r.json()
        if (!r.ok) {
          setStatus({ type: 'error', message: data.error || 'Save failed' })
          return
        }
        const stats = data.stats
        const counts = `${stats.rooms} rooms, ${stats.items} items, ${stats.monsters} monsters`
        setStatus({
          type: 'success',
          message: (draft
            ? `Saved as draft — not applied to the game (defines ${counts}).`
            : `Saved & applied: ${counts}.`) + withdrawnText(data.withdrawn),
        })
        setIsNew(false)
        loadScripts()
        // Reload the script to get updated data
        selectScript(filename, true)
      })
      .catch(err => setStatus({ type: 'error', message: err.message }))
      .finally(() => setSaving(false))
  }

  const deleteScript = () => {
    if (!selected) return
    const liveNote = isLive ? ' It is live, so its changes will also be removed from the game.' : ''
    if (!confirm(`Delete script "${selected.filename}"? This cannot be undone.${liveNote}`)) return
    fetch(`/api/gm/scripts/${encodeURIComponent(selected.filename)}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })
      .then(() => {
        setSelected(null)
        setEditing(false)
        setSidebarOpen(true)
        loadScripts()
      })
  }

  // Switch between draft and live without changing the stored content.
  const setDraft = (draft: boolean) => {
    if (!selected) return
    const unsaved = editorContent !== selected.content
    if (draft) {
      if (!confirm(`Disable "${selected.filename}"? It becomes a draft and its rooms, items, monsters and timed events are removed from the game until it is activated again.`)) return
    } else {
      const note = unsaved ? '\n\nNote: your unsaved edits are NOT included — the last saved version is what goes live. Use "Save & Apply" to apply your edits.' : ''
      if (!confirm(`Activate "${selected.filename}"? It will be applied to the live game now.${note}`)) return
    }
    setSaving(true)
    setStatus(null)
    fetch(`/api/gm/scripts/${encodeURIComponent(selected.filename)}/draft`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ draft }),
    })
      .then(async r => {
        const data = await r.json().catch(() => ({}))
        if (!r.ok) {
          setStatus({ type: 'error', message: data.error || 'Update failed' })
          return
        }
        setStatus({ type: 'success', message: (data.message || 'Updated') + (draft ? '.' + withdrawnText(data.withdrawn) : '.') })
        loadScripts()
        selectScript(selected.filename, true)
      })
      .catch(err => setStatus({ type: 'error', message: err.message }))
      .finally(() => setSaving(false))
  }

  const restoreVersion = (index: number) => {
    if (!selected) return
    if (!confirm(`Restore version from ${new Date(selected.history[index].uploadedAt).toLocaleString()}?`)) return
    setSaving(true)
    fetch(`/api/gm/scripts/${encodeURIComponent(selected.filename)}/restore/${index}`, {
      method: 'POST',
      headers: authHeaders(),
    })
      .then(async r => {
        const data = await r.json()
        if (!r.ok) {
          setStatus({ type: 'error', message: data.error || 'Restore failed' })
          return
        }
        setStatus({ type: 'success', message: (data.message || 'Version restored') + '.' + withdrawnText(data.withdrawn) })
        loadScripts()
        selectScript(selected.filename, true)
      })
      .catch(err => setStatus({ type: 'error', message: err.message }))
      .finally(() => setSaving(false))
  }

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const content = reader.result as string
      setEditorContent(content)
      if (!editorFilename) setEditorFilename(file.name)
      if (!editorName) setEditorName(file.name.replace(/\.[^.]+$/, ''))
    }
    reader.readAsText(file)
    // Reset so same file can be re-uploaded
    e.target.value = ''
  }

  const formatDate = (d: string) => {
    if (!d) return ''
    return new Date(d).toLocaleString()
  }

  return (
    <div className="flex flex-col h-full font-mono text-sm">
      {/* Tab bar */}
      <div className="flex items-center gap-2 px-3 py-2 bg-[#111] border-b border-[#333]">
        {view === 'scripts' && (
          <button
            className="sm:hidden text-gray-400 hover:text-white px-2 py-1"
            onClick={() => setSidebarOpen(!sidebarOpen)}
          >
            ☰
          </button>
        )}
        {([['scripts', 'GM Scripts'], ['reports', 'Reports'], ['ranges', 'Number Ranges']] as const).map(([v, label]) => (
          <button
            key={v}
            onClick={() => setView(v)}
            className={`px-2 py-1 rounded font-bold ${view === v ? 'text-amber-400 bg-[#1a1a2e]' : 'text-gray-500 hover:text-gray-300'}`}
          >
            {label}
            {v === 'reports' && !!unresolvedReports && (
              <span className="ml-1.5 px-1.5 rounded-full bg-red-700 text-white text-[10px] align-middle">{unresolvedReports}</span>
            )}
          </button>
        ))}
        <span className="text-gray-600 text-xs ml-auto hidden sm:inline">Character: {characterName}</span>
      </div>

      {view === 'ranges' ? (
        <GMNumberRanges authHeaders={authHeaders} characterName={characterName} />
      ) : view === 'reports' ? (
        <GMReports authHeaders={authHeaders} onUnresolvedCount={setUnresolvedReports} />
      ) : (
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar — script list */}
        <div className={`${sidebarOpen ? 'flex' : 'hidden'} sm:flex flex-col w-full sm:w-72 border-r border-[#333] bg-[#0d0d0d] overflow-y-auto`}>
          <div className="p-3 border-b border-[#333]">
            <button
              onClick={newScript}
              className="w-full px-3 py-2 bg-amber-700 hover:bg-amber-600 text-white rounded text-sm"
            >
              + New Script
            </button>
          </div>
          {scripts.length === 0 && (
            <div className="p-4 text-gray-600 text-center">No scripts uploaded yet</div>
          )}
          {scripts.map(s => (
            <button
              key={s.filename}
              onClick={() => selectScript(s.filename)}
              className={`w-full text-left px-3 py-2.5 border-b border-[#222] hover:bg-[#1a1a2e] transition-colors ${
                selected?.filename === s.filename ? 'bg-[#1a1a2e] border-l-2 border-l-amber-500' : ''
              }`}
            >
              <div className="flex items-center gap-2">
                <span className="text-amber-400 text-sm truncate">{s.name || s.filename}</span>
                <span className={`ml-auto shrink-0 px-1.5 rounded text-[10px] font-bold ${s.draft ? 'bg-[#333] text-gray-300' : 'bg-green-900 text-green-300'}`}>
                  {s.draft ? 'DRAFT' : 'LIVE'}
                </span>
              </div>
              <div className="text-gray-600 text-xs truncate">{s.filename}</div>
              <div className="flex justify-between text-gray-600 text-xs mt-1">
                <span>P:{s.priority}</span>
                <span>{s.uploadedBy}</span>
              </div>
              {s.parseStats && (
                <div className="text-gray-600 text-xs">
                  {s.parseStats.rooms}r {s.parseStats.items}i {s.parseStats.monsters}m
                </div>
              )}
            </button>
          ))}
        </div>

        {/* Editor pane */}
        <div className={`${sidebarOpen ? 'hidden sm:flex' : 'flex'} flex-col flex-1 overflow-hidden`}>
          {!editing ? (
            <div className="flex items-center justify-center h-full text-gray-600">
              Select a script or create a new one
            </div>
          ) : (
            <>
              {/* Metadata bar */}
              <div className="flex flex-wrap gap-2 items-end p-3 bg-[#111] border-b border-[#333]">
                <div className="flex flex-col gap-1">
                  <label className="text-gray-500 text-xs">Filename</label>
                  <input
                    value={editorFilename}
                    onChange={e => setEditorFilename(e.target.value)}
                    disabled={!isNew}
                    className="bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-44 disabled:opacity-50"
                    placeholder="my_script.scr"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-gray-500 text-xs">Name</label>
                  <input
                    value={editorName}
                    onChange={e => setEditorName(e.target.value)}
                    className="bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-48"
                    placeholder="Descriptive name"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-gray-500 text-xs">Priority</label>
                  <input
                    type="number"
                    value={editorPriority}
                    onChange={e => setEditorPriority(parseInt(e.target.value) || 0)}
                    className="bg-[#0a0a0a] border border-[#333] rounded px-2 py-1 text-gray-200 text-sm w-20"
                  />
                </div>
                {!isNew && selected && (
                  <div className="flex flex-col gap-1">
                    <label className="text-gray-500 text-xs">Status</label>
                    <span className={`px-2 py-1 rounded text-xs font-bold ${selected.draft ? 'bg-[#333] text-gray-300' : 'bg-green-900 text-green-300'}`}>
                      {selected.draft ? 'DRAFT — not in game' : 'LIVE'}
                    </span>
                  </div>
                )}
                <div className="flex flex-wrap gap-2 ml-auto">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".scr,.SCR,.txt"
                    onChange={handleFileUpload}
                    className="hidden"
                  />
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="px-3 py-1.5 bg-[#333] hover:bg-[#444] text-gray-300 rounded text-sm"
                  >
                    Upload .scr
                  </button>
                  <button
                    onClick={() => saveScript(true)}
                    disabled={saving}
                    title="Store the script without applying it to the game"
                    className="px-3 py-1.5 bg-[#444] hover:bg-[#555] text-gray-200 rounded text-sm disabled:opacity-50"
                  >
                    {saving ? 'Saving...' : 'Save as Draft'}
                  </button>
                  <button
                    onClick={() => saveScript(false)}
                    disabled={saving}
                    title="Store the script and apply it to the live game"
                    className="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white rounded text-sm disabled:opacity-50"
                  >
                    {saving ? 'Saving...' : 'Save & Apply'}
                  </button>
                  {!isNew && selected && (
                    <button
                      onClick={() => setDraft(!selected.draft)}
                      disabled={saving}
                      title={selected.draft ? 'Apply the saved version to the live game' : 'Remove from the game and keep as a draft'}
                      className={`px-3 py-1.5 rounded text-sm text-white disabled:opacity-50 ${selected.draft ? 'bg-blue-700 hover:bg-blue-600' : 'bg-amber-800 hover:bg-amber-700'}`}
                    >
                      {selected.draft ? 'Activate' : 'Disable'}
                    </button>
                  )}
                  {!isNew && (
                    <button
                      onClick={deleteScript}
                      className="px-3 py-1.5 bg-red-800 hover:bg-red-700 text-white rounded text-sm"
                    >
                      Delete
                    </button>
                  )}
                </div>
              </div>

              {/* Status message */}
              {status && (
                <div className={`px-3 py-2 text-sm ${status.type === 'success' ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
                  {status.message}
                </div>
              )}

              {/* Script editor */}
              <textarea
                value={editorContent}
                onChange={e => setEditorContent(e.target.value)}
                className="flex-1 bg-[#0a0a0a] text-gray-200 font-mono text-sm p-4 resize-none focus:outline-none"
                placeholder="; Enter script content here...
NUMBER 100
NAME The Great Hall
*DESCRIPTION_START
A vast hall stretches before you.
*DESCRIPTION_END"
                spellCheck={false}
                autoCorrect="off"
                autoCapitalize="off"
              />

              {/* History panel */}
              {selected && selected.history && selected.history.length > 0 && (
                <div className="border-t border-[#333] bg-[#111] max-h-40 overflow-y-auto">
                  <div className="px-3 py-1.5 text-gray-500 text-xs font-bold border-b border-[#222]">
                    Version History ({selected.history.length})
                  </div>
                  {selected.history.map((v, i) => (
                    <div key={i} className="flex items-center justify-between px-3 py-1.5 border-b border-[#222] text-xs">
                      <span className="text-gray-400">
                        {formatDate(v.uploadedAt)} by {v.uploadedBy}
                      </span>
                      <button
                        onClick={() => restoreVersion(i)}
                        className="text-amber-500 hover:text-amber-400"
                      >
                        Restore
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
      )}
    </div>
  )
}
