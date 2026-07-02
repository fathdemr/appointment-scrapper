import { useState } from 'react'
import { ChevronLeft, ChevronRight, X, ChevronDown, ChevronUp, CalendarDays } from 'lucide-react'

interface Props {
  selected: string[]
  onChange: (dates: string[]) => void
}

const DAYS = ['Pt', 'Sa', 'Ça', 'Pe', 'Cu', 'Ct', 'Pz']
const MONTHS = [
  'Ocak', 'Şubat', 'Mart', 'Nisan', 'Mayıs', 'Haziran',
  'Temmuz', 'Ağustos', 'Eylül', 'Ekim', 'Kasım', 'Aralık',
]

function fmt(d: Date): string {
  return `${String(d.getDate()).padStart(2, '0')}.${String(d.getMonth() + 1).padStart(2, '0')}.${d.getFullYear()}`
}

function parseDate(s: string): Date | null {
  const [d, m, y] = s.split('.')
  if (!d || !m || !y) return null
  return new Date(Number(y), Number(m) - 1, Number(d))
}

export function DatePicker({ selected, onChange }: Props) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const [open, setOpen] = useState(false)
  const [viewYear, setViewYear] = useState(today.getFullYear())
  const [viewMonth, setViewMonth] = useState(today.getMonth())

  const toggle = (dateStr: string) => {
    if (selected.includes(dateStr)) {
      onChange(selected.filter(d => d !== dateStr))
    } else {
      const sorted = [...selected, dateStr].sort((a, b) => {
        const da = parseDate(a)?.getTime() ?? 0
        const db = parseDate(b)?.getTime() ?? 0
        return da - db
      })
      onChange(sorted)
    }
  }

  const prevMonth = () => {
    if (viewMonth === 0) { setViewMonth(11); setViewYear(y => y - 1) }
    else setViewMonth(m => m - 1)
  }
  const nextMonth = () => {
    if (viewMonth === 11) { setViewMonth(0); setViewYear(y => y + 1) }
    else setViewMonth(m => m + 1)
  }

  const firstDay = new Date(viewYear, viewMonth, 1)
  let startOffset = firstDay.getDay() - 1
  if (startOffset < 0) startOffset = 6
  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate()
  const cells: (number | null)[] = [
    ...Array(startOffset).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => i + 1),
  ]
  while (cells.length % 7 !== 0) cells.push(null)

  const minBookDate = new Date(today)
  minBookDate.setDate(today.getDate() + 3)

  return (
    <div className="space-y-2">
      {/* Trigger */}
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between px-3 py-2.5 bg-slate-900 border border-slate-600 hover:border-slate-500 rounded-lg transition-colors"
      >
        <div className="flex items-center gap-2 text-sm">
          <CalendarDays className="w-4 h-4 text-slate-400 shrink-0" />
          {selected.length === 0
            ? <span className="text-slate-500">Tarih seç...</span>
            : <span className="text-white">{selected.length} tarih seçildi</span>
          }
        </div>
        {open ? <ChevronUp className="w-4 h-4 text-slate-400" /> : <ChevronDown className="w-4 h-4 text-slate-400" />}
      </button>

      {/* Takvim (collapsible) */}
      {open && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl overflow-hidden">
          {/* Başlık */}
          <div className="flex items-center justify-between px-3 py-2 border-b border-slate-700/50">
            <button type="button" onClick={prevMonth}
              className="p-1 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors">
              <ChevronLeft className="w-3.5 h-3.5" />
            </button>
            <span className="text-white text-xs font-semibold">{MONTHS[viewMonth]} {viewYear}</span>
            <button type="button" onClick={nextMonth}
              className="p-1 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors">
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>

          {/* Gün başlıkları */}
          <div className="grid grid-cols-7 px-2 pt-1">
            {DAYS.map(d => (
              <div key={d} className="py-1 text-center text-[10px] font-medium text-slate-600">{d}</div>
            ))}
          </div>

          {/* Günler */}
          <div className="grid grid-cols-7 px-2 pb-2 gap-0.5">
            {cells.map((day, idx) => {
              if (!day) return <div key={idx} />
              const date = new Date(viewYear, viewMonth, day)
              const dateStr = fmt(date)
              const isSelected = selected.includes(dateStr)
              const isPast = date < today
              const isTooClose = date < minBookDate && !isPast

              return (
                <button
                  key={idx}
                  type="button"
                  disabled={isPast}
                  onClick={() => toggle(dateStr)}
                  title={isTooClose ? '72 saat kuralı' : undefined}
                  className={`
                    aspect-square flex items-center justify-center rounded text-xs font-medium transition-colors
                    ${isPast ? 'text-slate-700 cursor-not-allowed' : ''}
                    ${isSelected ? 'bg-emerald-600 text-white' : ''}
                    ${!isSelected && !isPast ? 'text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer' : ''}
                    ${isTooClose && !isSelected ? 'text-amber-500/60' : ''}
                  `}
                >
                  {day}
                </button>
              )
            })}
          </div>
          <div className="px-3 pb-2 flex items-center gap-3 text-[10px] text-slate-600 border-t border-slate-800 pt-1.5">
            <span className="flex items-center gap-1"><span className="text-amber-500/60">■</span> 72s kuralı</span>
            <span className="flex items-center gap-1"><span className="text-emerald-500">■</span> Seçili</span>
          </div>
        </div>
      )}

      {/* Seçili tarihler */}
      {selected.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {selected.map(d => (
            <span key={d} className="inline-flex items-center gap-1 px-2 py-1 bg-emerald-900/40 border border-emerald-800/50 text-emerald-300 text-xs rounded-lg">
              {d}
              <button type="button" onClick={() => toggle(d)} className="text-emerald-600 hover:text-white transition-colors">
                <X className="w-3 h-3" />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
