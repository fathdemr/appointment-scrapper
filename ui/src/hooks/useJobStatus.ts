import { useEffect, useState, useCallback, useRef } from 'react'
import { api } from '../api'
import type { JobStatus_ } from '../types'

export function useJobStatus(jobId: string, active: boolean) {
  const [status, setStatus] = useState<JobStatus_ | null>(null)
  const [error, setError] = useState<string | null>(null)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const poll = useCallback(async () => {
    try {
      const s = await api.jobs.status(jobId)
      setStatus(s)
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Durum alınamadı')
    }
  }, [jobId])

  useEffect(() => {
    if (!active) return
    poll()
    timerRef.current = setInterval(poll, 3000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [active, poll])

  return { status, error, refresh: poll }
}
