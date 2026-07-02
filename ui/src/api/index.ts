import type { Job, JobStatus_, CreateJobRequest, UpdateJobRequest, SportType, Facility, Court } from '../types'

const BASE = 'https://api.app.blockcertify.uk/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    const body = await res.text()
    throw new Error(body || `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  jobs: {
    list: () => request<Job[]>('/jobs'),
    get: (id: string) => request<Job>(`/jobs/${id}`),
    create: (data: CreateJobRequest) =>
      request<Job>('/jobs', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: UpdateJobRequest) =>
      request<Job>(`/jobs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/jobs/${id}`, { method: 'DELETE' }),
    start: (id: string) => request<{ message: string }>(`/jobs/${id}/start`, { method: 'POST' }),
    stop: (id: string) => request<{ message: string }>(`/jobs/${id}/stop`, { method: 'POST' }),
    runNow: (id: string) => request<{ message: string }>(`/jobs/${id}/run`, { method: 'POST' }),
    status: (id: string) => request<JobStatus_>(`/jobs/${id}/status`),
    verifyTelegram: (id: string) =>
      request<{ message: string }>(`/jobs/${id}/telegram/verify`, { method: 'POST' }),
    submitSMS: (id: string, code: string) =>
      request<{ message: string }>(`/jobs/${id}/sms-reply`, {
        method: 'POST',
        body: JSON.stringify({ code }),
      }),
  },
  telegram: {
    verifyToken: (token: string) =>
      request<{ id: number; name: string; username: string; bot_link: string }>(
        `/telegram/verify-token?token=${encodeURIComponent(token)}`
      ),
    detectChat: (token: string) =>
      request<{ chat_id: string | null; found: boolean }>(
        `/telegram/detect-chat?token=${encodeURIComponent(token)}`
      ),
  },
  catalog: {
    sportTypes: () => request<SportType[]>('/catalog/sport-types'),
    facilities: (sportTypeId: string) =>
      request<Facility[]>(`/catalog/facilities?sport_type_id=${sportTypeId}`),
    courts: (facilityId: string) =>
      request<Court[]>(`/catalog/courts?facility_id=${facilityId}`),
  },
  health: () => request<{ status: string }>('/../../health'),
}
