import { useState, useEffect, useRef } from 'react'
import {
  CheckCircle2, ExternalLink, RefreshCw, AlertCircle,
  Bot, MessageSquare, Eye, EyeOff, Copy, Check,
} from 'lucide-react'
import { api } from '../api'

interface Props {
  botToken: string
  chatId: string
  onBotTokenChange: (v: string) => void
  onChatIdChange: (v: string) => void
}

type Step = 'create' | 'token' | 'chat'

interface BotInfo {
  name: string
  username: string
  bot_link: string
}

export function TelegramSetup({ botToken, chatId, onBotTokenChange, onChatIdChange }: Props) {
  const [step, setStep] = useState<Step>(botToken ? 'chat' : 'create')
  const [tokenInput, setTokenInput] = useState(botToken)
  const [showToken, setShowToken] = useState(false)
  const [botInfo, setBotInfo] = useState<BotInfo | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [detecting, setDetecting] = useState(false)
  const [tokenError, setTokenError] = useState('')
  const [detectStatus, setDetectStatus] = useState<'idle' | 'waiting' | 'found'>('idle')
  const [copied, setCopied] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Mevcut token varsa otomatik doğrula
  useEffect(() => {
    if (botToken && !botInfo) autoVerify(botToken)
    return () => stopPolling()
  }, [])

  const autoVerify = async (token: string) => {
    try {
      const info = await api.telegram.verifyToken(token)
      setBotInfo(info)
      setStep(chatId ? 'chat' : 'chat')
    } catch { /* sessizce geç */ }
  }

  const handleVerify = async () => {
    if (!tokenInput.trim()) return
    setVerifying(true)
    setTokenError('')
    try {
      const info = await api.telegram.verifyToken(tokenInput.trim())
      setBotInfo(info)
      onBotTokenChange(tokenInput.trim())
      setStep('chat')
    } catch (e) {
      setTokenError(e instanceof Error ? e.message : 'Token geçersiz')
    } finally {
      setVerifying(false)
    }
  }

  const startDetecting = () => {
    if (!botInfo) return
    setDetecting(true)
    setDetectStatus('waiting')
    pollRef.current = setInterval(async () => {
      try {
        const res = await api.telegram.detectChat(tokenInput.trim() || botToken)
        if (res.found && res.chat_id) {
          onChatIdChange(res.chat_id)
          setDetectStatus('found')
          stopPolling()
        }
      } catch { /* devam et */ }
    }, 2500)
  }

  const stopPolling = () => {
    if (pollRef.current) { clearInterval(pollRef.current); pollRef.current = null }
    setDetecting(false)
  }

  const copyToken = () => {
    navigator.clipboard.writeText(tokenInput || botToken)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const reset = () => {
    stopPolling()
    setBotInfo(null)
    setTokenInput('')
    onBotTokenChange('')
    onChatIdChange('')
    setDetectStatus('idle')
    setTokenError('')
    setStep('create')
  }

  return (
    <div className="space-y-3">
      {/* Adım göstergesi */}
      <div className="flex items-center gap-0">
        {([
          { id: 'create', label: 'Bot Oluştur' },
          { id: 'token',  label: 'Token Gir' },
          { id: 'chat',   label: 'Bağlan' },
        ] as const).map((s, i) => {
          const done = (s.id === 'create' && (step === 'token' || step === 'chat'))
                    || (s.id === 'token' && step === 'chat' && !!botInfo)
          const active = step === s.id
          return (
            <div key={s.id} className="flex items-center">
              {i > 0 && <div className={`h-px w-8 ${done || active ? 'bg-emerald-700' : 'bg-slate-700'}`} />}
              <button
                type="button"
                onClick={() => {
                  if (s.id === 'create') setStep('create')
                  if (s.id === 'token' && (step === 'chat' || botInfo)) setStep('token')
                }}
                className="flex items-center gap-1.5"
              >
                <span className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold transition-colors ${
                  done    ? 'bg-emerald-600 text-white'
                  : active ? 'bg-slate-600 text-white ring-2 ring-slate-400'
                  :          'bg-slate-800 text-slate-600 border border-slate-700'
                }`}>
                  {done ? <Check className="w-3 h-3" /> : i + 1}
                </span>
                <span className={`text-xs hidden sm:block ${active ? 'text-white' : done ? 'text-emerald-400' : 'text-slate-600'}`}>
                  {s.label}
                </span>
              </button>
            </div>
          )
        })}
        {(botInfo || chatId) && (
          <button type="button" onClick={reset} className="ml-auto text-[10px] text-slate-600 hover:text-slate-400 transition-colors">
            Sıfırla
          </button>
        )}
      </div>

      {/* ── Adım 1: Bot Oluştur ── */}
      {step === 'create' && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 space-y-3">
          <div className="flex items-start gap-3">
            <div className="w-8 h-8 rounded-full bg-sky-900/40 flex items-center justify-center shrink-0">
              <Bot className="w-4 h-4 text-sky-400" />
            </div>
            <div>
              <p className="text-white text-sm font-medium">Telegram botu oluştur</p>
              <p className="text-slate-400 text-xs mt-0.5">
                BotFather'a git, <code className="bg-slate-800 px-1 rounded">/newbot</code> yaz,
                bot adını ver, sana bir <b>token</b> verecek.
              </p>
            </div>
          </div>
          <a
            href="https://t.me/botfather"
            target="_blank"
            rel="noreferrer"
            className="flex items-center justify-center gap-2 w-full py-2 bg-sky-600 hover:bg-sky-500 text-white text-sm font-medium rounded-lg transition-colors"
          >
            <ExternalLink className="w-4 h-4" />
            BotFather'ı Aç
          </a>
          <button
            type="button"
            onClick={() => setStep('token')}
            className="w-full py-2 text-slate-400 hover:text-white text-xs transition-colors"
          >
            Botum var, token'ı gireceğim →
          </button>
        </div>
      )}

      {/* ── Adım 2: Token ── */}
      {step === 'token' && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 space-y-3">
          <p className="text-white text-sm font-medium">Bot token'ını gir</p>
          <p className="text-slate-400 text-xs">BotFather'ın verdiği token'ı yapıştır.</p>
          <div className="relative">
            <input
              type={showToken ? 'text' : 'password'}
              value={tokenInput}
              onChange={e => { setTokenInput(e.target.value); setTokenError('') }}
              placeholder="1234567890:ABCDef..."
              className="w-full bg-slate-800 border border-slate-600 focus:border-sky-500 rounded-lg px-3 py-2.5 pr-16 text-white text-sm placeholder-slate-600 outline-none transition-colors font-mono"
              onKeyDown={e => e.key === 'Enter' && handleVerify()}
            />
            <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
              <button type="button" onClick={() => setShowToken(s => !s)} className="p-1 text-slate-500 hover:text-slate-300">
                {showToken ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
              </button>
              <button type="button" onClick={copyToken} className="p-1 text-slate-500 hover:text-slate-300">
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>
          {tokenError && (
            <div className="flex items-center gap-2 text-red-400 text-xs">
              <AlertCircle className="w-3.5 h-3.5" />{tokenError}
            </div>
          )}
          <button
            type="button"
            onClick={handleVerify}
            disabled={verifying || !tokenInput.trim()}
            className="w-full flex items-center justify-center gap-2 py-2.5 bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
          >
            {verifying ? <RefreshCw className="w-4 h-4 animate-spin" /> : <CheckCircle2 className="w-4 h-4" />}
            {verifying ? 'Doğrulanıyor...' : 'Doğrula'}
          </button>
        </div>
      )}

      {/* ── Adım 3: Chat ID ── */}
      {step === 'chat' && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 space-y-3">
          {botInfo && (
            <div className="flex items-center gap-2 p-2.5 bg-emerald-900/20 border border-emerald-800/40 rounded-lg">
              <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
              <div>
                <p className="text-white text-xs font-medium">{botInfo.name}</p>
                <p className="text-slate-400 text-[11px]">@{botInfo.username}</p>
              </div>
            </div>
          )}

          <div>
            <p className="text-white text-sm font-medium mb-1">Chat ID'ni bağla</p>
            <p className="text-slate-400 text-xs">
              Bota herhangi bir mesaj at — sistem otomatik olarak senin Chat ID'ni algılayacak.
            </p>
          </div>

          {chatId ? (
            <div className="flex items-center gap-2 p-3 bg-emerald-900/20 border border-emerald-800/40 rounded-lg">
              <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
              <div className="flex-1">
                <p className="text-slate-400 text-xs">Chat ID</p>
                <p className="text-white font-mono text-sm">{chatId}</p>
              </div>
              <button type="button" onClick={() => { onChatIdChange(''); setDetectStatus('idle') }}
                className="text-xs text-slate-600 hover:text-slate-400 transition-colors">Sil</button>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="grid grid-cols-2 gap-2">
                {botInfo && (
                  <a
                    href={botInfo.bot_link}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center justify-center gap-1.5 py-2.5 bg-slate-700 hover:bg-slate-600 text-white text-sm font-medium rounded-lg transition-colors"
                  >
                    <MessageSquare className="w-4 h-4" />
                    Bota Mesaj At
                  </a>
                )}
                <button
                  type="button"
                  onClick={detecting ? stopPolling : startDetecting}
                  disabled={!botInfo}
                  className={`flex items-center justify-center gap-1.5 py-2.5 text-sm font-medium rounded-lg transition-colors disabled:opacity-40 ${
                    detecting
                      ? 'bg-amber-700/40 text-amber-300 hover:bg-amber-700/60'
                      : 'bg-sky-600 hover:bg-sky-500 text-white'
                  }`}
                >
                  <RefreshCw className={`w-4 h-4 ${detecting ? 'animate-spin' : ''}`} />
                  {detecting ? 'Bekleniyor...' : 'Algıla'}
                </button>
              </div>
              {detecting && (
                <p className="text-slate-500 text-xs text-center animate-pulse">
                  Telegram mesajı bekleniyor... Bota herhangi bir şey yaz.
                </p>
              )}
            </div>
          )}

          {/* Manuel giriş */}
          {!chatId && (
            <div>
              <p className="text-slate-600 text-xs mb-1.5">veya manuel gir</p>
              <input
                type="text"
                value={chatId}
                onChange={e => onChatIdChange(e.target.value)}
                placeholder="1933080456"
                className="w-full bg-slate-800 border border-slate-700 focus:border-sky-500 rounded-lg px-3 py-2 text-white text-sm placeholder-slate-600 outline-none transition-colors"
              />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
