import type { JobStatus } from '../types'

const config: Record<JobStatus, { label: string; className: string; dot: string }> = {
  pending:  { label: 'Bekliyor',    className: 'bg-slate-700 text-slate-300',    dot: 'bg-slate-400' },
  active:   { label: 'Aktif',       className: 'bg-emerald-900/60 text-emerald-300', dot: 'bg-emerald-400 animate-pulse' },
  running:  { label: 'Çalışıyor',   className: 'bg-blue-900/60 text-blue-300',   dot: 'bg-blue-400 animate-pulse-fast' },
  success:  { label: 'Başarılı',    className: 'bg-green-900/60 text-green-300', dot: 'bg-green-400' },
  failed:   { label: 'Hata',        className: 'bg-red-900/60 text-red-300',     dot: 'bg-red-400' },
  stopped:  { label: 'Durduruldu',  className: 'bg-orange-900/60 text-orange-300', dot: 'bg-orange-400' },
}

export function StatusBadge({ status }: { status: JobStatus }) {
  const { label, className, dot } = config[status] ?? config.pending
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${className}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${dot}`} />
      {label}
    </span>
  )
}
