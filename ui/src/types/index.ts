export type JobStatus = 'pending' | 'active' | 'running' | 'success' | 'failed' | 'stopped'

export interface Job {
  id: string
  name: string
  status: JobStatus
  tc_no: string
  password?: string
  sport_type: string
  facilities: string[]
  courts: string[]
  target_dates: string[]
  desired_times: string[]
  telegram_bot_token?: string
  telegram_chat_id?: string
  success_message?: string
  opening_offset_hours: number
  burst_before_seconds: number
  burst_after_seconds: number
  burst_interval_seconds: number
  poll_interval_seconds: number
  headless: boolean
  browser_timeout_seconds: number
  last_result?: string
  total_runs: number
  total_found: number
  last_run_at?: string
  created_at: string
  updated_at: string
}

export interface JobStatus_ {
  id: string
  name: string
  status: JobStatus
  is_running: boolean
  scheduler_mode?: 'idle' | 'normal' | 'burst'
  next_burst_at?: string
  next_burst_slot?: string
  sms_waiting: boolean
  total_runs: number
  total_found: number
  last_result?: string
  last_run_at?: string
}

export interface CreateJobRequest {
  name: string
  tc_no: string
  password: string
  sport_type: string
  facilities: string[]
  courts: string[]
  target_dates: string[]
  desired_times: string[]
  telegram_bot_token?: string
  telegram_chat_id?: string
  success_message?: string
  opening_offset_hours: number
  burst_before_seconds: number
  burst_after_seconds: number
  burst_interval_seconds: number
  poll_interval_seconds: number
  headless: boolean
  browser_timeout_seconds: number
}

export type UpdateJobRequest = Omit<CreateJobRequest, 'password'> & { password?: string }

// ─── Catalog ──────────────────────────────────────────────────────────────────

export interface SportType {
  id: string
  name: string
  site_value: string
  created_at: string
}

export interface Facility {
  id: string
  sport_type_id: string
  name: string
  site_value: string
  created_at: string
}

export interface Court {
  id: string
  facility_id: string
  name: string
  site_value: string
  created_at: string
}
