package model

import "time"

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusActive  JobStatus = "active"
	JobStatusRunning JobStatus = "running"
	JobStatusSuccess JobStatus = "success"
	JobStatusFailed  JobStatus = "failed"
	JobStatusStopped JobStatus = "stopped"
)

type BookingJob struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Status JobStatus `json:"status"`

	// spor.istanbul giriş bilgileri
	TCNo     string `json:"tc_no"`
	Password string `json:"password,omitempty"`

	// Rezervasyon parametreleri
	SportType    string   `json:"sport_type"`
	Facilities   []string `json:"facilities"`
	Courts       []string `json:"courts"`
	TargetDates  []string `json:"target_dates"`   // DD.MM.YYYY
	DesiredTimes []string `json:"desired_times"` // HH:MM

	// Telegram bildirim
	TelegramBotToken string `json:"telegram_bot_token,omitempty"`
	TelegramChatID   string `json:"telegram_chat_id,omitempty"`
	SuccessMessage   string `json:"success_message,omitempty"`

	// Zamanlayıcı ayarları
	OpeningOffsetHours   int `json:"opening_offset_hours"`
	BurstBeforeSeconds   int `json:"burst_before_seconds"`
	BurstAfterSeconds    int `json:"burst_after_seconds"`
	BurstIntervalSeconds int `json:"burst_interval_seconds"`
	PollIntervalSeconds  int `json:"poll_interval_seconds"`

	// Tarayıcı ayarları
	Headless              bool `json:"headless"`
	BrowserTimeoutSeconds int  `json:"browser_timeout_seconds"`

	// Çalışma durumu
	LastResult string     `json:"last_result,omitempty"`
	TotalRuns  int        `json:"total_runs"`
	TotalFound int        `json:"total_found"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
