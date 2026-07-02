import { X } from 'lucide-react'

interface Props {
  selected: string[]
  onChange: (times: string[]) => void
}

const ALL_SLOTS = [
  '07:00', '08:00', '09:00', '10:00', '11:00', '12:00',
  '13:00', '14:00', '15:00', '16:00', '17:00', '18:00',
  '19:00', '20:00', '21:00', '22:00', '23:00',
]

export function TimePicker({ selected, onChange }: Props) {
  const toggle = (time: string) => {
    if (selected.includes(time)) {
      onChange(selected.filter(t => t !== time))
    } else {
      onChange([...selected, time])
    }
  }

  const moveUp = (idx: number) => {
    if (idx === 0) return
    const next = [...selected]
    ;[next[idx - 1], next[idx]] = [next[idx], next[idx - 1]]
    onChange(next)
  }

  const moveDown = (idx: number) => {
    if (idx === selected.length - 1) return
    const next = [...selected]
    ;[next[idx], next[idx + 1]] = [next[idx + 1], next[idx]]
    onChange(next)
  }

  return (
    <div className="space-y-3">
      {/* Flat grid — no period grouping */}
      <div className="flex flex-wrap gap-2">
        {ALL_SLOTS.map(time => {
          const isSelected = selected.includes(time)
          const rank = selected.indexOf(time) + 1
          return (
            <button
              key={time}
              type="button"
              onClick={() => toggle(time)}
              className={`
                relative px-3 py-1.5 rounded-lg text-sm font-medium transition-all
                ${isSelected
                  ? 'bg-sky-600 text-white shadow-md shadow-sky-900/40'
                  : 'bg-slate-900 border border-slate-700 text-slate-300 hover:border-slate-500 hover:text-white'
                }
              `}
            >
              {time}
              {isSelected && (
                <span className="absolute -top-1.5 -right-1.5 w-4 h-4 bg-emerald-500 rounded-full flex items-center justify-center text-white text-[9px] font-bold leading-none">
                  {rank}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {/* Seçili saatler — öncelik sırası */}
      {selected.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {selected.map((time, idx) => (
            <span key={time} className="inline-flex items-center gap-1.5 px-2 py-1 bg-sky-900/40 border border-sky-800/50 text-sky-300 text-xs rounded-lg">
              <span className="w-3.5 h-3.5 bg-sky-600 rounded-full flex items-center justify-center text-white text-[9px] font-bold shrink-0">
                {idx + 1}
              </span>
              {time}
              <div className="flex items-center gap-0.5 text-sky-600">
                <button type="button" onClick={() => moveUp(idx)} disabled={idx === 0}
                  className="hover:text-white disabled:opacity-30 transition-colors leading-none">▲</button>
                <button type="button" onClick={() => moveDown(idx)} disabled={idx === selected.length - 1}
                  className="hover:text-white disabled:opacity-30 transition-colors leading-none">▼</button>
              </div>
              <button type="button" onClick={() => toggle(time)} className="text-sky-700 hover:text-red-400 transition-colors">
                <X className="w-3 h-3" />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
