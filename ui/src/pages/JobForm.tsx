import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Save, RefreshCw, AlertCircle, CheckCircle2, Eye, EyeOff,
} from 'lucide-react'

import { api } from '../api'
import { CatalogSelectors } from '../components/CatalogSelectors'
import { DatePicker } from '../components/DatePicker'
import { TimePicker } from '../components/TimePicker'
import { TelegramSetup } from '../components/TelegramSetup'
import type { CreateJobRequest } from '../types'

interface FormState {
  name: string
  tc_no: string
  password: string
  sport_type: string
  facilities: string[]
  courts: string[]
  target_dates: string[]
  desired_times: string[]
  telegram_bot_token: string
  telegram_chat_id: string
  success_message: string
  opening_offset_hours: number
  burst_before_seconds: number
  burst_after_seconds: number
  burst_interval_seconds: number
  poll_interval_seconds: number
  headless: boolean
  browser_timeout_seconds: number
}

const defaults: FormState = {
  name: '',
  tc_no: '',
  password: '',
  sport_type: '',
  facilities: [],
  courts: [],
  target_dates: [],
  desired_times: [],
  telegram_bot_token: '',
  telegram_chat_id: '',
  success_message: 'Randevu alındı! 🎉',
  opening_offset_hours: 72,
  burst_before_seconds: 60,
  burst_after_seconds: 300,
  burst_interval_seconds: 2,
  poll_interval_seconds: 30,
  headless: true,
  browser_timeout_seconds: 60,
}

export function JobForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const isEdit = Boolean(id)

  const [form, setForm] = useState<FormState>(defaults)
  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [showPass, setShowPass] = useState(false)
  const [telegramMsg, setTelegramMsg] = useState('')

  useEffect(() => {
    if (!isEdit) return
    api.jobs.get(id!).then(j => {
      setForm({
        name:                   j.name,
        tc_no:                  j.tc_no,
        password:               '',
        sport_type:             j.sport_type,
        facilities:             j.facilities ?? [],
        courts:                 j.courts ?? [],
        target_dates:           j.target_dates ?? [],
        desired_times:          j.desired_times ?? [],
        telegram_bot_token:     j.telegram_bot_token ?? '',
        telegram_chat_id:       j.telegram_chat_id ?? '',
        success_message:        j.success_message ?? '',
        opening_offset_hours:   j.opening_offset_hours,
        burst_before_seconds:   j.burst_before_seconds,
        burst_after_seconds:    j.burst_after_seconds,
        burst_interval_seconds: j.burst_interval_seconds,
        poll_interval_seconds:  j.poll_interval_seconds,
        headless:               j.headless,
        browser_timeout_seconds: j.browser_timeout_seconds,
      })
    }).catch(e => setError(e.message)).finally(() => setLoading(false))
  }, [id, isEdit])

  const set = <K extends keyof FormState>(key: K, value: FormState[K]) =>
    setForm(prev => ({ ...prev, [key]: value }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.sport_type) { setError('Spor tipi seçiniz'); return }
    if (form.facilities.length === 0) { setError('En az bir tesis seçiniz'); return }
    setSaving(true)
    setError('')

    const payload: CreateJobRequest = {
      name:                   form.name.trim(),
      tc_no:                  form.tc_no.trim(),
      password:               form.password,
      sport_type:             form.sport_type,
      facilities:             form.facilities,
      courts:                 form.courts,
      target_dates:           form.target_dates,
      desired_times:          form.desired_times,
      telegram_bot_token:     form.telegram_bot_token.trim(),
      telegram_chat_id:       form.telegram_chat_id.trim(),
      success_message:        form.success_message.trim(),
      opening_offset_hours:   form.opening_offset_hours,
      burst_before_seconds:   form.burst_before_seconds,
      burst_after_seconds:    form.burst_after_seconds,
      burst_interval_seconds: form.burst_interval_seconds,
      poll_interval_seconds:  form.poll_interval_seconds,
      headless:               form.headless,
      browser_timeout_seconds: form.browser_timeout_seconds,
    }

    try {
      if (isEdit) {
        const { password, ...rest } = payload
        await api.jobs.update(id!, password ? payload : rest)
        navigate(`/jobs/${id}`)
      } else {
        const job = await api.jobs.create(payload)
        navigate(`/jobs/${job.id}`)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Kayıt başarısız')
    } finally {
      setSaving(false)
    }
  }

  const verifyTelegram = async () => {
    if (!id) return
    try {
      const res = await api.jobs.verifyTelegram(id)
      setTelegramMsg(res.message || 'Mesaj gönderildi')
    } catch (e) {
      setTelegramMsg(e instanceof Error ? e.message : 'Hata')
    }
    setTimeout(() => setTelegramMsg(''), 4000)
  }

  if (loading) return (
    <div className="flex items-center justify-center py-32">
      <RefreshCw className="w-6 h-6 text-slate-500 animate-spin" />
    </div>
  )

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 py-8">
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={() => navigate(isEdit ? `/jobs/${id}` : '/')}
          className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div>
          <h1 className="text-xl sm:text-2xl font-bold text-white">
            {isEdit ? 'Job Düzenle' : 'Yeni Job Oluştur'}
          </h1>
          <p className="text-slate-400 text-sm mt-0.5">Rezervasyon parametrelerini girin</p>
        </div>
      </div>

      {error && (
        <div className="flex items-center gap-3 bg-red-900/20 border border-red-800/40 rounded-xl p-4 text-red-400 mb-6">
          <AlertCircle className="w-5 h-5 shrink-0" />{error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">

        {/* Temel Bilgiler */}
        <Section title="Temel Bilgiler">
          <Field label="Job Adı" required>
            <Input value={form.name} onChange={v => set('name', v)} placeholder="Florya Futbol Salı" required />
          </Field>
          <div className="grid sm:grid-cols-2 gap-4">
            <Field label="TC Kimlik No" required>
              <Input value={form.tc_no} onChange={v => set('tc_no', v)} placeholder="37414713752" required maxLength={11} />
            </Field>
            <Field label="Şifre" required={!isEdit}>
              <div className="relative">
                <Input
                  type={showPass ? 'text' : 'password'}
                  value={form.password}
                  onChange={v => set('password', v)}
                  placeholder={isEdit ? '(Değiştirmek için girin)' : '••••••••'}
                  required={!isEdit}
                />
                <button
                  type="button"
                  onClick={() => setShowPass(!showPass)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                >
                  {showPass ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
            </Field>
          </div>
        </Section>

        {/* Spor / Tesis / Salon (Katalog) */}
        <Section title="Tesis & Salon Seçimi">
          <CatalogSelectors
            sportType={form.sport_type}
            facilities={form.facilities}
            courts={form.courts}
            onSportTypeChange={v => set('sport_type', v)}
            onFacilitiesChange={v => set('facilities', v)}
            onCourtsChange={v => set('courts', v)}
          />

        </Section>

        {/* Tarih & Saat */}
        <Section title="Hedef Tarihler">
          <p className="text-slate-500 text-xs -mt-1 mb-2">
            Rezervasyon yapmak istediğin günleri seç. Birden fazla tarih seçebilirsin.
          </p>
          <DatePicker
            selected={form.target_dates}
            onChange={v => set('target_dates', v)}
          />
        </Section>

        <Section title="İstenen Saatler">
          <p className="text-slate-500 text-xs -mt-1 mb-2">
            İstediğin saat dilimlerini seç. Üzerindeki numara öncelik sırasını gösterir — önce yüksek öncelikli slotlar denenir.
          </p>
          <TimePicker
            selected={form.desired_times}
            onChange={v => set('desired_times', v)}
          />
        </Section>

        {/* Telegram */}
        <Section title="Telegram Bildirimleri">
          <TelegramSetup
            botToken={form.telegram_bot_token}
            chatId={form.telegram_chat_id}
            onBotTokenChange={v => set('telegram_bot_token', v)}
            onChatIdChange={v => set('telegram_chat_id', v)}
          />
          <Field label="Başarı Mesajı">
            <Input value={form.success_message} onChange={v => set('success_message', v)} placeholder="Randevu alındı! 🎉" />
          </Field>
          {isEdit && form.telegram_bot_token && form.telegram_chat_id && (
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={verifyTelegram}
                className="flex items-center gap-2 px-4 py-2 bg-sky-900/40 hover:bg-sky-900/60 text-sky-300 text-sm font-medium rounded-lg transition-colors"
              >
                <CheckCircle2 className="w-4 h-4" /> Test Mesajı Gönder
              </button>
              {telegramMsg && <p className="text-emerald-400 text-sm">{telegramMsg}</p>}
            </div>
          )}
        </Section>

        {/* Zamanlayıcı */}
        <Section title="Zamanlayıcı Ayarları">
          <div className="grid sm:grid-cols-3 gap-4">
            <Field label="Açılış Öncesi (saat)" hint="Genellikle 72">
              <NumberInput value={form.opening_offset_hours} onChange={v => set('opening_offset_hours', v)} min={1} max={168} />
            </Field>
            <Field label="Poll Aralığı (sn)">
              <NumberInput value={form.poll_interval_seconds} onChange={v => set('poll_interval_seconds', v)} min={5} max={300} />
            </Field>
            <Field label="Tarayıcı Timeout (sn)">
              <NumberInput value={form.browser_timeout_seconds} onChange={v => set('browser_timeout_seconds', v)} min={10} max={300} />
            </Field>
            <Field label="Burst Öncesi (sn)">
              <NumberInput value={form.burst_before_seconds} onChange={v => set('burst_before_seconds', v)} min={0} max={600} />
            </Field>
            <Field label="Burst Sonrası (sn)">
              <NumberInput value={form.burst_after_seconds} onChange={v => set('burst_after_seconds', v)} min={0} max={1800} />
            </Field>
            <Field label="Burst Aralığı (sn)">
              <NumberInput value={form.burst_interval_seconds} onChange={v => set('burst_interval_seconds', v)} min={1} max={60} />
            </Field>
          </div>
          <label className="flex items-center gap-3 cursor-pointer mt-2">
            <div
              onClick={() => set('headless', !form.headless)}
              className={`relative w-10 h-5 rounded-full transition-colors cursor-pointer ${form.headless ? 'bg-emerald-500' : 'bg-slate-600'}`}
            >
              <span className={`absolute top-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform ${form.headless ? 'translate-x-5' : 'translate-x-0.5'}`} />
            </div>
            <span className="text-slate-300 text-sm">Headless (görünmez tarayıcı)</span>
          </label>
        </Section>

        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="flex items-center gap-2 px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white font-medium rounded-lg transition-colors"
          >
            {saving ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            {isEdit ? 'Güncelle' : 'Oluştur'}
          </button>
          <button
            type="button"
            onClick={() => navigate(isEdit ? `/jobs/${id}` : '/')}
            className="px-5 py-2.5 text-slate-400 hover:text-white text-sm font-medium transition-colors"
          >
            İptal
          </button>
        </div>
      </form>
    </div>
  )
}

// ─── Küçük yardımcılar ────────────────────────────────────────────────────────

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-slate-800 border border-slate-700/50 rounded-xl p-5 space-y-4">
      <h2 className="text-white font-semibold border-b border-slate-700/50 pb-3">{title}</h2>
      {children}
    </div>
  )
}

function Field({ label, hint, required, children }: {
  label: string; hint?: string; required?: boolean; children: React.ReactNode
}) {
  return (
    <div>
      <label className="block text-slate-300 text-sm font-medium mb-1.5">
        {label}{required && <span className="text-red-400 ml-0.5">*</span>}
        {hint && <span className="text-slate-500 font-normal ml-1.5 text-xs">({hint})</span>}
      </label>
      {children}
    </div>
  )
}

function Input({ type = 'text', value, onChange, placeholder, required, maxLength }: {
  type?: string; value: string; onChange: (v: string) => void;
  placeholder?: string; required?: boolean; maxLength?: number
}) {
  return (
    <input
      type={type}
      value={value}
      onChange={e => onChange(e.target.value)}
      placeholder={placeholder}
      required={required}
      maxLength={maxLength}
      className="w-full bg-slate-900 border border-slate-600 focus:border-emerald-500 rounded-lg px-3 py-2.5 text-white text-sm placeholder-slate-600 outline-none transition-colors"
    />
  )
}


function NumberInput({ value, onChange, min, max }: {
  value: number; onChange: (v: number) => void; min?: number; max?: number
}) {
  return (
    <input
      type="number"
      value={value}
      onChange={e => onChange(Number(e.target.value))}
      min={min}
      max={max}
      className="w-full bg-slate-900 border border-slate-600 focus:border-emerald-500 rounded-lg px-3 py-2.5 text-white text-sm outline-none transition-colors"
    />
  )
}
