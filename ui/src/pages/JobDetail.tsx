import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft, Play, Square, Zap, Edit2, Trash2, RefreshCw,
  CheckCircle2, MessageSquare, Clock, BarChart2, AlertCircle,
  Wifi, WifiOff, Send, ChevronDown, ChevronUp,
} from 'lucide-react'
import { api } from '../api'
import { StatusBadge } from '../components/StatusBadge'
import { SMSModal } from '../components/SMSModal'
import { useJobStatus } from '../hooks/useJobStatus'
import type { Job } from '../types'

const modeLabel: Record<string, { text: string; color: string }> = {
  idle:   { text: 'Beklemede', color: 'text-slate-400' },
  normal: { text: 'Normal Polling', color: 'text-sky-400' },
  burst:  { text: 'Burst Modu 🔥', color: 'text-orange-400' },
}

function countdown(target: string | undefined): string {
  if (!target) return '—'
  const diff = new Date(target).getTime() - Date.now()
  if (diff <= 0) return 'Şimdi'
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1_000)
  if (h > 0) return `${h}s ${m}d`
  if (m > 0) return `${m}d ${s}sn`
  return `${s}sn`
}

export function JobDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [job, setJob] = useState<Job | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState(false)
  const [showSMS, setShowSMS] = useState(false)
  const [showDetails, setShowDetails] = useState(false)
  const [tick, setTick] = useState(0)
  const [telegramMsg, setTelegramMsg] = useState('')

  const isActive = job?.status === 'active' || job?.status === 'running'
  const { status, error: statusError, refresh: refreshStatus } = useJobStatus(id!, isActive)

  const loadJob = useCallback(async () => {
    if (!id) return
    try {
      const j = await api.jobs.get(id)
      setJob(j)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Job yüklenemedi')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { loadJob() }, [loadJob])

  // update countdown every second
  useEffect(() => {
    if (!isActive) return
    const t = setInterval(() => setTick(n => n + 1), 1000)
    return () => clearInterval(t)
  }, [isActive])

  // show SMS modal when waiting
  useEffect(() => {
    if (status?.sms_waiting) setShowSMS(true)
    else setShowSMS(false)
  }, [status?.sms_waiting])

  const act = async (action: 'start' | 'stop' | 'run') => {
    if (!id) return
    setActionLoading(true)
    try {
      if (action === 'start') await api.jobs.start(id)
      else if (action === 'stop') await api.jobs.stop(id)
      else await api.jobs.runNow(id)
      await loadJob()
      refreshStatus()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'İşlem başarısız')
    } finally {
      setActionLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!id || !confirm('Bu job silinsin mi?')) return
    await api.jobs.delete(id)
    navigate('/')
  }

  const handleVerifyTelegram = async () => {
    if (!id) return
    try {
      const res = await api.jobs.verifyTelegram(id)
      setTelegramMsg(res.message || 'Test mesajı gönderildi')
    } catch (e) {
      setTelegramMsg(e instanceof Error ? e.message : 'Hata')
    }
    setTimeout(() => setTelegramMsg(''), 4000)
  }

  const submitSMS = async (code: string) => {
    await api.jobs.submitSMS(id!, code)
  }

  if (loading) return (
    <div className="flex items-center justify-center py-32">
      <RefreshCw className="w-6 h-6 text-slate-500 animate-spin" />
    </div>
  )

  if (error || !job) return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="flex items-center gap-3 bg-red-900/20 border border-red-800/40 rounded-xl p-4 text-red-400">
        <AlertCircle className="w-5 h-5" />{error || 'Job bulunamadı'}
      </div>
    </div>
  )

  const liveMode = status?.scheduler_mode ? modeLabel[status.scheduler_mode] : null

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8">
      {showSMS && (
        <SMSModal
          jobId={id!}
          onSubmit={submitSMS}
          onClose={() => setShowSMS(false)}
        />
      )}

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <button onClick={() => navigate('/')} className="mt-1 p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors">
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-3 mb-1">
            <h1 className="text-xl sm:text-2xl font-bold text-white truncate">{job.name}</h1>
            <StatusBadge status={status?.status ?? job.status} />
          </div>
          <p className="text-slate-400 text-sm">{job.sport_type} · {job.facilities?.join(', ')}</p>
        </div>
        <Link to={`/jobs/${id}/edit`} className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors">
          <Edit2 className="w-4 h-4" />
        </Link>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left column: controls + stats */}
        <div className="lg:col-span-2 space-y-5">

          {/* Live status card */}
          <div className={`rounded-xl border p-5 ${
            status?.sms_waiting
              ? 'bg-violet-900/20 border-violet-500/50'
              : isActive
              ? 'bg-slate-800 border-emerald-700/40'
              : 'bg-slate-800 border-slate-700/50'
          }`}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-white font-semibold">Canlı Durum</h2>
              <div className="flex items-center gap-2">
                {isActive ? (
                  <span className="flex items-center gap-1.5 text-emerald-400 text-xs">
                    <Wifi className="w-3.5 h-3.5" /> Bağlı
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5 text-slate-500 text-xs">
                    <WifiOff className="w-3.5 h-3.5" /> İzlenmiyor
                  </span>
                )}
              </div>
            </div>

            {status?.sms_waiting && (
              <div className="mb-4 p-3 rounded-lg bg-violet-500/10 border border-violet-500/30 flex items-center gap-3">
                <MessageSquare className="w-5 h-5 text-violet-400 shrink-0 animate-pulse" />
                <div>
                  <p className="text-violet-300 text-sm font-medium">SMS Kodu Bekleniyor</p>
                  <p className="text-violet-400/70 text-xs">Telefonunuza gelen kodu Telegram'dan veya aşağıdan gönderin</p>
                </div>
                <button
                  onClick={() => setShowSMS(true)}
                  className="ml-auto flex items-center gap-1.5 px-3 py-1.5 bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium rounded-lg transition-colors shrink-0"
                >
                  <Send className="w-3.5 h-3.5" /> Kod Gir
                </button>
              </div>
            )}

            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <Stat
                label="Mod"
                value={liveMode ? <span className={liveMode.color}>{liveMode.text}</span> : '—'}
              />
              <Stat
                label="Burst'e Kalan"
                value={<span className="font-mono">{countdown(status?.next_burst_at) + (tick ? '' : '')}</span>}
              />
              <Stat label="Hedef Slot" value={status?.next_burst_slot ?? '—'} />
              <Stat label="Toplam Deneme" value={status?.total_runs ?? job.total_runs} />
              <Stat label="Bulunan" value={<span className="text-emerald-400">{status?.total_found ?? job.total_found}</span>} />
              <Stat
                label="Son Sonuç"
                value={<span className="text-xs text-slate-400 italic truncate">{status?.last_result || job.last_result || '—'}</span>}
              />
            </div>

            {statusError && (
              <p className="mt-3 text-yellow-500/80 text-xs">⚠ Durum alınamadı: {statusError}</p>
            )}
          </div>

          {/* Actions */}
          <div className="bg-slate-800 border border-slate-700/50 rounded-xl p-5">
            <h2 className="text-white font-semibold mb-4">İşlemler</h2>
            <div className="flex flex-wrap gap-3">
              {!isActive ? (
                <ActionBtn
                  icon={<Play className="w-4 h-4" />}
                  label="Başlat"
                  onClick={() => act('start')}
                  loading={actionLoading}
                  color="emerald"
                  disabled={job.status === 'success'}
                />
              ) : (
                <ActionBtn
                  icon={<Square className="w-4 h-4" />}
                  label="Durdur"
                  onClick={() => act('stop')}
                  loading={actionLoading}
                  color="red"
                />
              )}
              <ActionBtn
                icon={<Zap className="w-4 h-4" />}
                label="Şimdi Dene"
                onClick={() => act('run')}
                loading={actionLoading}
                color="sky"
              />
              <ActionBtn
                icon={<RefreshCw className="w-4 h-4" />}
                label="Yenile"
                onClick={loadJob}
                loading={false}
                color="slate"
              />
              <button
                onClick={handleDelete}
                className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-red-900/30 hover:bg-red-900/50 text-red-400 transition-colors"
              >
                <Trash2 className="w-4 h-4" /> Sil
              </button>
            </div>
          </div>

          {/* Job details (collapsible) */}
          <div className="bg-slate-800 border border-slate-700/50 rounded-xl overflow-hidden">
            <button
              onClick={() => setShowDetails(!showDetails)}
              className="w-full flex items-center justify-between p-5 text-white font-semibold hover:bg-slate-700/30 transition-colors"
            >
              Job Detayları
              {showDetails ? <ChevronUp className="w-4 h-4 text-slate-400" /> : <ChevronDown className="w-4 h-4 text-slate-400" />}
            </button>
            {showDetails && (
              <div className="px-5 pb-5 grid sm:grid-cols-2 gap-4 border-t border-slate-700/50 pt-4">
                <Detail label="TC No" value={job.tc_no} />
                <Detail label="Spor Tipi" value={job.sport_type} />
                <Detail label="Tesisler" value={job.facilities?.join(', ')} />
                <Detail label="Kortlar" value={job.courts?.join(', ')} />
                <Detail label="Tarihler" value={job.target_dates?.join(', ')} />
                <Detail label="Saatler" value={job.desired_times?.join(', ')} />
                <Detail label="Açılış Öncesi" value={`${job.opening_offset_hours}s`} />
                <Detail label="Poll Aralığı" value={`${job.poll_interval_seconds}sn`} />
                <Detail label="Burst Öncesi" value={`${job.burst_before_seconds}sn`} />
                <Detail label="Burst Sonrası" value={`${job.burst_after_seconds}sn`} />
                <Detail label="Burst Aralığı" value={`${job.burst_interval_seconds}sn`} />
                <Detail label="Tarayıcı" value={job.headless ? 'Headless' : 'Görünür'} />
              </div>
            )}
          </div>
        </div>

        {/* Right column: Telegram + timeline */}
        <div className="space-y-5">
          {/* Telegram */}
          <div className="bg-slate-800 border border-slate-700/50 rounded-xl p-5">
            <h2 className="text-white font-semibold mb-4">Telegram</h2>
            <div className="space-y-3">
              <Detail label="Bot Token" value={job.telegram_bot_token ? '••••••••' : '—'} />
              <Detail label="Chat ID" value={job.telegram_chat_id || '—'} />
              <Detail label="Başarı Mesajı" value={job.success_message || '—'} />
            </div>
            <button
              onClick={handleVerifyTelegram}
              disabled={!job.telegram_bot_token}
              className="mt-4 w-full flex items-center justify-center gap-2 px-3 py-2 bg-sky-900/40 hover:bg-sky-900/60 text-sky-300 text-sm font-medium rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <CheckCircle2 className="w-4 h-4" /> Test Mesajı Gönder
            </button>
            {telegramMsg && (
              <p className="mt-2 text-center text-xs text-emerald-400">{telegramMsg}</p>
            )}
          </div>

          {/* Timeline */}
          <div className="bg-slate-800 border border-slate-700/50 rounded-xl p-5">
            <h2 className="text-white font-semibold mb-4">Zaman Çizelgesi</h2>
            <ol className="relative border-l border-slate-700 ml-2 space-y-4">
              {[
                { label: 'Oluşturuldu', time: job.created_at, icon: '📝' },
                { label: 'Güncellendi', time: job.updated_at, icon: '🔄' },
                { label: 'Son Çalışma', time: job.last_run_at, icon: '⚡' },
              ].map(({ label, time, icon }) => (
                <li key={label} className="pl-5 relative">
                  <span className="absolute -left-2 w-4 h-4 flex items-center justify-center text-xs">{icon}</span>
                  <p className="text-slate-300 text-xs font-medium">{label}</p>
                  <p className="text-slate-500 text-xs">{time ? new Date(time).toLocaleString('tr-TR') : '—'}</p>
                </li>
              ))}
            </ol>
          </div>

          {/* Stats */}
          <div className="bg-slate-800 border border-slate-700/50 rounded-xl p-5">
            <h2 className="text-white font-semibold mb-4">İstatistikler</h2>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-slate-400 text-sm flex items-center gap-2"><BarChart2 className="w-4 h-4" /> Toplam Deneme</span>
                <span className="text-white font-semibold">{job.total_runs}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-slate-400 text-sm flex items-center gap-2"><Clock className="w-4 h-4" /> Bulunan Slot</span>
                <span className="text-emerald-400 font-semibold">{job.total_found}</span>
              </div>
              {job.total_runs > 0 && (
                <div className="mt-3">
                  <div className="flex justify-between text-xs text-slate-500 mb-1">
                    <span>Başarı Oranı</span>
                    <span>{Math.round((job.total_found / job.total_runs) * 100)}%</span>
                  </div>
                  <div className="h-1.5 bg-slate-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-emerald-500 rounded-full"
                      style={{ width: `${Math.min(100, Math.round((job.total_found / job.total_runs) * 100))}%` }}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="bg-slate-900/50 rounded-lg p-3">
      <p className="text-slate-500 text-xs mb-1">{label}</p>
      <p className="text-white text-sm font-medium">{value}</p>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-slate-500 text-xs mb-0.5">{label}</p>
      <p className="text-slate-200 text-sm">{value || '—'}</p>
    </div>
  )
}

function ActionBtn({
  icon, label, onClick, loading, color, disabled = false,
}: {
  icon: React.ReactNode
  label: string
  onClick: () => void
  loading: boolean
  color: 'emerald' | 'red' | 'sky' | 'slate'
  disabled?: boolean
}) {
  const colors = {
    emerald: 'bg-emerald-900/40 hover:bg-emerald-900/60 text-emerald-300',
    red:     'bg-red-900/40 hover:bg-red-900/60 text-red-300',
    sky:     'bg-sky-900/40 hover:bg-sky-900/60 text-sky-300',
    slate:   'bg-slate-700 hover:bg-slate-600 text-slate-300',
  }
  return (
    <button
      onClick={onClick}
      disabled={loading || disabled}
      className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${colors[color]}`}
    >
      {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : icon}
      {label}
    </button>
  )
}
