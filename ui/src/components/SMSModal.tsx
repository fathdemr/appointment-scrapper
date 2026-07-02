import { useState } from 'react'
import { MessageSquare, Send, X } from 'lucide-react'

interface Props {
  jobId: string
  onSubmit: (code: string) => Promise<void>
  onClose: () => void
}

export function SMSModal({ jobId: _, onSubmit, onClose }: Props) {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!code.trim()) return
    setLoading(true)
    setError('')
    try {
      await onSubmit(code.trim())
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Kod gönderilemedi')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-md bg-slate-800 rounded-2xl border border-violet-500/50 shadow-2xl shadow-violet-900/20 p-6 animate-pulse-fast">
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-violet-500/20 flex items-center justify-center">
              <MessageSquare className="w-5 h-5 text-violet-400" />
            </div>
            <div>
              <h2 className="text-white font-semibold text-lg">SMS Doğrulama</h2>
              <p className="text-slate-400 text-sm">Telefonunuza gelen kodu girin</p>
            </div>
          </div>
          <button onClick={onClose} className="text-slate-500 hover:text-white transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="mb-6 p-4 rounded-xl bg-violet-500/10 border border-violet-500/20">
          <p className="text-violet-300 text-sm text-center">
            Randevu rezervasyonu için SMS doğrulaması bekleniyor.
            Kodu Telegram üzerinden gönderebilir veya aşağıya girebilirsiniz.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            placeholder="123456"
            value={code}
            onChange={e => setCode(e.target.value)}
            autoFocus
            maxLength={8}
            className="w-full bg-slate-900 border border-slate-600 focus:border-violet-500 rounded-xl px-4 py-4 text-white text-center text-2xl font-mono tracking-widest placeholder-slate-600 outline-none transition-colors"
          />
          {error && <p className="text-red-400 text-sm text-center">{error}</p>}
          <button
            type="submit"
            disabled={loading || !code.trim()}
            className="w-full flex items-center justify-center gap-2 bg-violet-600 hover:bg-violet-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium py-3 rounded-xl transition-colors"
          >
            <Send className="w-4 h-4" />
            {loading ? 'Gönderiliyor...' : 'Kodu Gönder'}
          </button>
        </form>
      </div>
    </div>
  )
}
