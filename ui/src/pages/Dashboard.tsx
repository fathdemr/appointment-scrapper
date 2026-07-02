import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Play, Square, ChevronRight, Clock, Target, BarChart2,
  AlertCircle, RefreshCw, PlusCircle, Activity,
} from 'lucide-react'
import { api } from '../api'
import { StatusBadge } from '../components/StatusBadge'
import type { Job } from '../types'

export function Dashboard() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({})

  const load = async () => {
    try {
      setLoading(true)
      const data = await api.jobs.list()
      setJobs(data ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Job\'lar yüklenemedi')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const act = async (id: string, action: 'start' | 'stop') => {
    setActionLoading(prev => ({ ...prev, [id]: true }))
    try {
      if (action === 'start') await api.jobs.start(id)
      else await api.jobs.stop(id)
      await load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'İşlem başarısız')
    } finally {
      setActionLoading(prev => ({ ...prev, [id]: false }))
    }
  }

  const isRunning = (j: Job) => j.status === 'active' || j.status === 'running'

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="text-slate-400 text-sm mt-1">Tüm rezervasyon job'larını yönet</p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={load}
            className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
            title="Yenile"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          <Link
            to="/jobs/new"
            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium rounded-lg transition-colors"
          >
            <PlusCircle className="w-4 h-4" />
            Yeni Job
          </Link>
        </div>
      </div>

      {/* Stats row */}
      {jobs.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
          {[
            { label: 'Toplam',     value: jobs.length,                                              icon: Activity,  color: 'text-slate-300' },
            { label: 'Aktif',      value: jobs.filter(j => isRunning(j)).length,                   icon: Play,      color: 'text-emerald-400' },
            { label: 'Başarılı',   value: jobs.filter(j => j.status === 'success').length,         icon: Target,    color: 'text-green-400' },
            { label: 'Toplam Deneme', value: jobs.reduce((s, j) => s + j.total_runs, 0),           icon: BarChart2, color: 'text-blue-400' },
          ].map(({ label, value, icon: Icon, color }) => (
            <div key={label} className="bg-slate-800 rounded-xl p-4 border border-slate-700/50">
              <div className="flex items-center justify-between mb-2">
                <span className="text-slate-400 text-xs">{label}</span>
                <Icon className={`w-4 h-4 ${color}`} />
              </div>
              <span className={`text-2xl font-bold ${color}`}>{value}</span>
            </div>
          ))}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-24">
          <RefreshCw className="w-6 h-6 text-slate-500 animate-spin" />
        </div>
      ) : error ? (
        <div className="flex items-center gap-3 bg-red-900/20 border border-red-800/40 rounded-xl p-4 text-red-400">
          <AlertCircle className="w-5 h-5 shrink-0" />
          {error}
        </div>
      ) : jobs.length === 0 ? (
        <div className="text-center py-24">
          <div className="w-16 h-16 rounded-full bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <CalendarEmpty />
          </div>
          <h2 className="text-white font-semibold mb-2">Henüz job yok</h2>
          <p className="text-slate-400 text-sm mb-6">İlk rezervasyon job'ını oluşturarak başla</p>
          <Link to="/jobs/new" className="inline-flex items-center gap-2 px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium rounded-lg transition-colors">
            <PlusCircle className="w-4 h-4" /> Yeni Job Oluştur
          </Link>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {jobs.map(job => (
            <div key={job.id} className="bg-slate-800 border border-slate-700/50 rounded-xl p-5 hover:border-slate-600 transition-colors group">
              <div className="flex items-start justify-between mb-4">
                <div className="flex-1 min-w-0">
                  <h3 className="text-white font-semibold truncate">{job.name}</h3>
                  <p className="text-slate-400 text-xs mt-0.5 truncate">
                    {job.sport_type} · {job.facilities?.[0] ?? '—'}
                  </p>
                </div>
                <StatusBadge status={job.status} />
              </div>

              <div className="space-y-2 mb-4">
                <div className="flex items-center gap-2 text-slate-400 text-xs">
                  <Clock className="w-3.5 h-3.5 shrink-0" />
                  <span className="truncate">
                    {job.target_dates?.join(', ') || '—'} · {job.desired_times?.join(', ') || '—'}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-slate-400 text-xs">
                  <BarChart2 className="w-3.5 h-3.5 shrink-0" />
                  <span>{job.total_runs} deneme · {job.total_found} bulundu</span>
                </div>
                {job.last_result && (
                  <p className="text-slate-500 text-xs truncate italic">"{job.last_result}"</p>
                )}
              </div>

              <div className="flex items-center gap-2 pt-3 border-t border-slate-700/50">
                <button
                  onClick={() => act(job.id, isRunning(job) ? 'stop' : 'start')}
                  disabled={!!actionLoading[job.id] || job.status === 'success'}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${
                    isRunning(job)
                      ? 'bg-red-900/40 hover:bg-red-900/60 text-red-300'
                      : 'bg-emerald-900/40 hover:bg-emerald-900/60 text-emerald-300'
                  }`}
                >
                  {actionLoading[job.id] ? (
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  ) : isRunning(job) ? (
                    <Square className="w-3.5 h-3.5" />
                  ) : (
                    <Play className="w-3.5 h-3.5" />
                  )}
                  {isRunning(job) ? 'Durdur' : 'Başlat'}
                </button>
                <Link
                  to={`/jobs/${job.id}`}
                  className="flex items-center gap-1 ml-auto text-slate-400 hover:text-white text-xs font-medium transition-colors"
                >
                  Detay <ChevronRight className="w-3.5 h-3.5" />
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function CalendarEmpty() {
  return (
    <svg className="w-8 h-8 text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
        d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
    </svg>
  )
}
